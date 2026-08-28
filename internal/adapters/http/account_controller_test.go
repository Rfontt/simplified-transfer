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
	"event-driven-architecture/internal/application/account/command"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubCreateAccount struct {
	result *command.CreateAccountResult
	err    error
}

func (s *stubCreateAccount) Handle(_ context.Context, _ command.CreateAccountCommand) (*command.CreateAccountResult, error) {
	return s.result, s.err
}

func performRequest(t *testing.T, useCase command.CreateAccountUseCase, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := NewRouter(useCase)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	return rec
}

func TestCreateAccount_Success(t *testing.T) {
	id := uuid.New()
	ownerID := uuid.New()
	useCase := &stubCreateAccount{
		result: &command.CreateAccountResult{
			ID:       id.String(),
			OwnerID:  ownerID.String(),
			Currency: "BRL",
			Balance:  100,
		},
	}

	rec := performRequest(t, useCase, `{"owner_id":"`+ownerID.String()+`","currency":"BRL","balance":100}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp response.CreateAccountResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.ID != id.String() || resp.OwnerID != ownerID.String() {
		t.Errorf("unexpected response: %+v", resp)
	}
	if resp.Currency != "BRL" || resp.Balance != 100 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestCreateAccount_InvalidBody(t *testing.T) {
	useCase := &stubCreateAccount{}
	rec := performRequest(t, useCase, `not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateAccount_Conflict(t *testing.T) {
	useCase := &stubCreateAccount{err: command.ErrAccountAlreadyExists}
	body := `{"owner_id":"` + uuid.New().String() + `","currency":"BRL","balance":0}`
	rec := performRequest(t, useCase, body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestCreateAccount_OwnerNotFound(t *testing.T) {
	useCase := &stubCreateAccount{err: command.ErrOwnerNotFound}
	body := `{"owner_id":"` + uuid.New().String() + `","currency":"BRL","balance":0}`
	rec := performRequest(t, useCase, body)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCreateAccount_InternalError(t *testing.T) {
	useCase := &stubCreateAccount{err: errors.New("boom")}
	body := `{"owner_id":"` + uuid.New().String() + `","currency":"BRL","balance":0}`
	rec := performRequest(t, useCase, body)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
