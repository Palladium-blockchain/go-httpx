package httpx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadJSON_ReadsTypedPayload(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	got, err := ReadJSON[payload](strings.NewReader(`{"name":"alice","age":42}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Name != "alice" || got.Age != 42 {
		t.Fatalf("got %+v, want name=alice age=42", *got)
	}
}

func TestReadJSONWithOptions_DisallowUnknownFields(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	_, err := ReadJSONWithOptions[payload](
		strings.NewReader(`{"name":"alice","age":42}`),
		WithDisallowUnknownFields(),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReadJSONWithOptions_RejectsMultipleValues(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	_, err := ReadJSONWithOptions[payload](strings.NewReader(`{"name":"alice"}{"name":"bob"}`))
	if !errors.Is(err, ErrMultipleJSONValues) {
		t.Fatalf("expected ErrMultipleJSONValues, got %v", err)
	}
}

func TestReadJSONWithOptions_RejectsBodyTooLarge(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	_, err := ReadJSONWithOptions[payload](
		strings.NewReader(`{"name":"alice"}`),
		WithMaxBytes(8),
	)
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("expected ErrBodyTooLarge, got %v", err)
	}
}

func TestWriteJSON_WritesDefaultJSONResponse(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := WriteJSON(recorder, map[string]string{"ok": "yes"}, http.StatusCreated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if recorder.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusCreated)
	}
	if got := recorder.Header().Get("Content-Type"); got != ApplicationJSONUTF8 {
		t.Fatalf("got content-type %q, want %q", got, ApplicationJSONUTF8)
	}
	if body := recorder.Body.String(); body != "{\"ok\":\"yes\"}\n" {
		t.Fatalf("got body %q", body)
	}
}

func TestWriteJSONWithOptions_AppliesOptions(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := WriteJSONWithOptions(
		recorder,
		map[string]string{"html": "<b>safe</b>"},
		WithStatusCode(http.StatusAccepted),
		WithEscapeHTML(false),
		WithContentType(ApplicationJSON),
		WithHeader("X-Trace-Id", "abc-123"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusAccepted)
	}
	if got := recorder.Header().Get("Content-Type"); got != ApplicationJSON {
		t.Fatalf("got content-type %q, want %q", got, ApplicationJSON)
	}
	if got := recorder.Header().Get("X-Trace-Id"); got != "abc-123" {
		t.Fatalf("got X-Trace-Id %q, want %q", got, "abc-123")
	}
	if body := recorder.Body.String(); body != "{\"html\":\"<b>safe</b>\"}\n" {
		t.Fatalf("got body %q", body)
	}
}

func TestWriteError_WritesErrorEnvelope(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := WriteError(recorder, errors.New("boom"), http.StatusBadRequest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if body := recorder.Body.String(); body != "{\"error\":\"boom\"}\n" {
		t.Fatalf("got body %q", body)
	}
}

func TestWriteErrorWithOptions_RejectsNilError(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := WriteErrorWithOptions(recorder, nil)
	if !errors.Is(err, ErrNilError) {
		t.Fatalf("expected ErrNilError, got %v", err)
	}
}
