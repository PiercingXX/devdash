# Dates

Many widgets display data for a period, with the options `start_date` and `end_date`. Both accept:

* A date with the format `YYYY-MM-DD`, for example `2024-01-31`.
* An alias relative to the current day. An alias can mean a different day as `start_date` or as `end_date`.

Replace `x` with a number in the aliases below. The weeks begin on Monday.

| Alias          | As `start_date`                              | As `end_date`                                |
| -------------- | -------------------------------------------- | -------------------------------------------- |
| `today`        | Today                                        | Today                                        |
| `yesterday`    | Yesterday                                    | Yesterday                                    |
| `x_days_ago`   | `x` days before today                        | `x` days before today                        |
| `this_week`    | Monday of the current week                   | Sunday of the current week                   |
| `last_week`    | Monday of the previous week                  | Sunday of the previous week                  |
| `x_weeks_ago`  | Monday of the week `x` weeks ago             | Sunday of the week `x` weeks ago             |
| `this_month`   | First day of the current month               | Last day of the current month                |
| `last_month`   | First day of the previous month              | Last day of the previous month               |
| `x_months_ago` | First day of the month `x` months ago        | Last day of the month `x` months ago         |
| `this_year`    | January 1st of the current year              | December 31st of the current year            |
| `last_year`    | January 1st of the previous year             | December 31st of the previous year           |
| `x_years_ago`  | January 1st of the year `x` years ago        | December 31st of the year `x` years ago      |

For example, `start_date: 3_months_ago` and `end_date: this_month` cover the three previous months and the current one.

When they are not given, `start_date` is `7_days_ago` and `end_date` is `today`. Some widgets have other defaults, or don't accept every alias: see the page of their service.
