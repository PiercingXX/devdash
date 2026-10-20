# Configuration

A dashboard is a configuration file with two parts:

* `general`: the options of the whole dashboard.
* `projects`: one or more projects. Each project has a title, its services, its themes, and its widgets.

```yaml
general:
  refresh: 600

projects:
  - name: My blog
    name_options:
      color: yellow
    services:
      monitor:
        address: "https://example.com"
    themes:
      box:
        border_color: green
    widgets:
      - row:
          - col:
              size: 6
              elements:
                - name: mon.box_availability
                - name: lh.box_uptime
          - col:
              size: 6
              elements:
                - name: lh.gauge_cpu_rate
```

![structure of a DevDash configuration](./img/struct.png)

## General

| Option    | Description                                                                                      | Default   | Example         |
| --------- | ------------------------------------------------------------------------------------------------ | --------- | --------------- |
| `refresh` | Seconds between two refreshes of every widget                                                    | `60`      | `600`           |
| `editor`  | Editor opened by the `edit` key. The `devdash edit` command uses `$EDITOR` or its `-e` option instead. | `$EDITOR` | `vim`, `/usr/bin/nano` |
| `keys`    | Keyboard shortcuts, see below                                                                    |           |                 |

Each refresh calls the APIs of every widget. Keep it high enough not to reach their quotas: Google Analytics, Search Console and Github all limit the number of requests.

### Keys

| Option       | Description                                                                         | Default |
| ------------ | ----------------------------------------------------------------------------------- | ------- |
| `quit`       | Quit DevDash                                                                        | `C-c`   |
| `hot_reload` | Read the configuration again and refresh every widget                               | `C-r`   |
| `edit`       | Edit the configuration in the editor; the dashboard is reloaded when the editor exits | `C-e`   |

The keys use these notations:

* A character: `q`, `R`, `?`.
* `C-` for CTRL and `M-` for ALT (meta): `C-c`, `M-x`.
* Special keys: `<f1>` to `<f12>`, `<enter>`, `<tab>`, `<backtab>`, `<backspace>`, `<escape>`, `<space>`, `<delete>`, `<insert>`, `<home>`, `<end>`, `<previous>` (page up), `<next>` (page down), `<up>`, `<down>`, `<left>`, `<right>`.

```yaml
general:
  refresh: 600
  editor: nvim
  keys:
    quit: "q"
    hot_reload: "<f5>"
    edit: "C-e"
```

## Projects

A dashboard can display multiple projects, one after the other. A project has:

| Option         | Description                                                    |
| -------------- | -------------------------------------------------------------- |
| `name`         | The title of the project, displayed in a box above its widgets |
| `name_options` | The display of the title, see below                            |
| `services`     | The configuration of the services used by the widgets          |
| `themes`       | Reusable sets of widget options, see [Themes](#themes)         |
| `widgets`      | The rows of widgets                                            |

### Name options

| Option         | Description                             | Default |
| -------------- | --------------------------------------- | ------- |
| `size`         | Width of the title, see [sizes](./widgets.md#width) | `XXL` (the whole width) |
| `height`       | Height of the title                     | `3`     |
| `bold`         | Display the title in bold               | `true`  |
| `color`        | Color of the text and the border        | default |
| `text_color`   | Color of the text                       | default |
| `border_color` | Color of the border                     | default |

The colors are listed in [Colors](./widgets.md#colors).

### Services

Each service has its own configuration, documented on its page. A widget can only be used if its service is configured in the same project, except for the `lh` (local host) and `display` widgets, which don't need any configuration.

| Service                 | Page                                                         |
| ----------------------- | ------------------------------------------------------------ |
| `monitor`               | [Monitor](./services/monitor.md)                             |
| `remote_host`           | [Host](./services/host.md)                                   |
| `github`                | [Github](./services/github.md)                               |
| `google_analytics`      | [Google Analytics 4](./services/google-analytics.md)         |
| `google_search_console` | [Google Search Console](./services/google-search-console.md) |
| `feedly`                | [Feedly](./services/feedly.md)                               |
| `git`                   | [Git](./services/git.md)                                     |

## Rows and columns

The widgets are placed on a grid:

* `widgets` is a list of rows.
* A `row` is a list of columns.
* A `col` has a `size` (its width) and `elements`, the list of its widgets. The widgets of a column are displayed one below the other.

A row is 12 units wide. The sizes of the columns of a row shouldn't add up to more than 12. When they add up to exactly 12, the last column fills the remaining width of the terminal. The height of a row is the height of its highest column.

```yaml
widgets:
  - row:
      - col:
          size: 8      # two thirds of the width
          elements:
            - name: lh.bar_memory
            - name: lh.table_disk
      - col:
          size: S      # one third of the width
          elements:
            - name: lh.gauge_cpu_rate
            - name: lh.gauge_memory_rate
  - row:
      - col:
          size: XXL    # the whole width
          elements:
            - name: lh.table
```

Each element of a column is a widget:

| Option    | Description                                                                         |
| --------- | ----------------------------------------------------------------------------------- |
| `name`    | The widget, for example `mon.box_availability`. See [Widgets](./widgets.md)          |
| `options` | The options of the widget, described on each service's page and in [Widgets](./widgets.md#display-options) |
| `theme`   | The name of a theme to apply, see below                                             |

## Themes

A theme is a set of widget options defined once in the project, and applied to many widgets.

```yaml
projects:
  - name: DevDash
    services:
      github:
        owner: PiercingXX
        repository: devdash
    themes:
      bar:
        color: yellow
        title_color: red
        bar_color: green
        bar_gap: 1
      table:
        border_color: green
        row_limit: 10
      ocean:
        border_color: blue
        num_color: black
        bar_color: cyan
        title_color: magenta
    widgets:
      - row:
          - col:
              size: 12
              elements:
                # A bar widget: the theme "bar" is applied.
                - name: github.bar_stars
                # The theme "ocean" is applied instead of the theme "bar".
                - name: github.bar_views
                  theme: ocean
                  # The options of the widget override the options of its theme.
                  options:
                    bar_color: red
```

* A theme named after a widget type (`box`, `bar`, `gauge` or `table`) is applied to every widget of this type.
* A widget with a `theme` uses this theme instead of the theme of its type. The two are not merged.
* The options of the widget override the options of its theme.

A theme can hold any option, not only display options: `row_limit`, `start_date`...
