package command

import (
	"context"
	"errors"
	"testing"

	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/account"
	"event-driven-architecture/internal/domain/user"

	"github.com/google/uuid"
)

type stubHasher struct{}

func (stubHasher) Hash(plain string) (string, error) {
	return "hashed:" + plain, nil
}

type fakeUserRepository struct {
	users map[uuid.UUID]*user.User
}

func (f *fakeUserRepository) Add(_ context.Context, u *user.User) error {
	f.users[uuid.UUID(u.ID)] = u
	return nil
}

func (f *fakeUserRepository) ByID(_ context.Context, id user.ID) (*user.User, error) {
	u, ok := f.users[uuid.UUID(id)]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (f *fakeUserRepository) AllUsers(_ context.Context) ([]*user.User, error) {
	out := make([]*user.User, 0, len(f.users))
	for _, u := range f.users {
		out = append(out, u)
	}
	return out, nil
}

type fakeTransferRepository struct {
	saved         []*account.Transfer
	payerBalances []float64
	payeeBalances []float64
	saveErr       error
}

func (f *fakeTransferRepository) ByID(_ context.Context, id string) (*account.Transfer, error) {
	return nil, nil
}

func (f *fakeTransferRepository) Save(_ context.Context, transfer *account.Transfer, from, to *account.Account) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, transfer)
	f.payerBalances = append(f.payerBalances, from.Balance.Value)
	f.payeeBalances = append(f.payeeBalances, to.Balance.Value)
	return nil
}

func (f *fakeTransferRepository) AllTransfers(_ context.Context) ([]*account.Transfer, error) {
	return f.saved, nil
}

type stubAuthorizer struct {
	err error
}

func (s stubAuthorizer) Authorize(_ context.Context, _ string) error {
	return s.err
}

type recordingPublisher struct {
	events []domain.Event
	err    error
}

func (p *recordingPublisher) Publish(_ context.Context, event domain.Event) error {
	if p.err != nil {
		return p.err
	}
	p.events = append(p.events, event)
	return nil
}

func newTransferHandlerFixture(t *testing.T) (*fakeAccountRepository, *fakeUserRepository, *account.Account, *account.Account) {
	t.Helper()
	ownerID := user.ID(uuid.New())

	from, err := account.NewAccount(account.AccountID(uuid.New()), ownerID, domain.MonetaryAmount{Currency: "BRL", Value: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	to, err := account.NewAccount(account.AccountID(uuid.New()), user.ID(uuid.New()), domain.MonetaryAmount{Currency: "BRL", Value: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payer, err := user.NewUser(ownerID, stubHasher{}, "Rita Fontenele", "52998224725", "rita@example.com", "secret", user.COMMON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	accounts := newFakeAccountRepository()
	accounts.accounts[uuid.UUID(from.ID)] = from
	accounts.accounts[uuid.UUID(to.ID)] = to
	users := &fakeUserRepository{users: map[uuid.UUID]*user.User{uuid.UUID(payer.ID): payer}}

	return accounts, users, from, to
}

func assertPublishedEvents(t *testing.T, publisher *recordingPublisher, names ...string) {
	t.Helper()
	if len(publisher.events) != len(names) {
		t.Fatalf("expected %d published events, got %d", len(names), len(publisher.events))
	}
	for i, name := range names {
		if publisher.events[i].EventName() != name {
			t.Errorf("event %d: expected %s, got %s", i, name, publisher.events[i].EventName())
		}
	}
}

func TestTransferMoneyCommandHandler_Success(t *testing.T) {
	accounts, users, from, to := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	result, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Value:          40,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TransferID == "" {
		t.Error("expected a transfer id")
	}
	if result.Status != string(account.COMPLETED) {
		t.Errorf("expected COMPLETED, got %s", result.Status)
	}
	if result.Value != 40 || result.Currency != "BRL" {
		t.Errorf("unexpected result: %+v", result)
	}
	if from.Balance.Value != 60 || to.Balance.Value != 40 {
		t.Errorf("expected balances 60 and 40, got %v and %v", from.Balance.Value, to.Balance.Value)
	}
	if len(transfers.saved) != 1 {
		t.Fatalf("expected 1 saved transfer, got %d", len(transfers.saved))
	}
	if len(result.Events) != 2 {
		t.Errorf("expected 2 events in the result, got %d", len(result.Events))
	}
	assertPublishedEvents(t, publisher, "AccountTransferCreated", "AccountTransferSucceeded")
}

func TestTransferMoneyCommandHandler_InvalidPayerAccountID(t *testing.T) {
	accounts, users, _, _ := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: "not-a-uuid",
		PayeeAccountID: uuid.New().String(),
		Value:          10,
	})
	if !errors.Is(err, ErrInvalidPayerAccountID) {
		t.Fatalf("expected ErrInvalidPayerAccountID, got %v", err)
	}
	if len(transfers.saved) != 0 || len(publisher.events) != 0 {
		t.Error("expected no port to be called")
	}
}

func TestTransferMoneyCommandHandler_InvalidPayeeAccountID(t *testing.T) {
	accounts, users, _, _ := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.New().String(),
		PayeeAccountID: "not-a-uuid",
		Value:          10,
	})
	if !errors.Is(err, ErrInvalidPayeeAccountID) {
		t.Fatalf("expected ErrInvalidPayeeAccountID, got %v", err)
	}
	if len(transfers.saved) != 0 || len(publisher.events) != 0 {
		t.Error("expected no port to be called")
	}
}

func TestTransferMoneyCommandHandler_PayerAccountNotFound(t *testing.T) {
	accounts, users, _, to := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.New().String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Value:          10,
	})
	if !errors.Is(err, ErrPayerAccountNotFound) {
		t.Fatalf("expected ErrPayerAccountNotFound, got %v", err)
	}
	if len(transfers.saved) != 0 || len(publisher.events) != 0 {
		t.Error("expected nothing to be saved or published")
	}
}

func TestTransferMoneyCommandHandler_PayeeAccountNotFound(t *testing.T) {
	accounts, users, from, _ := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.New().String(),
		Value:          10,
	})
	if !errors.Is(err, ErrPayeeAccountNotFound) {
		t.Fatalf("expected ErrPayeeAccountNotFound, got %v", err)
	}
	if len(transfers.saved) != 0 || len(publisher.events) != 0 {
		t.Error("expected nothing to be saved or published")
	}
}

func TestTransferMoneyCommandHandler_NotAuthorized(t *testing.T) {
	accounts, users, from, to := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{err: errors.New("gateway denied")}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Value:          40,
	})
	if !errors.Is(err, ErrTransferNotAuthorized) {
		t.Fatalf("expected ErrTransferNotAuthorized, got %v", err)
	}
	if from.Balance.Value != 100 || to.Balance.Value != 0 {
		t.Errorf("expected the balances to be unchanged, got %v and %v", from.Balance.Value, to.Balance.Value)
	}
	if len(transfers.saved) != 1 {
		t.Fatalf("expected the attempt to be recorded, got %d saved transfers", len(transfers.saved))
	}
	if transfers.saved[0].Status != account.FAILED {
		t.Errorf("expected a FAILED transfer, got %s", transfers.saved[0].Status)
	}
	assertPublishedEvents(t, publisher, "AccountTransferCreated", "AccountTransferFailed")
}

func TestTransferMoneyCommandHandler_DomainRefusal(t *testing.T) {
	accounts, users, from, to := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Value:          500,
	})
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
	if from.Balance.Value != 100 || to.Balance.Value != 0 {
		t.Errorf("expected the balances to be unchanged, got %v and %v", from.Balance.Value, to.Balance.Value)
	}
	if len(transfers.saved) != 1 || transfers.saved[0].Status != account.FAILED {
		t.Error("expected exactly one saved FAILED transfer")
	}
	assertPublishedEvents(t, publisher, "AccountTransferCreated", "AccountTransferFailed")
}

func TestTransferMoneyCommandHandler_SaveFailure(t *testing.T) {
	accounts, users, from, to := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{saveErr: errors.New("db down")}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Value:          40,
	})
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, ErrInsufficientBalance) {
		t.Error("a persistence failure must not be reported as a domain refusal")
	}
	if len(publisher.events) != 0 {
		t.Error("expected no event to be published when persisting fails")
	}
}

func TestTransferMoneyCommandHandler_PublisherFailure(t *testing.T) {
	accounts, users, from, to := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{err: errors.New("broker down")}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	result, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Value:          40,
	})
	if !errors.Is(err, ErrTransferEventPublishFailed) {
		t.Fatalf("expected ErrTransferEventPublishFailed, got %v", err)
	}
	if result != nil {
		t.Error("expected no result when publishing fails")
	}
}

var _ TransferMoneyUseCase = (*TransferMoneyCommandHandler)(nil)

var _ account.TransferRepository = (*fakeTransferRepository)(nil)

var _ account.TransferAuthorizer = stubAuthorizer{}

var _ user.UserRepository = (*fakeUserRepository)(nil)

var _ TransferEventPublisher = (*recordingPublisher)(nil)
