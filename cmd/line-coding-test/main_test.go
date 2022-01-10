package main

import "testing"

func TestToTime(t *testing.T) {
	testCases := map[string]struct {
		timeLog  string
		wantErr  bool
		expected string
	}{
		"not spreads across days": {
			timeLog:  "23:00:00.000",
			expected: "2006-01-01 23:00:00 +0000 UTC",
		},
		"spreads across days": {
			timeLog:  "25:00:00.000",
			expected: "2006-01-02 01:00:00 +0000 UTC",
		},
		"over 99 hours": {
			timeLog: "100:00:00.000",
			wantErr: true,
		},
		"unanticipated time": {
			timeLog: "hh:mm:ss.fff",
			wantErr: true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			ti, err := toTime(tt.timeLog)
			if tt.wantErr {
				if err == nil {
					t.Fatal("toTime should be fail", err)
				}
			} else {
				if err != nil {
					t.Fatal("toTime should be successful", err)
				}

				if tt.expected != ti.String() {
					t.Fatalf("expected to be %v, actual %v", tt.expected, ti.String())
				}
			}
		})
	}
}
