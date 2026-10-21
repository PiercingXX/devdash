[![CI](https://github.com/PiercingXX/devdash/actions/workflows/ci.yml/badge.svg)](https://github.com/PiercingXX/devdash/actions/workflows/ci.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/PiercingXX/devdash)](https://goreportcard.com/report/github.com/PiercingXX/devdash) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENCE)

![logo of devdash with a gopher](./doc/img/logo.jpg)

DevDash is a configurable terminal dashboard for developers and creators: choose the metrics you need, from your computer, your servers, Github or Google, and display them in one place.

![a blog dashboard with Google Analytics and Search Console](./doc/img/blog.png)

> [!NOTE]
> This is the maintained fork of [Phantas0s/devdash](https://github.com/Phantas0s/devdash), archived in 2023. It supports Google Analytics 4 and runs on current Go and libraries. Coming from the original? See [Upgrading](./doc/upgrading.md).

## Features

* Display data from:
  * Your computer, or a remote one through SSH: memory, CPU, disks, network... and the output of your own commands and scripts.
  * Github: stars, views, commits, issues, pull requests, branches.
  * Google Analytics 4 and Google Search Console.
  * The availability of your websites, Feedly subscribers, the branches of a local Git repository.
* Place your widgets on a grid, and choose their size, colors and data. Everything has a default.
* Reuse options with themes.
* As many dashboards as you want, written in YAML, JSON or TOML.
* The data is refreshed automatically, or with a keystroke. Edit a dashboard without leaving DevDash.

## Installation

Download the binary for your OS from the [latest release](https://github.com/PiercingXX/devdash/releases/latest), or:

```shell
# Linux, into /usr/local/bin
curl -fsSL https://raw.githubusercontent.com/PiercingXX/devdash/master/install/linux.sh | bash

# With Go 1.26 or later
go install github.com/PiercingXX/devdash@latest
```

## Getting started

Run `devdash`. On the first run, it creates a default dashboard in `$XDG_CONFIG_HOME/devdash/default.yml` (`~/.config/devdash` on Linux) and displays it.

* `CTRL+e` edits the dashboard, `CTRL+r` reloads it, `CTRL+c` quits.
* `devdash -c <name>` displays another dashboard, for example one of the [examples](./doc/examples.md).
* `devdash list` lists your dashboards, `devdash generate` creates one for a blog or a Github project.

## Documentation

* [Getting started](./doc/getting-started.md): installation, commands, keyboard shortcuts.
* [Examples](./doc/examples.md), with their screenshots.
* [Configuration](./doc/configuration.md), [widgets and display options](./doc/widgets.md), [dates](./doc/dates.md).
* The services: [Monitor](./doc/services/monitor.md), [Host](./doc/services/host.md), [Github](./doc/services/github.md), [Google Analytics 4](./doc/services/google-analytics.md), [Google Search Console](./doc/services/google-search-console.md), [Feedly](./doc/services/feedly.md), [Git](./doc/services/git.md), [Display](./doc/services/display.md).
* [Upgrading from the original DevDash](./doc/upgrading.md).

## Contributing

Bug reports, ideas and pull requests are welcome: please open an issue first for new features.

```shell
go build .            # build
go test -race ./...   # test
go vet ./...
```

The release binaries are built by [GoReleaser](https://goreleaser.com) when a `v*` tag is pushed. The [changelog](./CHANGELOG.md) lists the changes of each version.

## Acknowledgements

DevDash was created by [Matthieu Cneude](https://github.com/Phantas0s). The gopher of the logo comes from the [free gophers pack](https://github.com/MariaLetta/free-gophers-pack) of MariaLetta. DevDash was inspired by [wtf](https://github.com/wtfutil/wtf) and [tdash](https://github.com/jessfraz/tdash).

## Licence

[Apache Licence 2.0](./LICENCE)
