package router

import (
	"errors"
	"net/http"
	"strings"

	"github.com/MrMiaoMIMI/goshared/db/dbspi"
	"github.com/MrMiaoMIMI/goshared/logger"
	"github.com/MrMiaoMIMI/goshared/util/serverresp"
	"github.com/gin-gonic/gin"

	authlib "github.com/MrMiaoMIMI/mockserver/internal/auth"
)

type AuthConfig struct {
	JWT authlib.Config
}

func jwtAuthMiddleware(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.JWT.JWTEnabled() {
			c.Next()
			return
		}

		token := bearerToken(c.Request)
		if token == "" {
			serverresp.UnauthorizedError(c, errors.New("jwt token is required"))
			c.Abort()
			return
		}

		claims, err := authlib.ValidateToken(config.JWT, token)
		if err != nil {
			message := "invalid jwt token"
			logger.Warn(c.Request.Context(), "JWT authentication failed", logger.Err(err), logger.String("reason", message))
			serverresp.UnauthorizedError(c, errors.New(message))
			c.Abort()
			return
		}

		c.Set(authlib.UserEmailContextKey, claims.Email)
		if _, ok := dbspi.OperatorFromContext(c.Request.Context()); !ok {
			c.Request = c.Request.WithContext(dbspi.WithOperator(c.Request.Context(), claims.Email))
		}
		c.Next()
	}
}

func bearerToken(r *http.Request) string {
	parts := strings.Fields(strings.TrimSpace(r.Header.Get("Authorization")))
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
