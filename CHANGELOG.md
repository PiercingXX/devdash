# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Maintained fork of [Phantas0s/devdash](https://github.com/Phantas0s/devdash), archived upstream in 2023.

### ADDED

* Google Analytics 4 support, with the Google Analytics Data API. The service `google_analytics` needs a `property_id` instead of a `view_id`.
* New metric aliases for Google Analytics: `engaged_sessions`, `new_users`, `active_users`, `engagement_rate`. New dimension aliases: `landing_page`, `referrer`.
* Install with `go install github.com/PiercingXX/devdash@latest`.
* Release binaries for arm64.
* The complete documentation in the `doc` directory, including a guide to upgrade from upstream DevDash.
* The Google Analytics `filters` beginning with `-` exclude their value, like for Google Search Console.
* `devdash edit` without argument edits the default dashboard.
* New example dashboards and screenshots: localhost, Github, Google Analytics, Search Console and a blog.

### UPDATED

* Replace the unmaintained termui/termbox with tcell. The size option of the project title is now honored.
* The terminal is given to the editor while editing a dashboard with `CTRL+e`.
* Port Google Search Console to the Search Console API, and Github to a current API client.
* The monitor widget pings on every refresh instead of once at startup, and displays ping errors.
* Several Google Analytics filters are now combined with OR instead of only using the first one.
* Upgrade Go and every dependency. CI now runs on GitHub Actions.
* The box option to wrap the text is now `multiline` (it was `false`), and the option of `github.bar_commits` to count every commit is now `scope` (it was `owner`), as documented.
* The default title of `mon.box_ping` is ` Ping `. The columns of `git.table_branches` describe the last commit of each branch.
* The default dashboard checks github.com and, on Linux, displays the metrics of the computer.
* The text of the boxes is wrapped between words.
* Bar charts too narrow for their bars display the last (most recent) ones, instead of the first ones.
* The widest columns of a table too wide for its widget are shortened.
* The numbers on a colored bar or gauge are black by default.
* `github.table_pull_requests` displays the author, and the state `merged`, instead of the `merged` and `commits` columns, which were always `unknown`.
* `rh.box_disk_io` only counts the physical disks: the I/O of a disk was counted again for each of its partitions and virtual devices.

### REMOVED

* The Travis CI service (travis-ci.org was shut down).
* The Google Analytics metrics `entrances` and `unique_page_views`, which don't exist in GA4.
* The GitHub funding link to the original author.
* The references to the original author's websites, in the default dashboard, the examples and the tests.
* Dead code: unused functions, options and configuration fields.

### FIXED

* `ga.bar_returning` displayed every user instead of the returning ones.
* Stacked bars changed colors on every refresh, and could crash with more than 5 values.
* The config generator swapped the address and the Google Analytics ID.
* The blog template generated invalid YAML and an ignored `reload` key.
* Google Search Console crashed on empty filters.
* The Linux install command didn't install anything.
* `gsc.table_pages` displayed queries instead of pages.
* `ga.bar_bounces` added " Bounces " to the title of the widget.
* `ga.table` without `dimension` had an empty header.
* On Sundays, the weeks began on Saturday.
* `end_date: this_year` and `end_date: x_years_ago` were the first day of the year instead of the last one.
* `rh.box_disk_io` and `lh.box_disk_io` displayed values 1024 times too big.
* The example dashboards used options that don't exist.
* `devdash version` displayed "gocket".
* A widget name without a dot crashed DevDash.
* `devdash list` ignored the JSON and TOML dashboards and the names with a dot, and `devdash edit` couldn't open the dashboards of the current directory.
* The `unit` option converted wrongly to a smaller unit, and uppercase units were considered as exabytes.
* `gsc.table_pages` removed the leading `/` of the pages, and displayed the home page as an empty line.
* The tables could cut a character in the middle with `character_limit`.
* The blog template of `devdash generate` used the unknown option `title_options` instead of `name_options`.

## [0.5.0] - 2021-04-25

### ADDED

* Automatically redraw the dashboard when the size of the terminal change.
* New command "list" - list all the dashboards available.
* New command "generate" - Generate dashboards according to templates.
* New command "edit" - Edit dashboard with the default $EDITOR or any other shell based editor you want.
* New command "version" - Display the version of DevDash.
* Add the possibility to edit the dashboard configuration file from the display with the keystroke `CTRL+e`.
* Add the possibility to set authorization tokens and files via environment variables:
    * "DEVDASH_GITHUB_TOKEN" - Token for the service Github.
    * "DEVDASH_GA_KEYFILE" - Keyfile for the service Google Analytics.
    * "DEVDASH_GSC_KEYFILE" - Keyfile for the service Google Search Console.

### UPDATED

* Fix errors linked to Google JWT files.
* Command version - Display the current version of DevDash
* Add possibility to edit file with an editor of your choice

## [0.4.2] - 2020-02-25

### Updated

* Possibility to overwrite the URL from the monitoring service in a monitoring widget (allows to monitor one address by box widget).
* Fix a small display bug (forgot to delete a useless debug output... oops).

## [0.4.1] - 2020-11-24

### Updated

* Fix filepaths when using DevDash on Windows

## [0.4.0] - 2020-06-08

### Added

* Google Analytics widgets:
    * ga.bar_countries
    * ga.bar_devices

* Remote host / local host service (grab data from command line via SSH on remote or local host)
    * rh.box_uptime 
    * rh.box_load
    * rh.box_net_io
    * rh.box_disk_io
    * rh.bar_memory
    * rh.gauge_cpu_rate
    * rh.gauge_memory_rate
    * rh.gauge_swap_rate
    * rh.bar_rates
    * rh.table_disk
    * rh.box
    * rh.gauge
    * rh.table
    * rh.bar

* Possibility to load dashboard configurations from `$XDG_CONFIG_HOME` only using filename.
    * Example: `devdash -c dashboard` if there is a file `$XDG_CONFIG_HOME/devdash/dashboard.yml/json`

### Updated 

* Increase performances

### Breaking Changes

* `devdash -config dashboad-config.yml` is not valid anymore. Replaced by `devdash -c dashboard-config.yml` or `devdash --config dashboard-config.yml`

## [0.3.0] - 2020-01-14

### Added

* Add Git service
    * Add widget table branches - display information about git branches
* Add Display service
    * Add widget box - display a widget box containing some text
* Add Travis CI service
    * Add Travis builds widget - display your last Travis CI builds
* Add Feedly service
    * Add Feedly subscribers widget - display number of Feedly subscribers for your website

## [0.2.0] - 2019-10-05

### Added

* Add Github widgets
  * Display count stars overtime
  * Display count commits overtime
  * Display issues
  * Display repositories in table with information
  * Display last week traffic on Github page

* Add `color` options to have same color for border, title and everything color related for one widget
* Add themes to simplify the configuration - possibility to use same options defined once, for multiple widgets

* Add possibility to hot reload any dashboard via a keystroke - no need to restart DevDash when changing a dashboard configuration

* Create the [official DevDash website](https://thedevdash.com)

### Updated 

* Replace `title_options` by `name_options` for project's config (breaking change)

## [0.1.1] - 2019-07-21

### Added

* use goreleaser for relases

## [0.1.0] - 2019-05-28

### Added

* Write README documentation
* Add Github widgets
  * github.box_stars
  * github.box_watchers
  * github.box_open_issues
  * github.table_branches
  * github.table_issues
  * github.table_repositories
* Add Github API
* Google Search Console widgets
  * gsc.table_pages
  * gsc.table_queries
  * gsc.table
* Add Google Search Console API
* Create ToTime library
* Google Analytics widgets:
  * ga.box_real_time
  * ga.box_total
  * ga.bar_sessions
  * ga.bar_bounces
  * ga.bar_users
  * ga.bar_returning
  * ga.bar_pages
  * ga.bar
  * ga.bar_new_returning
  * ga.table_pages
  * ga.table_traffic_sources
  * ga.table
* Google Analytics service
* Google Analytics API
* Monitoring service
* Dashboard refreshing system 
* Projects / Services / Widget system
* Display and Grid system
* YAML configuration system
