package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"event-driven-architecture/internal/adapters/http/response"
	usercommand "event-driven-architecture/internal/application/user/command"
	"event-driven-architecture/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubCreateUser struct {
	result *usercommand.CreateUserResult
	err    error
}

func (s *stubCreateUser) Handle(_ context.Context, _ usercommand.CreateUserCommand) (*usercommand.CreateUserResult, error) {
	return s.result, s.err
}

func performUserRequest(t *testing.T, useCase usercommand.CreateUserUseCase, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := NewRouter(&stubCreateAccount{}, useCase)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	return rec
}

func TestCreateUser_Success(t *testing.T) {
	id := uuid.New()
	useCase := &stubCreateUser{
		result: &usercommand.CreateUserResult{
			ID:       id.String(),
			FullName: "Rita Fontenele",
			Document: "52998224725",
			Email:    "rita@example.com",
			Type:     "common",
		},
	}

	rec := performUserRequest(t, useCase, `{"full_name":"Rita Fontenele","document":"52998224725","email":"rita@example.com","password":"secret","type":"common"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp response.CreateUserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.ID != id.String() || resp.FullName != "Rita Fontenele" {
		t.Errorf("unexpected response: %+v", resp)
	}
	if resp.Document != "52998224725" || resp.Email != "rita@example.com" || resp.Type != "common" {
		t.Errorf("unexpected response: %+v", resp)
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Errorf("response must not contain the password, got %s", rec.Body.String())
	}
}

func TestCreateUser_InvalidBody(t *testing.T) {
	useCase := &stubCreateUser{}
	rec := performUserRequest(t, useCase, `not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateUser_MissingRequiredField(t *testing.T) {
	useCase := &stubCreateUser{}
	body := `{"full_name":"Rita Fontenele","document":"52998224725","email":"rita@example.com","type":"common"}`
	rec := performUserRequest(t, useCase, body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateUser_Conflict(t *testing.T) {
	useCase := &stubCreateUser{err: usercommand.ErrUserAlreadyExists}
	body := `{"full_name":"Rita Fontenele","document":"52998224725","email":"rita@example.com","password":"secret","type":"common"}`
	rec := performUserRequest(t, useCase, body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestCreateUser_InvalidDocument(t *testing.T) {
	useCase := &stubCreateUser{err: &domain.ConstraintValidationError{Field: "document"}}
	body := `{"full_name":"Rita Fontenele","document":"123","email":"rita@example.com","password":"secret","type":"common"}`
	rec := performUserRequest(t, useCase, body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateUser_InternalError(t *testing.T) {
	useCase := &stubCreateUser{err: errors.New("boom")}
	body := `{"full_name":"Rita Fontenele","document":"52998224725","email":"rita@example.com","password":"secret","type":"common"}`
	rec := performUserRequest(t, useCase, body)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
