# Documentation screenshots

Use [repo-helper](https://github.com/jamesread/repo-common) (`repo-helper screenshot`) to keep Antora doc images up to date. Each documented UI feature gets its own folder under `docs/modules/ROOT/images/`.

Reference implementation: `args/suggestions/`.

## Folder layout

Create one folder per doc page (or logical screenshot group):

```
docs/modules/ROOT/images/<topic>/<name>/
├── screenshots.ini      # batch capture config for repo-helper
├── config.yaml          # minimal OliveTin config for this screenshot only
├── setup_<variant>.py   # Selenium setup script(s); each defines run(driver)
├── Makefile             # start OliveTin, capture, stop
├── .gitignore           # runtime artifacts (see below)
└── *.png                # output images (committed)
```

Wire images in the matching `.adoc` page:

```asciidoc
image::<topic>/<name>/my-screenshot.png[]
```

Paths are relative to `docs/modules/ROOT/images/`.

## Port and OliveTin instance

- Doc screenshots use a **dedicated port** (not 1337) so they do not clash with a dev server.
- All screenshot folders share port **11337**.
- Set the same port in `config.yaml` (`listenAddressSingleHTTPFrontend`) and `screenshots.ini` (`base_url`).
- Start OliveTin from `service/` so the webui is found:

  ```bash
  cd service && ./OliveTin -configdir /path/to/screenshot/folder/
  ```

The Makefile handles start/wait/capture/stop.

## screenshots.ini

Each `[section]` with a `url` is one PNG. Section `name` (or section title) becomes the filename (`name.png`).

```ini
[DEFAULT]
base_url = http://localhost:11337/
dir = .
width = 640
height = 480
post_script_sleep = 0.5

[my-screenshot]
url = .
name = my-screenshot
script = setup_my_screenshot.py
```

Notes:

- `--config` must point at the real `screenshots.ini` in the folder; relative paths (`script`, `dir`) resolve from the INI directory.
- `url = .` loads the dashboard at `base_url`.
- Override `width`, `height`, `script`, etc. per section when needed.

Capture:

```bash
cd docs/modules/ROOT/images/<topic>/<name>
make update-screenshots
# or, if OliveTin is already running on that port:
repo-helper screenshot --config screenshots.ini
```

## config.yaml

Keep configs **minimal**: only actions, dashboards, and settings required for the screenshot.

- Match YAML examples shown in the doc page.
- Disable noise: `checkForUpdates: false`, `showFooter: false`, `logLevel: "WARN"`.
- Set argument `type` explicitly to avoid startup warnings.
- Omit `icon` on actions unless the screenshot needs a specific glyph. OliveTin 3k applies a default action icon (`defaultIconForActions`, currently the neutral CLI glyph) when `icon` is not set.

Reuse integration-test patterns where possible (`integration-tests/tests/*/config.yaml`).

## Setup scripts (Python)

repo-helper loads each `--script` file and calls `run(driver)`. Scripts run **in isolation** (no imports from sibling modules unless you add `sys.path` yourself); prefer one self-contained file per variant.

UI setup should mirror integration tests (`integration-tests/lib/elements.js`):

1. Wait for `body[loaded-dashboard]` before clicking actions.
2. Click action buttons via `[title="Action Title"]` or `.action-button button`.
3. For argument forms, wait for `body[loaded-argument-form]`.
4. Use `#argument-popup`, input ids (`#container`), etc.

Example skeleton:

```python
import time
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait

def run(driver):
    WebDriverWait(driver, 15).until(
        lambda d: d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard")
    )
    driver.find_element(By.CSS_SELECTOR, '[title="My Action"]').click()
    WebDriverWait(driver, 15).until(
        lambda d: d.find_element(By.TAG_NAME, "body").get_attribute("loaded-argument-form")
    )
    # optional: frame the form, open menus, inject overlays — see args/suggestions/
    time.sleep(0.2)
```

### Headless Chrome limitations

repo-helper uses headless Chrome only. Native browser UI (e.g. `<datalist>` dropdowns, date pickers) often **does not appear** in screenshots. When needed, use `driver.execute_script(...)` in the setup script to render a representative overlay after opening the real form. See `args/suggestions/setup_chrome.py` and `setup_firefox.py`.

## Makefile template

Each screenshot folder needs only a thin `Makefile` that sets `CONFIGDIR` and includes the shared rules:

```makefile
CONFIGDIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
include ../../screenshots.mk
```

Shared targets (defined in `docs/modules/ROOT/images/screenshots.mk`):

- `make start` — background OliveTin with this folder's `config.yaml`
- `make` or `make update-screenshots` — stop any instance on 11337, start, run `repo-helper screenshot --config screenshots.ini`, stop
- `make stop` — kill whatever is listening on port 11337

## .gitignore (per folder)

```
custom-webui/
__pycache__/
```

## Checklist for a new doc screenshot

1. Create `docs/modules/ROOT/images/<topic>/<name>/` with the files above.
2. Add `config.yaml` that reproduces the doc example in the UI.
3. Write `setup_*.py` to reach the desired UI state; test selectors against the Vue UI.
4. Add sections to `screenshots.ini`; output PNG names match what the `.adoc` will reference.
5. Update the `.adoc` page: `image::<topic>/<name>/<png>[]`.
6. Run `make update-screenshots` and commit PNGs plus config/scripts.
7. Remove obsolete PNGs from `images/` if paths moved.

## Prompt template (for agents)

Use or adapt this when asking to add or refresh doc screenshots:

> Update the documentation screenshot(s) for `<page.adoc>`.
>
> - Put everything in `docs/modules/ROOT/images/<topic>/<name>/`: `screenshots.ini`, `config.yaml`, setup script(s), `Makefile`, `.gitignore`, and output PNGs.
> - Follow `docs/modules/ROOT/images/SCREENSHOTS.md` and copy structure from `docs/modules/ROOT/images/args/suggestions/`.
> - Use a dedicated OliveTin port (not 1337); start from `service/` with `-configdir` pointing at the screenshot folder.
> - Setup scripts should wait for `loaded-dashboard` / `loaded-argument-form` like integration tests; reuse selectors from `integration-tests/lib/elements.js` where applicable.
> - Update `image::` paths in the `.adoc` page to match the new folder.
> - Run `make update-screenshots` and verify the PNGs before finishing.
