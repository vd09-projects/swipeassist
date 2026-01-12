package extractor

import "github.com/vd09-projects/vision-traits/config"

type ExtractorConfig struct {
	BehaviourCfgPath string
	PersonaCfgPath   string
	// Optional preloaded configs; when set, they take precedence over paths.
	BehaviourCfg *config.Config
	PersonaCfg   *config.Config
}
