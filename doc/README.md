# DevDash Documentation

DevDash is a terminal dashboard. Each dashboard is a configuration file (YAML, JSON or TOML) describing which data to fetch and how to display it.

## Getting started

* [Installation and first dashboard](./getting-started.md): install DevDash, run the default dashboard, the commands and the keyboard shortcuts.
* [Examples](./examples.md): complete dashboards with their screenshot.
* [Upgrading from upstream DevDash](./upgrading.md): what changed compared to [Phantas0s/devdash](https://github.com/Phantas0s/devdash).

## Reference

* [Configuration](./configuration.md): the structure of a dashboard, the general options, projects, the grid of rows and columns, and themes.
* [Widgets and display options](./widgets.md): widget names and types, sizes, heights, colors, and every display option.
* [Dates](./dates.md): the values accepted by `start_date` and `end_date`.
* [Checking a dashboard](./configuration.md#checking-a-dashboard): `devdash check`, and the JSON schema for the completion of your editor.

## Services

Each widget belongs to a service. The prefix of the widget name is the service ID.

| Service                                                     | Widget prefix | What it displays                                         |
| ----------------------------------------------------------- | ------------- | -------------------------------------------------------- |
| [Monitor](./services/monitor.md)                            | `mon`         | Availability and ping of a website                       |
| [Host](./services/host.md)                                  | `lh`, `rh`    | System metrics and command output, local or through SSH  |
| [Github](./services/github.md)                              | `github`      | Stars, views, commits, issues, pull requests, Actions    |
| [Google Analytics 4](./services/google-analytics.md)        | `ga`          | Sessions, users, pages, traffic sources, countries...    |
| [Google Search Console](./services/google-search-console.md) | `gsc`         | Clicks, impressions, CTR and position of queries / pages |
| [Feedly](./services/feedly.md)                              | `feedly`      | Number of subscribers of a website                       |
| [Git](./services/git.md)                                    | `git`         | Branches of a local repository                           |
| [Display](./services/display.md)                            | `display`     | A box with any text                                      |
| [JSON](./services/json.md)                                  | `json`        | Values of any API answering in JSON                      |
