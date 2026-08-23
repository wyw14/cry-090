package moderation

import (
	"testing"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

func TestResolvedReportCannotPenalizeReputationTwice(t *testing.T) {
	now := time.Date(2026, 8, 23, 14, 0, 0, 0, time.UTC)
	report, err := NewReport("report-1", "reporter", "target", "harassment", "repeated unwanted invitations", now)
	if err != nil {
		t.Fatal(err)
	}
	reputation := &Reputation{UserID: "target", Score: 50}
	if err := report.Resolve("moderator", ReportResolved, "evidence upheld", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	reputation.ApplyUpheldReport()
	if err := report.Resolve("moderator-2", ReportResolved, "processed again", now.Add(2*time.Minute)); common.CodeOf(err) != common.CodeAlreadyProcessed {
		reputation.ApplyUpheldReport()
		t.Fatalf("closed report was processed twice: err=%v report=%+v reputation=%+v", err, report, reputation)
	}
	if report.Version != 2 || report.ResolvedBy != "moderator" || reputation.ReportsUpheld != 1 || reputation.Score != 40 {
		t.Fatalf("resolution or reputation is inconsistent: report=%+v reputation=%+v", report, reputation)
	}
}
