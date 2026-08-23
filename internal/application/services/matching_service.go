package services

import (
	"context"
	"time"

	"github.com/wyw14/cry-090/internal/application/ports"
	"github.com/wyw14/cry-090/internal/domain/common"
	"github.com/wyw14/cry-090/internal/domain/identity"
	"github.com/wyw14/cry-090/internal/domain/matching"
)

type MatchingService struct {
	deps    ports.Dependencies
	matcher matching.Matcher
}

func NewMatchingService(deps ports.Dependencies) *MatchingService {
	return &MatchingService{deps: deps}
}

func (s *MatchingService) Find(ctx context.Context, needID string, confirmed map[string]bool) ([]matching.Result, error) {
	subject, err := s.deps.Needs.Get(ctx, needID)
	if err != nil {
		return nil, err
	}
	if !subject.Active || !time.Now().Before(subject.ExpiresAt) {
		return nil, common.NewError(common.CodeExpired, "partner need is expired")
	}
	needs, err := s.deps.Needs.ListForEvent(ctx, subject.EventID)
	if err != nil {
		return nil, err
	}
	candidates := make([]matching.Candidate, 0, len(needs))
	for _, need := range needs {
		profile, err := s.deps.Users.GetProfile(ctx, need.OwnerID)
		if err != nil {
			continue
		}
		candidates = append(candidates, matching.Candidate{Need: need, Profile: profile})
	}
	return s.matcher.Match(subject, candidates, s.deps.Clock.Now(), confirmed), nil
}

func (s *MatchingService) PublicProfile(ctx context.Context, userID string, confirmed bool) (identity.PublicProfile, error) {
	profile, err := s.deps.Users.GetProfile(ctx, userID)
	if err != nil {
		return identity.PublicProfile{}, err
	}
	return profile.PublicView(confirmed), nil
}
