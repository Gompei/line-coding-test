package main

import (
	"testing"
	"time"
)

func TestCalcTaxiFare(t *testing.T) {
	// ログ内容は適当
	testCases := map[string]struct {
		driveLogs []string
		wantErr   bool
		expected  int
	}{
		"1527m run": {
			driveLogs: []string{
				"13:00:00.100 0.0",
				"13:01:00.100 500.0",
				"13:02:00.100 500.0",
				"13:03:00.100 527.0",
			},
			expected: 650,
		},
		"240 seconds low speed driving": {
			driveLogs: []string{
				"13:00:00.100 0.0",
				"13:01:00.100 100.0",
				"13:02:00.100 100.0",
				"13:03:00.100 100.0",
				"13:04:00.100 100.0",
			},
			expected: 570,
		},
		"1200m run at midnight": {
			driveLogs: []string{
				"22:00:00.100 0.0",
				"22:01:00.100 400.0",
				"22:02:00.100 400.0",
				"22:03:00.100 400.0",
			},
			expected: 570,
		},
		"120 seconds low speed driving": {
			driveLogs: []string{
				"22:00:00.100 0.0",
				"22:01:00.100 100.0",
				"22:02:00.100 100.0",
			},
			expected: 490,
		},
	}

	for name, tt := range testCases {
		driveLogs := make([]DriveLog, 0, len(tt.driveLogs))
		for _, log := range tt.driveLogs {
			driveLog, err := parseDriveLog(log)
			if err != nil {
				t.Fatal("parse failure", err)
			}

			driveLogs = append(driveLogs, driveLog)
		}

		t.Run(name, func(t *testing.T) {
			fare, err := calcTaxiFare(driveLogs)
			if tt.wantErr {
				if err == nil {
					t.Fatal("calculateTaxiFare should be fail", err)
				}
			} else {
				if err != nil {
					t.Fatal("calculateTaxiFare should be successful", err)
				}

				if tt.expected != fare {
					t.Fatalf("expected to be %d, actual %d", tt.expected, fare)
				}
			}
		})
	}
}

func TestCalcTaxiMileageFare(t *testing.T) {
	testCases := map[string]struct {
		mileage  float64
		expected int
	}{
		// 以下
		"less than 1052m": {
			mileage:  1052.0,
			expected: 410,
		},
		"more than 1053m": {
			mileage:  1053.0,
			expected: 490,
		},
		"less than 1289m": {
			mileage:  1289.0,
			expected: 490,
		},
		"more than 1290m": {
			mileage:  1290.0,
			expected: 570,
		},
		"less than 1526m": {
			mileage:  1526.0,
			expected: 570,
		},
		"more than 1527m": {
			mileage:  1527.0,
			expected: 650,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			fare := calcTaxiDistanceFare(tt.mileage)
			if tt.expected != fare {
				t.Fatalf("expected to be %v, actual %v", tt.expected, fare)
			}
		})
	}
}

func TestCalcTaxiLowSpeedDriveFare(t *testing.T) {
	testCases := map[string]struct {
		lowSpeedDrivingTotalTime float64
		expected                 int
	}{
		// 未満
		"less than 90 seconds": {
			lowSpeedDrivingTotalTime: 89.0,
			expected:                 0,
		},
		"more than 90 seconds": {
			lowSpeedDrivingTotalTime: 90.0,
			expected:                 80,
		},
		"less than 180 seconds": {
			lowSpeedDrivingTotalTime: 179.0,
			expected:                 80,
		},
		"more than 180 seconds": {
			lowSpeedDrivingTotalTime: 180.0,
			expected:                 160,
		},
		"less than 270 seconds": {
			lowSpeedDrivingTotalTime: 269.0,
			expected:                 160,
		},
		"more than 270 seconds": {
			lowSpeedDrivingTotalTime: 270.0,
			expected:                 240,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			additionalFare := calcTaxiLowSpeedDriveFare(tt.lowSpeedDrivingTotalTime)
			if tt.expected != additionalFare {
				t.Fatalf("expected to be %v, actual %v", tt.expected, additionalFare)
			}
		})
	}
}

func TestCalcAverageSpeed(t *testing.T) {
	const TimeFormat = "2006-01-02 15:04:05.000"

	testCases := map[string]struct {
		beforeMileage float64
		afterMileage  float64
		beforeTime    string
		afterTime     string
		expected      float64
	}{
		// 今回は異常系なし
		"30km/h": {
			// 単位:メートル
			beforeMileage: 0.0,
			afterMileage:  30000.0,
			beforeTime:    "2006-01-02 01:00:00.000",
			afterTime:     "2006-01-02 02:00:00.000",
			expected:      30.0,
		},
	}

	for name, tt := range testCases {
		beforeTime, err := time.Parse(TimeFormat, tt.beforeTime)
		if err != nil {
			t.Fatal("conversion failure", err)
		}

		afterTime, err := time.Parse(TimeFormat, tt.afterTime)
		if err != nil {
			t.Fatal("conversion failure", err)
		}

		t.Run(name, func(t *testing.T) {
			averageSpeed := calcAverageSpeed(tt.beforeMileage, tt.afterMileage, beforeTime, afterTime)
			if tt.expected != averageSpeed {
				t.Fatalf("expected to be %v, actual %v", tt.expected, averageSpeed)
			}
		})
	}
}

func TestIsMidnight(t *testing.T) {
	testCases := map[string]struct {
		hour     int
		expected bool
	}{
		"5 am": {
			hour:     5,
			expected: false,
		},
		// 今回の場合、24,25,26..形式での使用は想定外
		"24 am": {
			hour:     24,
			expected: false,
		},
		"0 am": {
			hour:     0,
			expected: true,
		},
		"22 pm": {
			hour:     22,
			expected: true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			if tt.expected != isMidnight(tt.hour) {
				t.Fatalf("expected to be %v, actual %v", tt.expected, !tt.expected)
			}
		})
	}
}

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
