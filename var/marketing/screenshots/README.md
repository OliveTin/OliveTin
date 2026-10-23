# Marketing screenshots

The screenshots used in the project README, captured from a throwaway
OliveTin running the fictional config in [`seed/`](seed/).

```bash
make screenshots                      # from the repo root
make -C var/marketing/screenshots     # or from anywhere
```

This builds OliveTin and the webui, starts OliveTin on `127.0.0.1:18337`
with a copy of `seed/` in `tmp/`, runs a few actions through the API
(`seed.sh`) so the logs have history, captures every section of
`screenshots.ini` with `repo-helper screenshot`, regenerates
`screenshot-manifest.yaml`, and then stops OliveTin, even if a step fails.

## Requirements

- Go and Node.js, to build OliveTin and the webui
- [`repo-helper`](https://github.com/jamesread/repo-common), with Chrome and a
  matching `chromedriver`
- `curl` and `jq`

No containers are needed: OliveTin keeps its state in config files and memory.

## Files

| File | Purpose |
|------|---------|
| `seed/config.yaml`, `seed/entities/` | Fictional config. Commands print static, made-up output, so images never show details of the machine running `make`. |
| `seed.sh` | Runs actions with fixed execution tracking ids, so the logs and execution URLs are the same on every run. |
| `screenshots.ini` | One section per image, plus the `title`/`description`/`theme` used for the manifest. |
| `scripts/` | Selenium setup scripts. `light.py`/`dark.py` pin `prefers-color-scheme`, because repo-helper shares one browser session across sections. |
| `generate-manifest.py` | Writes `screenshot-manifest.yaml`. |
| `screenshot-manifest.yaml` | Generated index of the screenshots for indexers. Do not edit by hand. |

To add or change a screenshot, edit `screenshots.ini` and run `make`. If only
a title or description changed, `make manifest` is enough.

To browse the demo instance yourself, run `make build up seed`, open
<http://127.0.0.1:18337/>, and run `make down` when finished.
