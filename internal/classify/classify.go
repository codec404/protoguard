package classify

import (
	"strings"

	"github.com/codec404/protoguard/internal/model"
)

// AssignImpact applies rule-based classification.
func AssignImpact(c model.Change) model.Impact {
	p := strings.ToLower(c.Path)
	sum := strings.ToLower(c.Summary)

	switch c.Kind {
	case model.ChangeRemove:
		switch {
		case strings.Contains(p, "openapi.paths."):
			return model.ImpactBreaking
		case strings.HasPrefix(p, "openapi.components.schemas."):
			return model.ImpactBreaking
		case strings.Contains(p, ".responses.") || strings.Contains(p, "responses."):
			return model.ImpactRisky
		case strings.HasPrefix(p, "grpc.field.") || strings.HasPrefix(p, "grpc.message.") || strings.HasPrefix(p, "grpc.service."):
			return model.ImpactBreaking
		case strings.HasPrefix(p, "protobuf.file."):
			return model.ImpactBreaking
		default:
			return model.ImpactRisky
		}

	case model.ChangeAdd:
		if strings.Contains(p, "parameter") && strings.Contains(sum, "required") {
			return model.ImpactBreaking
		}
		return model.ImpactNonBreaking

	case model.ChangeTypeChange:
		return model.ImpactBreaking

	case model.ChangeModify:
		switch {
		case strings.Contains(p, "parameter") && strings.Contains(sum, "required"):
			return model.ImpactBreaking
		case strings.HasPrefix(p, "grpc.service.") && strings.Contains(p, ".method."):
			return model.ImpactBreaking
		case strings.Contains(p, ".operation"):
			return model.ImpactRisky
		default:
			return model.ImpactRisky
		}

	case model.ChangeRenameSuspected:
		return model.ImpactRisky
	}

	return model.ImpactRisky
}

// ApplyToReport sets Impact on all changes.
func ApplyToReport(r *model.DiffReport) {
	for i := range r.Changes {
		r.Changes[i].Impact = AssignImpact(r.Changes[i])
	}
}
