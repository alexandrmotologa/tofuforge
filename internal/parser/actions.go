package parser

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

// ClassifyAction maps a slice of action strings from terraform-json to a canonical Action.
func ClassifyAction(actions []string) model.Action {
	if len(actions) == 0 {
		return model.ActionNoOp
	}
	if len(actions) == 1 {
		switch actions[0] {
		case "no-op":
			return model.ActionNoOp
		case "create":
			return model.ActionCreate
		case "read":
			return model.ActionRead
		case "update":
			return model.ActionUpdate
		case "delete":
			return model.ActionDelete
		default:
			return model.ActionNoOp
		}
	}

	if len(actions) == 2 {
		if (actions[0] == "delete" && actions[1] == "create") ||
			(actions[0] == "create" && actions[1] == "delete") {
			return model.ActionReplace
		}
	}

	return model.ActionUpdate
}

// IsDestructive returns true if the action deletes or recreates a resource.
func IsDestructive(action model.Action) bool {
	return action == model.ActionDelete || action == model.ActionReplace
}

// ComputeAttributeDiffs compares before and after maps and associates replacement paths.
func ComputeAttributeDiffs(change model.RawChange) []model.AttributeDiff {
	replaceKeySet := make(map[string]bool)
	for _, pathItems := range change.ReplacePaths {
		var parts []string
		for _, item := range pathItems {
			parts = append(parts, fmt.Sprintf("%v", item))
		}
		if len(parts) > 0 {
			replaceKeySet[parts[0]] = true
			replaceKeySet[strings.Join(parts, ".")] = true
		}
	}

	diffMap := make(map[string]model.AttributeDiff)

	// Scan keys in before
	for k, beforeVal := range change.Before {
		afterVal, hasAfter := change.After[k]
		if !hasAfter {
			diffMap[k] = model.AttributeDiff{
				Path:              k,
				Before:            beforeVal,
				After:             nil,
				ForcesReplacement: replaceKeySet[k],
				Type:              "removed",
			}
		} else if !reflect.DeepEqual(beforeVal, afterVal) {
			diffMap[k] = model.AttributeDiff{
				Path:              k,
				Before:            beforeVal,
				After:             afterVal,
				ForcesReplacement: replaceKeySet[k],
				Type:              "modified",
			}
		}
	}

	// Scan keys in after that weren't in before
	for k, afterVal := range change.After {
		if _, hasBefore := change.Before[k]; !hasBefore {
			diffMap[k] = model.AttributeDiff{
				Path:              k,
				Before:            nil,
				After:             afterVal,
				ForcesReplacement: replaceKeySet[k],
				Type:              "added",
			}
		}
	}

	diffs := make([]model.AttributeDiff, 0, len(diffMap))
	for _, diff := range diffMap {
		diffs = append(diffs, diff)
	}

	sort.Slice(diffs, func(i, j int) bool {
		// Put replacement-forcing changes first
		if diffs[i].ForcesReplacement != diffs[j].ForcesReplacement {
			return diffs[i].ForcesReplacement
		}
		return diffs[i].Path < diffs[j].Path
	})

	return diffs
}
