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
)

func main() {
	var (
		cfgPath   = flag.String("config", "input/configs/ui_text_extractor_config_v1.yaml", "Path to extractor config YAML")
		imagesCSV = flag.String("images", "", "Comma-separated list of image paths. Defaults to bundled sample screenshots.")
		outPath   = flag.String("out", "", "Optional file to write the JSON result (prints to stdout when empty)")
		timeout   = flag.Duration("timeout", 2*time.Minute, "Context timeout for the extraction request")
	)
	flag.Parse()

	images := parseImageList(*imagesCSV)
	if len(images) == 0 {
		images = defaultImages()
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	ext, err := extractor.New(*cfgPath)
	if err != nil {
		log.Fatalf("init extractor: %v", err)
	}

	results, err := ext.ExtractText(ctx, images)
	if err != nil {
		log.Fatalf("extract: %v", err)
	}

	payload, err := json.MarshalIndent(results, "", "  ")
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
	fmt.Printf("wrote extractor output to %s\n", *outPath)
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

func defaultImages() []string {
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
