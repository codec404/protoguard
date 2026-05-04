package protoguard_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/codec404/protoguard/pkg/protoguard"
)

func TestDiffOpenAPI_SkipLLM(t *testing.T) {
	root := filepath.Join("..", "..", "testdata")
	res, err := protoguard.Diff(context.Background(), protoguard.Options{
		OldPath: filepath.Join(root, "openapi_old.yaml"),
		NewPath: filepath.Join(root, "openapi_new.yaml"),
		Spec:    protoguard.SpecOpenAPI,
		SkipLLM: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Breaking || len(res.Report.Changes) == 0 {
		t.Fatalf("report: %+v", res.Report)
	}
}
