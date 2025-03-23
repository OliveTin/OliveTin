# OliveTin

<img alt = "project logo" src = "https://github.com/OliveTin/OliveTin/blob/main/webui.dev/OliveTinLogo.png" align = "right" width = "160px" />

OliveTin gives **safe** and **simple** access to predefined shell commands from a web interface.

[![Maturity Badge](https://img.shields.io/badge/maturity-Production-brightgreen)](#none)
[![Discord](https://img.shields.io/discord/846737624960860180?label=Discord%20Server)](https://discord.gg/jhYWWpNJ3v)
[![Awesome](https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/awesome-selfhosted/awesome-selfhosted#automation)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/5050/badge)](https://bestpractices.coreinfrastructure.org/projects/5050)

[![Go Report Card](https://goreportcard.com/badge/github.com/Olivetin/OliveTin)](https://goreportcard.com/report/github.com/OliveTin/OliveTin)
[![Build Snapshot](https://github.com/OliveTin/OliveTin/actions/workflows/build-snapshot.yml/badge.svg)](https://github.com/OliveTin/OliveTin/actions/workflows/build-snapshot.yml)

<img alt = "screenshot" src = "https://github.com/OliveTin/OliveTin/blob/main/var/marketing/screenshotDesktop.png" />
<a href = "#screenshots">More screenshots below</a>

All documentation can be found at [docs.olivetin.app](https://docs.olivetin.app). This includes installation and usage guide, etc.

## Use cases

**Safely** give access to commands, for less technical people;

* eg: Give your family a button to `podman restart plex`
* eg: Give junior admins a simple web form with dropdowns, to start your custom script. `backupScript.sh --folder {{ customerName }}`
* eg: Enable SSH access to the server for the next 20 mins `firewall-cmd --add-service ssh --timeout 20m`

**Simplify** complex commands, make them accessible and repeatable;

* eg: Expose complex commands on touchscreen tablets stuck on walls around your house. `wake-on-lan aa:bb:cc:11:22:33`
* eg: Run long-lived commands on your servers from your cell phone. `dnf update -y`
* eg: Define complex commands with lots of preset arguments, and turn a few arguments into dropdown select boxes. `docker rm {{ container }} && docker create {{ container }} && docker start {{ container }}`

[Join the community on Discord](https://discord.gg/jhYWWpNJ3v) to talk with other users about use cases, or to ask for support in getting started.

## YouTube demo video

[![YouTube demo video](https://raw.githubusercontent.com/OliveTin/OliveTin/main/var/marketing/YouTubeBanner.png)](https://www.youtube.com/watch?v=UBgOfNrzId4)

## Features

* **Responsive, touch-friendly UI** - great for tablets and mobile
* **Super simple config in YAML** - because if it's not YAML now-a-days, it's not "cloud native" :-)
* **Dark mode** - for those of you that roll that way.
* **Accessible** - passes all the accessibility checks in Firefox, and issues with accessibility are taken seriously.
* **Container** - available for quickly testing and getting it up and running, great for the selfhosted community.
* **Integrate with anything** - OliveTin just runs Linux shell commands, so theoretially you could integrate with a bunch of stuff just by using curl, ping, etc. However, writing your own shell scripts is a great way to extend OliveTin.
* **Lightweight on resources** - uses only a few MB of RAM and barely any CPU. Written in Go, with a web interface written as a modern, responsive, Single Page App that uses the REST/gRPC API.
* **Good amount of unit tests and style checks** - helps potential contributors be consistent, and helps with maintainability.

## Screenshots

Desktop web browser;

<p align = "center">
<img alt = "screenshot" src = "https://github.com/OliveTin/OliveTin/blob/main/var/marketing/screenshotDesktop.png" />
</p>

Desktop web browser (dark mode);

<p align = "center">
<img alt = "screenshot" src = "https://github.com/OliveTin/OliveTin/blob/main/var/marketing/screenshotDesktopDark.png" />
</p>

Mobile screen size (responsive layout);

<p align = "center">
<img alt = "screenshot" src = "https://github.com/OliveTin/OliveTin/blob/main/var/marketing/screenshotMobile.png" style = "height: 700px;" />
</p>

## Documentation

All documentation can be found at [docs.olivetin.app](https://docs.olivetin.app). This includes installation and usage guide, etc.

You can find instructions in the docs on how to install as a **Linux package**, **Linux Container**, on **FreeBSD**, **Windows**, **MacOS** and other platforms, too!
