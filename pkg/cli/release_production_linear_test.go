package cli

import (
	"context"
	"github.com/jamesonstone/mint/pkg/promotion"
	"strings"
	"testing"
)

func TestExternalHotfixCannotSkipAuthenticatedActiveControlRun(t *testing.T) {
	o := productionOperation{snapshot: promotion.JournalSnapshot{State: promotion.State{Baseline: &promotion.Baseline{ID: "verified"}}}}
	_, _, err := o.prepareHotfix(context.Background(), 0, "verified", nil, "https://linear.app/acme/issue/ENG-42")
	if err == nil || !strings.Contains(err.Error(), "trusted Actions dispatch") {
		t.Fatal("missing run authorization accepted", err)
	}
}
