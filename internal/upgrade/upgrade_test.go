package upgrade_test

import (
	"testing"

	"github.com/hashload/boss/internal/upgrade"
)

func TestBossUpgrade(t *testing.T) {
	if err := upgrade.BossUpgrade(true); err != nil {
		t.Errorf("failed to upgrade boss: %s", err.Error())
	}
}
