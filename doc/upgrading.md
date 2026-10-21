# Upgrading from Upstream DevDash

This fork is compatible with the dashboards of [Phantas0s/devdash](https://github.com/Phantas0s/devdash) 0.5.0, except for the changes below. The complete list is in the [changelog](../CHANGELOG.md).

The website thedevdash.com documents the original version: use this documentation instead.

## Google Analytics

Google shut down Universal Analytics, and its API. The `google_analytics` service now uses Google Analytics 4:

1. Enable the [Google Analytics Data API](https://console.cloud.google.com/apis/library/analyticsdata.googleapis.com) in the Google Cloud project of your service account. Your keyfile still works.
2. Give the service account access to your GA4 property, see [Google Analytics](./services/google-analytics.md#give-the-service-account-access-to-your-property).
3. Replace `view_id` with the `property_id` of your GA4 property:

```yaml
services:
  google_analytics:
    keyfile: my-keyfile.json
    property_id: 123456789   # was: view_id: 12345678
```

The aliases (`sessions`, `page_views`, `users`, `page_path`...) work as before. The Universal Analytics names prefixed with `ga:` are converted when GA4 has an equivalent: `ga:pageviews` becomes `screenPageViews`, `ga:users` becomes `totalUsers`... Other names are sent to GA4 without the `ga:` prefix, so `ga:sessionsPerUser` still works. Some metrics don't exist anymore:

* `entrances` and `unique_page_views` display an error.
* `bounces` is computed as the sessions without engagement. GA4 doesn't count bounces like Universal Analytics did, so the numbers are different.

The default metrics of the table widgets are now `sessions,page_views,users`, instead of `sessions,page_views,entrances,unique_page_views`.

## Travis CI

travis-ci.org was shut down: the `travis` service and the widget `travis.table_builds` are removed.

## Options renamed

These options were documented, but didn't work with their documented name:

| Widgets                   | Before                  | Now                    |
| ------------------------- | ----------------------- | ---------------------- |
| Box widgets               | `false: true`           | `multiline: true`      |
| `github.bar_commits`      | `owner: all`            | `scope: all`           |
| Templates of `devdash generate -t blog` | `reload`   | `hot_reload`           |

## Behavior fixed

* `gsc.table_pages` displayed queries instead of pages.
* `ga.bar_returning` displayed every user instead of the returning ones.
* `ga.bar_bounces` added ` Bounces ` to your title.
* On Sundays, the weeks (`this_week`, `last_week`, `x_weeks_ago`) began on Saturday.
* `end_date: this_year` and `end_date: x_years_ago` were the first day of the year.
* `rh.box_disk_io` displayed values 1024 times too big, and counted the I/O of a disk again for each of its partitions and virtual devices.
* `mon.box_ping` only pinged once, when DevDash started.
* The Google Analytics `filters` beginning with `-` now exclude their value, as they already did for Google Search Console.

## Terminal

The terminal interface uses [tcell](https://github.com/gdamore/tcell) instead of termui and termbox. The widgets look the same, but:

* The `size` option of `name_options` is now applied.
* `CTRL+e` gives the terminal to your editor until you quit it, then reloads the dashboard.
* The keys keep the same notation (`C-c`, `M-x`, `<f5>`...).
* The text of the boxes is wrapped between words.
* When a bar chart is too narrow for its bars, the last (most recent) ones are displayed instead of the first ones.
* The columns of a table too wide for its widget are shortened.
* The numbers on a colored bar or gauge are black by default, to be readable.

## Other changes

* The default dashboard, created on the first run, checks `https://github.com` and displays the metrics of your computer (on Linux), instead of a fixed blog.
* `github.table_pull_requests` displays the author instead of the `merged` and `commits` columns, which were always `unknown`. Merged pull requests have the state `merged`.
* The pages of `gsc.table_pages` keep their leading `/`, and the home page is `/`.
* `devdash list` and `devdash edit` find the JSON and TOML dashboards, and the dashboards of the current directory.
