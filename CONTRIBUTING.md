# Contributing to terraform-provider-statusgator

Thank you for your interest in contributing!

## Development Setup

### Requirements

- [Go](https://golang.org/doc/install) >= 1.25
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [golangci-lint](https://golangci-lint.run/usage/install/)

### Clone and Build

```shell
git clone https://github.com/arslanbekov/terraform-provider-statusgator.git
cd terraform-provider-statusgator
make build
```

### Running Tests

```shell
# Unit tests
make test

# Acceptance tests (requires API credentials)
export STATUSGATOR_API_TOKEN="your-token"
export STATUSGATOR_TEST_BOARD_ID="your-board-id"
make testacc
```

### Local Installation

```shell
make install
```

Then add to your `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "arslanbekov/statusgator" = "/path/to/go/bin"
  }
  direct {}
}
```

## Making Changes

### Code Style

- Run `gofmt -s -w .` before committing
- Run `make lint` to check for issues
- Follow existing code patterns

### Adding a New Resource

1. Create `internal/provider/<resource>_resource.go`
2. Create `internal/provider/<resource>_resource_test.go`
3. Add the resource to `provider.go` in the `Resources()` function
4. Create an example in `examples/resources/statusgator_<resource>/resource.tf`
5. Run `make docs` to generate documentation

### Adding a New Data Source

1. Create `internal/provider/<datasource>_data_source.go`
2. Create `internal/provider/<datasource>_data_source_test.go`
3. Add the data source to `provider.go` in the `DataSources()` function
4. Create an example in `examples/data-sources/statusgator_<datasource>/data-source.tf`
5. Run `make docs` to generate documentation

### Documentation

Documentation is auto-generated from:
- Schema descriptions in the Go code
- Examples in `examples/` directory
- Template in `templates/index.md.tmpl`

To regenerate documentation:

```shell
make docs
```

### Testing

- All new features must include tests
- All bug fixes should include a regression test
- Aim for high test coverage

### Commit Messages

Follow conventional commit format:

- `feat: add new resource`
- `fix: handle nil pointer in read`
- `docs: update examples`
- `test: add acceptance tests`
- `refactor: simplify error handling`
- `deps: update terraform-plugin-framework`
- `ci: add dependabot`

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linter (`make lint`)
6. Update documentation (`make docs`)
7. Commit your changes
8. Push to your fork
9. Create a Pull Request

### PR Checklist

- [ ] Tests pass
- [ ] Linter passes
- [ ] Documentation updated
- [ ] CHANGELOG.md updated (if applicable)
- [ ] PR description explains the changes

## Release Process

Releases are automated via GitHub Actions when a tag is pushed:

```shell
git tag v1.0.0
git push origin v1.0.0
```

## Getting Help

- Open an issue for bugs or feature requests
- Check existing issues before creating a new one
- Provide as much detail as possible

## Code of Conduct

Be respectful and constructive in all interactions.
