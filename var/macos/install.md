# Installing OliveTin on macOS

> **Draft** — local Markdown draft kept in sync with the AsciiDoc docs at
> <https://docs.olivetin.app/install/macos.html> and
> <https://docs.olivetin.app/install/macos_service.html>
> (`docs/modules/ROOT/pages/install/macos.adoc` and `macos_service.adoc`).

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
properly, see step 6 — you can install it **as your own user (no root)** or
**system-wide**.

---

## 3. Gatekeeper and notarization

Current release binaries are **Developer ID signed and notarized** by Apple.
After extract, you should be able to run `./OliveTin` normally.

If Gatekeeper still blocks an older (unsigned) build, or you see a prompt that
Apple cannot check the binary for malicious software, clear the quarantine
attribute:

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
system log, and Docker — see the **`config.macos.yaml`** that ships alongside
this guide. Copy it in place with:

```sh
cp config.macos.yaml config.yaml
```

---

## 5. Run OliveTin

From the folder that contains both `OliveTin` and `config.yaml`:

```sh
./OliveTin
```

Then open the web interface at:

```text
http://localhost:1337
```

(or `http://<your-mac-hostname>:1337` from another device on your network).

Press **Ctrl-C** in the Terminal to stop it.

---

## 6. Run OliveTin as a background service (launchd)

On Linux, OliveTin is managed by **systemd**. The macOS equivalent is
**launchd**. launchd offers two ways to run a background service, and which one
you pick decides whether you need root:

* **LaunchAgent (local user)** — runs as *your* user and starts when you log in.
  **No `sudo` required**, and everything lives under your home folder. Best for a
  desktop Mac. See [Local user installation](#local-user-installation-no-root).
* **LaunchDaemon (system-wide)** — runs as `root` and starts at boot, before any
  user logs in. Requires `sudo`. Best for a headless, always-on Mac. See
  [System-wide installation](#system-wide-installation-requires-root).

You only need to follow **one** of the two sections below.

### Local user installation (no root)

Everything — the binary, configuration, the `var` data folder, and the `webui`
folder — is kept together under `~/Library/Application Support/OliveTin`, so you
never need `sudo`.

**Install the files** (run from the extracted archive directory):

```sh
# Create the application folder and a place for logs
mkdir -p ~/Library/Application\ Support/OliveTin/var
mkdir -p ~/Library/Logs/OliveTin

# Copy in the binary, your config, and the bundled web UI
cp OliveTin    ~/Library/Application\ Support/OliveTin/
cp config.yaml ~/Library/Application\ Support/OliveTin/
cp -R webui    ~/Library/Application\ Support/OliveTin/
```

This gives you the following layout, all owned by your user:

```
~/Library/Application Support/OliveTin/
├── OliveTin          # the binary
├── config.yaml       # your configuration
├── webui/            # the web interface assets (shipped in the archive)
└── var/              # runtime data OliveTin writes (logs, etc.)

~/Library/Logs/OliveTin/olivetin.log   # service stdout/stderr
```

**Create the service definition.** Create a file named
`app.olivetin.olivetin.plist` with the contents below.

> **Important:** launchd does *not* expand `~`, so the paths must be absolute.
> Replace `YOUR_USERNAME` with the output of `whoami` in every path.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>app.olivetin.olivetin</string>

    <key>ProgramArguments</key>
    <array>
        <string>/Users/YOUR_USERNAME/Library/Application Support/OliveTin/OliveTin</string>
        <string>-configdir</string>
        <string>/Users/YOUR_USERNAME/Library/Application Support/OliveTin</string>
    </array>

    <key>WorkingDirectory</key>
    <string>/Users/YOUR_USERNAME/Library/Application Support/OliveTin</string>

    <key>KeepAlive</key>
    <true/>

    <key>RunAtLoad</key>
    <true/>

    <key>StandardOutPath</key>
    <string>/Users/YOUR_USERNAME/Library/Logs/OliveTin/olivetin.log</string>
    <key>StandardErrorPath</key>
    <string>/Users/YOUR_USERNAME/Library/Logs/OliveTin/olivetin.log</string>
</dict>
</plist>
```

`WorkingDirectory` makes the relative `webui` and `var` folders resolve inside
the application folder, `KeepAlive` restarts OliveTin if it exits (like systemd's
`Restart=always`), and `RunAtLoad` starts it as soon as the service is loaded.

**Register and start the service:**

```sh
cp app.olivetin.olivetin.plist ~/Library/LaunchAgents/
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/app.olivetin.olivetin.plist
```

> `bootstrap`/`bootout` replace the deprecated `launchctl load`/`unload`. They
> take a *domain target*: `gui/$(id -u)` is your own per-user GUI domain
> (`id -u` is your numeric user ID).

To stop and disable it:

```sh
launchctl bootout gui/$(id -u) ~/Library/LaunchAgents/app.olivetin.olivetin.plist
```

**Restart after a change.** After editing `config.yaml` or replacing the
binary, restart the service so the change takes effect. To restart in place:

```sh
launchctl kickstart -k gui/$(id -u)/app.olivetin.olivetin
```

If you changed the *plist* itself, `kickstart` is not enough — boot the service
out and back in so launchd re-reads it (`bootstrap` errors if the service is
still loaded):

```sh
launchctl bootout gui/$(id -u) ~/Library/LaunchAgents/app.olivetin.olivetin.plist
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/app.olivetin.olivetin.plist
```

**Verify** — open <http://localhost:1337>. If the page does not load, check the
service log:

```sh
tail -f ~/Library/Logs/OliveTin/olivetin.log
```

### System-wide installation (requires root)

Use this for a headless or shared Mac that should start OliveTin at boot, before
anyone logs in. It installs the binary on the system `PATH` and runs as `root`
via a LaunchDaemon, so the commands use `sudo`.

**Install the files:**

```sh
sudo cp OliveTin /usr/local/bin/OliveTin

sudo mkdir -p /usr/local/etc/OliveTin
sudo cp config.yaml /usr/local/etc/OliveTin/config.yaml
sudo cp -R webui /usr/local/etc/OliveTin/
```

> OliveTin looks for `config.yaml` in the directory given by the `-configdir`
> flag, which defaults to the current directory. The service definition below
> passes `-configdir /usr/local/etc/OliveTin` explicitly, and sets
> `WorkingDirectory` so the `webui` and `var` folders resolve there.

**Create the service definition.** Create a file named
`app.olivetin.olivetin.plist` with the following contents. Adjust the paths if
you installed OliveTin elsewhere.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>app.olivetin.olivetin</string>

    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/OliveTin</string>
        <string>-configdir</string>
        <string>/usr/local/etc/OliveTin</string>
    </array>

    <key>WorkingDirectory</key>
    <string>/usr/local/etc/OliveTin</string>

    <key>KeepAlive</key>
    <true/>

    <key>RunAtLoad</key>
    <true/>

    <key>StandardOutPath</key>
    <string>/usr/local/var/log/olivetin.log</string>
    <key>StandardErrorPath</key>
    <string>/usr/local/var/log/olivetin.log</string>
</dict>
</plist>
```

`KeepAlive` restarts OliveTin if it exits (like systemd's `Restart=always`), and
`RunAtLoad` starts it as soon as the service is loaded.

**Register and start the service:**

```sh
sudo mkdir -p /usr/local/var/log
sudo cp app.olivetin.olivetin.plist /Library/LaunchDaemons/
sudo chown root:wheel /Library/LaunchDaemons/app.olivetin.olivetin.plist
sudo launchctl bootstrap system /Library/LaunchDaemons/app.olivetin.olivetin.plist
```

> `bootstrap`/`bootout` replace the deprecated `launchctl load`/`unload`. The
> domain target for a LaunchDaemon is `system`.

To stop and disable it:

```sh
sudo launchctl bootout system /Library/LaunchDaemons/app.olivetin.olivetin.plist
```

**Restart after a change.** After editing `config.yaml` or replacing the
binary, restart the service so the change takes effect. To restart in place:

```sh
sudo launchctl kickstart -k system/app.olivetin.olivetin
```

If you changed the *plist* itself, `kickstart` is not enough — boot the service
out and back in so launchd re-reads it (`bootstrap` errors if the service is
still loaded):

```sh
sudo launchctl bootout system /Library/LaunchDaemons/app.olivetin.olivetin.plist
sudo launchctl bootstrap system /Library/LaunchDaemons/app.olivetin.olivetin.plist
```

**Verify** — open <http://localhost:1337>. If the page does not load, check the
service log:

```sh
tail -f /usr/local/var/log/olivetin.log
```

---

## Troubleshooting

**"Bad CPU type in executable"** — you downloaded the wrong architecture. Get
the `arm64` build for Apple Silicon, `amd64` for Intel (see step 1).

**Gatekeeper still blocks it** — for older unsigned builds, re-run
`xattr -dr com.apple.quarantine ./OliveTin` (see step 3), or approve the app
under **System Settings → Privacy & Security**. Current signed releases should
not need this.

**It runs but the page won't load** — check that nothing else is using port
1337 (`lsof -i :1337`), and that you're browsing to `http://` (not `https://`).

**Reading the logs**

* Running in Terminal: the log is printed directly to the window.
* Running under launchd as a local user: `tail -f ~/Library/Logs/OliveTin/olivetin.log`
* Running under launchd system-wide: `tail -f /usr/local/var/log/olivetin.log`
* You can raise detail by setting `logLevel: "DEBUG"` in `config.yaml`.

**Still stuck?** Ask in the
[OliveTin Discord](https://discord.gg/jhYWWpNJ3v) or open an issue on
[GitHub](https://github.com/OliveTin/OliveTin/issues).

---

## Next steps

* [Create your first action](https://docs.olivetin.app/action_execution/create_your_first.html)
* [Configuration reference](https://docs.olivetin.app/)
* [Security & authentication](https://docs.olivetin.app/security/local.html)
