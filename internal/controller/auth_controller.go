package controller

import (
	"errors"
	"strings"

	"github.com/MrMiaoMIMI/goshared/util/serverresp"
	"github.com/gin-gonic/gin"

	authlib "github.com/MrMiaoMIMI/mockserver/internal/auth"
	"github.com/MrMiaoMIMI/mockserver/internal/model/request"
	"github.com/MrMiaoMIMI/mockserver/internal/model/response"
)

type AuthController struct {
	config authlib.Config
}

func NewAuthController(config authlib.Config) *AuthController {
	return &AuthController{config: config}
}

func (c *AuthController) DebugLogin(ctx *gin.Context) {
	if !c.config.JWTEnabled() || !c.config.DebugLoginEnabled {
		serverresp.ForbiddenError(ctx, errors.New("debug login is not enabled"))
		return
	}

	var req request.DebugLoginRequest
	if !bindJSON(ctx, &req) {
		return
	}
	email := strings.TrimSpace(req.Email)
	token, err := authlib.GenerateToken(c.config, email)
	if err != nil {
		serverresp.InternalServerError(ctx, err)
		return
	}
	serverresp.Success(ctx, response.LoginResponse{
		Token: token,
		User: response.UserInfo{
			Email: email,
		},
	})
}

func (c *AuthController) CurrentUser(ctx *gin.Context) {
	emailValue, ok := ctx.Get(authlib.UserEmailContextKey)
	if !ok {
		serverresp.UnauthorizedError(ctx, errors.New("user is not authenticated"))
		return
	}
	email, ok := emailValue.(string)
	if !ok || strings.TrimSpace(email) == "" {
		serverresp.UnauthorizedError(ctx, errors.New("user email is missing"))
		return
	}
	serverresp.Success(ctx, response.UserInfo{
		Email: strings.TrimSpace(email),
	})
}
