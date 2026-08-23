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
	Version      int64
}

func NewReport(id, reporterID, targetID, reason, details string, now time.Time) (*Report, error) {
	if id == "" || reporterID == "" || targetID == "" {
		return nil, common.FieldError("report", "identities are required")
	}
	if reporterID == targetID {
		return nil, common.NewError(common.CodeInvalid, "cannot report yourself")
	}
	if strings.TrimSpace(reason) == "" {
		return nil, common.FieldError("reason", "is required")
	}
	return &Report{ID: id, ReporterID: reporterID, TargetUserID: targetID, Reason: strings.TrimSpace(reason),
		Details: strings.TrimSpace(details), Status: ReportOpen, Version: 1, CreatedAt: now.UTC()}, nil
}

func (r *Report) Resolve(actorID string, status ReportStatus, resolution string, now time.Time) error {
	if actorID == "" || strings.TrimSpace(resolution) == "" {
		return common.FieldError("resolution", "actor and resolution are required")
	}
	if r.Status != ReportOpen && r.Status != ReportInvestigate {
		return common.NewError(common.CodeAlreadyProcessed, "report is already closed")
	}
	if status != ReportResolved && status != ReportDismissed {
		return common.FieldError("status", "must resolve or dismiss the report")
	}
	at := now.UTC()
	r.Status = status
	r.ResolvedBy = actorID
	r.Resolution = strings.TrimSpace(resolution)
	r.ResolvedAt = &at
	r.Version++
	return nil
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

func (r *Reputation) ApplyCompleted()    { r.Completed++; r.recompute() }
func (r *Reputation) ApplyNoShow()       { r.NoShows++; r.recompute() }
func (r *Reputation) ApplyUpheldReport() { r.ReportsUpheld++; r.recompute() }

func (r *Reputation) recompute() {
	r.Score = 50 + min(40, r.Completed*2) - min(35, r.NoShows*5) - min(20, r.ReportsUpheld*10)
	if r.Score < 0 {
		r.Score = 0
	}
	if r.Score > 100 {
		r.Score = 100
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
