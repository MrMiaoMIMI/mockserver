package eo

const (
	ProtocolHTTP  = "http"
	ProtocolSPEX  = "spex"
	ProtocolCache = "cache"
)

const (
	ActionTypeRespond = "respond"
	ActionTypeForward = "forward"
)

const (
	ActionRendererStatic   = "static"
	ActionRendererTemplate = "template"
	ActionRendererCEL      = "cel"
	ActionRendererSequence = "sequence"
	ActionRendererWebhook  = "webhook"
)

const (
	SequenceStrategyLoop = "loop"
	SequenceStrategyLast = "last"
)

const (
	DecisionKindResponse = "response"
	DecisionKindForward  = "forward"
)

const (
	FallbackReasonRulesetMiss = "ruleset_miss"
	FallbackReasonRuleMiss    = "rule_miss"
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
