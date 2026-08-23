package matching

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/wyw14/cry-090/internal/domain/event"
	"github.com/wyw14/cry-090/internal/domain/identity"
)

type Candidate struct {
	Need    Need
	Profile identity.Profile
}

type Reason struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Weight  int    `json:"weight"`
}

type Result struct {
	NeedID  string                 `json:"need_id"`
	User    identity.PublicProfile `json:"user"`
	Score   int                    `json:"score"`
	Reasons []Reason               `json:"reasons"`
}

type Matcher struct{}

func (Matcher) Match(subject Need, candidates []Candidate, now time.Time, confirmed map[string]bool) []Result {
	if !subject.Active || !now.UTC().Before(subject.ExpiresAt) {
		return nil
	}
	results := make([]Result, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Need.OwnerID == subject.OwnerID || !candidate.Need.Active {
			continue
		}
		if candidate.Need.EventID != subject.EventID || !candidate.Need.Window.Overlaps(subject.Window) {
			continue
		}
		if !roleCompatible(subject.WantedRole, candidate.Profile.Roles) {
			continue
		}
		// A match result is shown before the two parties have confirmed each
		// other, so it must never carry contact details or precise coordinates.
		// confirmed is keyed by the candidate owner: only once both sides have
		// agreed does the profile's PublicView expose those sensitive fields.
		isConfirmed := confirmed[candidate.Need.OwnerID]
		public := candidate.Profile.PublicView(isConfirmed)
		reasons := make([]Reason, 0, 4)
		score := 0
		if stylesOverlap(subject.Styles, candidate.Need.Styles) {
			reasons = append(reasons, Reason{Code: "style", Message: "dance styles overlap", Weight: 35})
			score += 35
		}
		if public.DistanceBucket != nil && *public.DistanceBucket <= subject.MaxDistanceBucket {
			reasons = append(reasons, Reason{Code: "distance", Message: fmt.Sprintf("public distance band %d is acceptable", *public.DistanceBucket), Weight: 20})
			score += 20
		}
		// Only the candidate's explicitly-public tags can be shared in a reason;
		// private tags never surface, even as part of a match explanation.
		shared := sharedTags(subject.Tags, public.Tags)
		if len(shared) > 0 {
			weight := min(30, len(shared)*10)
			reasons = append(reasons, Reason{Code: "tags", Message: "shared public tags: " + strings.Join(shared, ", "), Weight: weight})
			score += weight
		}
		if candidate.Need.Level == subject.Level {
			reasons = append(reasons, Reason{Code: "level", Message: "requested levels are aligned", Weight: 15})
			score += 15
		}
		if score == 0 {
			continue
		}
		results = append(results, Result{NeedID: candidate.Need.ID, User: public, Score: score, Reasons: reasons})
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].NeedID < results[j].NeedID
		}
		return results[i].Score > results[j].Score
	})
	return results
}

func roleCompatible(wanted identity.DanceRole, offered []identity.DanceRole) bool {
	for _, role := range offered {
		if role == identity.RoleSwitch || wanted == identity.RoleSwitch || role == wanted {
			return true
		}
	}
	return false
}

func stylesOverlap(left, right []event.SalsaStyle) bool {
	seen := make(map[event.SalsaStyle]struct{}, len(left))
	for _, style := range left {
		seen[style] = struct{}{}
	}
	for _, style := range right {
		if _, ok := seen[style]; ok {
			return true
		}
	}
	return false
}

func sharedTags(left, right []string) []string {
	seen := make(map[string]struct{}, len(left))
	for _, tag := range left {
		seen[strings.ToLower(tag)] = struct{}{}
	}
	result := make([]string, 0)
	for _, tag := range right {
		tag = strings.ToLower(tag)
		if _, ok := seen[tag]; ok {
			result = append(result, tag)
		}
	}
	sort.Strings(result)
	return result
}
