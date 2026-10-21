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

| Option        | Description                                                  | Default               |
| ------------- | ------------------------------------------------------------ | --------------------- |
| `username`    | The SSH user                                                 |                       |
| `address`     | The host and its SSH port; the port is required              |                       |
| `known_hosts` | The file with the keys of the known hosts                    | `~/.ssh/known_hosts`  |

DevDash authenticates with the keys of your [ssh-agent](https://www.ssh.com/academy/ssh/agent): the agent needs to run (`SSH_AUTH_SOCK` set) with a key accepted by the remote host. Passwords are not supported.

DevDash verifies the key of the remote host with your known hosts, like `ssh` does. If the host isn't known yet, connect to it once with `ssh -p <port> <username>@<host>` to add it. The connection is kept between the refreshes, and a command which doesn't finish within 30 seconds is displayed as an error.

## Widgets

Replace `rh` with `lh` for the local host.

| Widget                 | Type  | Description                                                                 |
| ---------------------- | ----- | --------------------------------------------------------------------------- |
| `rh.box_uptime`        | box   | Time since the last boot                                                    |
| `rh.box_load`          | box   | Load averages over 1, 5 and 15 minutes                                      |
| `rh.box_processes`     | box   | Running processes / total processes                                         |
| `rh.box_cpu_rate`      | box   | CPU usage (%), measured during half a second                                |
| `rh.box_memory_rate`   | box   | Memory usage (%)                                                            |
| `rh.box_swap_rate`     | box   | Swap usage (%)                                                              |
| `rh.box_net_io`        | box   | Data received / sent since the last boot, or per second with `rate`. The loopback, bridges, container and VM interfaces, and VPN tunnels are not counted: their traffic goes through a physical interface too. |
| `rh.box_disk_io`       | box   | Data read / written on the physical disks since the last boot, or per second with `rate`. The partitions and virtual devices are not counted twice. |
| `rh.gauge_cpu_rate`    | gauge | CPU usage (%), measured during half a second                                |
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
| `rate`    | `true` to display the throughput per second, instead of the total since the boot | `false`                   | `box_net_io`, `box_disk_io` |
| `command` | The command to run                                                  | An example                                      | `box`, `gauge`, `bar`, `table` |

## Commands

The widgets `box`, `gauge`, `bar` and `table` display the output of `command`:

| Widget  | Expected output                                                                                               |
| ------- | ------------------------------------------------------------------------------------------------------------- |
| `box`   | Any text                                                                                                      |
| `gauge` | A number between 0 and 100 on the first line                                                                  |
| `bar`   | Positive integers separated with spaces or new lines, one per bar. Give their labels with `headers`.          |
| `table` | One row per line, the cells separated with spaces. Without `headers`, the first line is used as headers. The cells can't contain commas. |

The command is run by a shell: `sh` on the local host (`cmd` on Windows), the shell of the SSH user on a remote host. Pipes, quotes, variables and redirections work. A command which doesn't finish within 30 seconds is displayed as an error.

On the local host, a command writing anything on its error output is displayed as an error: add `2>/dev/null` to ignore it.

## Example

A complete dashboard is in [`example/localhost.yml`](../../example/localhost.yml):

![localhost dashboard](../img/localhost.png)

The widgets displaying your own commands:

```yaml
- name: lh.table
  options:
    title: " uname -srmo "
    command: "uname -srmo"
    headers: "kernel,release,machine,os"
- name: lh.bar
  options:
    title: " Output of a command "
    command: "echo 3 7 5 9 4"
    headers: "mon,tue,wed,thu,fri"
- name: lh.gauge
  options:
    title: " Root filesystem usage "
    command: "df --output=pcent / | tail -n 1 | tr -d ' %'"
- name: rh.box
  options:
    title: " Last login "
    command: "last -n 1 | head -n 1"
```
