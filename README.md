# Simple Budget Tracker

[![Actions Status](https://github.com/<github_username>/<repo>/workflows/build/badge.svg)](https://github.com/w-k-s/simple-budget-tracker/actions)

[![codecov](https://codecov.io/gh/<github_username>/<repo>/branch/master/graph/badge.svg)](https://codecov.io/gh/w-k-s/simple-budget-tracker)

## Setting up Project

**Githooks**

This project uses pre-commit githooks to run `go fmt` and `golangci-lint`
The githook is located in the `.githooks/` directory and you'll need to update your git configuration to look for githooks in this directory. This can be done by running the following command:

```sh
git config core.hooksPath .githooks
```

**Golangci-lint**

You'll also need to install `golangci-lint`; on an OS-X machine, this can be installed using:

```sh
brew install golangci-lint
brew upgrade golangci-lint
```

For other operating systems, refer to the `golangci-lint`'s [documentation](https://golangci-lint.run/usage/install/#local-installation)

**Dependencies**

To install dependencies, run

```
go get
```

In ordeer to host the swagger documentation, `statik` is used to create a static file system that hosts the editor pages. Ensure that statik is installed when installing dependencies. If it isn't you can run:

```
go get github.com/rakyll/statik
```

## Useful Resources

- [Project Layout](https://github.com/golang-standards/project-layout)
- [Setting up `codecov` with Github actions](https://gist.github.com/Harold2017/d98607f242659ca65e731c688cb92707)
- [Serving Swagger UI as static file system](https://ribice.medium.com/serve-swaggerui-within-your-golang-application-5486748a5ed4)
