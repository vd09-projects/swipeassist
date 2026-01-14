package extractor

import (
	"time"

	"github.com/vd09-projects/vision-traits/config"
)

type ExtractorConfig struct {
	BehaviourCfgPath string
	PersonaCfgPath   string
	// Optional preloaded configs; when set, they take precedence over paths.
	BehaviourCfg *config.Config
	PersonaCfg   *config.Config

	// Optional retry configuration. Zero values fall back to sensible defaults.
	RetryAttempts int
	RetryDelay    time.Duration
}
