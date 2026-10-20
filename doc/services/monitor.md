# Monitor

Check that a website is up.

## Configuration

```yaml
services:
  monitor:
    address: "https://example.com"
```

| Option    | Description                                         |
| --------- | --------------------------------------------------- |
| `address` | The URL to monitor, beginning with `http://` or `https://` |

## Widgets

| Widget                 | Type | Description                                                                                                  |
| ---------------------- | ---- | ------------------------------------------------------------------------------------------------------------ |
| `mon.box_availability` | box  | Send a GET request to the address. Display `online (200)` for a status code 200, `offline (<status code>)` otherwise, or `offline (0)` without any answer. Redirections are followed. |
| `mon.box_ping`         | box  | Ping the host of the address once, and display the packets sent and received, and the round-trip time. |

## Options

| Option    | Description                       | Default                 | Widgets  |
| --------- | --------------------------------- | ----------------------- | -------- |
| `address` | Override the address of the service | The service's `address` | All      |
| `title`   | Title                             | ` Availability `, ` Ping ` | All   |

The [box display options](../widgets.md#box) are accepted too.

## Ping without root

`mon.box_ping` uses unprivileged ICMP sockets. On Linux, your group needs to be allowed to open them by the `net.ipv4.ping_group_range` sysctl. For example, to allow every group:

```shell
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

Otherwise the widget displays a "permission denied" error.

## Example

```yaml
projects:
  - name: Monitoring
    services:
      monitor:
        address: "https://example.com"
    widgets:
      - row:
          - col:
              size: M
              elements:
                - name: mon.box_availability
                  options:
                    color: green
                - name: mon.box_availability
                  options:
                    title: " Status page "
                    address: "https://status.example.com"
                - name: mon.box_ping
```
