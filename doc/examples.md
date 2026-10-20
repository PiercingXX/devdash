# Examples

Each example is a complete dashboard: copy it into `$XDG_CONFIG_HOME/devdash`, replace the credentials and addresses with yours, and run `devdash -c <name>`.

The screenshots were taken with the original DevDash, with the same configurations.

## Monitoring

[`monitor.yml`](../example/monitor.yml): the availability of a website.

![monitoring dashboard](../example/img/monitor.png)

## Localhost

The system metrics of your computer, and the output of your own commands. The configuration is in [Host](./services/host.md#example).

## Google Analytics

[`ga-1.yml`](../example/ga-1.yml): bar charts of users, a stacked bar chart of new and returning users, real time users and the availability of the website. The columns use numbers and T-shirt sizes.

![google analytics dashboard](../example/img/ga-1.png)

[`ga-2.yml`](../example/ga-2.yml): the `general` options, monthly bars, and tables of pages and traffic sources.

![google analytics dashboard with tables](../example/img/ga-2.png)

## Google Search Console

[`gsc-1.yml`](../example/gsc-1.yml): the queries and pages of last month.

![google search console dashboard](../example/img/gsc-1.png)

## Github

[`git-1.yml`](../example/git-1.yml): repositories, branches and issues.

![github dashboard](../example/img/git-1.png)

A dashboard with themes, stars, views and commits is in [Github](./services/github.md#example).

## A blog

[`mix-1.yml`](../example/mix-1.yml): Google Analytics, Google Search Console and monitoring together, to follow a blog.

![blog dashboard](../example/img/mix-1.png)

### Page views of chosen pages

```yaml
projects:
  - name: https://thevaluable.dev - Pages
    services:
      google_analytics:
        keyfile: my-keyfile.json
        property_id: 123456789
    widgets:
      - row:
          - col:
              size: M
              elements:
                - name: ga.bar_pages
                  options:
                    title: " Developer stress "
                    start_date: 18_days_ago
                    color: blue
                    filters: "/developer-stress/"
          - col:
              size: M
              elements:
                # Equivalent to ga.bar_pages: ga.bar can query any metric and dimension.
                - name: ga.bar
                  options:
                    title: " DevDash case study "
                    color: cyan
                    start_date: 18_days_ago
                    dimensions: page_path
                    metric: page_views
                    filters: "/programming-side-project-example-devdash/"
      - row:
          - col:
              size: M
              elements:
                - name: ga.bar_pages
                  options:
                    title: " Vim search "
                    start_date: 18_days_ago
                    # Each element of a widget can have its own color.
                    title_color: green
                    border_color: yellow
                    bar_color: red
                    filters: "/vim-search/"
          - col:
              size: M
              elements:
                - name: ga.bar_pages
                  options:
                    title: " DRY principle "
                    start_date: 18_days_ago
                    color: red
                    filters: "/dry-principle-explained/"
```

![page views dashboard](../example/img/thevaluabledev-2.png)

### Pages per day

```yaml
projects:
  - name: https://thevaluable.dev - Pages per day
    services:
      google_analytics:
        keyfile: my-keyfile.json
        property_id: 123456789
    themes:
      table:
        metrics: page_views
        row_limit: 30
    widgets:
      - row:
          - col:
              size: XS
              elements:
                - name: ga.table_pages
                  options:
                    title: " 3 days ago "
                    start_date: 3_days_ago
                    end_date: 3_days_ago
                    color: yellow
          - col:
              size: XS
              elements:
                - name: ga.table_pages
                  options:
                    title: " 2 days ago "
                    start_date: 2_days_ago
                    end_date: 2_days_ago
                    color: blue
          - col:
              size: XS
              elements:
                - name: ga.table_pages
                  options:
                    title: " Yesterday "
                    start_date: yesterday
                    end_date: yesterday
                    color: magenta
          - col:
              size: XS
              elements:
                - name: ga.table_pages
                  options:
                    title: " Today "
                    start_date: today
                    end_date: today
                    color: green
```

![pages per day dashboard](../example/img/thevaluabledev-3.png)

More dashboards are in the [dotfiles of the original author](https://github.com/Phantas0s/.dotfiles/tree/master/devdash). They use Universal Analytics: see [Upgrading](./upgrading.md) to convert them.
