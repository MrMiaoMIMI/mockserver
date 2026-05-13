package mockprotocol

import "strings"

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
	return spec
}

func HTTPSpec() ProtocolSpec {
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
