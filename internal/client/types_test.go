package client

import (
	"encoding/json"
	"testing"
)

func TestFlexIntUnmarshal(t *testing.T) {
	cases := []struct {
		in   string
		want FlexInt
	}{
		{`24`, 24},
		{`"24"`, 24},
		{`"24 mo."`, 24},
		{`"36 months"`, 36},
		{`"-5"`, -5},
		{`null`, 0},
		{`""`, 0},
		{`"months"`, 0},
	}
	for _, tc := range cases {
		var got FlexInt
		if err := json.Unmarshal([]byte(tc.in), &got); err != nil {
			t.Errorf("FlexInt(%s): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("FlexInt(%s) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestFlexBoolUnmarshal(t *testing.T) {
	cases := []struct {
		in   string
		want FlexBool
	}{
		{`true`, true},
		{`false`, false},
		{`1`, true},
		{`0`, false},
		{`"1"`, true},
		{`"0"`, false},
		{`null`, false},
	}
	for _, tc := range cases {
		var got FlexBool
		if err := json.Unmarshal([]byte(tc.in), &got); err != nil {
			t.Errorf("FlexBool(%s): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("FlexBool(%s) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeMoney(t *testing.T) {
	cases := map[string]string{
		"1,234.56":  "1234.56",
		" 99.00 ":   "99.00",
		"1,000,000": "1000000",
		"":          "",
	}
	for in, want := range cases {
		if got := NormalizeMoney(in); got != want {
			t.Errorf("NormalizeMoney(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRefIDOrZero(t *testing.T) {
	if got := (*Ref)(nil).IDOrZero(); got != 0 {
		t.Errorf("nil ref = %d", got)
	}
	if got := (&Ref{ID: 9}).IDOrZero(); got != 9 {
		t.Errorf("ref = %d", got)
	}
}
