package aichat

import (
	"errors"
	"testing"
	"time"

	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/internal/domain"
)

type quotaTestRepository struct {
	allowed     bool
	err         error
	quotaDay    string
	identityKey string
	ipKey       string
	limit       int
}

func (r *quotaTestRepository) FindUser(string) (*domain.UserEntity, error) {
	return nil, nil
}

func (r *quotaTestRepository) FindUserByID(uint) (*domain.UserEntity, error) {
	return nil, nil
}

func (r *quotaTestRepository) SaveAIUsage(*domain.AIUsageEntity) error {
	return nil
}

func (r *quotaTestRepository) ReserveAIUsageDailyQuota(quotaDay string, identityKey string, ipKey string, limit int) (bool, error) {
	r.quotaDay = quotaDay
	r.identityKey = identityKey
	r.ipKey = ipKey
	r.limit = limit
	return r.allowed, r.err
}

func (r *quotaTestRepository) ListAIUsage(*uint, *time.Time, *time.Time, string, string) (*[]domain.AIUsageEntity, error) {
	rows := []domain.AIUsageEntity{}
	return &rows, nil
}

func TestReservePersistentDailyBudgetUsesBothPseudonymousKeys(t *testing.T) {
	repository := &quotaTestRepository{allowed: true}
	service := &AIChatService{Repo: repository}
	identity := requestIdentity{IdentityKey: "identity-hash", IPKey: "ip-hash"}
	denial, err := service.reservePersistentDailyBudget(identity, config.AIChatLimits{DailyRequestBudget: 50})
	if err != nil || denial != nil {
		t.Fatalf("reservation denial=%+v err=%v", denial, err)
	}
	if repository.quotaDay == "" || repository.identityKey != identity.IdentityKey || repository.ipKey != identity.IPKey || repository.limit != 50 {
		t.Fatalf("unexpected reservation: %+v", repository)
	}
}

func TestReservePersistentDailyBudgetFailsClosed(t *testing.T) {
	identity := requestIdentity{IdentityKey: "identity-hash", IPKey: "ip-hash"}
	limits := config.AIChatLimits{DailyRequestBudget: 50}

	service := &AIChatService{Repo: &quotaTestRepository{allowed: false}}
	denial, err := service.reservePersistentDailyBudget(identity, limits)
	if err != nil || denial == nil || denial.Reason != guardDeniedDaily || denial.RetryAfter <= 0 {
		t.Fatalf("exhausted quota denial=%+v err=%v", denial, err)
	}

	databaseErr := errors.New("database unavailable")
	service = &AIChatService{Repo: &quotaTestRepository{err: databaseErr}}
	if _, err = service.reservePersistentDailyBudget(identity, limits); !errors.Is(err, databaseErr) {
		t.Fatalf("database error = %v", err)
	}
}

func TestUsageCaptureKeepsRealUsernameSeparateFromDifyUser(t *testing.T) {
	service := &AIChatService{Config: &config.Config{AIChatConfig: config.AIChatConfig{Model: "model"}}}
	identity := requestIdentity{
		UserID:      42,
		Username:    "alice",
		IdentityKey: "identity-hash",
		IPKey:       "ip-hash",
		DifyUser:    "user_pseudonymous",
	}
	usage := service.newUsageCapture(sendMessageReq{ConversationID: "conversation"}, identity)
	if usage.Username != "alice" || usage.Username == identity.DifyUser {
		t.Fatalf("usage username = %q, Dify user = %q", usage.Username, identity.DifyUser)
	}
	if usage.UserID != 42 || usage.IdentityKey != identity.IdentityKey || usage.IPKey != identity.IPKey {
		t.Fatalf("unexpected usage identity: %+v", usage)
	}
}
