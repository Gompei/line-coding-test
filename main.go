package main

import (
	"bufio"
	"errors"
	"fmt"
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

func main() {
	// 走行ログ取得(最大5万行なので、あらかじめメモリを確保しておく)
	// Scan後に計算しても良いが、今回は別関数に切り出し
	driveLogs := make([]string, 0, 500000)
	reader := bufio.NewScanner(os.Stdin)
	for reader.Scan() {
		driveLogs = append(driveLogs, reader.Text())
	}
	if err := reader.Err(); err != nil {
		log.Fatalf("Failed to Scan: %v", err)
	}

	// 運賃計算
	fare, err := calcTaxiFare(driveLogs)
	if err != nil {
		log.Fatalf("Failed to calculate taxi fare: %v", err)
	}

	fmt.Println(fare)
}

// calcTaxiFare タクシー走行ログを元に、乗車料金を算出する
func calcTaxiFare(driveLogs []string) (int, error) {
	// 走行距離, 低速走行総時間
	var totalMileage, lowSpeedDrivingTotalTime float64
	// 前レコードの走行時刻,距離
	var previousDriveTime time.Time
	var previousDriveMileage float64

	// 走行距離,低速走行時間算出
	for _, v := range driveLogs {
		driveLog := strings.Split(v, " ")

		// 走行時刻
		driveTime, err := toTime(driveLog[0])
		if err != nil {
			return 0, err
		}

		// 移動距離
		mileage, err := strconv.ParseFloat(driveLog[1], 64)
		if err != nil {
			return 0, err
		}

		// 深夜料金計算
		if isMidnight(driveTime.Hour()) {
			totalMileage += mileage * 1.25
		} else {
			totalMileage += mileage
		}

		if mileage != 0.0 {
			averageSpeed := calcAverageSpeed(previousDriveMileage, mileage+previousDriveMileage, previousDriveTime, driveTime)
			// 低速走行基準: 10km/h以下
			if averageSpeed <= 10.0 {
				// 深夜であれば、1.25倍にする
				s := driveTime.Sub(previousDriveTime).Seconds()
				if isMidnight(driveTime.Hour()) {
					s = s * 1.25
				}
				lowSpeedDrivingTotalTime += s
			}
		}

		previousDriveTime = driveTime
		previousDriveMileage = mileage
	}

	// 距離料金計算
	fare := calcTaxiMileageFare(totalMileage)

	// 低速運賃計算
	fare += calcTaxiLowSpeedDriveFare(lowSpeedDrivingTotalTime)

	return fare, nil
}

// calcTaxiMileageFare タクシーの走行距離を元に運賃を計算する
func calcTaxiMileageFare(mileage float64) int {
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

func calcAverageSpeed(bm, am float64, bt, at time.Time) float64 {
	// 移動距離(km)
	driveDistance := (am - bm) * 0.001

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
