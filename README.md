# go-onfido

<!-- [![CircleCI](https://circleci.com/gh/getground/go-onfido.svg?style=svg)](https://circleci.com/gh/getground/go-onfido) [![Go Report Card](https://goreportcard.com/badge/github.com/getground/go-onfido)](https://goreportcard.com/report/github.com/getground/go-onfido) -->

Client for the [Onfido API](https://documentation.onfido.com/)

[![go-doc](https://godoc.org/github.com/getground/go-onfido?status.svg)](https://godoc.org/github.com/getground/go-onfido)

> This library was built for Utility Warehouse internal projects, so priority was given to supporting the
features we needed. If the library is missing a feature from the API, raise an issue or ideally open a PR.

## Important

This library is used by [Module Core](https://github.com/getground/module-core). Do `go get -u github.com/getground/go-onfido` there if you wish to update the module version used.

## Installation

To install go-onfido, use `go get`:

```
go get github.com/getground/go-onfio
```

## Usage

First you're going to need to instantiate a client (grab your [sandbox API key](https://onfido.com/dashboard/v2/#/api/tokens))

```golang
client := onfido.NewClient("test_123")
```

Or you can instantiate usign the env variable `ONFIDO_TOKEN`

```golang
client, err := onfido.NewClientFromEnv()
```

Now checkout some of the [examples](https://github.com/getground/go-onfido/tree/master/examples)
