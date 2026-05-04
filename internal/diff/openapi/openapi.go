package openapi

import (
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/codec404/protoguard/internal/model"
)

var httpMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD", "TRACE"}

func opPtr(pi *openapi3.PathItem, m string) **openapi3.Operation {
	switch strings.ToUpper(m) {
	case "GET":
		return &pi.Get
	case "POST":
		return &pi.Post
	case "PUT":
		return &pi.Put
	case "PATCH":
		return &pi.Patch
	case "DELETE":
		return &pi.Delete
	case "OPTIONS":
		return &pi.Options
	case "HEAD":
		return &pi.Head
	case "TRACE":
		return &pi.Trace
	default:
		return nil
	}
}

func diffOps(piOld, piNew *openapi3.PathItem, path string, changes *[]model.Change) {
	for _, m := range httpMethods {
		pp := opPtr(piOld, m)
		pn := opPtr(piNew, m)
		if pp == nil || pn == nil {
			continue
		}
		o := *pp
		n := *pn
		if o == nil && n == nil {
			continue
		}
		if o != nil && n == nil {
			*changes = append(*changes, model.Change{
				Path:    fmt.Sprintf("openapi.paths.%s.%s", path, m),
				Kind:    model.ChangeRemove,
				Old:     snippetOp(o),
				Summary: "removed HTTP operation",
			})
			continue
		}
		if o == nil && n != nil {
			*changes = append(*changes, model.Change{
				Path:    fmt.Sprintf("openapi.paths.%s.%s", path, m),
				Kind:    model.ChangeAdd,
				New:     snippetOp(n),
				Summary: "added HTTP operation",
			})
			continue
		}
		diffOperation(o, n, fmt.Sprintf("openapi.paths.%s.%s", path, m), changes)
	}
}

func snippetOp(o *openapi3.Operation) any {
	if o == nil {
		return nil
	}
	params := make([]string, 0, len(o.Parameters))
	for _, p := range o.Parameters {
		if p.Value == nil {
			continue
		}
		required := p.Value.Required
		params = append(params, fmt.Sprintf("%s:%s required=%v", p.Value.In, p.Value.Name, required))
	}
	sort.Strings(params)
	bodyRequired := false
	if o.RequestBody != nil && o.RequestBody.Value != nil {
		bodyRequired = o.RequestBody.Value.Required
	}
	respCodes := make([]string, 0, len(o.Responses.Map()))
	for code := range o.Responses.Map() {
		respCodes = append(respCodes, code)
	}
	sort.Strings(respCodes)
	return map[string]any{
		"operationId":    o.OperationID,
		"parameters":     params,
		"requestBodyReq": bodyRequired,
		"responseCodes":  respCodes,
	}
}

func diffOperation(oldO, newO *openapi3.Operation, base string, changes *[]model.Change) {
	sOld := snippetOp(oldO)
	sNew := snippetOp(newO)
	if fmt.Sprintf("%v", sOld) != fmt.Sprintf("%v", sNew) {
		*changes = append(*changes, model.Change{
			Path:    base + ".operation",
			Kind:    model.ChangeModify,
			Old:     sOld,
			New:     sNew,
			Summary: "operation signature changed",
		})
	}
	// Parameter-level required transitions
	oldReq := paramRequiredSet(oldO)
	newReq := paramRequiredSet(newO)
	for k, req := range newReq {
		if req && !oldReq[k] {
			*changes = append(*changes, model.Change{
				Path:    fmt.Sprintf("%s.parameter[%s]", base, k),
				Kind:    model.ChangeModify,
				New:     true,
				Summary: "parameter became required",
			})
		}
	}
	// Response codes removed
	oldCodes := codesSet(oldO)
	newCodes := codesSet(newO)
	for c := range oldCodes {
		if !newCodes[c] {
			*changes = append(*changes, model.Change{
				Path:    fmt.Sprintf("%s.responses.%s", base, c),
				Kind:    model.ChangeRemove,
				Old:     c,
				Summary: "response status removed",
			})
		}
	}
}

func paramRequiredSet(o *openapi3.Operation) map[string]bool {
	out := map[string]bool{}
	for _, p := range o.Parameters {
		if p.Value == nil {
			continue
		}
		key := strings.ToLower(p.Value.In) + ":" + p.Value.Name
		out[key] = p.Value.Required
	}
	return out
}

func codesSet(o *openapi3.Operation) map[string]bool {
	out := map[string]bool{}
	for code := range o.Responses.Map() {
		out[code] = true
	}
	return out
}

func schemaKeys(comp *openapi3.Components) []string {
	if comp == nil || comp.Schemas == nil {
		return nil
	}
	keys := make([]string, 0, len(comp.Schemas))
	for k := range comp.Schemas {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func diffSchemas(oldC, newC *openapi3.Components, changes *[]model.Change) {
	oldKeys := schemaKeys(oldC)
	newKeys := schemaKeys(newC)
	oldSet := map[string]bool{}
	for _, k := range oldKeys {
		oldSet[k] = true
	}
	newSet := map[string]bool{}
	for _, k := range newKeys {
		newSet[k] = true
	}
	for _, k := range oldKeys {
		if !newSet[k] {
			*changes = append(*changes, model.Change{
				Path:    fmt.Sprintf("openapi.components.schemas.%s", k),
				Kind:    model.ChangeRemove,
				Old:     k,
				Summary: "schema component removed",
			})
		}
	}
	for _, k := range newKeys {
		if !oldSet[k] {
			*changes = append(*changes, model.Change{
				Path:    fmt.Sprintf("openapi.components.schemas.%s", k),
				Kind:    model.ChangeAdd,
				New:     k,
				Summary: "schema component added",
			})
			continue
		}
		var oldS, newS *openapi3.SchemaRef
		if oldC != nil && oldC.Schemas != nil {
			oldS = oldC.Schemas[k]
		}
		if newC != nil && newC.Schemas != nil {
			newS = newC.Schemas[k]
		}
		if schemaRefChanged(oldS, newS) {
			*changes = append(*changes, model.Change{
				Path:    fmt.Sprintf("openapi.components.schemas.%s", k),
				Kind:    model.ChangeModify,
				Old:     schemaSnippet(oldS),
				New:     schemaSnippet(newS),
				Summary: "schema definition changed",
			})
		}
	}
}

func schemaRefChanged(oldS, newS *openapi3.SchemaRef) bool {
	return fmt.Sprintf("%v", schemaSnippet(oldS)) != fmt.Sprintf("%v", schemaSnippet(newS))
}

func schemaSnippet(sr *openapi3.SchemaRef) any {
	if sr == nil {
		return nil
	}
	if sr.Ref != "" {
		return map[string]any{"$ref": sr.Ref}
	}
	s := sr.Value
	if s == nil {
		return nil
	}
	required := append([]string(nil), s.Required...)
	sort.Strings(required)
	props := make([]string, 0, len(s.Properties))
	for k := range s.Properties {
		props = append(props, k)
	}
	sort.Strings(props)
	enumAny := append([]any(nil), s.Enum...)
	sort.Slice(enumAny, func(i, j int) bool {
		return fmt.Sprint(enumAny[i]) < fmt.Sprint(enumAny[j])
	})
	return map[string]any{
		"type":        s.Type,
		"required":    required,
		"properties":  props,
		"enum":        enumAny,
		"format":      s.Format,
		"nullable":    s.Nullable,
		"description": s.Description,
	}
}

// Diff compares two OpenAPI documents.
func Diff(oldDoc, newDoc *openapi3.T) *model.DiffReport {
	var ch []model.Change

	oldPaths := pathsKeys(oldDoc)
	newPaths := pathsKeys(newDoc)
	pathSet := map[string]bool{}
	for _, p := range oldPaths {
		pathSet[p] = true
	}
	for _, p := range newPaths {
		pathSet[p] = true
	}
	allPaths := make([]string, 0, len(pathSet))
	for p := range pathSet {
		allPaths = append(allPaths, p)
	}
	sort.Strings(allPaths)

	for _, path := range allPaths {
		var piOld, piNew *openapi3.PathItem
		if oldDoc.Paths != nil {
			piOld = oldDoc.Paths.Find(path)
		}
		if newDoc.Paths != nil {
			piNew = newDoc.Paths.Find(path)
		}
		switch {
		case piOld != nil && piNew == nil:
			ch = append(ch, model.Change{
				Path:    fmt.Sprintf("openapi.paths.%s", path),
				Kind:    model.ChangeRemove,
				Old:     path,
				Summary: "removed path",
			})
		case piOld == nil && piNew != nil:
			ch = append(ch, model.Change{
				Path:    fmt.Sprintf("openapi.paths.%s", path),
				Kind:    model.ChangeAdd,
				New:     path,
				Summary: "added path",
			})
		default:
			diffOps(piOld, piNew, path, &ch)
		}
	}

	diffSchemas(oldDoc.Components, newDoc.Components, &ch)

	return &model.DiffReport{
		SchemaVersion: model.SchemaVersion,
		SpecKind:      model.SpecOpenAPI,
		Changes:       ch,
	}
}

func pathsKeys(doc *openapi3.T) []string {
	if doc == nil || doc.Paths == nil {
		return nil
	}
	out := doc.Paths.InMatchingOrder()
	sort.Strings(out)
	return out
}
