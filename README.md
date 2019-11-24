# mux

Package mux implements HTTP application lifecycle helpers.

## Overview

This package centers around the `Handler`, a flexible implementation of the
standard `http.Handler` interface that is so much more than a HTTP request
multiplexer, but less than a framework.

Use of this package gives you the following:

- request router/dispatcher
- middleware (top-level and per-route)
- automatic HEAD responses
- automatic OPTIONS responses
- 400 Bad Request responses on decode errors
- 405 Method Not Allowed responses
- 406 Not Acceptable plain text error
- 415 Unsupported Media Type responses on content type errors
- 422 Unprocessable Entity responses on form validation errors
- 500 Internal Server Error responses on panic
- response buffer pool to eliminate partially rendered responses
- request identifiers for instrumentation
- locale detection for internationalization
- export routes to static files

Individual components of this package are customizable and/or replacable where
possible. Application structure is not imposed. Defaults for a greenfield JSON
API are preconfigured.

Route patterns are mapped to by a `Router`. The `Router` may also be a
`Builder` or `Walker` for additional functionality.

`HandlerFunc`s are like `http.HandlerFunc`s but return `error`s that are
resolved to HTTP error responses through a `Resolver` that transforms `error`s
to `Error` views. The error views are encoded and written as a response.

The error views are encoded with `Encode`, also available for use within
`HandlerFunc`s for any encoded response. The handler can be configured with a
mapping of request `Accept` headers to `Encoder`s. A request that is not mapped
to an encoder will be served a plain text HTTP 406 Not Acceptable error.

Use `Decode` to decode incoming request data to some validated data structure.
The handler can be configured with a mapping of request `Content-Type` headers
to `Decoder`s. A request that is not mapped to a decoder will be served a HTTP
415 Unsupported Media Type error, encoded per the request `Accept` header.

The unique request identifier can be retrieved with `RequestID`. This may be
useful for implementing custom `Logger`s or `Error` views.

The detected locale can be retrieved with `Locale`. The supported locales can
be set when creating the handler using `WithLocales` and defaults to English.