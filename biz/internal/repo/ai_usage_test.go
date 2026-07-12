package repo

import (
	"reflect"
	"sync"
	"testing"

	"gorm.io/gorm/schema"
	"personal-page-be/biz/internal/domain"
)

func TestDailyQuotaKeysUseStableLockOrder(t *testing.T) {
	got := dailyQuotaKeys("ffff", "0000")
	want := []string{"identity:ffff", "ip:0000"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("quota keys = %#v, want %#v", got, want)
	}
}

func TestDailyQuotaModelHasCompositeUniqueIndex(t *testing.T) {
	parsed, err := schema.Parse(&domain.AIUsageDailyQuotaEntity{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, index := range parsed.ParseIndexes() {
		if index.Name != "idx_ai_usage_daily_quota" || index.Class != "UNIQUE" || len(index.Fields) != 2 {
			continue
		}
		if index.Fields[0].DBName == "quota_day" && index.Fields[1].DBName == "quota_key" {
			found = true
		}
	}
	if !found {
		t.Fatal("daily quota table must have a unique (quota_day, quota_key) index")
	}
}
