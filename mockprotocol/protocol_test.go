package mockprotocol

import "testing"

func TestDefaultRegistryContainsHTTPAndCacheAndSPEXSpecs(t *testing.T) {
	registry := DefaultRegistry()
	if _, ok := registry.Get("http"); !ok {
		t.Fatalf("expected http spec")
	}
	if _, ok := registry.Get("cache"); !ok {
		t.Fatalf("expected cache spec")
	}
	if _, ok := registry.Get("spex"); !ok {
		t.Fatalf("expected spex spec")
	}
	if len(registry.List()) != 3 {
		t.Fatalf("unexpected registry size: %d", len(registry.List()))
	}
}

func TestFieldForPathSupportsDynamicPaths(t *testing.T) {
	field, ok := FieldForPath("http", "request.body.status")
	if !ok {
		t.Fatalf("expected dynamic body field")
	}
	if field.Path != "request.body" || field.Type != FieldTypeJSON {
		t.Fatalf("unexpected field: %+v", field)
	}
	headerField, ok := FieldForPath("http", "request.headers.x-env[0]")
	if !ok {
		t.Fatalf("expected dynamic header field")
	}
	if headerField.Path != "request.headers" || headerField.Type != FieldTypeJSON {
		t.Fatalf("unexpected header field: %+v", headerField)
	}
	if _, ok := FieldForPath("cache", "request.unknown"); ok {
		t.Fatalf("unexpected unknown cache field")
	}
	spexReqField, ok := FieldForPath("spex", "request.req.user.id")
	if !ok {
		t.Fatalf("expected dynamic spex req field")
	}
	if spexReqField.Path != "request.req" || spexReqField.Type != FieldTypeJSON {
		t.Fatalf("unexpected spex req field: %+v", spexReqField)
	}
}

func TestOperatorAllowedUsesFieldOverridesAndTypeDefaults(t *testing.T) {
	if OperatorAllowed("http", "request.method", OperatorRegex) {
		t.Fatalf("method should not allow regex because it has field-level operators")
	}
	if !OperatorAllowed("http", "request.path", OperatorRegex) {
		t.Fatalf("path should allow regex through string defaults")
	}
	if !OperatorAllowed("http", "request.body.score", OperatorGTE) {
		t.Fatalf("json dynamic fields should allow numeric comparison")
	}
	if !OperatorAllowed("http", "request.query.q1[*]", OperatorContains) {
		t.Fatalf("object dynamic child fields should use json operator defaults")
	}
	if !OperatorAllowed("cache", "request.ttl_ms", OperatorLTE) {
		t.Fatalf("number fields should allow numeric comparison")
	}
}

func TestRegisteredSpecsExposeEffectiveOperators(t *testing.T) {
	specs := RegisteredSpecs()
	var httpSpec ProtocolSpec
	for _, spec := range specs {
		if spec.Name == "http" {
			httpSpec = spec
			break
		}
	}
	if httpSpec.Name == "" {
		t.Fatalf("expected http spec")
	}
	for _, field := range httpSpec.Fields {
		if len(field.Operators) == 0 {
			t.Fatalf("field %s should expose effective operators", field.Path)
		}
	}
	for _, selector := range httpSpec.Selectors {
		if len(selector.Operators) == 0 {
			t.Fatalf("selector %s should expose effective operators", selector.Path)
		}
	}
	if !SelectorOperatorAllowed("http", "request.path", OperatorPrefix) {
		t.Fatalf("http path selector should allow prefix")
	}
	if SelectorOperatorAllowed("http", "request.path", OperatorRegex) {
		t.Fatalf("http path selector should not allow regex")
	}
	if !SelectorOperatorAllowed("spex", "request.cmd", OperatorPrefix) {
		t.Fatalf("spex cmd selector should allow prefix")
	}
}

func TestSPEXResponseSpecUsesDirectJSONResp(t *testing.T) {
	spec, ok := DefaultRegistry().Get("spex")
	if !ok {
		t.Fatalf("expected spex spec")
	}
	if got := spec.Response.Defaults["resp"]; got == nil {
		t.Fatalf("expected resp default")
	} else if _, ok := got.(map[string]any); !ok {
		t.Fatalf("expected resp default to be JSON object, got %#v", got)
	}
	var respField ResponseFieldSpec
	for _, field := range spec.Response.Fields {
		if field.Path == "resp" {
			respField = field
			break
		}
	}
	if respField.Path == "" {
		t.Fatalf("expected resp field")
	}
	if respField.Type != FieldTypeJSON || respField.Format != "" {
		t.Fatalf("unexpected resp field: %+v", respField)
	}
	if _, err := spec.NormalizeResponsePayload(map[string]any{
		"code": 0,
		"resp": map[string]any{"order_status": "mocked"},
	}); err != nil {
		t.Fatalf("NormalizeResponsePayload() error = %v", err)
	}
}
