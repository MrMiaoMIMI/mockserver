package engine

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"default": func(value, fallback any) any {
			if value == nil || value == "" {
				return fallback
			}
			return value
		},
		"field": func(root map[string]any, path string) any {
			return templateField(root, path)
		},
		"query": func(root map[string]any, key string) any {
			return templateField(root, "request.query."+key+"[0]")
		},
		"queryAll": func(root map[string]any, key string) any {
			return templateField(root, "request.query."+key+"[*]")
		},
		"header": func(root map[string]any, key string) any {
			return templateField(root, "request.headers."+strings.ToLower(key)+"[0]")
		},
		"headerAll": func(root map[string]any, key string) any {
			return templateField(root, "request.headers."+strings.ToLower(key)+"[*]")
		},
		"body": func(root map[string]any, path string) any {
			return templateField(root, "request.body."+path)
		},
		"first": func(value any) any {
			switch typed := value.(type) {
			case []any:
				if len(typed) == 0 {
					return nil
				}
				return typed[0]
			case []string:
				if len(typed) == 0 {
					return nil
				}
				return typed[0]
			default:
				return value
			}
		},
		"toJSON": func(value any) string {
			raw, err := json.Marshal(value)
			if err != nil {
				return fmt.Sprintf("%v", value)
			}
			return string(raw)
		},
	}
}

func templateField(root map[string]any, path string) any {
	values, err := resolvePath(root, path)
	if err != nil || len(values) == 0 {
		return nil
	}
	if len(values) == 1 {
		return values[0]
	}
	return values
}
