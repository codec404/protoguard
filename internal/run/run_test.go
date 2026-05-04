package run_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/codec404/protoguard/internal/run"
)

func TestDiffOpenAPIBreakingPathRemoval(t *testing.T) {
	root := filepath.Join("..", "..", "testdata")
	oldP := filepath.Join(root, "openapi_old.yaml")
	newP := filepath.Join(root, "openapi_new.yaml")

	res, err := run.Diff(context.Background(), run.Options{
		OldPath: oldP,
		NewPath: newP,
		Spec:    run.SpecOpenAPI,
		SkipLLM: true,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Breaking {
		t.Fatalf("expected breaking changes, report: %+v", res.Report.Changes)
	}
	found := false
	for _, c := range res.Report.Changes {
		if c.Path == "openapi.paths./pets" && c.Kind == "remove" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected removal of /pets, got %#v", res.Report.Changes)
	}
}

func TestDiffProtobufFieldRemovalBreaking(t *testing.T) {
	root := filepath.Join("..", "..", "testdata")
	oldP := filepath.Join(root, "proto_old.pb")
	newP := filepath.Join(root, "proto_new.pb")

	res, err := run.Diff(context.Background(), run.Options{
		OldPath: oldP,
		NewPath: newP,
		Spec:    run.SpecProtobuf,
		SkipLLM: true,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Report.SpecKind != "protobuf" {
		t.Fatalf("spec kind: got %q", res.Report.SpecKind)
	}
	if !res.Breaking {
		t.Fatalf("expected breaking changes, got %#v", res.Report.Changes)
	}
	foundRemove := false
	for _, c := range res.Report.Changes {
		if c.Kind == "remove" && c.Path == "grpc.field.pgtest.Item.1" {
			foundRemove = true
			break
		}
	}
	if !foundRemove {
		t.Fatalf("expected removal of field 1 on pgtest.Item, got %#v", res.Report.Changes)
	}
}
