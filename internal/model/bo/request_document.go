package bo

import (
	"fmt"
	"strings"
)

func RequestString(request EventRequest, key string) string {
	value, ok := request[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}

func RequestStringMap(request EventRequest, key string) map[string][]string {
	value, ok := request[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case map[string][]string:
		return typed
	case map[string]any:
		result := make(map[string][]string, len(typed))
		for itemKey, itemValue := range typed {
			result[strings.ToLower(itemKey)] = anyToStringSlice(itemValue)
		}
		return result
	default:
		return nil
	}
}

func anyToStringSlice(value any) []string {
	switch typed := value.(type) {
	case nil:
		return nil
	case []string:
		return append([]string(nil), typed...)
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			result = append(result, RequestValueString(item))
		}
		return result
	case string:
		return []string{typed}
	default:
		return []string{RequestValueString(typed)}
	}
}

func RequestValueString(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprint(value)
}
