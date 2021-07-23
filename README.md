# Simple Budget Tracker

## Setting up Project

### Setting up Git hooks

This project uses pre-commit githooks to run `go fmt` and `golangci-lint`
The githook is located in the `.githooks/` directory and you'll need to update your git configuration to look for githooks in this directory. This can be done by running the following command:

```sh
git config core.hooksPath .githooks
```

You'll also need to install `golangci-lint`; on an OS-X machine, this can be installed using:

```sh
brew install golangci-lint
brew upgrade golangci-lint
```

For other operating systems, refer to the `golangci-lint`'s [documentation](https://golangci-lint.run/usage/install/#local-installation)

### Installing Dependencies

To install dependencies, run

```
go get
```

## Useful Resources

- [Project Layout](https://github.com/golang-standards/project-layout)
