# Contributing to DevDash

Thanks for your help! Bug reports, ideas and pull requests are welcome. For a new feature, please open an issue first, to agree on what it should do.

## Development

You need Go 1.26 or later.

```shell
go build .                      # build ./devdash
go test -race ./...             # test
go vet ./...
golangci-lint run               # lint, with the version of the CI (see .github/workflows/ci.yml)
./devdash check my-dashboard    # check a dashboard
```

The tests also check that every example dashboard, template and dashboard of the documentation is valid, and that the published JSON schema is up to date.

## Structure

| Directory           | Content                                                                                       |
| ------------------- | --------------------------------------------------------------------------------------------- |
| `cmd/`              | The commands, the loading of the dashboards, the default dashboard                            |
| `internal/`         | The widgets of each service (`*_widget.go`), the layout of the projects, the registry of the widgets and their options, the checks and the JSON schema |
| `internal/platform` | The clients of the services (Google, Github, SSH...) and the terminal interface (tcell)      |
| `doc/`              | The documentation, its images, and the JSON schema                                            |
| `example/`          | Example dashboards                                                                            |
| `scripts/`          | The generation of the screenshots                                                             |

## Adding a widget

1. Add its name as a constant and its case in `CreateWidgets`, in the `*_widget.go` file of its service.
2. Add it with its options to `internal/registry.go`: `devdash check` and the JSON schema use it.
3. Generate the JSON schema again: `go run . schema > doc/devdash.schema.json`.
4. Document it on the page of its service in `doc/services/`, and add it to the [changelog](./CHANGELOG.md).

## Screenshots

The screenshots of the documentation are generated with [`scripts/screenshots`](./scripts/screenshots/README.md), with sample data for the Google services.

## Releases

Push a tag `vX.Y.Z` on `master`: the release workflow builds the binaries with GoReleaser and publishes the release. Move the `[Unreleased]` changes of the changelog to the new version first.
