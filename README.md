[![CI](https://github.com/PiercingXX/devdash/actions/workflows/ci.yml/badge.svg)](https://github.com/PiercingXX/devdash/actions/workflows/ci.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/PiercingXX/devdash)](https://goreportcard.com/report/github.com/PiercingXX/devdash) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![logo of devdash with a gopher](./doc/img/logo.jpg) 
DevDash is a highly configurable terminal dashboard for developers and creators who want to choose and display the most up-to-date metrics they need, at one place.

> [!NOTE]
> This is a maintained fork of [Phantas0s/devdash](https://github.com/Phantas0s/devdash), which was archived in 2023. Compared to the last upstream release, it:
>
> * Supports **Google Analytics 4** (Universal Analytics was shut down by Google). Use `property_id` instead of `view_id`; see [Google Analytics 4](./doc/services/google-analytics.md).
> * Uses [tcell](https://github.com/gdamore/tcell) for the terminal UI instead of the abandoned termui/termbox.
> * Removes the Travis CI service (travis-ci.org was shut down).
> * Is built with a current Go toolchain and up-to-date dependencies.
>
> See [Upgrading from upstream DevDash](./doc/upgrading.md) and the [changelog](./CHANGELOG.md) for the details.

[![google analytics example DevDash configuration](./example/img/mix-1.png)](./example/img/mix-1.png)

## Why using DevDash?

* Pull and display the data in nice widgets using YAML, JSON or TOML config files, from:
    * Your own computer. From your own scripts and command lines too!
    * A remote computer via SSH.
    * Github.
    * Google Analytics 4.
    * Google Search Console.
    * Feedly.
    * A local Git repository.

* A lot of flexibility compared to other terminal dashboards:
  * Choose the widgets you want.
  * Place your widgets where you want.
  * Choose the data you want to display, the colors you want to use, and a lot of other things for each widget.
  * Don't want to personalize everything? Don't overwrite the defaults, then.
* Unlimited amount of different dashboards with different configurations.
* Data refreshed automatically via time ticks, or via a keyboard shortcut (Ctrl + r by default).

## Video

You can watch a [quick demo on Youtube](https://www.youtube.com/watch?v=aYQVJRuBrC8&feature=youtu.be).

## Menu

* [Installation](#installation)
* [Getting Started](#getting-started)
* [Documentation](#documentation)
* [Acknowledgement](#acknowledgement)
* [Contribute](#contribute)
* [Licence](#licence)
* [Making of](#making-of)
* [Showcase](#showcase)

## Installation

You can simply grab the [latest released binary file](https://github.com/PiercingXX/devdash/releases/latest) and download the version you need, depending on your OS.

### Linux script

To download DevDash in `/usr/local/bin`, run the following in a terminal:

```shell
curl -fsSL https://raw.githubusercontent.com/PiercingXX/devdash/master/install/linux.sh | bash
```

### With Go

With Go 1.26 or later:

```shell
go install github.com/PiercingXX/devdash@latest
```

### Manual installation

You need to: 
1. Clone this repository 
2. In your terminal, run in the root directory of the project `go build .`

## Getting started

* Run `devdash`. If it doesn't exist yet, DevDash creates a default dashboard, `default.yml`, in `$XDG_CONFIG_HOME/devdash` (`~/.config/devdash` on Linux), and displays it.
* Modify this dashboard, or create your own from the [examples](./doc/examples.md). Press `CTRL+r` to reload the dashboard after modifying it, or `CTRL+e` to edit it directly.
* Display another dashboard with `devdash -c <name>`: DevDash looks for `<name>.yml` (or `.json`, `.toml`) in `$XDG_CONFIG_HOME/devdash`, then in the current directory. A path to the file works too.

See [Getting started](./doc/getting-started.md) for the commands, the keyboard shortcuts and the environment variables.

## Documentation

[The complete documentation is in the `doc` directory](./doc/README.md):

* [Getting started](./doc/getting-started.md)
* [Examples](./doc/examples.md)
* [Configuration reference](./doc/configuration.md): general options, projects, rows and columns, themes.
* [Widgets and display options](./doc/widgets.md): sizes, colors, and the options of each widget type.
* [Dates](./doc/dates.md)
* The services: [Monitor](./doc/services/monitor.md), [Host](./doc/services/host.md), [Github](./doc/services/github.md), [Google Analytics 4](./doc/services/google-analytics.md), [Google Search Console](./doc/services/google-search-console.md), [Feedly](./doc/services/feedly.md), [Git](./doc/services/git.md), [Display](./doc/services/display.md).
* [Upgrading from upstream DevDash](./doc/upgrading.md)

The website [thedevdash.com](https://thedevdash.com) documents the original, unmaintained version.

## Acknowledgement

Thanks to [Matthieu Cneude](https://github.com/Phantas0s), who created DevDash, and to [MariaLetta](https://github.com/MariaLetta/free-gophers-pack) for the awesome and beautiful Gopher pack used for the logo on top.

DevDash was inspired from other open source projects:

* [wtf](https://github.com/wtfutil/wtf)
* [tdash](https://github.com/jessfraz/tdash)

## Bugs and Ideas

If you have any idea to improve DevDash or if you see any bug, please open an issue.

## Contribute

Thanks a lot if you want to contribute to DevDash!

If you want to implement a new feature, let's speak about it first (by opening an issue) and decide if it fits DevDash scope.

## Making Of

The original author wrote about [developing DevDash on the side of a full-time job](https://thevaluable.dev/programming-side-project-example-devdash/): how they organized their time and kept their motivation.

## Licence

[Apache Licence 2.0](https://choosealicense.com/licenses/apache-2.0/)

## Showcase

![google analytics example DevDash configuration](./example/img/thevaluabledev-2.png)
-------
![google analytics example DevDash configuration](./example/img/terminal.png)
-------
![google analytics example DevDash configuration](./example/img/thevaluabledev-3.png)
-------
![github example DevDash configuration](./example/img/devdash-1.png)
