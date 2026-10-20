# Widgets and Display Options

## Widget names

A widget name has the form `<service>.<type>_<data>`. For example, `ga.bar_pages` is a widget of the service Google Analytics (`ga`), displayed as a `bar` chart.

The service prefixes are listed in the [documentation index](./README.md#services). The widgets of each service are listed on its page.

There are four types of widgets:

| Type    | Display                                                                             |
| ------- | ----------------------------------------------------------------------------------- |
| `box`   | A text in a box                                                                     |
| `gauge` | A percentage, as a horizontal gauge                                                 |
| `bar`   | A bar chart. A few Google Analytics bar widgets are stacked bar charts.             |
| `table` | A table, with a line between each row                                               |

## Display options

Every widget accepts the options of its type, in addition to the options of its service.

### Box

| Option         | Description                                                            | Default                 |
| -------------- | ---------------------------------------------------------------------- | ----------------------- |
| `title`        | Title, displayed on the border                                         | Depends on the widget   |
| `height`       | Height, borders included                                               | `3`                     |
| `color`        | Color of the title, the border and the text                            | default                 |
| `title_color`  | Color of the title                                                     | default                 |
| `border_color` | Color of the border                                                    | default                 |
| `text_color`   | Color of the text                                                      | default                 |
| `bold`         | Display the text in bold                                               | `false`                 |
| `multiline`    | Wrap the text and make the box as high as needed to display all of it. Without it, a text too long for the box ends with `…`. | `false` |

### Gauge

| Option         | Description                                          | Default               |
| -------------- | ---------------------------------------------------- | --------------------- |
| `title`        | Title                                                | Depends on the widget |
| `height`       | Height, borders included                             | `3`                   |
| `color`        | Color of the title, the border, the text and the bar | default               |
| `title_color`  | Color of the title                                   | default               |
| `border_color` | Color of the border                                  | default               |
| `text_color`   | Color of the percentage                              | default               |
| `bar_color`    | Color of the gauge                                   | default (reverse video) |

### Bar

| Option            | Description                                                            | Default               |
| ----------------- | ---------------------------------------------------------------------- | --------------------- |
| `title`           | Title                                                                  | Depends on the widget |
| `height`          | Height, borders and labels included                                    | `10`                  |
| `bar_width`       | Width of each bar                                                      | `6`                   |
| `bar_gap`         | Space between two bars                                                 | `0`                   |
| `color`           | Color of the title, the border, the labels and the bars. The numbers become black. | default   |
| `title_color`     | Color of the title                                                     | default               |
| `border_color`    | Color of the border                                                    | default               |
| `text_color`      | Color of the labels below the bars                                     | default               |
| `bar_color`       | Color of the bars                                                      | default (reverse video) |
| `num_color`       | Color of the numbers displayed on the bars                             | default               |
| `empty_num_color` | Color of the numbers of empty bars                                     | default               |

The bars that don't fit in the width of the widget are not displayed: reduce `bar_width` or `bar_gap`, or make the column wider.

#### Stacked bar

`ga.bar_new_returning` and `ga.bar_devices` stack several series in each bar. They accept the bar options, except `bar_color` and `empty_num_color`, and the colors of the series:

| Option         | Description                 | Default   |
| -------------- | --------------------------- | --------- |
| `first_color`  | Color of the first series   | `blue`    |
| `second_color` | Color of the second series  | `green`   |
| `third_color`  | Color of the third series   | `yellow`  |
| `fourth_color` | Color of the fourth series  | `red`     |
| `fifth_color`  | Color of the fifth series   | `magenta` |

The series are sorted by name, so they keep the same color between refreshes. At most five series are displayed.

### Table

| Option         | Description                                 | Default               |
| -------------- | ------------------------------------------- | --------------------- |
| `title`        | Title                                       | Depends on the widget |
| `color`        | Color of the title, the border and the text | default               |
| `title_color`  | Color of the title                          | default               |
| `border_color` | Color of the border                         | default               |
| `text_color`   | Color of the text                           | default               |

The height of a table depends on its number of rows. Most table widgets limit the rows with the option `row_limit`.

## Sizes

### Width

The width of a widget is the `size` of its column. The grid has 12 units; a size is a number from `1` to `12`, or one of these T-shirt sizes (case doesn't matter):

| Size  | Units | Width         |
| ----- | ----- | ------------- |
| `XXS` | 1     | 1/12          |
| `XS`  | 2     | 1/6           |
| `S`   | 4     | 1/3           |
| `M`   | 6     | 1/2           |
| `L`   | 8     | 2/3           |
| `XL`  | 10    | 5/6           |
| `XXL` | 12    | the whole row |

See [Rows and columns](./configuration.md#rows-and-columns).

### Height

The option `height` is a number of terminal lines, borders included. `tput lines` gives the number of lines of your terminal.

## Colors

The colors are the colors of your terminal's palette:

| Name      |
| --------- |
| `default` |
| `black`   |
| `red`     |
| `green`   |
| `yellow`  |
| `blue`    |
| `magenta` |
| `cyan`    |
| `white`   |

`default` is the default foreground color of your terminal. An unknown color is replaced by `default`.
