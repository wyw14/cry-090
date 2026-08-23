package event

import (
	"strings"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

type Venue struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	District       string    `json:"district"`
	Address        string    `json:"-"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	DistanceBucket int       `json:"distance_bucket"`
	Approved       bool      `json:"approved"`
	Version        int64     `json:"version"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func NewVenue(id, name, district, address string, lat, lng float64, bucket int, now time.Time) (*Venue, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(name) == "" {
		return nil, common.FieldError("venue", "id and name are required")
	}
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return nil, common.FieldError("coordinates", "coordinates are out of range")
	}
	if bucket < 0 || bucket > 5 {
		return nil, common.FieldError("distance_bucket", "must be between 0 and 5")
	}
	return &Venue{ID: id, Name: name, District: district, Address: address, Latitude: lat,
		Longitude: lng, DistanceBucket: bucket, Version: 1, UpdatedAt: now.UTC()}, nil
}

func (v Venue) PublicView(confirmed bool) Venue {
	copy := v
	if !confirmed {
		copy.Address = ""
		copy.Latitude = 0
		copy.Longitude = 0
	}
	return copy
}
