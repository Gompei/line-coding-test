package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// TimeFormat 走行ログの時間フォーマット
	TimeFormat = "2006-01-02 15:04:05.000"
	// AdditionalFare 追加運賃(単位:円)
	AdditionalFare = 80
)

type DriveLog struct {
	Time     time.Time
	Distance float64
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Failed to calculate taxi fare: %v", err)
	}
}

func run() error {
	driveLogs, err := scanDriveLogs(os.Stdin)
	if err != nil {
		return err
	}

	fare, err := calcTaxiFare(driveLogs)
	if err != nil {
		return err
	}
	fmt.Println(fare)

	return nil
}

func scanDriveLogs(input io.Reader) ([]DriveLog, error) {
	driveLogs := make([]DriveLog, 0, 500000)
	reader := bufio.NewScanner(input)
	for reader.Scan() {
		driveLog, err := parseDriveLog(reader.Text())
		if err != nil {
			return nil, err
		}

		driveLogs = append(driveLogs, driveLog)
	}

	return driveLogs, reader.Err()
}

// calcTaxiFare タクシー走行ログを元に、乗車料金を算出する
func calcTaxiFare(driveLogs []DriveLog) (int, error) {
	// 走行距離, 低速走行総時間
	var totalDistance, lowSpeedDrivingTotalTime float64

	// 走行距離,低速走行時間算出
	for i, driveLog := range driveLogs {
		totalDistance += driveLog.calcDistance()

		if i > 0 {
			// 平均移動速度
			averageSpeed := calcAverageSpeed(driveLogs[i-1].Distance, driveLogs[i-1].Distance+driveLog.Distance, driveLogs[i-1].Time, driveLog.Time)
			// 低速時間算出
			lowSpeedDrivingTotalTime += driveLog.calcLowSpeedDrivingTime(averageSpeed, driveLogs[i-1].Time)
		}
	}

	// 距離料金計算
	fare := calcTaxiDistanceFare(totalDistance)

	// 低速運賃計算
	fare += calcTaxiLowSpeedDriveFare(lowSpeedDrivingTotalTime)

	return fare, nil
}

// calcTaxiDistanceFare タクシーの走行距離を元に運賃を計算する
func calcTaxiDistanceFare(mileage float64) int {
	// 初乗り運賃
	fare := 410

	// 1052mまでは初乗り運賃で走行可能、 後237m毎に運賃加算
	if mileage >= 1053.0 {
		fare += (int(mileage-1053.0)/237 + 1) * AdditionalFare
	}

	return fare
}

// calcTaxiLowSpeedDriveFare タクシーの低速走行総時間を元に加算運賃を計算する
func calcTaxiLowSpeedDriveFare(lowSpeedDrivingTotalTime float64) int {
	// 加算運賃
	additionalFare := 0

	// 最初の89秒は無料、後90秒毎に運賃加算
	if lowSpeedDrivingTotalTime >= 90 {
		additionalFare += (int(lowSpeedDrivingTotalTime-90)/90 + 1) * AdditionalFare
	}

	return additionalFare
}

func (d DriveLog) calcLowSpeedDrivingTime(averageSpeed float64, previousTimeLog time.Time) float64 {
	// 低速走行基準: 10km/h以下
	if averageSpeed <= 10.0 {
		// 深夜であれば、1.25倍にする
		s := d.Time.Sub(previousTimeLog).Seconds()
		if isMidnight(d.Time.Hour()) {
			return s * 1.25
		}
		return s
	}
	return 0
}

func (d DriveLog) calcDistance() float64 {
	if isMidnight(d.Time.Hour()) {
		return d.Distance * 1.25
	}
	return d.Distance
}

func calcAverageSpeed(bd, ad float64, bt, at time.Time) float64 {
	// 移動距離(km)
	driveDistance := (ad - bd) * 0.001

	// 所要時間
	time := at.Sub(bt).Hours()

	return driveDistance / time
}

func isMidnight(hour int) bool {
	// 今回の場合、24,25,26..形式での使用は想定外
	// 0時~4時59分59秒... 又は 22時~23時59分59秒...
	if (5 > hour && hour >= 0) || (24 > hour && hour >= 22) {
		return true
	}
	return false
}

func toTime(str string) (time.Time, error) {
	day := 1

	// hh:mm:ss.fff
	hour, err := strconv.Atoi(strings.Split(str, ":")[0])
	if err != nil {
		return time.Time{}, err
	}

	// 想定外の走行時間
	if hour > 99 {
		return time.Time{}, errors.New("unexpected running time")
	}

	for i := hour; 24 <= i; i -= 24 {
		hour -= 24
		day++
	}

	t, err := time.Parse(TimeFormat, fmt.Sprintf("2006-01-0%d %d:%s", day, hour, str[3:]))
	if err != nil {
		return t, err
	}

	return t, nil
}

func parseDriveLog(log string) (DriveLog, error) {
	driveLog := strings.Split(log, " ")

	time, err := parseLogTime(driveLog[0])
	if err != nil {
		return DriveLog{}, err
	}

	distance, err := parseLogDistance(driveLog[1])
	if err != nil {
		return DriveLog{}, err
	}

	return DriveLog{
		Time:     time,
		Distance: distance,
	}, nil
}

func parseLogTime(timeLog string) (time.Time, error) {
	t, err := toTime(timeLog)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func parseLogDistance(distanceLog string) (float64, error) {
	distance, err := strconv.ParseFloat(distanceLog, 64)
	if err != nil {
		return 0, err
	}
	return distance, nil
}
