package bigduration

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type testCase struct {
	Input   string      `json:"-"`
	Compact string      `json:"-"`
	Dur     BigDuration `json:"duration"`
}

var (
	someDurStr       = "2h45m0s"
	someDurStr2      = "200h45m0s"
	someDuration, _  = time.ParseDuration(someDurStr)
	someDuration2, _ = time.ParseDuration(someDurStr2)
)

var testsCases = []testCase{
	{
		Input:   "1year",
		Compact: "1year",
		Dur: BigDuration{
			Years: 1,
		},
	},
	{
		Input:   "2month",
		Compact: "2month",
		Dur: BigDuration{
			Months: 2,
		},
	},
	{
		Input:   "30day",
		Compact: "1month",
		Dur: BigDuration{
			Days: 30,
		},
	},
	{
		Input:   "61day",
		Compact: "2month1day",
		Dur: BigDuration{
			Days: 61,
		},
	},
	{
		Input:   "61day" + someDurStr,
		Compact: "2month1day" + someDurStr,
		Dur: BigDuration{
			Days:  61,
			Nanos: someDuration,
		},
	},
	{
		Input:   "1year61day" + someDurStr,
		Compact: "1year2month1day" + someDurStr,
		Dur: BigDuration{
			Years: 1,
			Days:  61,
			Nanos: someDuration,
		},
	},
	{
		Input:   "426day" + someDurStr,
		Compact: "1year2month1day" + someDurStr,
		Dur: BigDuration{
			Days:  426,
			Nanos: someDuration,
		},
	},
	{
		Input:   "1year2month1day" + someDurStr,
		Compact: "1year2month1day" + someDurStr,
		Dur: BigDuration{
			Years:  1,
			Months: 2,
			Days:   1,
			Nanos:  someDuration,
		},
	},
	{
		Input:   "2year23month103day" + someDurStr,
		Compact: "4year2month3day" + someDurStr,
		Dur: BigDuration{
			Years:  2,
			Months: 23,
			Days:   103,
			Nanos:  someDuration,
		},
	},
	{
		Input:   someDurStr2,
		Compact: "8day8h45m0s",
		Dur: BigDuration{
			Nanos: someDuration2,
		},
	},
	{
		Input:   "2year23month29day" + someDurStr2, // 200h = 8day8h
		Compact: "3year12month2day8h45m0s",         // 12month = 360days vs 1year = 365days
		Dur: BigDuration{
			Years:  2,
			Months: 23,
			Days:   29,
			Nanos:  someDuration2,
		},
	},
	{
		Input:   "2year23month40day",
		Compact: "4year", // 12month = 360days vs 1year = 365days
		Dur: BigDuration{
			Years:  2,
			Months: 23,
			Days:   40,
		},
	},
}

func TestInOut(t *testing.T) {
	for _, tc := range testsCases {
		assertParser(t, tc.Dur, tc.Input)
		assertStringer(t, tc.Dur, tc.Input)
		assertCompact(t, tc.Dur, tc.Compact)
	}
}

func TestJSON(t *testing.T) {
	for _, tc := range testsCases {
		assertMarshal(t, tc)
		assertUnmarshal(t, tc)
	}
}

func assertStringer(t *testing.T, d BigDuration, s string) {
	if d.String() != s {
		t.Errorf("Unexpected string representation %s vs expected: %s", d.String(), s)
	}
}

func assertCompact(t *testing.T, d BigDuration, s string) {
	if d.Compact() != s {
		t.Errorf("Unexpected compact string representation %s vs expected: %s", d.Compact(), s)
	}
}

func assertParser(t *testing.T, d BigDuration, s string) {
	p, err := ParseBigDuration(s)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(d, p) {
		t.Errorf("Unexpected parsed struct %+v vs expected: %+v", d, p)
	}
}

func assertMarshal(t *testing.T, tc testCase) {
	b, _ := json.Marshal(tc)
	actual := string(b)
	expected := fmt.Sprintf(`{"duration":"%s"}`, tc.Input)
	if actual != expected {
		t.Errorf("Unexpected JSON representation %s vs expected: %s", actual, expected)
	}
}

func assertUnmarshal(t *testing.T, tc testCase) {
	var tcj testCase
	data := fmt.Sprintf(`{"duration":"%s"}`, tc.Input)
	err := json.Unmarshal([]byte(data), &tcj)
	if err != nil {
		panic(err)
	}

	if !reflect.DeepEqual(tc.Dur, tcj.Dur) {
		t.Errorf("Unexpected struct from JSON %+v vs expected: %+v", tcj.Dur, tc.Dur)
	}
}
