package ocr

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestNoopEngineExtractText(t *testing.T) {
	engine := NoopEngine{}

	res, err := engine.ExtractText(context.Background(), []byte("ignored"))
	if err != nil {
		t.Fatalf("NoopEngine returned error: %v", err)
	}
	if res.Engine != engine.Name() {
		t.Fatalf("expected engine %q, got %q", engine.Name(), res.Engine)
	}
	if res.Text != "" {
		t.Fatalf("expected empty text, got %q", res.Text)
	}
	if res.Confidence != 0.0 {
		t.Fatalf("expected zero confidence, got %f", res.Confidence)
	}
}

func TestTesseractEngineExtractTextSuccess(t *testing.T) {
	tests := []struct {
		name         string
		lang         string
		expectedLang string
		stdout       string
		wantConf     float64
	}{
		{name: "default language", lang: "", expectedLang: "eng", stdout: "recognized text", wantConf: 0.70},
		{name: "custom language", lang: "deu", expectedLang: "deu", stdout: "hallo welt", wantConf: 0.70},
		{name: "empty output lowers confidence", lang: "", expectedLang: "eng", stdout: "   ", wantConf: 0.0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			installFakeTesseract(t, fakeTesseractConfig{
				stdout:       tt.stdout,
				stderr:       "",
				shouldFail:   false,
				expectedLang: tt.expectedLang,
			})

			engine := TesseractEngine{Lang: tt.lang}
			res, err := engine.ExtractText(context.Background(), []byte("img"))
			if err != nil {
				t.Fatalf("ExtractText returned error: %v", err)
			}
			if res.Engine != engine.Name() {
				t.Fatalf("expected engine %q, got %q", engine.Name(), res.Engine)
			}
			if res.Text != tt.stdout {
				t.Fatalf("expected text %q, got %q", tt.stdout, res.Text)
			}
			if diff := math.Abs(res.Confidence - tt.wantConf); diff > 1e-9 {
				t.Fatalf("unexpected confidence: got %f want %f", res.Confidence, tt.wantConf)
			}
		})
	}
}

func TestTesseractEngineExtractTextFailure(t *testing.T) {
	const stderrMsg = "tesseract crashed"

	installFakeTesseract(t, fakeTesseractConfig{
		stdout:       "",
		stderr:       stderrMsg,
		shouldFail:   true,
		expectedLang: "eng",
	})

	engine := TesseractEngine{}
	res, err := engine.ExtractText(context.Background(), []byte("img"))
	if err == nil {
		t.Fatal("expected error from ExtractText, got nil")
	}
	if res != (Result{}) {
		t.Fatalf("expected zero result on error, got %+v", res)
	}
	if !strings.Contains(err.Error(), "tesseract failed") {
		t.Fatalf("expected error to mention tesseract failure, got %v", err)
	}
	if !strings.Contains(err.Error(), stderrMsg) {
		t.Fatalf("expected error to include stderr output, got %v", err)
	}
}

type fakeTesseractConfig struct {
	stdout       string
	stderr       string
	shouldFail   bool
	expectedLang string
}

func installFakeTesseract(t *testing.T, cfg fakeTesseractConfig) {
	t.Helper()

	original := execCommandContext
	t.Cleanup(func() {
		execCommandContext = original
	})

	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, args...)

		cmd := exec.CommandContext(ctx, os.Args[0], cs...)
		env := append([]string{}, os.Environ()...)
		env = append(env,
			"GO_WANT_HELPER_PROCESS=1",
			"HELPER_EXPECT_LANG="+cfg.expectedLang,
			"HELPER_STDOUT="+cfg.stdout,
			"HELPER_STDERR="+cfg.stderr,
		)
		if cfg.shouldFail {
			env = append(env, "HELPER_SHOULD_FAIL=1")
		}
		cmd.Env = env
		return cmd
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	for i, a := range args {
		if a == "--" {
			args = args[i+1:]
			break
		}
	}

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "no command provided")
		os.Exit(2)
	}

	if args[0] != "tesseract" {
		fmt.Fprintf(os.Stderr, "unexpected command: %s\n", args[0])
		os.Exit(2)
	}

	if len(args) < 5 {
		fmt.Fprintf(os.Stderr, "unexpected arg list: %v\n", args)
		os.Exit(2)
	}

	if args[1] != "stdin" || args[2] != "stdout" || args[3] != "-l" {
		fmt.Fprintf(os.Stderr, "unexpected args: %v\n", args)
		os.Exit(2)
	}

	if expectedLang := os.Getenv("HELPER_EXPECT_LANG"); expectedLang != "" {
		if args[4] != expectedLang {
			fmt.Fprintf(os.Stderr, "unexpected lang: got %s want %s\n", args[4], expectedLang)
			os.Exit(2)
		}
	}

	fmt.Fprint(os.Stdout, os.Getenv("HELPER_STDOUT"))
	fmt.Fprint(os.Stderr, os.Getenv("HELPER_STDERR"))

	if os.Getenv("HELPER_SHOULD_FAIL") == "1" {
		os.Exit(1)
	}
	os.Exit(0)
}
