# Installing OliveTin on macOS

OliveTin runs natively on macOS on both **Apple Silicon (M1/M2/M3/M4)** and
**Intel** Macs. It is a single self-contained binary written in Go — there is no
installer and no background dependencies to install.

---

## 1. Choose the right download

macOS builds are published on the
[GitHub releases page](https://github.com/OliveTin/OliveTin/releases). Pick the
archive that matches your Mac's processor:

| Your Mac | Archive |
|---|---|
| Apple Silicon (M-series) | `OliveTin-darwin-arm64.tar.gz` |
| Intel | `OliveTin-darwin-amd64.tar.gz` |

Not sure which you have? Run this in Terminal:

```sh
uname -m
```

`arm64` → Apple Silicon, `x86_64` → Intel.

> If you download the wrong architecture, macOS will refuse to run it with a
> "Bad CPU type in executable" error.

---

## 2. Extract and place the binary

```sh
# Move to your Downloads folder (adjust if needed)
cd ~/Downloads

# Extract — replace arm64 with amd64 on Intel
tar -xzf OliveTin-darwin-arm64.tar.gz
cd OliveTin-darwin-arm64
```

For a quick try-out you can run it straight from this folder. To install it
system-wide, copy the binary somewhere on your `PATH`:

```sh
sudo cp OliveTin /usr/local/bin/OliveTin
```

---

## 3. Clear the Gatekeeper quarantine

Because the binary is downloaded from the internet and is **not notarized by
Apple**, macOS Gatekeeper will block the first run with a message like
*"OliveTin can't be opened because Apple cannot check it for malicious
software."*

Remove the quarantine attribute so it will run:

```sh
xattr -dr com.apple.quarantine ./OliveTin
```

Alternatively, the first time only, you can right-click the binary in Finder →
**Open**, or approve it under **System Settings → Privacy & Security**.

---

## 4. Create a configuration file

OliveTin looks for a file named `config.yaml` in its **config directory**, which
defaults to the current directory (`.`). You can point elsewhere with
`-configdir /path/to/dir`.

A minimal `config.yaml` to confirm everything works:

```yaml
listenAddressSingleHTTPFrontend: 0.0.0.0:1337
logLevel: "INFO"

actions:
  - title: Hello macOS
    icon: terminal
    shell: echo "Hello from $(scutil --get ComputerName)!"
    popupOnStart: execution-dialog-stdout-only
```

For a fuller, macOS-tuned starting point — with working examples for
notifications (`osascript`), `caffeinate`, `pmset`, disk usage, the unified
system log, and Docker — see the **`macos.config.yaml`** that ships alongside
this guide. Copy it in place with:

```sh
cp macos.config.yaml config.yaml
```

---

## 5. Run OliveTin

From the folder that contains both `OliveTin` and `config.yaml`:

```sh
./OliveTin
```

Or, if you installed it to `/usr/local/bin` and keep your config elsewhere:

```sh
OliveTin -configdir /usr/local/etc/OliveTin
```

Then open the web interface at:

```
http://localhost:1337
```

(or `http://<your-mac-hostname>:1337` from another device on your network).

Press **Ctrl-C** in the Terminal to stop it.

---

## 6. Run OliveTin as a background service (launchd)

On Linux, OliveTin is managed by **systemd**. The macOS equivalent is
**launchd**. A ready-to-use service definition ships next to this guide as
[`app.olivetin.olivetin.plist`](app.olivetin.olivetin.plist).

You have two choices:

* **LaunchAgent** (recommended for a desktop Mac) — runs as *your* user, starts
  when you log in, no root required.
* **LaunchDaemon** — runs as root, starts at boot before any user logs in. Use
  this for a headless / always-on Mac.

### As a LaunchAgent (per-user)

```sh
# Edit the two paths inside the plist to match your install first!
cp app.olivetin.olivetin.plist ~/Library/LaunchAgents/

# Start now and on every login
launchctl load ~/Library/LaunchAgents/app.olivetin.olivetin.plist
```

To stop and disable it:

```sh
launchctl unload ~/Library/LaunchAgents/app.olivetin.olivetin.plist
```

### As a LaunchDaemon (system-wide, at boot)

```sh
sudo cp app.olivetin.olivetin.plist /Library/LaunchDaemons/
sudo chown root:wheel /Library/LaunchDaemons/app.olivetin.olivetin.plist
sudo launchctl load /Library/LaunchDaemons/app.olivetin.olivetin.plist
```

The plist sets `KeepAlive` (restart on crash, equivalent to systemd's
`Restart=always`) and `RunAtLoad` (start immediately). It also writes logs to
`/usr/local/var/log/olivetin.log`.

---

## Troubleshooting

**"Bad CPU type in executable"** — you downloaded the wrong architecture. Get
the `arm64` build for Apple Silicon, `amd64` for Intel (see step 1).

**Gatekeeper still blocks it** — re-run the `xattr -dr com.apple.quarantine`
command in step 3, or approve the app under **System Settings → Privacy &
Security**.

**It runs but the page won't load** — check that nothing else is using port
1337 (`lsof -i :1337`), and that you're browsing to `http://` (not `https://`).

**Reading the logs**

* Running in Terminal: the log is printed directly to the window.
* Running under launchd: tail the log file configured in the plist:

  ```sh
  tail -f /usr/local/var/log/olivetin.log
  ```

* You can raise detail by setting `logLevel: "DEBUG"` in `config.yaml`.

**Still stuck?** Ask in the
[OliveTin Discord](https://discord.gg/jhYWWpNJ3v) or open an issue on
[GitHub](https://github.com/OliveTin/OliveTin/issues).

---

## Next steps

* [Create your first action](https://docs.olivetin.app/action_execution/create_your_first.html)
* [Configuration reference](https://docs.olivetin.app/)
* [Security & authentication](https://docs.olivetin.app/security/local.html)
