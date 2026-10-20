# Host

Display system metrics, or the output of your own commands and scripts, from your computer or from a remote host through SSH.

The system metrics are read from `/proc` and `df`: they need a Linux host. The command widgets work on any system.

## Local host

The local host doesn't need any configuration: use the widgets prefixed with `lh`, for example `lh.box_uptime`.

## Remote host

Use the widgets prefixed with `rh`, for example `rh.box_uptime`, and configure the service `remote_host`:

```yaml
services:
  remote_host:
    username: my-user
    address: "203.0.113.10:22"
```

| Option     | Description                                      |
| ---------- | ------------------------------------------------ |
| `username` | The SSH user                                     |
| `address`  | The host and its SSH port; the port is required  |

DevDash authenticates with the keys of your [ssh-agent](https://www.ssh.com/academy/ssh/agent): the agent needs to run (`SSH_AUTH_SOCK` set) with a key accepted by the remote host. Passwords are not supported.

> [!WARNING]
> DevDash doesn't verify the key of the remote host. Only connect to hosts on networks you trust.

## Widgets

Replace `rh` with `lh` for the local host.

| Widget                 | Type  | Description                                                                 |
| ---------------------- | ----- | --------------------------------------------------------------------------- |
| `rh.box_uptime`        | box   | Time since the last boot                                                    |
| `rh.box_load`          | box   | Load averages over 1, 5 and 15 minutes                                      |
| `rh.box_processes`     | box   | Running processes / total processes                                         |
| `rh.box_cpu_rate`      | box   | CPU usage (%), averaged since the last boot                                 |
| `rh.box_memory_rate`   | box   | Memory usage (%)                                                            |
| `rh.box_swap_rate`     | box   | Swap usage (%)                                                              |
| `rh.box_net_io`        | box   | Data received / sent on every network interface except `lo`, since the last boot |
| `rh.box_disk_io`       | box   | Data read / written on every disk, since the last boot                      |
| `rh.gauge_cpu_rate`    | gauge | CPU usage (%), averaged since the last boot                                 |
| `rh.gauge_memory_rate` | gauge | Memory usage (%)                                                            |
| `rh.gauge_swap_rate`   | gauge | Swap usage (%)                                                              |
| `rh.bar_rates`         | bar   | CPU, memory and swap usage (%)                                              |
| `rh.bar_memory`        | bar   | Values of `/proc/meminfo`                                                   |
| `rh.table_disk`        | table | Size, usage and mount point of the filesystems (without `tmpfs`, `devtmpfs` and `debugfs`) |
| `rh.box`               | box   | Output of a command                                                         |
| `rh.gauge`             | gauge | Output of a command, as a percentage                                        |
| `rh.bar`               | bar   | Output of a command, as bars                                                |
| `rh.table`             | table | Output of a command, as a table                                             |

## Options

Every widget accepts `title`, and the [display options](../widgets.md#display-options) of its type.

| Option    | Description                                                         | Default                                         | Widgets |
| --------- | ------------------------------------------------------------------- | ----------------------------------------------- | ------- |
| `unit`    | Unit of the values: `b`, `kb`, `mb`, `gb`, `tb`, `pb` or `eb` (powers of 1024) | `kb`, `gb` for `table_disk`           | `box_net_io`, `box_disk_io`, `bar_memory`, `table_disk` |
| `metrics` | Fields of `/proc/meminfo` to display, separated with commas         | `MemTotal,MemFree,MemAvailable`                 | `bar_memory` |
| `headers` | Labels of the bars, or columns of the table, separated with commas  | The metrics for `bar_memory`; `Filesystem,Size,Used,Available,Use%,Mount` for `table_disk` | `bar_memory`, `table_disk`, `bar`, `table` |
| `command` | The command to run                                                  | An example                                      | `box`, `gauge`, `bar`, `table` |

## Commands

The widgets `box`, `gauge`, `bar` and `table` display the output of `command`:

| Widget  | Expected output                                                                                               |
| ------- | ------------------------------------------------------------------------------------------------------------- |
| `box`   | Any text                                                                                                      |
| `gauge` | A number between 0 and 100 on the first line                                                                  |
| `bar`   | Positive integers separated with spaces or new lines, one per bar. Give their labels with `headers`.          |
| `table` | One row per line, the cells separated with spaces. Without `headers`, the first line is used as headers. The cells can't contain commas. |

On the local host, DevDash runs the command without a shell: pipes (`|`) work, but quotes are removed and other shell features (redirections, variables, `&&`...) don't. Put anything more complex in a script, and run it with `command: "sh ./my-script.sh"`. On a remote host, the command is run by the shell of the SSH user.

On the local host, a command writing anything on its error output is displayed as an error.

## Example

```yaml
projects:
  - name: My Sweet Localhost
    widgets:
      - row:
          - col:
              size: 8
              elements:
                - name: lh.bar_memory
                  options:
                    unit: "mb"
                    color: yellow
                    bar_gap: 1
                    bar_width: 10
                - name: lh.table_disk
                  options:
                    color: magenta
                - name: lh.table
                  options:
                    title: " ls -lah "
                    command: "ls -lah | tail -n +2"
                    headers: "rights,links,user,group,size,month,day,time,name"
          - col:
              size: 4
              elements:
                - name: lh.gauge_cpu_rate
                  options:
                    bar_color: green
                - name: lh.gauge_memory_rate
                  options:
                    bar_color: green
                - name: lh.box_uptime
                  options:
                    color: blue
                - name: lh.box_load
                  options:
                    color: blue
                - name: lh.box_net_io
                  options:
                    color: green
                    unit: "mb"
                - name: lh.bar
                  options:
                    headers: "a,b,c"
                    command: "echo 10 20 30"
```
