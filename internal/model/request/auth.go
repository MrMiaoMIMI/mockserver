package request

type DebugLoginRequest struct {
	Email string `json:"email" binding:"required,email"`
}
