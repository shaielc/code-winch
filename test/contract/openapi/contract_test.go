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
			// code-winch.yaml assembles paths/ by $ref, so the loader must
			// follow them; validation then covers the assembled document
			// rather than the root file alone.
			loader := openapi3.NewLoader()
			loader.IsExternalRefsAllowed = true
			document, err := loader.LoadFromFile(relative)
			if err != nil {
				t.Fatalf("load OpenAPI fixture %q: %v", relative, err)
			}
			if err := document.Validate(context.Background()); err != nil {
				t.Fatalf("validate OpenAPI fixture %q: %v", relative, err)
			}
		})
	}
}
