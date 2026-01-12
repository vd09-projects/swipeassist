package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vd09-projects/swipeassist/extractor"
	"github.com/vd09-projects/vision-traits/traits"
)

func main() {
	var (
		behaviourCfg = flag.String("behaviour-config", "input/configs/ui_text_extractor_config_v1.yaml", "Path to behaviour extractor config YAML (required for extractor init)")
		personaCfg   = flag.String("persona-config", "input/configs/persona_photo_extractor_config_v1.yaml", "Path to persona photo extractor config YAML")
		imagesCSV    = flag.String("images", "", "Comma-separated list of persona photo paths. Defaults to bundled sample screenshots.")
		outPath      = flag.String("out", "", "Optional file to write the JSON result (prints to stdout when empty)")
		timeout      = flag.Duration("timeout", 5*time.Minute, "Context timeout for the extraction request")
	)
	flag.Parse()

	images := parseImageList(*imagesCSV)
	if len(images) == 0 {
		images = defaultPersonaImages()
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	ext, err := extractor.NewVisionExtractor(&extractor.ExtractorConfig{
		BehaviourCfgPath: *behaviourCfg,
		PersonaCfgPath:   *personaCfg,
	})
	if err != nil {
		log.Fatalf("init extractor: %v", err)
	}

	personaByPhoto := make([]*traits.ExtractedTraits, 0, len(images))
	for _, image := range images {
		tr, err := ext.ExtractPhotoPersona(ctx, "", []string{image})
		if err != nil {
			log.Fatalf("extract persona for %s: %v", image, err)
		}
		personaByPhoto = append(personaByPhoto, tr)
	}

	bundle := extractor.MapPhotosToPersonaBundle(personaByPhoto)

	payload, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		log.Fatalf("marshal results: %v", err)
	}

	if *outPath == "" {
		fmt.Println(string(payload))
		return
	}

	if err := ensureDir(*outPath); err != nil {
		log.Fatalf("prepare output dir: %v", err)
	}
	if err := os.WriteFile(*outPath, payload, 0o644); err != nil {
		log.Fatalf("write output: %v", err)
	}
	fmt.Printf("wrote persona bundle to %s\n", *outPath)
}

func parseImageList(csv string) []string {
	csv = strings.TrimSpace(csv)
	if csv == "" {
		return nil
	}

	parts := strings.Split(csv, ",")
	images := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		images = append(images, trimmed)
	}
	return images
}

func defaultPersonaImages() []string {
	return []string{
		"input/images/BMVD1.png",
		"input/images/BMVD2.png",
		"input/images/BMVD3.png",
	}
}

func ensureDir(outPath string) error {
	dir := filepath.Dir(outPath)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
