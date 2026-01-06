package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vd09-projects/swipeassist/analytics"
	"github.com/vd09-projects/swipeassist/apps"
	"github.com/vd09-projects/swipeassist/decisionengine"
	"github.com/vd09-projects/swipeassist/decisionengine/policies"
	"github.com/vd09-projects/swipeassist/domain"
	"github.com/vd09-projects/swipeassist/extractor"
	"github.com/vd09-projects/swipeassist/internal/persistence"
	"github.com/vd09-projects/swipeassist/utils"
	"github.com/vd09-projects/vision-traits/traits"
)

type Config struct {
	App               domain.AppName
	LoginURL          string
	Headless          bool
	ControlURL        string
	BehaviourCfgPath  string
	PersonaCfgPath    string
	ProfileCount      int // 0 means run until timeout
	ShotsPerProfile   int
	ScreenshotPattern string
	Timeout           time.Duration
	DryRun            bool
	PolicyName        policies.PolicyName
	ProbLikeWeight    int
	ProbPassWeight    int
	DBURL             string
}

const (
	settleDelay          = 5 * time.Second
	betweenShotsDelay    = 500 * time.Millisecond
	betweenProfilesDelay = 3 * time.Second
)

func main() {
	cfg := parseFlags()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := run(ctx, cfg); err != nil {
		log.Fatalf("decision_engine: %v", err)
	}
}

func parseFlags() *Config {
	var (
		appName       = flag.String("app", string(domain.Bumble), "App name (only BUMBLE is supported today)")
		loginURL      = flag.String("login-url", "", "App entry URL; defaults to adapter's value when empty")
		headless      = flag.Bool("headless", false, "Run browser headless")
		control       = flag.String("remote-url", "", "Rod ControlURL (optional). If empty, launches a new browser")
		behaviourCfg  = flag.String("behaviour-config", "input/configs/ui_text_extractor_config_v1.yaml", "Path to behaviour extractor config YAML")
		personaCfg    = flag.String("persona-config", "input/configs/persona_photo_extractor_config_v1.yaml", "Path to persona photo extractor config YAML")
		profileCount  = flag.Int("profiles", 0, "Number of profiles to process (0 = run until timeout)")
		shotsPerProf  = flag.Int("shots-per-profile", 1, "Screenshots to capture per profile (album images)")
		screenshotTpl = flag.String("screenshot-pattern", "out/decision_engine/profile_%02d_img_%02d.png", "Printf-style pattern for screenshots; args: profile index (1-based), shot index (1-based)")
		timeout       = flag.Duration("timeout", 10*time.Minute, "Overall timeout for the pipeline")
		dryRun        = flag.Bool("dry-run", false, "Print the decision but do not click Like/Pass/Superswipe")
		policyName    = flag.String("policy", string(policies.QACyclePolicyName), "Decision policy to use (qa_cycle_v1, probabilistic_ratio_v1)")
		dbURL         = flag.String("db-url", "postgres://postgres:postgres@localhost:55432/swipeassist?sslmode=disable", "Postgres connection URL; when set, behaviour traits and decisions are persisted")
	)
	flag.Parse()

	app := domain.AppName(strings.ToUpper(strings.TrimSpace(*appName)))

	return &Config{
		App:               app,
		LoginURL:          *loginURL,
		Headless:          *headless,
		ControlURL:        *control,
		BehaviourCfgPath:  *behaviourCfg,
		PersonaCfgPath:    *personaCfg,
		ProfileCount:      *profileCount,
		ShotsPerProfile:   *shotsPerProf,
		ScreenshotPattern: *screenshotTpl,
		Timeout:           *timeout,
		DryRun:            *dryRun,
		PolicyName:        normalizePolicyName(*policyName),
		DBURL:             strings.TrimSpace(*dbURL),
	}
}

func run(ctx context.Context, cfg *Config) (retErr error) {
	store, err := persistence.NewStore(ctx, cfg.DBURL)
	if err != nil {
		return fmt.Errorf("init db store: %w", err)
	}
	defer store.Close(ctx)

	session := analytics.NewSession(analytics.Config{
		App:    cfg.App,
		Policy: cfg.PolicyName,
		Sinks:  []analytics.Sink{analytics.NewLogSink()},
	})
	defer func() {
		session.Close(ctx, retErr)
	}()

	client, err := makeClient(cfg)
	if err != nil {
		return fmt.Errorf("init app client: %w", err)
	}
	defer client.Close()

	if err := client.Open(ctx); err != nil {
		return fmt.Errorf("open app: %w", err)
	}

	time.Sleep(settleDelay)

	// ext, err := extractor.NewVisionExtractor(&extractor.ExtractorConfig{
	// 	BehaviourCfgPath: cfg.BehaviourCfgPath,
	// 	PersonaCfgPath:   cfg.PersonaCfgPath,
	// })
	// if err != nil {
	// 	return fmt.Errorf("init extractor: %w", err)
	// }
	ext := extractor.NewNoopExtractor(5 * time.Second)

	engine, err := makeDecisionEngine(cfg)
	if err != nil {
		return fmt.Errorf("init decision engine: %w", err)
	}

	for profile := 1; ; profile++ {
		if cfg.ProfileCount > 0 && profile > cfg.ProfileCount {
			break
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if profile > 1 {
			if err := utils.SleepCtx(ctx, betweenProfilesDelay); err != nil {
				return err
			}
		}

		session.ProfileAttempt()
		if err := processProfile(ctx, profile, cfg, client, ext, engine, session, store); err != nil {
			return fmt.Errorf("profile %d: %w", profile, err)
		}
		session.ProfileComplete()
	}
	return nil
}

func makeClient(cfg *Config) (*apps.GenericClient, error) {
	return apps.New(apps.Config{
		AppName:    cfg.App,
		EntryURL:   cfg.LoginURL,
		Headless:   cfg.Headless,
		ControlURL: cfg.ControlURL,
	})
}

func makeDecisionEngine(cfg *Config) (*decisionengine.DecisionEngine, error) {
	reg := decisionengine.NewRegistry()
	policy, err := reg.Resolve(cfg.PolicyName)
	if err != nil {
		return nil, err
	}
	return decisionengine.NewDecisionEngine(reg, policy.Name()), nil
}

func normalizePolicyName(name string) policies.PolicyName {
	normalized := policies.PolicyName(strings.ToLower(strings.TrimSpace(name)))
	if normalized == "" {
		return policies.QACyclePolicyName
	}
	return normalized
}

func processProfile(
	ctx context.Context,
	profileIdx int,
	cfg *Config,
	client *apps.GenericClient,
	ext extractor.Extractor,
	engine *decisionengine.DecisionEngine,
	session *analytics.Session,
	store persistence.Store,
) error {
	profileKey := client.GetProfileId(ctx)
	imagePaths, err := captureProfileScreens(ctx, client, profileIdx, cfg.ShotsPerProfile, cfg.ScreenshotPattern)
	if err != nil {
		return err
	}
	if len(imagePaths) == 0 {
		return fmt.Errorf("no screenshots captured")
	}
	if session != nil {
		session.AddScreenshots(len(imagePaths))
	}

	behaviour, err := ext.ExtractBehaviour(ctx, imagePaths)
	if err != nil {
		return fmt.Errorf("extract behaviour: %w", err)
	}
	if err := store.SaveBehaviour(ctx, profileKey, cfg.App, behaviour); err != nil {
		return fmt.Errorf("store behaviour traits: %w", err)
	}

	photoPersona, err := ext.ExtractPhotoPersona(ctx, imagePaths[0:1])
	if err != nil {
		return fmt.Errorf("extract behaviour: %w", err)
	}

	decision, err := engine.Decide(ctx, &policies.DecisionContext{
		App:             cfg.App,
		BehaviourTraits: behaviour,
		PhotoPersona:    extractor.MapPhotosToPersonaBundle([]*traits.ExtractedTraits{photoPersona}),
		ProfileKey:      profileKey,
	})
	if err != nil {
		return fmt.Errorf("decision engine: %w", err)
	}
	if session != nil {
		session.RecordAction(decision.Action.Kind)
	}
	if err := store.SaveDecision(ctx, profileKey, decision); err != nil {
		return fmt.Errorf("store decision: %w", err)
	}

	log.Printf("profile %d: decision=%s score=%d policy=%s reason=%s", profileIdx, decision.Action.Kind, decision.Score, decision.PolicyName, decision.Reason)

	if cfg.DryRun {
		return nil
	}

	if err := client.Act(ctx, decision.Action); err != nil {
		return fmt.Errorf("apply action: %w", err)
	}
	log.Printf("profile %d: applied action %s", profileIdx, decision.Action.Kind)
	return nil
}

func captureProfileScreens(
	ctx context.Context,
	client *apps.GenericClient,
	profileIdx int,
	shots int,
	pattern string,
) ([]string, error) {
	paths := make([]string, 0, shots)

	for s := 1; s <= shots; s++ {
		if err := ctx.Err(); err != nil {
			return paths, err
		}
		path := fmt.Sprintf(pattern, client.GetProfileId(ctx), s)
		if err := client.Screenshot(ctx, path); err != nil {
			return paths, fmt.Errorf("capture screenshot %d: %w", s, err)
		}
		log.Printf("profile %d: saved screenshot %s", profileIdx, path)
		paths = append(paths, path)

		if s < shots {
			if err := client.NextMedia(ctx); err != nil {
				log.Printf("profile %d: NextMedia stopped after %d shot(s): %v", profileIdx, s, err)
				break
			}
			if err := utils.SleepCtx(ctx, betweenShotsDelay); err != nil {
				return paths, err
			}
		}
	}

	return paths, nil
}
