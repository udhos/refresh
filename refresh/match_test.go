package refresh

import (
	"testing"
)

type testMatch struct {
	notification   string
	application    string
	expectedResult bool
}

var testMatchTable = []testMatch{
	{"feature:flags:**", "feature-flags", true},
	{"feature-flags:**", "feature-flags", true},
	{"badge:color-gateway:**", "badge-color-gateway", true},
	{"badge-color-gateway:**", "badge-color-gateway", true},
	{":**", "feature-flags", false},
	{":**", "badge-color-gateway", false},
	{"feature:flags:**", "XXX", false},
	{"feature-flags:**", "XXX", false},
	{"badge:color-gateway:**", "XXX", false},
	{"badge-color-gateway:**", "XXX", false},
	{"feature:flags:**", "", false},
	{"feature-flags:**", "", false},
	{"badge:color-gateway:**", "", false},
	{"badge-color-gateway:**", "", false},
	{"badge-color-gateway:**", "#", true},
}

// go test -v ./refresh -run=TestMatch
func TestMatch(t *testing.T) {
	for _, data := range testMatchTable {
		result := DefaultMatchApplication(data.notification, data.application)
		if result != data.expectedResult {
			t.Errorf("notification='%s' application='%s' expected=%t got=%t",
				data.notification, data.application, data.expectedResult, result)
		}
	}
}
