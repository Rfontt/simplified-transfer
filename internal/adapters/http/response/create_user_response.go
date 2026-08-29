package response

import "event-driven-architecture/internal/application/user/command"

type CreateUserResponse struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Document string `json:"document"`
	Email    string `json:"email"`
	Type     string `json:"type"`
}

func NewCreateUserResponse(result *command.CreateUserResult) CreateUserResponse {
	return CreateUserResponse{
		ID:       result.ID,
		FullName: result.FullName,
		Document: result.Document,
		Email:    result.Email,
		Type:     result.Type,
	}
}
