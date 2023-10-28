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
	{"feature:toggles:**", "feature-toggles", true},
	{"feature-toggles:**", "feature-toggles", true},
	{"cartao:branco-gateway:**", "cartao-branco-gateway", true},
	{"cartao-branco-gateway:**", "cartao-branco-gateway", true},
	{":**", "feature-toggles", false},
	{":**", "cartao-branco-gateway", false},
	{"feature:toggles:**", "XXX", false},
	{"feature-toggles:**", "XXX", false},
	{"cartao:branco-gateway:**", "XXX", false},
	{"cartao-branco-gateway:**", "XXX", false},
	{"feature:toggles:**", "", false},
	{"feature-toggles:**", "", false},
	{"cartao:branco-gateway:**", "", false},
	{"cartao-branco-gateway:**", "", false},
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
