package common

import "strings"

type Money struct {
	Minor    int64  `json:"minor"`
	Currency string `json:"currency"`
}

func NewMoney(minor int64, currency string) (Money, error) {
	if minor < 0 {
		return Money{}, FieldError("price_minor", "must not be negative")
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if len(currency) != 3 {
		return Money{}, FieldError("currency", "must be a three-letter ISO code")
	}
	return Money{Minor: minor, Currency: currency}, nil
}
