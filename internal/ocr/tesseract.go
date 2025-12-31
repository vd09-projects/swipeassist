package ocr

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

var execCommandContext = exec.CommandContext

type TesseractEngine struct {
	Lang string // e.g. "eng"
}

func (t TesseractEngine) Name() string { return "tesseract" }

func (t TesseractEngine) ExtractText(ctx context.Context, imageBytes []byte) (Result, error) {
	// tesseract reads from file or stdin; easiest: stdin to stdout using "-" and "stdout"
	// Many installs support: tesseract stdin stdout -l eng
	lang := t.Lang
	if lang == "" {
		lang = "eng"
	}

	cmd := execCommandContext(ctx, "tesseract", "stdin", "stdout", "-l", lang)
	cmd.Stdin = bytes.NewReader(imageBytes)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{}, fmt.Errorf("tesseract failed: %w (output=%s)", err, string(out))
	}

	// Tesseract doesn't provide a simple confidence here without extra flags;
	// We'll set a conservative default and refine later.
	txt := string(out)
	conf := 0.70
	if len(bytes.TrimSpace(out)) == 0 {
		conf = 0.0
	}

	return Result{Text: txt, Confidence: conf, Engine: t.Name()}, nil
}
