package proto

import (
	"fmt"
	"sort"

	dpb "google.golang.org/protobuf/types/descriptorpb"

	"github.com/codec404/protoguard/internal/model"
)

// Diff compares two FileDescriptorSets (typically single-file sets).
func Diff(oldFDS, newFDS *dpb.FileDescriptorSet) *model.DiffReport {
	oldIdx := indexFiles(oldFDS)
	newIdx := indexFiles(newFDS)
	names := unionKeys(oldIdx, newIdx)
	sort.Strings(names)

	var ch []model.Change
	for _, fn := range names {
		of, okOld := oldIdx[fn]
		nf, okNew := newIdx[fn]
		switch {
		case okOld && !okNew:
			ch = append(ch, model.Change{
				Path:    fmt.Sprintf("protobuf.file.%s", fn),
				Kind:    model.ChangeRemove,
				Old:     fn,
				Summary: "proto file removed from descriptor set",
			})
		case !okOld && okNew:
			ch = append(ch, model.Change{
				Path:    fmt.Sprintf("protobuf.file.%s", fn),
				Kind:    model.ChangeAdd,
				New:     fn,
				Summary: "proto file added to descriptor set",
			})
		default:
			diffFile(of, nf, &ch)
		}
	}

	return &model.DiffReport{
		SchemaVersion: model.SchemaVersion,
		SpecKind:      model.SpecProtobuf,
		Changes:       ch,
	}
}

func indexFiles(fds *dpb.FileDescriptorSet) map[string]*dpb.FileDescriptorProto {
	m := map[string]*dpb.FileDescriptorProto{}
	if fds == nil {
		return m
	}
	for _, f := range fds.File {
		if f.GetName() != "" {
			m[f.GetName()] = f
		}
	}
	return m
}

func unionKeys(a, b map[string]*dpb.FileDescriptorProto) []string {
	set := map[string]bool{}
	for k := range a {
		set[k] = true
	}
	for k := range b {
		set[k] = true
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}

func diffFile(oldF, newF *dpb.FileDescriptorProto, ch *[]model.Change) {
	pkg := oldF.GetPackage()
	if pkg == "" {
		pkg = newF.GetPackage()
	}

	oldMsgs := indexMessages(oldF.GetMessageType(), pkg, "")
	newMsgs := indexMessages(newF.GetMessageType(), pkg, "")
	diffMessages(oldMsgs, newMsgs, ch)

	oldSvc := indexServices(oldF.GetService(), pkg)
	newSvc := indexServices(newF.GetService(), pkg)
	diffServices(oldSvc, newSvc, ch)
}

func indexMessages(messages []*dpb.DescriptorProto, pkg, parent string) map[string]*dpb.DescriptorProto {
	out := map[string]*dpb.DescriptorProto{}
	var walk func(prefix string, msgs []*dpb.DescriptorProto)
	walk = func(prefix string, msgs []*dpb.DescriptorProto) {
		for _, m := range msgs {
			name := m.GetName()
			fqn := name
			if prefix != "" {
				fqn = prefix + "." + name
			}
			full := fqn
			if pkg != "" {
				full = pkg + "." + fqn
			}
			out[full] = m
			if len(m.GetNestedType()) > 0 {
				walk(fqn, m.GetNestedType())
			}
		}
	}
	walk("", messages)
	return out
}

func diffMessages(oldM, newM map[string]*dpb.DescriptorProto, ch *[]model.Change) {
	keys := unionStringKeys(oldM, newM)
	sort.Strings(keys)
	for _, k := range keys {
		om, okO := oldM[k]
		nm, okN := newM[k]
		switch {
		case okO && !okN:
			*ch = append(*ch, model.Change{
				Path:    fmt.Sprintf("grpc.message.%s", k),
				Kind:    model.ChangeRemove,
				Old:     k,
				Summary: "message removed",
			})
		case !okO && okN:
			*ch = append(*ch, model.Change{
				Path:    fmt.Sprintf("grpc.message.%s", k),
				Kind:    model.ChangeAdd,
				New:     k,
				Summary: "message added",
			})
		default:
			diffMessageFields(k, om, nm, ch)
		}
	}
}

func unionStringKeys(oldM, newM map[string]*dpb.DescriptorProto) []string {
	set := map[string]bool{}
	for k := range oldM {
		set[k] = true
	}
	for k := range newM {
		set[k] = true
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}

func diffMessageFields(fqn string, om, nm *dpb.DescriptorProto, ch *[]model.Change) {
	oldByNum := fieldsByNumber(om.GetField())
	newByNum := fieldsByNumber(nm.GetField())
	nums := map[int32]bool{}
	for n := range oldByNum {
		nums[n] = true
	}
	for n := range newByNum {
		nums[n] = true
	}
	list := make([]int32, 0, len(nums))
	for n := range nums {
		list = append(list, n)
	}
	sort.Slice(list, func(i, j int) bool { return list[i] < list[j] })

	for _, num := range list {
		of, oOk := oldByNum[num]
		nf, nOk := newByNum[num]
		base := fmt.Sprintf("grpc.field.%s.%d", fqn, num)
		switch {
		case oOk && !nOk:
			*ch = append(*ch, model.Change{
				Path:    base,
				Kind:    model.ChangeRemove,
				Old:     fieldSnippet(of),
				Summary: "field removed (reserve number for compatibility)",
			})
		case !oOk && nOk:
			*ch = append(*ch, model.Change{
				Path:    base,
				Kind:    model.ChangeAdd,
				New:     fieldSnippet(nf),
				Summary: "field added",
			})
		default:
			if !fieldsEqual(of, nf) {
				*ch = append(*ch, model.Change{
					Path:    base,
					Kind:    model.ChangeTypeChange,
					Old:     fieldSnippet(of),
					New:     fieldSnippet(nf),
					Summary: "field type or label changed",
				})
			}
		}
	}
}

func fieldsByNumber(fields []*dpb.FieldDescriptorProto) map[int32]*dpb.FieldDescriptorProto {
	m := map[int32]*dpb.FieldDescriptorProto{}
	for _, f := range fields {
		m[f.GetNumber()] = f
	}
	return m
}

func fieldSnippet(f *dpb.FieldDescriptorProto) any {
	if f == nil {
		return nil
	}
	return map[string]any{
		"name":        f.GetName(),
		"number":      f.GetNumber(),
		"label":       f.GetLabel().String(),
		"type":        f.GetType().String(),
		"type_name":   f.GetTypeName(),
		"json_name":   f.GetJsonName(),
		"oneof_index": f.GetOneofIndex(),
	}
}

func fieldsEqual(a, b *dpb.FieldDescriptorProto) bool {
	if a.GetName() != b.GetName() {
		return false
	}
	if a.GetNumber() != b.GetNumber() {
		return false
	}
	if a.GetLabel() != b.GetLabel() {
		return false
	}
	if a.GetType() != b.GetType() {
		return false
	}
	if a.GetTypeName() != b.GetTypeName() {
		return false
	}
	if a.GetJsonName() != b.GetJsonName() {
		return false
	}
	return true
}

func indexServices(svcs []*dpb.ServiceDescriptorProto, pkg string) map[string]*dpb.ServiceDescriptorProto {
	out := map[string]*dpb.ServiceDescriptorProto{}
	for _, s := range svcs {
		fqn := s.GetName()
		if pkg != "" {
			fqn = pkg + "." + fqn
		}
		out[fqn] = s
	}
	return out
}

func diffServices(oldS, newS map[string]*dpb.ServiceDescriptorProto, ch *[]model.Change) {
	keys := map[string]bool{}
	for k := range oldS {
		keys[k] = true
	}
	for k := range newS {
		keys[k] = true
	}
	list := make([]string, 0, len(keys))
	for k := range keys {
		list = append(list, k)
	}
	sort.Strings(list)

	for _, svcName := range list {
		os, okO := oldS[svcName]
		ns, okN := newS[svcName]
		switch {
		case okO && !okN:
			*ch = append(*ch, model.Change{
				Path:    fmt.Sprintf("grpc.service.%s", svcName),
				Kind:    model.ChangeRemove,
				Old:     svcName,
				Summary: "service removed",
			})
		case !okO && okN:
			*ch = append(*ch, model.Change{
				Path:    fmt.Sprintf("grpc.service.%s", svcName),
				Kind:    model.ChangeAdd,
				New:     svcName,
				Summary: "service added",
			})
		default:
			diffMethods(svcName, os.GetMethod(), ns.GetMethod(), ch)
		}
	}
}

func diffMethods(svcName string, om, nm []*dpb.MethodDescriptorProto, ch *[]model.Change) {
	oldByName := map[string]*dpb.MethodDescriptorProto{}
	for _, m := range om {
		oldByName[m.GetName()] = m
	}
	newByName := map[string]*dpb.MethodDescriptorProto{}
	for _, m := range nm {
		newByName[m.GetName()] = m
	}
	names := map[string]bool{}
	for k := range oldByName {
		names[k] = true
	}
	for k := range newByName {
		names[k] = true
	}
	list := make([]string, 0, len(names))
	for k := range names {
		list = append(list, k)
	}
	sort.Strings(list)

	for _, mn := range list {
		oMet, oOk := oldByName[mn]
		nMet, nOk := newByName[mn]
		path := fmt.Sprintf("grpc.service.%s.method.%s", svcName, mn)
		switch {
		case oOk && !nOk:
			*ch = append(*ch, model.Change{
				Path:    path,
				Kind:    model.ChangeRemove,
				Old:     methodSnippet(oMet),
				Summary: "RPC removed",
			})
		case !oOk && nOk:
			*ch = append(*ch, model.Change{
				Path:    path,
				Kind:    model.ChangeAdd,
				New:     methodSnippet(nMet),
				Summary: "RPC added",
			})
		default:
			if !methodsEqual(oMet, nMet) {
				*ch = append(*ch, model.Change{
					Path:    path,
					Kind:    model.ChangeModify,
					Old:     methodSnippet(oMet),
					New:     methodSnippet(nMet),
					Summary: "RPC signature or streaming mode changed",
				})
			}
		}
	}
}

func methodSnippet(m *dpb.MethodDescriptorProto) any {
	if m == nil {
		return nil
	}
	return map[string]any{
		"name":             m.GetName(),
		"input_type":       m.GetInputType(),
		"output_type":      m.GetOutputType(),
		"client_streaming": m.GetClientStreaming(),
		"server_streaming": m.GetServerStreaming(),
	}
}

func methodsEqual(a, b *dpb.MethodDescriptorProto) bool {
	return a.GetInputType() == b.GetInputType() &&
		a.GetOutputType() == b.GetOutputType() &&
		a.GetClientStreaming() == b.GetClientStreaming() &&
		a.GetServerStreaming() == b.GetServerStreaming()
}
