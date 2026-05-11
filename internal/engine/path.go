package engine

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
)

type pathSegment struct {
	Key      string
	HasIndex bool
	Index    int
	Wildcard bool
}

func buildEventDocument(event bo.Event) map[string]any {
	return map[string]any{
		"protocol":  event.Protocol,
		"namespace": event.Namespace,
		"request":   map[string]any(event.Request),
		"meta": map[string]any{
			"trace_id": event.Meta.TraceID,
			"source":   event.Meta.Source,
			"extra":    event.Meta.Extra,
		},
	}
}

func resolveField(event bo.Event, field string) ([]any, bool) {
	values, err := resolvePath(buildEventDocument(event), field)
	if err != nil {
		return nil, false
	}
	return values, len(values) > 0
}

func resolvePath(root any, path string) ([]any, error) {
	if root == nil {
		return nil, nil
	}
	if strings.TrimSpace(path) == "" {
		return []any{root}, nil
	}

	segments, err := parsePath(path)
	if err != nil {
		return nil, err
	}

	current := []any{root}
	for _, segment := range segments {
		next := make([]any, 0)
		for _, item := range current {
			resolved, resolveErr := applyPathSegment(item, segment)
			if resolveErr != nil {
				return nil, resolveErr
			}
			next = append(next, resolved...)
		}
		current = next
	}
	return current, nil
}

func parsePath(path string) ([]pathSegment, error) {
	rawSegments := strings.Split(path, ".")
	segments := make([]pathSegment, 0, len(rawSegments))
	for _, raw := range rawSegments {
		if raw == "" {
			return nil, fmt.Errorf("invalid empty path segment in %q", path)
		}
		segment := pathSegment{Key: raw}
		leftBracket := strings.Index(raw, "[")
		if leftBracket < 0 {
			segments = append(segments, segment)
			continue
		}

		if !strings.HasSuffix(raw, "]") {
			return nil, fmt.Errorf("invalid bracket syntax in %q", raw)
		}
		segment.Key = raw[:leftBracket]
		indexToken := raw[leftBracket+1 : len(raw)-1]
		switch indexToken {
		case "*":
			segment.Wildcard = true
		default:
			index, err := strconv.Atoi(indexToken)
			if err != nil {
				return nil, fmt.Errorf("invalid index %q in %q", indexToken, raw)
			}
			segment.HasIndex = true
			segment.Index = index
		}
		segments = append(segments, segment)
	}
	return segments, nil
}

func applyPathSegment(item any, segment pathSegment) ([]any, error) {
	value, found := lookupSegmentValue(item, segment.Key)
	if !found {
		return nil, nil
	}

	if !segment.Wildcard && !segment.HasIndex {
		return []any{value}, nil
	}

	array, ok := toSlice(value)
	if !ok {
		return nil, nil
	}
	if segment.Wildcard {
		return array, nil
	}
	if segment.Index < 0 || segment.Index >= len(array) {
		return nil, nil
	}
	return []any{array[segment.Index]}, nil
}

func lookupSegmentValue(item any, key string) (any, bool) {
	if key == "" {
		return item, true
	}

	switch typed := item.(type) {
	case map[string]any:
		value, ok := typed[key]
		return value, ok
	case map[string][]string:
		value, ok := typed[key]
		if ok {
			return stringSliceToAny(value), true
		}
		return nil, false
	case map[string]string:
		value, ok := typed[key]
		return value, ok
	default:
		return nil, false
	}
}

func toSlice(value any) ([]any, bool) {
	switch typed := value.(type) {
	case []any:
		return typed, true
	case []string:
		return stringSliceToAny(typed), true
	default:
		return nil, false
	}
}
