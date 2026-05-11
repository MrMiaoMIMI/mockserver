package eo

const (
	ProtocolHTTP = "http"
)

const (
	ActionTypeStaticResponse   = "static_response"
	ActionTypeTemplateResponse = "template_response"
	ActionTypeCELResponse      = "cel_response"
	ActionTypeSequenceResponse = "sequence_response"
	ActionTypeWebhookResponse  = "webhook_response"
)

const (
	SequenceStrategyLoop = "loop"
	SequenceStrategyLast = "last"
)

const (
	NamespaceFallbackTypeResponse = "response"
	NamespaceFallbackTypeForward  = "forward"
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
