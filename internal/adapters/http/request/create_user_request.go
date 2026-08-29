package request

type CreateUserRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Document string `json:"document" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Type     string `json:"type" binding:"required"`
}
