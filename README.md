# Simple Budget Tracker

[![Actions Status](https://github.com/<github_username>/<repo>/workflows/build/badge.svg)](https://github.com/w-k-s/simple-budget-tracker/actions)
[![codecov](https://codecov.io/gh/<github_username>/<repo>/branch/master/graph/badge.svg)](https://codecov.io/gh/w-k-s/simple-budget-tracker)

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
- [Setting up `codecov` with Github actions](https://gist.github.com/Harold2017/d98607f242659ca65e731c688cb92707)
