[![CI](https://github.com/PiercingXX/devdash/actions/workflows/ci.yml/badge.svg)](https://github.com/PiercingXX/devdash/actions/workflows/ci.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/PiercingXX/devdash)](https://goreportcard.com/report/github.com/PiercingXX/devdash) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![logo of devdash with a gopher](./doc/img/logo.jpg) 
DevDash is a highly configurable terminal dashboard for developers and creators who want to choose and display the most up-to-date metrics they need, at one place.

> [!NOTE]
> This is a maintained fork of [Phantas0s/devdash](https://github.com/Phantas0s/devdash), which was archived in 2023. Compared to the last upstream release, it:
>
> * Supports **Google Analytics 4** (Universal Analytics was shut down by Google). Use `property_id` instead of `view_id`; see [Google Analytics 4](#google-analytics-4).
> * Uses [tcell](https://github.com/gdamore/tcell) for the terminal UI instead of the abandoned termui/termbox.
> * Removes the Travis CI service (travis-ci.org was shut down).
> * Is built with a current Go toolchain and up-to-date dependencies.
>
> See the [changelog](./CHANGELOG.md) for the details.

[![google analytics example DevDash configuration](./example/img/mix-1.png)](./example/img/mix-1.png)

## Why using DevDash?

* Pull and display the data in nice widgets using YAML (or JSON) config files, from: 
    * Your own computer. From your own scripts and command lines too!
    * A remote computer via SSH.
    * Github.
    * Google Analytics.
    * Google Search Console.
    * Feedly.

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
* [Making of](#Making-of)
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

* If you run `devdash` without giving a dashboard configuration, it will:
    * Create a config for the default dashboard (`default.yml`) in `$XDG_CONFIG_HOME/devdash`.
    * Display this default dashboard.
* You have [many examples here](https://thedevdash.com/getting-started/examples/) to modify this default dashboard or create your own from scratch.
* Each dashboard need a config file. To pass this config file to DevDash, you need to use the option `-c`. For example: `devdash -c default`.
    * If you only pass the filename of the dashboard's config, it needs to be in the working directory or in `$XDG_CONFIG_HOME/devdash`.
    * You can also pass a relative or absolute path to the config.
    * The dashboard's config can be written in YAML, JSON, or TOML.

The value of `$XDG_CONFIG_HOME` depends of your OS. Here are the defaults (if you didn't modify it):

* **Unix systems**: `~/.config`
* **macOS**: `~/Library/Application Support`
* **Windows**: `%LOCALAPPDATA%`


## Google Analytics 4

The `google_analytics` service uses the [Google Analytics Data API](https://developers.google.com/analytics/devguides/reporting/data/v1). You need:

1. A service account with a JSON keyfile, with the [Google Analytics Data API](https://console.cloud.google.com/apis/library/analyticsdata.googleapis.com) enabled for its project.
2. The service account's email added as a *Viewer* of your GA4 property (Admin > Property access management).
3. The ID of your GA4 property (Admin > Property details).

```yaml
services:
  google_analytics:
    keyfile: my-keyfile.json
    property_id: 123456789
```

The metric and dimension aliases (`sessions`, `page_views`, `users`, `bounces`, `bounce_rate`, `page_path`, `traffic_source`, `country`...) are mapped to GA4. You can also use any [GA4 metric or dimension](https://developers.google.com/analytics/devguides/reporting/data/v1/api-schema) directly, like `metric: "engagementRate"`. Universal Analytics names prefixed with `ga:` are converted when GA4 has an equivalent; `entrances` and `unique_page_views` don't exist anymore in GA4.

## Documentation

[The complete DevDash documentation is here](https://thedevdash.com). It was written for the upstream project: the Travis CI service doesn't exist anymore, and the Google Analytics service now needs a `property_id` (see above).

You'll find in there:

* [Installation / getting started](https://thedevdash.com/getting-started/installation/)
* [Simple examples](https://thedevdash.com/getting-started/examples/)
* [Real-life use cases](https://thedevdash.com/getting-started/use-cases/devdash/)
* [Complete reference of dashboard's config](https://thedevdash.com/reference/)

More real-life examples are in the [dotfiles of the original author](https://github.com/Phantas0s/.dotfiles/tree/master/devdash).

## Acknowledgement

Thanks to [MariaLetta](https://github.com/MariaLetta/free-gophers-pack) for the awesome and beautiful Gopher pack! I used it for my logo on top.

DevDash was inspired from other open source projects:

* [wtf](https://github.com/wtfutil/wtf)
* [tdash](https://github.com/jessfraz/tdash)

## Bugs and Ideas

If you have any idea to improve DevDash or if you see any bug, please open an issue.

## Contribute

Thanks a lot if you want to contribute to DevDash!

If you want to implement a new feature, let's speak about it first (by opening an issue) and decide if it fits DevDash scope.

## Making Of

For anybody interested how I managed to develop DevDash on side of a full time job, and how I organized my time and kept my motivation, [I wrote an article about that on my blog](https://thevaluable.dev/programming-side-project-example-devdash/).

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
