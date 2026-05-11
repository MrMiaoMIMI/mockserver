package httpadapter

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"

	"mockserver/internal/model/bo"
	"mockserver/internal/model/eo"
)

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
