package httpx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func ReadJSON[T any](r io.Reader) (*T, error) {
	return ReadJSONWithOptions[T](r)
}

type ReadConfig struct {
	MaxBytes              int64
	DisallowUnknownFields bool
}

type ReadOption func(*ReadConfig)

func WithMaxBytes(limit int64) ReadOption {
	return func(c *ReadConfig) {
		c.MaxBytes = limit
	}
}

func WithDisallowUnknownFields() ReadOption {
	return func(c *ReadConfig) {
		c.DisallowUnknownFields = true
	}
}

func ReadJSONWithOptions[T any](r io.Reader, opts ...ReadOption) (*T, error) {
	if r == nil {
		return nil, ErrNilReader
	}

	cfg := ReadConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	reader := r
	var limited *io.LimitedReader
	if cfg.MaxBytes > 0 {
		limited = &io.LimitedReader{R: r, N: cfg.MaxBytes + 1}
		reader = limited
	}

	var target T
	decoder := json.NewDecoder(reader)
	if cfg.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if err := decoder.Decode(&target); err != nil {
		if limited != nil && limited.N <= 0 && errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, ErrBodyTooLarge
		}
		return nil, err
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if len(extra) != 0 {
		return nil, ErrMultipleJSONValues
	}
	if limited != nil && limited.N <= 0 {
		return nil, ErrBodyTooLarge
	}

	return &target, nil
}

func WriteJSON(w http.ResponseWriter, v any, code int) error {
	return WriteJSONWithOptions(w, v, WithStatusCode(code))
}

func WriteError(w http.ResponseWriter, err error, code int) error {
	return WriteErrorWithOptions(w, err, WithStatusCode(code))
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type WriteConfig struct {
	StatusCode  int
	ContentType string
	EscapeHTML  bool
	Headers     http.Header
}

type WriteOption func(*WriteConfig)

func WithStatusCode(code int) WriteOption {
	return func(c *WriteConfig) {
		c.StatusCode = code
	}
}

func WithContentType(contentType string) WriteOption {
	return func(c *WriteConfig) {
		c.ContentType = contentType
	}
}

func WithEscapeHTML(escape bool) WriteOption {
	return func(c *WriteConfig) {
		c.EscapeHTML = escape
	}
}

func WithHeader(key, value string) WriteOption {
	return func(c *WriteConfig) {
		if c.Headers == nil {
			c.Headers = make(http.Header)
		}
		c.Headers.Set(key, value)
	}
}

func WriteJSONWithOptions(w http.ResponseWriter, v any, opts ...WriteOption) error {
	if w == nil {
		return ErrNilResponseWriter
	}

	cfg := WriteConfig{
		StatusCode:  http.StatusOK,
		ContentType: ApplicationJSONUTF8,
		EscapeHTML:  true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	encoder.SetEscapeHTML(cfg.EscapeHTML)
	if err := encoder.Encode(v); err != nil {
		return fmt.Errorf("encode JSON response: %w", err)
	}

	header := w.Header()
	for key, values := range cfg.Headers {
		for _, value := range values {
			header.Add(key, value)
		}
	}
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", cfg.ContentType)
	}

	w.WriteHeader(cfg.StatusCode)
	if _, err := w.Write(body.Bytes()); err != nil {
		return fmt.Errorf("write JSON response: %w", err)
	}
	return nil
}

func WriteErrorWithOptions(w http.ResponseWriter, err error, opts ...WriteOption) error {
	if err == nil {
		return ErrNilError
	}
	return WriteJSONWithOptions(w, ErrorResponse{Error: err.Error()}, opts...)
}
