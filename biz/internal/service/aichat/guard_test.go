package aichat

import (
	"testing"
	"time"

	"personal-page-be/biz/infra/config"
)

func TestRequestGuardEnforcesIdentityAndIPConcurrency(t *testing.T) {
	guard := testRequestGuard()
	lease, denial := guard.acquire("visitor-a", "shared-ip")
	if denial != nil {
		t.Fatalf("first request denied: %+v", denial)
	}
	if _, denial = guard.acquire("visitor-b", "shared-ip"); denial == nil || denial.Reason != guardDeniedConcurrent {
		t.Fatalf("shared IP concurrent request should be denied, got %+v", denial)
	}
	lease.release()
	if lease, denial = guard.acquire("visitor-b", "shared-ip"); denial != nil {
		t.Fatalf("request should pass after release: %+v", denial)
	}
	lease.release()
}

func TestRequestGuardEnforcesAndResetsMinuteLimit(t *testing.T) {
	guard := testRequestGuard()
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	guard.now = func() time.Time { return now }

	for i := 0; i < 2; i++ {
		lease, denial := guard.acquire("visitor", "ip")
		if denial != nil {
			t.Fatalf("request %d denied: %+v", i, denial)
		}
		lease.release()
	}
	if _, denial := guard.acquire("visitor", "ip"); denial == nil || denial.Reason != guardDeniedMinute {
		t.Fatalf("third request should hit minute limit, got %+v", denial)
	}

	now = now.Add(time.Minute)
	lease, denial := guard.acquire("visitor", "ip")
	if denial != nil {
		t.Fatalf("minute limit should reset: %+v", denial)
	}
	lease.release()
}

func testRequestGuard() *requestGuard {
	return newRequestGuard(config.AIChatLimits{
		RequestsPerMinute: 2,
		MaxConcurrent:     1,
	})
}
