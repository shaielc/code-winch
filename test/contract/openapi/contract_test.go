package openapi_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOpenAPIContract(t *testing.T) {
	t.Parallel()

	for _, relative := range []string{
		"../../../api/openapi/code-winch.yaml",
		"v1.yaml",
	} {
		relative := relative
		t.Run(filepath.Base(relative), func(t *testing.T) {
			t.Parallel()
			document, err := openapi3.NewLoader().LoadFromFile(relative)
			if err != nil {
				t.Fatalf("load OpenAPI fixture %q: %v", relative, err)
			}
			if err := document.Validate(context.Background()); err != nil {
				t.Fatalf("validate OpenAPI fixture %q: %v", relative, err)
			}
		})
	}
}
