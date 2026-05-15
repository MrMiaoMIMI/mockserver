package httpadapter

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
)

type HTTPResponse struct {
	Status  int
	Headers map[string][]string
	Body    any
}

func NormalizeHTTPRequest(r *http.Request, namespace string) (bo.Event, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return bo.Event{}, err
	}
	r.Body.Close()

	var parsedBody any
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &parsedBody); err != nil {
			parsedBody = nil
		}
	}

	headers := make(map[string][]string, len(r.Header))
	for key, values := range r.Header {
		headers[strings.ToLower(key)] = append([]string(nil), values...)
	}

	query := make(map[string][]string, len(r.URL.Query()))
	for key, values := range r.URL.Query() {
		query[key] = append([]string(nil), values...)
	}

	originalHost := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if originalHost == "" {
		originalHost = r.Host
	}
	host := originalHost
	if strings.Contains(host, ":") {
		if parsedHost, _, splitErr := net.SplitHostPort(host); splitErr == nil {
			host = parsedHost
		}
	}

	request := bo.EventRequest{
		"method":        r.Method,
		"scheme":        schemeFor(r),
		"host":          strings.ToLower(host),
		"original_host": strings.ToLower(originalHost),
		"path":          r.URL.Path,
		"query":         query,
		"headers":       headers,
		"client_ip":     clientIPFor(r),
	}
	if len(bodyBytes) > 0 {
		request["raw_body"] = string(bodyBytes)
		if parsedBody != nil {
			request["body"] = parsedBody
		}
	}

	return bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Operation: "request",
		Namespace: namespace,
		Request:   request,
		Meta: bo.EventMeta{
			TraceID: r.Header.Get("X-Trace-ID"),
		},
	}, nil
}

func HTTPResponseFromProtocolResponse(response bo.ProtocolResponse) (HTTPResponse, error) {
	if response.Protocol != "" && !strings.EqualFold(response.Protocol, eo.ProtocolHTTP) {
		return HTTPResponse{}, fmt.Errorf("response protocol %q cannot be applied by HTTP adapter", response.Protocol)
	}
	payload := response.Payload
	status, err := statusFromPayload(payload["status"])
	if err != nil {
		return HTTPResponse{}, err
	}
	headers, err := headersFromPayload(payload["headers"])
	if err != nil {
		return HTTPResponse{}, err
	}
	return HTTPResponse{
		Status:  status,
		Headers: headers,
		Body:    payload["body"],
	}, nil
}

func statusFromPayload(raw any) (int, error) {
	switch value := raw.(type) {
	case nil:
		return 200, nil
	case int:
		if value < 100 || value > 599 {
			return 0, fmt.Errorf("http response status must be between 100 and 599")
		}
		return value, nil
	case float64:
		status := int(value)
		if float64(status) != value || status < 100 || status > 599 {
			return 0, fmt.Errorf("http response status must be an integer between 100 and 599")
		}
		return status, nil
	case json.Number:
		parsed, err := value.Int64()
		if err != nil || parsed < 100 || parsed > 599 {
			return 0, fmt.Errorf("http response status must be an integer between 100 and 599")
		}
		return int(parsed), nil
	default:
		return 0, fmt.Errorf("http response status must be a number")
	}
}

func headersFromPayload(raw any) (map[string][]string, error) {
	switch value := raw.(type) {
	case nil:
		return nil, nil
	case map[string][]string:
		return cloneHeaders(value), nil
	case map[string]string:
		headers := make(map[string][]string, len(value))
		for key, item := range value {
			headers[key] = []string{item}
		}
		return headers, nil
	case map[string]any:
		headers := make(map[string][]string, len(value))
		for key, item := range value {
			switch typed := item.(type) {
			case string:
				headers[key] = []string{typed}
			case []string:
				headers[key] = append([]string(nil), typed...)
			case []any:
				items := make([]string, 0, len(typed))
				for _, candidate := range typed {
					text, ok := candidate.(string)
					if !ok {
						return nil, fmt.Errorf("http response header %s contains non-string value", key)
					}
					items = append(items, text)
				}
				headers[key] = items
			default:
				return nil, fmt.Errorf("http response header %s has unsupported value type", key)
			}
		}
		return headers, nil
	default:
		return nil, fmt.Errorf("http response headers must be an object")
	}
}

func schemeFor(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if forwarded := r.Header.Get("X-Forwarded-Proto"); forwarded != "" {
		return strings.ToLower(forwarded)
	}
	return "http"
}

func clientIPFor(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func cloneHeaders(headers map[string][]string) map[string][]string {
	if len(headers) == 0 {
		return nil
	}
	cloned := make(map[string][]string, len(headers))
	for key, values := range headers {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}
