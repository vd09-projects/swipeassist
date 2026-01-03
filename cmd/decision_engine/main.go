package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vd09-projects/swipeassist/apps"
	"github.com/vd09-projects/swipeassist/decisionengine"
	"github.com/vd09-projects/swipeassist/decisionengine/policies"
	"github.com/vd09-projects/swipeassist/domain"
	"github.com/vd09-projects/swipeassist/extractor"
)

type Config struct {
	App               domain.AppName
	LoginURL          string
	Headless          bool
	ControlURL        string
	BehaviourCfgPath  string
	PersonaCfgPath    string
	ProfileCount      int
	ShotsPerProfile   int
	ScreenshotPattern string
	Timeout           time.Duration
	DryRun            bool
}

const (
	settleDelay          = 10 * time.Second
	betweenShotsDelay    = 500 * time.Millisecond
	betweenProfilesDelay = 10 * time.Second
)

func main() {
	cfg := parseFlags()

	// ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	fmt.Println(cfg.Timeout, 10*time.Minute)
	// return

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
		profileCount  = flag.Int("profiles", 1, "Number of profiles to process in one run")
		shotsPerProf  = flag.Int("shots-per-profile", 1, "Screenshots to capture per profile (album images)")
		screenshotTpl = flag.String("screenshot-pattern", "out/decision_engine/profile_%02d_img_%02d.png", "Printf-style pattern for screenshots; args: profile index (1-based), shot index (1-based)")
		timeout       = flag.Duration("timeout", 10*time.Minute, "Overall timeout for the pipeline")
		dryRun        = flag.Bool("dry-run", false, "Print the decision but do not click Like/Pass/Superswipe")
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
	}
}

func run(ctx context.Context, cfg *Config) error {
	client, err := makeClient(cfg)
	if err != nil {
		return fmt.Errorf("init app client: %w", err)
	}
	defer client.Close()

	if err := client.Open(ctx); err != nil {
		return fmt.Errorf("open app: %w", err)
	}

	time.Sleep(settleDelay)

	ext, err := extractor.New(&extractor.ExtractorConfig{
		BehaviourCfgPath: cfg.BehaviourCfgPath,
		PersonaCfgPath:   cfg.PersonaCfgPath,
	})
	if err != nil {
		return fmt.Errorf("init extractor: %w", err)
	}

	engine := makeDecisionEngine(cfg.App)

	for profile := 1; profile <= cfg.ProfileCount; profile++ {
		if profile > 1 {
			time.Sleep(betweenProfilesDelay)
		}

		if err := processProfile(ctx, profile, cfg, client, ext, engine); err != nil {
			return fmt.Errorf("profile %d: %w", profile, err)
		}
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

func makeDecisionEngine(app domain.AppName) *decisionengine.DecisionEngine {
	reg := decisionengine.NewRegistry()
	reg.Register(app, policies.NewQACyclePolicy(policies.DefaultQACyclePolicyConfig()))
	return decisionengine.NewDecisionEngine(reg)
}

func processProfile(
	ctx context.Context,
	profileIdx int,
	cfg *Config,
	client *apps.GenericClient,
	ext extractor.Extractor,
	engine *decisionengine.DecisionEngine,
) error {
	imagePaths, err := captureProfileScreens(ctx, client, profileIdx, cfg.ShotsPerProfile, cfg.ScreenshotPattern)
	if err != nil {
		return err
	}
	if len(imagePaths) == 0 {
		return fmt.Errorf("no screenshots captured")
	}

	behaviour, err := ext.ExtractBehaviour(ctx, imagePaths)
	if err != nil {
		return fmt.Errorf("extract behaviour: %w", err)
	}

	decision, err := engine.Decide(ctx, &policies.DecisionContext{
		App:             cfg.App,
		BehaviourTraits: behaviour,
		ProfileKey:      fmt.Sprintf("profile_%02d", profileIdx),
	})
	if err != nil {
		return fmt.Errorf("decision engine: %w", err)
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
		path := fmt.Sprintf(pattern, profileIdx, s)
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
			time.Sleep(betweenShotsDelay)
		}
	}

	return paths, nil
}
