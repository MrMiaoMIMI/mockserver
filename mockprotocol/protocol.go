package mockprotocol

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

type FieldType string

const (
	FieldTypeString FieldType = "string"
	FieldTypeNumber FieldType = "number"
	FieldTypeBool   FieldType = "bool"
	FieldTypeObject FieldType = "object"
	FieldTypeArray  FieldType = "array"
	FieldTypeJSON   FieldType = "json"
)

const (
	OperatorEQ         = "eq"
	OperatorNE         = "ne"
	OperatorIn         = "in"
	OperatorNotIn      = "not_in"
	OperatorContains   = "contains"
	OperatorNotContain = "not_contains"
	OperatorExists     = "exists"
	OperatorNotExists  = "not_exists"
	OperatorIsNull     = "is_null"
	OperatorIsNotNull  = "is_not_null"
	OperatorPrefix     = "prefix"
	OperatorSuffix     = "suffix"
	OperatorRegex      = "regex"
	OperatorGT         = "gt"
	OperatorGTE        = "gte"
	OperatorLT         = "lt"
	OperatorLTE        = "lte"
)

type ProtocolSpec struct {
	Name      string         `json:"name"`
	Fields    []FieldSpec    `json:"fields"`
	Selectors []SelectorSpec `json:"selectors,omitempty"`
	Response  ResponseSpec   `json:"response"`
	Actions   []ActionSpec   `json:"actions,omitempty"`
}

type FieldSpec struct {
	Path        string    `json:"path"`
	Type        FieldType `json:"type"`
	DynamicPath bool      `json:"dynamic_path,omitempty"`
	Operators   []string  `json:"operators,omitempty"`
}

type SelectorSpec struct {
	Path        string   `json:"path"`
	Operators   []string `json:"operators,omitempty"`
	DynamicPath bool     `json:"dynamic_path,omitempty"`
}

type ResponseSpec struct {
	Fields   []ResponseFieldSpec `json:"fields,omitempty"`
	Defaults map[string]any      `json:"defaults,omitempty"`
}

type ResponseFieldSpec struct {
	Path        string    `json:"path"`
	Type        FieldType `json:"type"`
	Required    bool      `json:"required,omitempty"`
	DynamicPath bool      `json:"dynamic_path,omitempty"`
	Default     any       `json:"default,omitempty"`
	Min         *float64  `json:"min,omitempty"`
	Max         *float64  `json:"max,omitempty"`
	Format      string    `json:"format,omitempty"`
}

type ActionSpec struct {
	Type      string   `json:"type"`
	Renderers []string `json:"renderers,omitempty"`
}

type Registry struct {
	specs map[string]ProtocolSpec
	order []string
}

func NewRegistry(specs ...ProtocolSpec) Registry {
	registry := Registry{
		specs: make(map[string]ProtocolSpec, len(specs)),
		order: make([]string, 0, len(specs)),
	}
	for _, spec := range specs {
		registry.Register(spec)
	}
	return registry
}

func (r *Registry) Register(spec ProtocolSpec) {
	name := strings.ToLower(strings.TrimSpace(spec.Name))
	if name == "" {
		return
	}
	spec.Name = name
	if _, exists := r.specs[name]; !exists {
		r.order = append(r.order, name)
	}
	r.specs[name] = spec
}

func (r Registry) Get(name string) (ProtocolSpec, bool) {
	spec, ok := r.specs[strings.ToLower(strings.TrimSpace(name))]
	return spec, ok
}

func (r Registry) List() []ProtocolSpec {
	items := make([]ProtocolSpec, 0, len(r.order))
	for _, name := range r.order {
		items = append(items, r.specs[name])
	}
	return items
}

func DefaultRegistry() Registry {
	return NewRegistry(HTTPSpec(), SPEXSpec(), CacheSpec())
}

func RegisteredSpecs() []ProtocolSpec {
	return EffectiveSpecs(DefaultRegistry().List())
}

func EffectiveSpecs(specs []ProtocolSpec) []ProtocolSpec {
	items := make([]ProtocolSpec, 0, len(specs))
	for _, spec := range specs {
		items = append(items, EffectiveSpec(spec))
	}
	return items
}

func EffectiveSpec(spec ProtocolSpec) ProtocolSpec {
	fields := make([]FieldSpec, 0, len(spec.Fields))
	for _, field := range spec.Fields {
		if len(field.Operators) == 0 {
			field.Operators = DefaultOperators(field.Type)
		} else {
			field.Operators = append([]string(nil), field.Operators...)
		}
		fields = append(fields, field)
	}
	spec.Fields = fields

	selectors := make([]SelectorSpec, 0, len(spec.Selectors))
	for _, selector := range spec.Selectors {
		if len(selector.Operators) == 0 {
			if field, ok := spec.FieldForPath(selector.Path); ok {
				selector.Operators = OperatorsForField(field)
			}
		} else {
			selector.Operators = append([]string(nil), selector.Operators...)
		}
		selectors = append(selectors, selector)
	}
	spec.Selectors = selectors
	spec.Response = effectiveResponseSpec(spec.Response)
	spec.Actions = effectiveActionSpecs(spec.Actions)
	return spec
}

func HTTPSpec() ProtocolSpec {
	statusMin, statusMax := 100.0, 599.0
	return ProtocolSpec{
		Name: "http",
		Fields: append(commonFields(), []FieldSpec{
			{Path: "request.method", Type: FieldTypeString, Operators: []string{OperatorEQ, OperatorNE, OperatorIn, OperatorNotIn, OperatorExists, OperatorNotExists, OperatorIsNull, OperatorIsNotNull}},
			{Path: "request.scheme", Type: FieldTypeString},
			{Path: "request.host", Type: FieldTypeString},
			{Path: "request.original_host", Type: FieldTypeString},
			{Path: "request.path", Type: FieldTypeString},
			{Path: "request.query", Type: FieldTypeObject, DynamicPath: true},
			{Path: "request.headers", Type: FieldTypeObject, DynamicPath: true},
			{Path: "request.body", Type: FieldTypeJSON, DynamicPath: true},
			{Path: "request.raw_body", Type: FieldTypeString},
			{Path: "request.client_ip", Type: FieldTypeString},
		}...),
		Selectors: []SelectorSpec{
			{Path: "request.host"},
			{Path: "request.path", Operators: []string{OperatorEQ, OperatorPrefix}},
		},
		Response: ResponseSpec{
			Defaults: map[string]any{
				"status":  200,
				"headers": map[string]any{"content-type": []any{"application/json"}},
				"body":    map[string]any{},
			},
			Fields: []ResponseFieldSpec{
				{Path: "status", Type: FieldTypeNumber, Required: true, Default: 200, Min: &statusMin, Max: &statusMax},
				{Path: "headers", Type: FieldTypeObject, Default: map[string]any{"content-type": []any{"application/json"}}},
				{Path: "body", Type: FieldTypeJSON, Default: map[string]any{}},
			},
		},
		Actions: defaultResponseActionSpecs(),
	}
}

func CacheSpec() ProtocolSpec {
	return ProtocolSpec{
		Name: "cache",
		Fields: append(commonFields(), []FieldSpec{
			{Path: "request.operation", Type: FieldTypeString, Operators: []string{OperatorEQ, OperatorNE, OperatorIn, OperatorNotIn, OperatorExists, OperatorNotExists, OperatorIsNull, OperatorIsNotNull}},
			{Path: "request.key", Type: FieldTypeString},
			{Path: "request.ttl_ms", Type: FieldTypeNumber},
			{Path: "request.value", Type: FieldTypeJSON, DynamicPath: true},
		}...),
		Selectors: []SelectorSpec{
			{Path: "request.operation"},
			{Path: "request.key", Operators: []string{OperatorEQ, OperatorPrefix}},
		},
		Response: ResponseSpec{
			Defaults: map[string]any{
				"hit":   true,
				"value": nil,
			},
			Fields: []ResponseFieldSpec{
				{Path: "hit", Type: FieldTypeBool, Required: true, Default: true},
				{Path: "value", Type: FieldTypeJSON},
			},
		},
		Actions: defaultResponseActionSpecs(),
	}
}

// SPEX is a self-developed RPC framework
func SPEXSpec() ProtocolSpec {
	return ProtocolSpec{
		Name: "spex",
		Fields: append(commonFields(), []FieldSpec{
			{Path: "request.cmd", Type: FieldTypeString},
			{Path: "request.req", Type: FieldTypeJSON, DynamicPath: true},
			{Path: "request.param", Type: FieldTypeString},
		}...),
		Selectors: []SelectorSpec{
			{Path: "request.cmd", Operators: []string{OperatorPrefix}},
		},
		Response: ResponseSpec{
			Defaults: map[string]any{
				"code": 0,
				"resp": map[string]any{},
			},
			Fields: []ResponseFieldSpec{
				{Path: "code", Type: FieldTypeNumber, Required: true, Default: 0},
				{Path: "resp", Type: FieldTypeJSON, Required: true, Default: map[string]any{}},
			},
		},
		Actions: defaultResponseActionSpecs(),
	}
}

func commonFields() []FieldSpec {
	return []FieldSpec{
		{Path: "protocol", Type: FieldTypeString},
		{Path: "namespace", Type: FieldTypeString},
		{Path: "meta.trace_id", Type: FieldTypeString},
		{Path: "meta.source", Type: FieldTypeString},
		{Path: "meta.extra", Type: FieldTypeObject, DynamicPath: true},
	}
}

func FieldForPath(protocol string, fieldPath string) (FieldSpec, bool) {
	spec, ok := DefaultRegistry().Get(protocol)
	if !ok {
		return FieldSpec{}, false
	}
	return spec.FieldForPath(fieldPath)
}

func (s ProtocolSpec) FieldForPath(fieldPath string) (FieldSpec, bool) {
	fieldPath = strings.TrimSpace(fieldPath)
	for _, field := range s.Fields {
		if field.Path == fieldPath {
			return field, true
		}
		if field.DynamicPath && isDynamicChildPath(field.Path, fieldPath) {
			child := field
			child.Type = FieldTypeJSON
			child.Operators = nil
			return child, true
		}
	}
	return FieldSpec{}, false
}

func OperatorAllowed(protocol string, fieldPath string, operator string) bool {
	field, ok := FieldForPath(protocol, fieldPath)
	if !ok {
		return false
	}
	return operatorAllowedForField(field, operator)
}

func SelectorForPath(protocol string, fieldPath string) (SelectorSpec, bool) {
	spec, ok := DefaultRegistry().Get(protocol)
	if !ok {
		return SelectorSpec{}, false
	}
	fieldPath = strings.TrimSpace(fieldPath)
	for _, selector := range spec.Selectors {
		if selector.Path == fieldPath {
			return selector, true
		}
		if selector.DynamicPath && isDynamicChildPath(selector.Path, fieldPath) {
			child := selector
			child.Operators = nil
			return child, true
		}
	}
	return SelectorSpec{}, false
}

func SelectorOperatorAllowed(protocol string, fieldPath string, operator string) bool {
	selector, ok := SelectorForPath(protocol, fieldPath)
	if !ok {
		return false
	}
	operator = strings.TrimSpace(operator)
	if len(selector.Operators) > 0 {
		for _, item := range selector.Operators {
			if item == operator {
				return true
			}
		}
		return false
	}
	return OperatorAllowed(protocol, fieldPath, operator)
}

func DefaultResponsePayload(protocol string) map[string]any {
	spec, ok := DefaultRegistry().Get(protocol)
	if !ok {
		return nil
	}
	return spec.DefaultResponsePayload()
}

func (s ProtocolSpec) DefaultResponsePayload() map[string]any {
	return cloneStringAnyMap(s.Response.Defaults)
}

func NormalizeResponsePayload(protocol string, payload map[string]any) (map[string]any, error) {
	spec, ok := DefaultRegistry().Get(protocol)
	if !ok {
		return nil, fmt.Errorf("unsupported protocol %q", protocol)
	}
	return spec.NormalizeResponsePayload(payload)
}

func ValidateResponsePayload(protocol string, payload map[string]any) error {
	_, err := NormalizeResponsePayload(protocol, payload)
	return err
}

func (s ProtocolSpec) NormalizeResponsePayload(payload map[string]any) (map[string]any, error) {
	normalized := s.DefaultResponsePayload()
	if normalized == nil {
		normalized = make(map[string]any)
	}
	for key, value := range payload {
		normalized[key] = value
	}
	for _, field := range s.Response.Fields {
		value, exists := valueAtResponsePath(normalized, field.Path)
		if !exists {
			if field.Required {
				return nil, fmt.Errorf("response.%s is required", field.Path)
			}
			continue
		}
		if err := validateResponseFieldValue(field, value); err != nil {
			return nil, fmt.Errorf("response.%s %s", field.Path, err.Error())
		}
	}
	return normalized, nil
}

func effectiveResponseSpec(spec ResponseSpec) ResponseSpec {
	if spec.Defaults != nil {
		spec.Defaults = cloneStringAnyMap(spec.Defaults)
	}
	fields := make([]ResponseFieldSpec, 0, len(spec.Fields))
	for _, field := range spec.Fields {
		fields = append(fields, field)
	}
	spec.Fields = fields
	return spec
}

func effectiveActionSpecs(specs []ActionSpec) []ActionSpec {
	if len(specs) == 0 {
		return defaultResponseActionSpecs()
	}
	items := make([]ActionSpec, 0, len(specs))
	for _, spec := range specs {
		spec.Renderers = append([]string(nil), spec.Renderers...)
		items = append(items, spec)
	}
	return items
}

func defaultResponseActionSpecs() []ActionSpec {
	return []ActionSpec{
		{
			Type:      "respond",
			Renderers: []string{"static", "template", "cel", "sequence", "webhook"},
		},
		{Type: "forward"},
	}
}

func validateResponseFieldValue(field ResponseFieldSpec, value any) error {
	switch field.Type {
	case FieldTypeString:
		text, ok := value.(string)
		if !ok {
			return fmt.Errorf("must be a string")
		}
		if field.Format == "json_string" && strings.TrimSpace(text) != "" && !json.Valid([]byte(text)) {
			return fmt.Errorf("must be a valid JSON string")
		}
	case FieldTypeNumber:
		number, ok := numberValue(value)
		if !ok || math.IsNaN(number) || math.IsInf(number, 0) {
			return fmt.Errorf("must be a number")
		}
		if field.Min != nil && number < *field.Min {
			return fmt.Errorf("must be >= %v", trimFloat(*field.Min))
		}
		if field.Max != nil && number > *field.Max {
			return fmt.Errorf("must be <= %v", trimFloat(*field.Max))
		}
	case FieldTypeBool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("must be a boolean")
		}
	case FieldTypeObject:
		switch value.(type) {
		case map[string]any, map[string]string, map[string][]string:
		default:
			return fmt.Errorf("must be an object")
		}
	case FieldTypeArray:
		switch value.(type) {
		case []any, []string:
		default:
			return fmt.Errorf("must be an array")
		}
	case FieldTypeJSON:
		return nil
	default:
		return fmt.Errorf("has unsupported field type %q", field.Type)
	}
	return nil
}

func valueAtResponsePath(payload map[string]any, path string) (any, bool) {
	if path == "" {
		return payload, true
	}
	parts := strings.Split(path, ".")
	var current any = payload
	for _, part := range parts {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func numberValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case float32:
		return float64(typed), true
	case float64:
		return typed, true
	default:
		return 0, false
	}
}

func trimFloat(value float64) any {
	if value == math.Trunc(value) {
		return int64(value)
	}
	return value
}

func cloneStringAnyMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	cloned := make(map[string]any, len(input))
	for key, value := range input {
		cloned[key] = cloneAny(value)
	}
	return cloned
}

func cloneAny(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneStringAnyMap(typed)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = cloneAny(item)
		}
		return out
	case map[string][]string:
		out := make(map[string][]string, len(typed))
		for key, values := range typed {
			out[key] = append([]string(nil), values...)
		}
		return out
	default:
		return typed
	}
}

func operatorAllowedForField(field FieldSpec, operator string) bool {
	operator = strings.TrimSpace(operator)
	for _, item := range OperatorsForField(field) {
		if item == operator {
			return true
		}
	}
	return false
}

func OperatorsForField(field FieldSpec) []string {
	if len(field.Operators) > 0 {
		return append([]string(nil), field.Operators...)
	}
	return DefaultOperators(field.Type)
}

func DefaultOperators(fieldType FieldType) []string {
	switch fieldType {
	case FieldTypeNumber:
		return []string{OperatorEQ, OperatorNE, OperatorIn, OperatorNotIn, OperatorGT, OperatorGTE, OperatorLT, OperatorLTE, OperatorExists, OperatorNotExists, OperatorIsNull, OperatorIsNotNull}
	case FieldTypeBool:
		return []string{OperatorEQ, OperatorNE, OperatorExists, OperatorNotExists, OperatorIsNull, OperatorIsNotNull}
	case FieldTypeObject:
		return []string{OperatorExists, OperatorNotExists, OperatorIsNull, OperatorIsNotNull}
	case FieldTypeArray:
		return []string{OperatorContains, OperatorNotContain, OperatorExists, OperatorNotExists, OperatorIsNull, OperatorIsNotNull}
	case FieldTypeJSON:
		return []string{OperatorEQ, OperatorNE, OperatorIn, OperatorNotIn, OperatorContains, OperatorNotContain, OperatorExists, OperatorNotExists, OperatorIsNull, OperatorIsNotNull, OperatorPrefix, OperatorSuffix, OperatorRegex, OperatorGT, OperatorGTE, OperatorLT, OperatorLTE}
	default:
		return []string{OperatorEQ, OperatorNE, OperatorIn, OperatorNotIn, OperatorContains, OperatorNotContain, OperatorPrefix, OperatorSuffix, OperatorRegex, OperatorExists, OperatorNotExists, OperatorIsNull, OperatorIsNotNull}
	}
}

func isDynamicChildPath(root string, candidate string) bool {
	return strings.HasPrefix(candidate, root+".") || strings.HasPrefix(candidate, root+"[")
}
