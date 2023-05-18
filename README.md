# Flipt OpenFeature Provider (Go)

[![CI](https://github.com/flipt-io/openfeature-provider-go/actions/workflows/ci.yml/badge.svg)](https://github.com/flipt-io/openfeature-provider-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/flipt-io/openfeature-provider-go/branch/main/graph/badge.svg?token=0X8OWMEV16)](https://codecov.io/gh/flipt-io/openfeature-provider-go)
![status](https://img.shields.io/badge/status-experimental-orange.svg)
![license](https://img.shields.io/github/license/flipt-io/flipt-openfeature-provider-go)
[![Go Reference](https://pkg.go.dev/badge/go.flipt.io/flipt-openfeature-provider.svg)](https://pkg.go.dev/go.flipt.io/flipt-openfeature-provider)

[![OpenFeature Specification](https://img.shields.io/static/v1?label=OpenFeature%20Specification&message=v0.5.1&color=yellow)](https://github.com/open-feature/spec/tree/v0.5.1)
[![OpenFeature SDK](https://img.shields.io/static/v1?label=OpenFeature%20Golang%20SDK&message=v1.0.0&color=green)](https://github.com/open-feature/go-sdk)

This repository and package provides a [Flipt](https://github.com/flipt-io/flipt) [OpenFeature Provider](https://docs.openfeature.dev/docs/specification/sections/providers) for interacting with the Flipt service backend using the [OpenFeature Go SDK](https://github.com/open-feature/go-sdk).

From the [OpenFeature Specification](https://docs.openfeature.dev/docs/specification/sections/providers):

> Providers are the "translator" between the flag evaluation calls made in application code, and the flag management system that stores flags and in some cases evaluates flags.

## Requirements

- Go 1.20+
- A running instance of [Flipt](https://www.flipt.io/docs/installation)

## Usage

### Installation

```bash
go get go.flipt.io/flipt-openfeature-provider
```

### Example

```go
package main

import (
    "context"

    "go.flipt.io/flipt-openfeature-provider/pkg/provider/flipt"
    "github.com/open-feature/go-sdk/pkg/openfeature"
)


func main() {
    // http://localhost:8080 is the default Flipt address
    openfeature.SetProvider(flipt.NewProvider())

    client := openfeature.NewClient("my-app")
    value, err := client.BooleanValue(context.Background(), "v2_enabled", false, openfeature.EvaluationContext{
        TargetingKey: "tim@apple.com",
        Attributes: map[string]interface{}{
            "favorite_color": "blue",
        },
    })

    if err != nil {
        panic(err)
    }

    if value {
        // do something
    } else {
        // do something else
    }
}
```

## Configuration

The Flipt provider allows you communicate with Flipt over either HTTP(S) or GRPC, depending on the address provided.

### HTTP(S)

```go
provider := flipt.NewProvider(flipt.WithAddress("https://localhost:443"))
```

#### Unix Socket

```go
provider := flipt.NewProvider(flipt.WithAddress("unix:///path/to/socket"))
```

### GRPC

#### HTTP/2

```go
type Token struct {}

func (t *Token) ClientToken() (string, error) {
    return "a-client-token", nil
}

provider := flipt.NewProvider(
    flipt.WithAddress("grpc://localhost:9000"),
    flipt.WithCertificatePath("/path/to/cert.pem"), // optional
    flipt.WithClientProvider(&Token{}), // optional
)
```

#### Unix Socket

```go
provider := flipt.NewProvider(
    flipt.WithAddress("unix:///path/to/socket"),
)
```
