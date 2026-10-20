# Feedly

Display the number of [Feedly](https://feedly.com) subscribers of a website. It doesn't need any account.

## Configuration

```yaml
services:
  feedly:
    address: https://thevaluable.dev
```

| Option    | Description                                    |
| --------- | ---------------------------------------------- |
| `address` | The address of the website or of its feed      |

## Widgets

| Widget                   | Type | Description                                                                              |
| ------------------------ | ---- | ---------------------------------------------------------------------------------------- |
| `feedly.box_subscribers` | box  | Subscribers of the first feed Feedly finds for the address, `0` if it doesn't find any  |

It accepts `title`, and the [box display options](../widgets.md#box).
