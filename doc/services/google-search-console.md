# Google Search Console

Display the clicks, impressions, click-through rate and position of your website in Google Search, with the [Search Console API](https://developers.google.com/webmaster-tools).

## Authorization

DevDash connects to Google with a service account and its JSON keyfile. You can use the same service account and keyfile as for [Google Analytics](./google-analytics.md#authorization).

1. Create a service account and its JSON keyfile, as described in [Google Analytics](./google-analytics.md#create-the-service-account-and-its-keyfile), but enable the [Google Search Console API](https://console.cloud.google.com/apis/library/searchconsole.googleapis.com) for the project.
2. In [Google Search Console](https://search.google.com/search-console/), select your property, and go to *Settings > Users and permissions*.
3. Add the email address of the service account as a user, with the *Restricted* permission.

## Configuration

```yaml
services:
  google_search_console:
    keyfile: my-keyfile.json
    address: https://example.com/
```

| Option    | Description                                                                                                         |
| --------- | ------------------------------------------------------------------------------------------------------------------- |
| `keyfile` | Path of the JSON keyfile. A relative path is searched in the current directory, then in `$XDG_CONFIG_HOME/devdash`. |
| `address` | The property, exactly as in Search Console: `https://example.com/` for a URL-prefix property, `sc-domain:example.com` for a domain property. |

You can also set the keyfile with the environment variable `DEVDASH_GSC_KEYFILE`.

## Widgets

| Widget              | Type  | Description                                     |
| ------------------- | ----- | ----------------------------------------------- |
| `gsc.table_queries` | table | Metrics per search query                        |
| `gsc.table_pages`   | table | Metrics per page. The address of the property is removed from the URLs. |
| `gsc.table`         | table | Metrics per any dimension, see `dimension`      |

## Options

Every widget accepts `title`, and the [table display options](../widgets.md#table).

| Option            | Description                                                                                       | Default                           |
| ----------------- | ------------------------------------------------------------------------------------------------- | --------------------------------- |
| `start_date`      | Beginning of the period, see [Dates](../dates.md)                                                 | `7_days_ago`                      |
| `end_date`        | End of the period                                                                                 | `today`                           |
| `dimension`       | `query`, `page`, `country`, `device`, `searchAppearance` or `date`. Not used by `gsc.table_pages`. | `query`                           |
| `metrics`         | Columns, separated with commas: `clicks`, `impressions`, `ctr`, `position`                        | `clicks,impressions,ctr,position` |
| `filters`         | See [Filters](#filters)                                                                           |                                   |
| `row_limit`       | Maximum number of rows                                                                            | `5`                               |
| `character_limit` | Maximum number of characters of the first column                                                  | `1000`                            |

The rows are ordered by clicks. Search Console data is usually available after two or three days: the most recent days of a period can be empty.

### Filters

The `filters` are values separated with commas. The rows need to match every one of them:

* `value`: the dimension contains `value`.
* `-value`: the dimension doesn't contain `value`.
* `*dimension value`: filter another dimension than the one displayed. For example, `*country fra` on `gsc.table_queries` keeps the queries from France (the countries are [ISO 3166-1 alpha-3](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-3) codes). `*page -/tag/` keeps the queries of every page except the tags.

## Example

```yaml
projects:
  - name: https://example.com
    services:
      google_search_console:
        keyfile: my-keyfile.json
        address: 'https://example.com/'
    widgets:
      - row:
          - col:
              size: 6
              elements:
                - name: gsc.table_queries
                  options:
                    title: " Last month queries (without my brand) "
                    start_date: last_month
                    end_date: last_month
                    character_limit: 25
                    row_limit: 10
                    metrics: "clicks,impressions"
                    filters: "-example"
          - col:
              size: 6
              elements:
                - name: gsc.table_pages
                  options:
                    title: " Last month pages "
                    start_date: last_month
                    end_date: last_month
                    row_limit: 10
                - name: gsc.table
                  options:
                    title: " Clicks per device "
                    dimension: device
                    metrics: clicks
```

![google search console dashboard](../../example/img/gsc-1.png)
