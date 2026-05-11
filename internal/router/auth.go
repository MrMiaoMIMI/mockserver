package router

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/MrMiaoMIMI/goshared/util/serverresp"
	"github.com/gin-gonic/gin"
)

type adminAuthRole string

const (
	adminAuthRoleRead    adminAuthRole = "read"
	adminAuthRoleWrite   adminAuthRole = "write"
	adminAuthRolePublish adminAuthRole = "publish"
)

type AdminAuthConfig struct {
	AdminToken   string
	ReadToken    string
	WriteToken   string
	PublishToken string
}

func (c AdminAuthConfig) enabled() bool {
	return c.AdminToken != "" || c.ReadToken != "" || c.WriteToken != "" || c.PublishToken != ""
}

func adminAuthMiddleware(config AdminAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.enabled() {
			c.Next()
			return
		}

		r := c.Request
		requiredRole, ok := requiredAdminAuthRole(r)
		if !ok {
			c.Next()
			return
		}

		token := bearerOrHeaderToken(r)
		if token == "" {
			serverresp.UnauthorizedError(c, errors.New("admin token is required"))
			c.Abort()
			return
		}
		if !config.hasToken(token) {
			serverresp.UnauthorizedError(c, errors.New("invalid admin token"))
			c.Abort()
			return
		}
		if !config.authorized(token, requiredRole) {
			serverresp.ForbiddenError(c, errors.New("admin token does not have required permission"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func requiredAdminAuthRole(r *http.Request) (adminAuthRole, bool) {
	if !strings.HasPrefix(r.URL.Path, "/mockserver/api/v1/admin/") && r.URL.Path != "/mockserver/api/v1/admin/rulesets" {
		return "", false
	}
	if r.Method == http.MethodGet {
		return adminAuthRoleRead, true
	}

	normalizedPath := strings.TrimRight(r.URL.Path, "/")
	switch {
	case normalizedPath == "/mockserver/api/v1/admin/published/simulate":
		return adminAuthRoleRead, true
	case strings.HasSuffix(normalizedPath, "/validate"):
		return adminAuthRoleRead, true
	case strings.HasSuffix(normalizedPath, "/simulate"):
		return adminAuthRoleRead, true
	case strings.HasSuffix(normalizedPath, "/rollback/preview"):
		return adminAuthRoleRead, true
	case strings.HasSuffix(normalizedPath, "/publish"):
		return adminAuthRolePublish, true
	case strings.HasSuffix(normalizedPath, "/rollback"):
		return adminAuthRolePublish, true
	default:
		return adminAuthRoleWrite, true
	}
}

func bearerOrHeaderToken(r *http.Request) string {
	if token := strings.TrimSpace(r.Header.Get("X-Mockserver-Admin-Token")); token != "" {
		return token
	}
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	const bearerPrefix = "Bearer "
	if strings.HasPrefix(authorization, bearerPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(authorization, bearerPrefix))
	}
	return ""
}

func (c AdminAuthConfig) hasToken(token string) bool {
	return tokenMatches(token, c.AdminToken) ||
		tokenMatches(token, c.ReadToken) ||
		tokenMatches(token, c.WriteToken) ||
		tokenMatches(token, c.PublishToken)
}

func (c AdminAuthConfig) authorized(token string, role adminAuthRole) bool {
	if tokenMatches(token, c.AdminToken) {
		return true
	}
	switch role {
	case adminAuthRoleRead:
		return tokenMatches(token, c.ReadToken) ||
			tokenMatches(token, c.WriteToken) ||
			tokenMatches(token, c.PublishToken)
	case adminAuthRoleWrite:
		return tokenMatches(token, c.WriteToken)
	case adminAuthRolePublish:
		return tokenMatches(token, c.PublishToken)
	default:
		return false
	}
}

func tokenMatches(got string, want string) bool {
	if got == "" || want == "" || len(got) != len(want) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}
