# gocheckcov

## Status
Alpha

## Description
gocheckcov allows users to assert that a set of golang packages meets a minimum level of test covreage. Users supply a coverage profile generated by `go test -coverprofile` and a path to a tree of golang packages. Users can specify a minimum coverage percentage for all packages or specify a minimum for each package via a configuration file. If each package does not meet the specified minimum coverage gocheckcov with exit with code 1.

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

### Configuration
Specify minimum coverage for each package via a configuration file
```
#.gocheckcov-config.yaml
packages:
- name: github.com/bar/foo/pkg/baz
  mininum_coverage_percentage: 66.6
```
