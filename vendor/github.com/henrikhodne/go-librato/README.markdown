# go-librato

go-librato is a Go client library for accessing the [Librato API][].

**Documentation**: [![GoDoc](https://godoc.org/github.com/henrikhodne/go-librato/librato?status.svg)](https://godoc.org/github.com/henrikhodne/go-librato/librato)  
**Build Status:** [![Build Status](https://travis-ci.org/henrikhodne/go-librato.svg?branch=master)](https://travis-ci.org/henrikhodne/go-librato)  

[Librato API]: http://dev.librato.com/v1

go-travis requires Go version 1.5 or later. It may work with earlier versions,
but this is not supported and new functionality that breaks this will not be
considered a breaking change.

## Usage

```go
import "github.com/henrikhodne/go-librato"
```

Construct a new Librato client, then use the various services on the client to
access different parts of the Librato API. For example, to list all spaces:

```go
client := librato.NewClient("me@example.com", "your-librato-token")
spaces, _, err := client.Spaces.List(nil)
```

Some API methods have optional parameters that can be passed. For example, to
list spaces with names that match "foobar":

```go
client := librato.NewClient("me@example.com", "your-librato-token")
opt := &librato.SpaceListOptions{Name: "foobar"}
spaces, _, err := client.Spaces.List(opt)
```

## Authentication

You need to get an API token for the account you want to access Librato as
before using the API. You can generate API tokens in your [account settings][].
You then pass the email address of your account and the API token into the
`NewClient` function:

```go
client := librato.NewClient("your-email", "your-librato-token")
```

[account settings]: https://metrics.librato.com/account/api_tokens

## Roadmap

API methods will likely be implemented as they are needed, but the plan is to
eventually have full coverage of the API, so contributions are
[always welcome][contributing].

[contributing]: CONTRIBUTING.markdown

## License

This library is distributes under the MIT license found in the
[LICENSE](./LICENSE) file.
