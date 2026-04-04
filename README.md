# Go HTTPX

Lightweight Go helpers for reading and writing JSON over HTTP.

## Overview

`go-httpx` provides a small `httpx` package focused on the most common JSON response and request tasks in HTTP handlers:

- decode request bodies into typed Go values
- encode JSON responses with status codes and headers
- return a standard JSON error envelope

The package keeps the default API very small, while also exposing option-based helpers for stricter decoding and configurable response writing.

## Installation

```bash
go get github.com/Palladium-blockchain/go-httpx
```

## Quick Start

```go
package main

import (
	"errors"
	"net/http"

	"github.com/Palladium-blockchain/go-httpx/pkg/httpx"
)

type createUserRequest struct {
	Name string `json:"name"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.ReadJSON[createUserRequest](r.Body)
	if err != nil {
		_ = httpx.WriteError(w, err, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		_ = httpx.WriteError(w, errors.New("name is required"), http.StatusBadRequest)
		return
	}

	_ = httpx.WriteJSON(w, map[string]string{"status": "created"}, http.StatusCreated)
}
```

## Usage

### Reading JSON

Use `ReadJSON` for the default case:

```go
payload, err := httpx.ReadJSON[MyPayload](r.Body)
```

If you need stricter decoding, use `ReadJSONWithOptions`:

```go
payload, err := httpx.ReadJSONWithOptions[MyPayload](
	r.Body,
	httpx.WithDisallowUnknownFields(),
	httpx.WithMaxBytes(1<<20),
)
```

Available read options:

- `WithDisallowUnknownFields()` rejects unexpected JSON fields
- `WithMaxBytes(limit)` rejects payloads larger than the configured limit

### Writing JSON

Use `WriteJSON` for the common case:

```go
err := httpx.WriteJSON(w, response, http.StatusOK)
```

For more control, use `WriteJSONWithOptions`:

```go
err := httpx.WriteJSONWithOptions(
	w,
	response,
	httpx.WithStatusCode(http.StatusAccepted),
	httpx.WithContentType(httpx.ApplicationJSON),
	httpx.WithHeader("X-Request-Id", "abc-123"),
	httpx.WithEscapeHTML(false),
)
```

### Writing Errors

`WriteError` returns a standard JSON error payload:

```go
err := httpx.WriteError(w, errors.New("invalid payload"), http.StatusBadRequest)
```

Response body:

```json
{"error":"invalid payload"}
```
