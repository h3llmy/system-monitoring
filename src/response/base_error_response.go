package response

type BaseErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}
