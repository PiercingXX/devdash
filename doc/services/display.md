# Display

Display a box with your own text. It doesn't need any configuration.

## Widgets

| Widget        | Type | Description           |
| ------------- | ---- | --------------------- |
| `display.box` | box  | A box with any text   |

## Options

| Option    | Description         | Default      |
| --------- | ------------------- | ------------ |
| `content` | The text of the box | `No content` |
| `title`   | Title               | No title     |

The [box display options](../widgets.md#box) are accepted too. With `multiline: true`, the box grows to display the whole text.

## Example

```yaml
- name: display.box
  options:
    title: " TODO "
    content: "Write the next blog post"
    multiline: true
    bold: true
    color: cyan
```
