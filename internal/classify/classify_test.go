package classify_test

import (
	"testing"

	"github.com/codec404/protoguard/internal/classify"
	"github.com/codec404/protoguard/internal/model"
)

func TestAssignImpactRemovePath(t *testing.T) {
	c := model.Change{Path: "openapi.paths./x", Kind: model.ChangeRemove}
	if classify.AssignImpact(c) != model.ImpactBreaking {
		t.Fatal()
	}
}

func TestAssignImpactAddPath(t *testing.T) {
	c := model.Change{Path: "openapi.paths./x", Kind: model.ChangeAdd}
	if classify.AssignImpact(c) != model.ImpactNonBreaking {
		t.Fatal()
	}
}
