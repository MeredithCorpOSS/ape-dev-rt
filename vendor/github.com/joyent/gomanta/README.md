gomanta
=======

The gomanta package enables Go programs to interact with the Joyent Manta service.

## Installation

Use `go-get` to install gomanta
```
go get github.com/joyent/gomanta
```

## Packages

The gomanta package is structured as follow:

	- gomanta/localservices. This package provides local services to be used for testing.
	- gomanta/manta. This package interacts with the Manta API (http://apidocs.joyent.com/manta/).


## Documentation

Documentation can be found on godoc.

- [http://godoc.org/github.com/joyent/gomanta](http://godoc.org/github.com/joyent/gomanta)
- [http://godoc.org/github.com/joyent/gomanta/localservices](http://godoc.org/github.com/joyent/gomanta/localservices)
- [http://godoc.org/github.com/joyent/gomanta/manta](http://godoc.org/github.com/joyent/gomanta/manta)

## Testing

Make sure you have the dependencies

```
go get "launchpad.net/gocheck"
```

To Run all tests
```
go test ./...
```

## License
Licensed under [MPLv2](LICENSE).

Copyright (c) 2016 Joyent Inc.
Written by Daniele Stroppa <daniele.stroppa@joyent.com>