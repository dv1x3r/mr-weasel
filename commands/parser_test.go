package commands

import "testing"

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		Input    string
		Prefix   string
		Expected []string
	}{
		{Input: "", Prefix: "", Expected: []string{}},
		{Input: "/cmd", Prefix: "/cmd", Expected: []string{}},
		{Input: "/cmd sub", Prefix: "/cmd", Expected: []string{"sub"}},
		{Input: "/cmd sub 123", Prefix: "/cmd", Expected: []string{"sub", "123"}},
		{Input: "/cmd sub 123 456", Prefix: "/cmd sub", Expected: []string{"123", "456"}},
	}
	for _, test := range tests {
		actual := splitCommand(test.Input, test.Prefix)
		if len(actual) != len(test.Expected) {
			t.Fatalf("len does not match, actual [%+v], expected [%+v]\n", actual, test)
		}
		for i := range actual {
			if actual[i] != test.Expected[i] {
				t.Fatalf("elements do not match, actual [%+v], expected [%+v]\n", actual, test)
			}
		}
	}
}
