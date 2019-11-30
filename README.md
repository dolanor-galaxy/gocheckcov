# gocheckcov
![GitHub Workflow Status (branch)](https://img.shields.io/github/workflow/status/cvgw/gocheckcov/Go/master?style=plastic)
![Coveralls github branch](https://img.shields.io/coveralls/github/cvgw/gocheckcov/master?style=plastic)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/cvgw/gocheckcov?style=plastic)
[![license](https://img.shields.io/github/license/cvgw/gocheckcov?style=plastic)](./LICENSE)

* [Status](#status)
* [Description](#description)
* [Usage](#usage)
  * [Supported Golang Versions](#supported-golang-versions)
  * [Install](#install)
  * [Configuration](#configuration)
* [Development](#development)
* [Contributing](#contributing)

## Status
Alpha

## Description
gocheckcov allows users to assert that a set of golang packages meets a minimum level of test coverage. Users supply a coverage profile generated by `go test -coverprofile` and a path to a tree of golang packages. Users can specify a minimum coverage percentage for all packages or specify a minimum for each package via a configuration file. If each package does not meet the specified minimum coverage gocheckcov will exit with code 1.

```
$ gocheckcov check --profile-file cp.out --minimum-coverage 66.6 $GOPATH/src/github.com/bar/foo/pkg/baz

pkg github.com/bar/foo/pkg/baz coverage is 10% (10/100 statements)
coverage 10% for package github.com/bar/foo/pkg/baz did not meet minimum 66.6%

$ echo $?
1
```

## Usage
```
gocheckcov help
```

### Supported Golang Versions
* 1.11.x
* 1.12.x
* 1.13.x

### Install
`go get github.com/cvgw/gocheckcov`

### Configuration
Specify minimum coverage for each package via a configuration file
```
#.gocheckcov-config.yaml
min_coverage_percentage: 25
packages:
- name: github.com/bar/foo/pkg/baz
  # this overrides the global val of min_coverage_percentage for only this package
  mininum_coverage_percentage: 66.6
```

## Development
gocheckcov uses `dep` for dependency management and `golangci-lint` for linting. See the [development guide](./DEVELOPMENT.md) for more info.

## Contributing
Contributors are welcome and appreciated. Please read the [contributing guide](./CONTRIBUTING.md) for more info.
