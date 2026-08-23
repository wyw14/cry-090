package audit

import "time"

type Entry struct {
	ID         string         `json:"id"`
	ActorID    string         `json:"actor_id"`
	Source     string         `json:"source"`
	EntityType string         `json:"entity_type"`
	EntityID   string         `json:"entity_id"`
	Before     map[string]any `json:"before"`
	After      map[string]any `json:"after"`
	Reason     string         `json:"reason"`
	CreatedAt  time.Time      `json:"created_at"`
}

func New(id, actor, source, entityType, entityID, reason string, before, after map[string]any, now time.Time) Entry {
	return Entry{ID: id, ActorID: actor, Source: source, EntityType: entityType, EntityID: entityID,
		Before: cloneMap(before), After: cloneMap(after), Reason: reason, CreatedAt: now.UTC()}
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	result := make(map[string]any, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}
