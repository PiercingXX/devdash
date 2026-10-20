# Git

Display the branches of a local Git repository. It needs `git` in your `PATH`.

## Configuration

```yaml
services:
  git:
    path: /home/me/projects/devdash
```

| Option | Description                         |
| ------ | ----------------------------------- |
| `path` | The directory of the repository. `~` isn't expanded: use an absolute path, or a path relative to the directory where you run DevDash. |

## Widgets

| Widget               | Type  | Description                                                                                      |
| -------------------- | ----- | ------------------------------------------------------------------------------------------------ |
| `git.table_branches` | table | Local branches, from the oldest to the most recently updated, with the author, email and date of their last commit |

It accepts `title`, and the [table display options](../widgets.md#table).
