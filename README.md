# keybd

[![license](https://badgen.net/static/license/MIT/blue?cache-3600)](https://spdx.org/licenses/MIT.html)

## Overview

`keybd` is a [Go](https://go.dev) module that can perform keyboard synthesization on both MacOS and Windows desktops.

**License**: [MIT](LICENSE)

## Usage

### Getting Started

From your shell:

```text
go get "github.com/kamaranl/keybd"
```

In your code:

```go
//go:build (windows || darwin)

package myapp

import "github.com/kamaranl/keybd"

```

## TODO

* Add detailed examples.
