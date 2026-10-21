# Github

Display data about your Github repositories.

## Token

A token isn't required for public repositories, but Github only allows 60 requests per hour without one. To create a token:

1. On Github, go to *Settings > Developer settings > Personal access tokens > Fine-grained tokens*.
2. Click *Generate new token*.
3. For public repositories only, keep *Public repositories* (read-only access). For private repositories, choose them in *Repository access*, and give the *Metadata*, *Contents*, *Issues* and *Pull requests* permissions (read-only).
4. `github.bar_views` needs the *Administration* permission (read-only) on the repository: Github only gives the traffic of a repository to the users who can push to it.

You can also set the token with the environment variable `DEVDASH_GITHUB_TOKEN` instead of writing it in the dashboard.

## Configuration

```yaml
services:
  github:
    token: github_pat_123
    owner: PiercingXX
    repository: devdash
```

| Option       | Description                                                                          |
| ------------ | ------------------------------------------------------------------------------------ |
| `token`      | The Github token                                                                     |
| `owner`      | The owner of the repositories (a user or an organization)                            |
| `repository` | The default repository of the widgets. It can be overridden in each widget with the option `repository`. |

## Widgets

| Widget                       | Type  | Description                                                                                           |
| ---------------------------- | ----- | ----------------------------------------------------------------------------------------------------- |
| `github.box_stars`           | box   | Number of stars of the repository                                                                     |
| `github.box_watchers`        | box   | Number of watchers of the repository                                                                  |
| `github.box_open_issues`     | box   | Number of open issues and pull requests of the repository                                             |
| `github.table_repositories`  | table | Public repositories of the owner, with their stars, watchers, forks and open issues                   |
| `github.table_branches`      | table | Branches of the repository                                                                            |
| `github.table_issues`        | table | Issues and pull requests of the repository, open or closed, with their state                          |
| `github.table_pull_requests` | table | Pull requests of the repository, with their state (open, closed or merged), author and creation date |
| `github.bar_views`           | bar   | Views of the repository's page, per day, for the last 14 days                                         |
| `github.bar_commits`         | bar   | Commits per week, for the last 52 weeks at most                                                       |
| `github.bar_stars`           | bar   | Stars given per day. The days without a new star are not displayed.                                   |

## Options

Every widget accepts `title`, and the [display options](../widgets.md#display-options) of its type.

| Option       | Description                                                                         | Default                                | Widgets                                       |
| ------------ | ----------------------------------------------------------------------------------- | -------------------------------------- | --------------------------------------------- |
| `repository` | Override the repository of the service                                              | The service's `repository`             | Every widget except `table_repositories`      |
| `row_limit`  | Maximum number of rows                                                              | `5`                                    | Table widgets                                 |
| `metrics`    | Columns, separated with commas: `stars`, `watchers`, `forks`, `open_issues`         | `stars,watchers,forks,open_issues`     | `table_repositories`                          |
| `order`      | Order of the repositories: `created`, `updated`, `pushed` or `full_name`            | `pushed`                               | `table_repositories`                          |
| `start_date` | Beginning of the period                                                             | `7_days_ago`; `7_weeks_ago` for `bar_commits` | `bar_stars`, `bar_commits`             |
| `end_date`   | End of the period                                                                   | `today`; `0_weeks_ago` for `bar_commits` | `bar_stars`, `bar_commits`                  |
| `scope`      | Count the commits of the `owner` only, or of everybody (`all`)                      | `owner`                                | `bar_commits`                                 |

The dates are described in [Dates](../dates.md). For `github.bar_commits`, both dates need to be written `x_weeks_ago`, with `x` between `0` (the current week) and `52`.

## Example

[`example/github.yml`](../../example/github.yml) works without a token:

```yaml
projects:
  - name: gdamore/tcell
    name_options:
      color: green
    services:
      github:
        # token: or the environment variable DEVDASH_GITHUB_TOKEN
        owner: gdamore
        repository: tcell
    themes:
      table:
        color: yellow
        row_limit: 5
    widgets:
      - row:
          - col:
              size: 8
              elements:
                - name: github.bar_commits
                  options:
                    title: " Commits per week (everybody) "
                    start_date: 15_weeks_ago
                    end_date: 0_weeks_ago
                    scope: all
                    bar_width: 5
                    bar_gap: 1
                    color: green
          - col:
              size: 2
              elements:
                - name: github.box_stars
                  options:
                    title: " Stars "
                    color: yellow
                - name: github.box_watchers
                  options:
                    color: yellow
                - name: github.box_open_issues
                  options:
                    title: " Open issues "
                    color: red
          - col:
              size: 2
              elements:
                - name: github.table_branches
      - row:
          - col:
              size: 6
              elements:
                - name: github.table_pull_requests
          - col:
              size: 6
              elements:
                - name: github.table_repositories
                  options:
                    metrics: stars,forks
```

![github dashboard](../img/github.png)
