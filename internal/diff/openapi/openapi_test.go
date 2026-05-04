package openapi_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	odiff "github.com/codec404/protoguard/internal/diff/openapi"
)

func TestDiffDetectsNewOperation(t *testing.T) {
	l := openapi3.NewLoader()
	oldD, err := l.LoadFromData([]byte(`openapi: "3.0.3"
info: {title: T, version: "1"}
paths:
  /a:
    get:
      responses:
        "200": {description: ok}
`))
	if err != nil {
		t.Fatal(err)
	}
	newD, err := l.LoadFromData([]byte(`openapi: "3.0.3"
info: {title: T, version: "1"}
paths:
  /a:
    get:
      responses:
        "200": {description: ok}
    post:
      responses:
        "201": {description: created}
`))
	if err != nil {
		t.Fatal(err)
	}
	r := odiff.Diff(oldD, newD)
	if len(r.Changes) < 1 {
		t.Fatalf("expected changes, got %d", len(r.Changes))
	}
}
