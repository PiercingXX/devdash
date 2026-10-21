# JSON

Display values from any API answering in JSON: your own services, a status page, Plausible, GitLab... It doesn't need any service configuration.

## Widgets

| Widget       | Type  | Description                                                         |
| ------------ | ----- | ------------------------------------------------------------------- |
| `json.box`   | box   | The value at the `path`                                             |
| `json.gauge` | gauge | The number at the `path`, as a percentage                           |
| `json.bar`   | bar   | One bar per element of the array at the `path`                      |
| `json.table` | table | One row per element of the array at the `path`, with the `columns`  |

## Options

Every widget accepts `title`, and the [display options](../widgets.md#display-options) of its type.

| Option      | Description                                                                                        | Widgets       |
| ----------- | -------------------------------------------------------------------------------------------------- | ------------- |
| `url`       | The URL to request with GET. Required.                                                             | All           |
| `path`      | Where the value is in the JSON, see below. Without a path, the whole JSON is used.                 | All           |
| `header`    | A header of the request, written `Name: value`. `${VARIABLE}` is replaced by the environment variable, to keep the tokens out of the dashboard. | All |
| `label`     | The path of the label in each element. By default, the bars are numbered.                         | `json.bar`    |
| `value`     | The path of the value in each element. By default, the element itself.                            | `json.bar`    |
| `columns`   | The paths of the cells in each element, separated with commas. Required.                          | `json.table`  |
| `headers`   | The headers of the columns, separated with commas. By default, the `columns`.                      | `json.table`  |
| `row_limit` | Maximum number of rows                                                                             | `json.table`  |

The requests time out after 30 seconds. An answer which isn't a success (2xx) or isn't JSON is displayed as an error.

### Paths

The paths use the [GJSON syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md):

| JSON                                         | Path              | Value      |
| -------------------------------------------- | ----------------- | ---------- |
| `{"stats": {"visitors": 42}}`                | `stats.visitors`  | `42`       |
| `{"items": [{"name": "a"}, {"name": "b"}]}`  | `items.1.name`    | `b`        |
| `{"items": [{"name": "a"}, {"name": "b"}]}`  | `items.#`         | `2`        |
| `[{"up": true}, {"up": false}]`              | `#(up==false)#`   | The elements down |

## Examples

```yaml
projects:
  - name: JSON APIs
    widgets:
      - row:
          - col:
              size: 4
              elements:
                - name: json.box
                  options:
                    title: " tcell stars "
                    url: https://api.github.com/repos/gdamore/tcell
                    path: stargazers_count
                - name: json.table
                  options:
                    title: " tcell releases "
                    url: https://api.github.com/repos/gdamore/tcell/releases?per_page=3
                    columns: tag_name,published_at
                    headers: release,date
          - col:
              size: 8
              elements:
                - name: json.bar
                  options:
                    title: " tcell commits per week "
                    url: https://api.github.com/repos/gdamore/tcell/stats/participation
                    path: all
                    bar_color: green
```

With a token in the environment variable `PLAUSIBLE_TOKEN`, the visitors of the last 30 days in [Plausible](https://plausible.io/docs/stats-api):

```yaml
- name: json.box
  options:
    title: " Visitors - last 30 days "
    url: https://plausible.io/api/v1/stats/aggregate?site_id=example.com&period=30d
    header: "Authorization: Bearer ${PLAUSIBLE_TOKEN}"
    path: results.visitors.value
```
