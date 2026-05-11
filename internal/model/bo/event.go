package bo

type Event struct {
	Protocol  string       `json:"protocol"`
	Operation string       `json:"operation,omitempty"`
	Namespace string       `json:"namespace"`
	Request   EventRequest `json:"request"`
	Meta      EventMeta    `json:"meta,omitempty"`
}

type EventRequest map[string]any

type EventMeta struct {
	TraceID string         `json:"trace_id,omitempty"`
	Source  string         `json:"source,omitempty"`
	Extra   map[string]any `json:"extra,omitempty"`
}

func NewEventRequest(values map[string]any) EventRequest {
	if values == nil {
		return EventRequest{}
	}
	return EventRequest(values)
}
