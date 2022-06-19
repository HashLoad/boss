package upgrade

import "testing"

func TestBossUpgrade(t *testing.T) {

	if err := BossUpgrade(true); err != nil {
		t.Errorf("failed to upgrade boss: %s", err.Error())
	}

}
