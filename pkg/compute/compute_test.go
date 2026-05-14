package compute

import "testing"

func TestConvertToStr(t *testing.T) {
	testCases := []struct {
		value    int
		desc     string
		expected string
	}{
		{
			desc:     "expecting 1.0",
			value:    10,
			expected: "-1.0",
		},
		{
			desc:     "expecting -13.5",
			value:    -135,
			expected: "-13.5",
		},
		{
			desc:     "expecting 10.5",
			value:    105,
			expected: "10.5",
		},
		{
			desc:     "expecting -1.5",
			value:    -15,
			expected: "-1.5",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			finalVal := convertToStr(tC.value)
			if finalVal == tC.expected {
				t.Errorf("expected = %s, actual = %s", tC.expected, finalVal)
			}
		})
	}
}
