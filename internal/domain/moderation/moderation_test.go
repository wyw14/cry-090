package moderation

import (
	"testing"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

func TestReportResolveOnceAndSingleReputationPenalty(t *testing.T) {
	now := time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)
	report, err := NewReport("r1", "reporter", "target", "harassment", "details", now)
	if err != nil {
		t.Fatal(err)
	}
	rep := &Reputation{UserID: "target"}

	// First reviewer adjudicates and the target's reputation is docked once.
	if err := report.Resolve("reviewer-1", ReportResolved, "evidence upheld", now); err != nil {
		t.Fatalf("first resolve failed: %v", err)
	}
	if !report.ApplyReputationPenalty(rep) {
		t.Fatal("first penalty should be applied")
	}
	originalHandler := report.ResolvedBy
	originalResolution := report.Resolution
	originalVersion := report.Version
	if rep.Score != -10 || rep.ReportsUpheld != 1 {
		t.Fatalf("expected single -10 penalty, got score=%d upheld=%d", rep.Score, rep.ReportsUpheld)
	}

	// A second reviewer attempts to re-adjudicate the already-closed dispute.
	err = report.Resolve("reviewer-2", ReportDismissed, "no further action", now.Add(time.Hour))
	if code := common.CodeOf(err); code != common.CodeAlreadyProcessed {
		t.Fatalf("expected already_processed error on second resolve, got %v (code=%s)", err, code)
	}
	// The original handler, reasoning and version must be preserved (not overwritten).
	if report.ResolvedBy != originalHandler || report.Resolution != originalResolution || report.Version != originalVersion {
		t.Fatalf("closed report was mutated: handler=%q reason=%q version=%d",
			report.ResolvedBy, report.Resolution, report.Version)
	}

	// Re-running the penalty step must be a no-op: reputation is compensated at most once.
	if report.ApplyReputationPenalty(rep) {
		t.Fatal("penalty must not be applied twice for the same report")
	}
	if rep.Score != -10 || rep.ReportsUpheld != 1 {
		t.Fatalf("reputation was double-counted: score=%d upheld=%d", rep.Score, rep.ReportsUpheld)
	}
}

func TestReportResolveValidatesInputs(t *testing.T) {
	now := time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)
	report, _ := NewReport("r1", "reporter", "target", "harassment", "details", now)

	cases := []struct {
		name       string
		actor      string
		status     ReportStatus
		resolution string
		wantCode   common.ErrorCode
	}{
		{"empty actor", "", ReportResolved, "reason", common.CodeInvalid},
		{"empty resolution", "rev", ReportResolved, "", common.CodeInvalid},
		{"non-terminal target status", "rev", ReportOpen, "reason", common.CodeInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := report.Resolve(tc.actor, tc.status, tc.resolution, now)
			if code := common.CodeOf(err); code != tc.wantCode {
				t.Fatalf("expected code %s, got %v (%s)", tc.wantCode, err, code)
			}
		})
	}
}
