# DevDash

![logo of devdash with a gopher](./doc/img/logo.jpg)

[![CI](https://github.com/PiercingXX/devdash/actions/workflows/ci.yml/badge.svg)](https://github.com/PiercingXX/devdash/actions/workflows/ci.yml) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENCE)

A configurable terminal dashboard: pick the metrics you care about and see them in one place.

![a blog dashboard with Google Analytics and Search Console](./doc/img/blog.png)

Maintained fork of the archived [Phantas0s/devdash](https://github.com/Phantas0s/devdash). Coming from the original? See [Upgrading](./doc/upgrading.md).

## What it shows

* **Your computer or a server (SSH):** CPU, memory, disks, network, and the output of your own commands.
* **Github:** stars, commits, issues, pull requests, Actions runs.
* **Google Analytics 4** and **Search Console**.
* **Anything else** with a JSON API, plus website availability, Feedly and local Git branches.

## Install

```shell
git clone https://github.com/PiercingXX/devdash.git
cd devdash
go build .   # needs Go 1.26+
```

Or download a binary from the [releases](https://github.com/PiercingXX/devdash/releases/latest).

## Use

Run `devdash`. The first run creates `~/.config/devdash/default.yml` and displays it.

| Key / command            | Does                                   |
| ------------------------ | -------------------------------------- |
| `CTRL+e`                 | Edit the dashboard                     |
| `CTRL+r`                 | Reload it                              |
| `CTRL+n` / `CTRL+p`      | Next / previous dashboard              |
| `CTRL+c`                 | Quit                                   |
| `devdash -c <name>`      | Open another dashboard                 |
| `devdash check <name>`   | Find mistakes in a dashboard           |

## Documentation

Everything is in [`doc/`](./doc/README.md): [getting started](./doc/getting-started.md), [examples](./doc/examples.md), [configuration](./doc/configuration.md), and one page per service.

## Credits

Created by [Matthieu Cneude](https://github.com/Phantas0s). Gopher from MariaLetta's [free gophers pack](https://github.com/MariaLetta/free-gophers-pack). Licensed under [Apache 2.0](./LICENCE).
