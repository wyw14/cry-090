package moderation

import (
	"strings"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

type ReportStatus string

const (
	ReportOpen        ReportStatus = "open"
	ReportInvestigate ReportStatus = "investigating"
	ReportResolved    ReportStatus = "resolved"
	ReportDismissed   ReportStatus = "dismissed"
)

// terminalStatuses are the ReportStatus values that mark a report as closed.
// Once a report reaches one of these it must never be resolved again.
var terminalStatuses = map[ReportStatus]bool{
	ReportResolved:  true,
	ReportDismissed: true,
}

type Report struct {
	ID           string
	ReporterID   string
	TargetUserID string
	Reason       string
	Details      string
	Status       ReportStatus
	Resolution   string
	ResolvedBy   string
	CreatedAt    time.Time
	ResolvedAt   *time.Time
	// ReputationApplied records whether the target's reputation has already been
	// adjusted for this report. It guarantees the penalty is applied at most once
	// even if resolution is retried or replayed.
	ReputationApplied bool
	Version           int64
}

func NewReport(id, reporterID, targetID, reason, details string, now time.Time) (*Report, error) {
	return &Report{ID: id, ReporterID: reporterID, TargetUserID: targetID, Reason: reason,
		Details: details, Status: ReportOpen, Version: 1, CreatedAt: now.UTC()}, nil
}

// Resolve closes the report with the given terminal status. A report may be
// resolved exactly once: once it is resolved or dismissed a second reviewer
// cannot re-adjudicate it, so the original handler and reasoning are preserved.
func (r *Report) Resolve(actorID string, status ReportStatus, resolution string, now time.Time) error {
	if r.IsTerminal() {
		return common.NewError(common.CodeAlreadyProcessed, "report has already been resolved")
	}
	if actorID == "" {
		return common.FieldError("actor_id", "reviewer identity is required")
	}
	if resolution == "" {
		return common.FieldError("resolution", "resolution reason is required")
	}
	if !terminalStatuses[status] {
		return common.FieldError("status", "resolution must be resolved or dismissed")
	}
	at := now.UTC()
	r.Status = status
	r.ResolvedBy = actorID
	r.Resolution = resolution
	r.ResolvedAt = &at
	r.Version++
	return nil
}

// IsTerminal reports whether the dispute has already been closed and can no
// longer be modified.
func (r *Report) IsTerminal() bool {
	return terminalStatuses[r.Status]
}

// ApplyReputationPenalty deducts the target's reputation for an upheld report.
// It is a no-op once the penalty has already been applied for this report, so a
// single report can never reduce a user's reputation more than once.
func (r *Report) ApplyReputationPenalty(rep *Reputation) bool {
	if r.ReputationApplied {
		return false
	}
	r.ReputationApplied = true
	rep.ApplyUpheldReport()
	return true
}

type Block struct {
	BlockerID string
	BlockedID string
	Reason    string
	CreatedAt time.Time
}

func NewBlock(blockerID, blockedID, reason string, now time.Time) (*Block, error) {
	if blockerID == "" || blockedID == "" || blockerID == blockedID {
		return nil, common.FieldError("block", "valid distinct users are required")
	}
	return &Block{BlockerID: blockerID, BlockedID: blockedID, Reason: strings.TrimSpace(reason), CreatedAt: now.UTC()}, nil
}

type Reputation struct {
	UserID        string
	Completed     int
	NoShows       int
	ReportsUpheld int
	Score         int
	UpdatedAt     time.Time
}

func (r *Reputation) ApplyCompleted()    { r.Completed++; r.Score += 2 }
func (r *Reputation) ApplyNoShow()       { r.NoShows++; r.Score -= 5 }
func (r *Reputation) ApplyUpheldReport() { r.ReportsUpheld++; r.Score -= 10 }

func (r *Reputation) recompute() {
	r.Score = r.Completed - r.NoShows - r.ReportsUpheld
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
