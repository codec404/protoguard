package proto_test

import (
	"testing"

	"google.golang.org/protobuf/proto"
	dpb "google.golang.org/protobuf/types/descriptorpb"

	pdiff "github.com/codec404/protoguard/internal/diff/proto"
)

func fileWithMessage(pkg, msg string, fields []*dpb.FieldDescriptorProto) *dpb.FileDescriptorProto {
	return &dpb.FileDescriptorProto{
		Name:    proto.String("t.proto"),
		Package: proto.String(pkg),
		MessageType: []*dpb.DescriptorProto{
			{
				Name:  proto.String(msg),
				Field: fields,
			},
		},
	}
}

func TestProtoFieldRemovalBreaking(t *testing.T) {
	old := &dpb.FileDescriptorSet{
		File: []*dpb.FileDescriptorProto{
			fileWithMessage("x", "M", []*dpb.FieldDescriptorProto{
				{Name: proto.String("a"), Number: proto.Int32(1), Type: dpb.FieldDescriptorProto_TYPE_STRING.Enum(), Label: dpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum()},
			}),
		},
	}
	newF := &dpb.FileDescriptorSet{
		File: []*dpb.FileDescriptorProto{
			fileWithMessage("x", "M", nil),
		},
	}
	r := pdiff.Diff(old, newF)
	if len(r.Changes) != 1 {
		t.Fatalf("changes: %+v", r.Changes)
	}
	if r.Changes[0].Kind != "remove" {
		t.Fatalf("%+v", r.Changes[0])
	}
}
