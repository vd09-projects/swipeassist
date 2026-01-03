# SwipeAssist

Utilities for capturing dating-app screenshots and extracting structured UI text.

## Prerequisites

- Go 1.25+
- Local Ollama endpoint configured as referenced in `input/configs/ui_text_extractor_config_v1.yaml`
- Chrome installed (for the Bumble automation client)

## Run the UI text extractor

From the repo root:

```bash
go run ./cmd/ui_text_extractor \
  -config input/configs/ui_text_extractor_config_v1.yaml \
  -images input/images/BMVD1.png,input/images/BMVD2.png,input/images/BMVD3.png \
  -out out/ui_traits.json
```

Flags:

- `-config` points to the YAML config that wires prompts, taxonomy, and Ollama connection.
- `-images` accepts a comma-separated list of screenshot paths (defaults to the sample `input/images/BMVD*.png` files when omitted).
- `-out` writes the JSON output to a file; omit it to print to stdout.

## Run the Bumble app automation client

1. Launch Chrome with remote debugging enabled (only once per session):

   ```bash
   /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome \
     --remote-debugging-port=9222 \
     --user-data-dir=/tmp/rod-profile
   ```

2. Fetch the DevTools browser ID (needed for the `-remote-url` flag) from the running Chrome instance:

   ```bash
   curl http://127.0.0.1:9222/json/version | grep "devtools/browser"
   ```

3. Run the Bumble client against that Chrome instance:

   ```bash
   go run ./cmd/bumble_app_client \
     -login-url="https://gew3.bumble.com/app" \
     -remote-url="ws://127.0.0.1:9222/devtools/browser/55467fc4-17a0-438d-8ac2-934eef36e9b0" \
     -action=LIKE
   ```

   - Add `-headless=true` if you launch Chrome separately and want the Rod-controlled browser hidden.
   - Update `-remote-url` to match the WebSocket endpoint printed in Chrome’s terminal (swap in the devtools/browser ID from step 2).
   - Available actions: `PASS`, `LIKE`, `SUPERSWIPE`.

## Run the end-to-end decision engine

1) Launch Chrome with remote debugging (once per session):

```bash
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome \
  --remote-debugging-port=9222 \
  --user-data-dir=/tmp/rod-profile
```

2) Fetch the DevTools browser ID:

```bash
curl http://127.0.0.1:9222/json/version | grep "devtools/browser"
```

3) Run the integrated pipeline (attach to the running Chrome; fill in your devtools ID):

```bash
go run ./cmd/decision_engine \
  -remote-url "ws://127.0.0.1:9222/devtools/browser/<id>" \
  -login-url "https://bumble.com/app" \
  -profiles 0 \ # 0 = run until timeout; set >0 to cap profiles
  -shots-per-profile 3 \
  -screenshot-pattern "out/decision_engine/profile_%02d_img_%02d.png" \
  -behaviour-config input/configs/ui_text_extractor_config_v1.yaml \
  -persona-config input/configs/persona_photo_extractor_config_v1.yaml \
  -timeout 3m \
  -headless=false \
  -dry-run=false
```

Flags and tips:
- `-remote-url`: attach to existing Chrome with your Bumble session; avoids re-login prompts.
- `-profiles`: number of profiles to process; use `0` to keep processing until the `-timeout` elapses. `-shots-per-profile`: album screenshots to take per profile.
- `-screenshot-pattern`: printf pattern for saved images (`profile index`, `shot index`).
- `-behaviour-config` / `-persona-config`: extractor YAMLs (defaults point to bundled configs).
- `-dry-run`: skip clicking actions; only log decisions.
- Outputs: screenshots under `out/decision_engine` and logged decisions (action, score, policy, reason).
