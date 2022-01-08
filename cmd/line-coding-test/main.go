package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// TODO: 変数名の統一性
const (
	// TimeFormat 走行ログの時間フォーマット
	TimeFormat = "2006-01-02 15:04:05.000"
	// BaseFare 初乗り料金(単位:円)
	BaseFare = 410
	// AdditionalFare 追加運賃(単位:円)
	AdditionalFare = 80
	// FareAdditionDistance 運賃加算距離(単位:m)
	FareAdditionDistance = 237
	// MaxTravelLogCount タクシー最大走行ログ件数(1乗客)
	MaxTravelLogCount = 500000
	// FirstRideDistance 初乗り料金最大走行距離(単位:km)
	FirstRideDistance = 1052.0
	// LowSpeedDrivingStandard 低速走行基準(単位:10km/h)
	LowSpeedDrivingStandard = 10.0
	// SlowFareAdditionTime 低速運賃加算時間(単位:秒)
	SlowFareAdditionTime = 90.0
)

func main() {
	// 走行ログ取得(最大5万行なので、あらかじめメモリを確保しておく)
	driveLogs := make([]string, 0, MaxTravelLogCount)
	reader := bufio.NewScanner(os.Stdin)
	for reader.Scan() {
		driveLogs = append(driveLogs, reader.Text())
	}
	if err := reader.Err(); err != nil {
		log.Fatalf("Failed to Scan: %v", err)
	}

	// 運賃計算
	fare, err := calculateTaxiFare(driveLogs)
	if err != nil {
		log.Fatalf("Failed to calculate taxi fare: %v", err)
	}

	fmt.Println(fare)
}

// calculateTaxiFare タクシー走行ログを元に、乗車料金を算出する
func calculateTaxiFare(driveLogs []string) (int, error) {
	// 運賃
	fare := BaseFare
	// 走行距離, 低速走行時間
	var mileage, lowSpeedRunningTime float64
	// 前レコード(1レコード分)
	var PreviousDriveLog []string

	// 走行距離,低速走行時間算出
	for _, v := range driveLogs {
		travelLog := strings.Split(v, " ")

		// 走行時刻
		ti, err := toTime(travelLog[0])
		if err != nil {
			return 0, err
		}

		// 移動距離
		mi, err := strconv.ParseFloat(travelLog[1], 64)
		if err != nil {
			return 0, err
		}

		// 深夜料金計算
		// 00:00:00.000 〜 04:59:59.999 又は 22:00:00.000 〜 23:59.59.999
		hour := ti.Hour()
		if (5 > hour && hour >= 0) || (24 > hour && hour >= 22) {
			mileage += mi * 1.25
		} else {
			mileage += mi
		}

		if mileage != 0.0 && len(PreviousDriveLog) > 0 {
			f, err := strconv.ParseFloat(PreviousDriveLog[1], 64)
			if err != nil {
				return 0, err
			}
			// 移動距離(km)
			travelDistance := (mi - f) * 0.001

			t, err := toTime(PreviousDriveLog[0])
			if err != nil {
				return 0, err
			}
			duration := ti.Sub(t)
			// 移動にかかった時間
			travelTime := duration.Hours()
			// 深夜料金
			if (5 > hour && hour >= 0) || (24 > hour && hour >= 22) {
				travelTime *= 1.25
			}

			// 平均速度
			averageSpeed := travelDistance / travelTime
			if averageSpeed < LowSpeedDrivingStandard {
				lowSpeedRunningTime += duration.Seconds()
			}
		}

		PreviousDriveLog = travelLog
	}

	// 距離料金計算
	mileage -= FirstRideDistance
	// forでは回さない
	if mileage > 0.0 {
		fare += AdditionalFare
		fare += (int(mileage) / FareAdditionDistance) * AdditionalFare
		if int(mileage)%FareAdditionDistance > 0 {
			fare += AdditionalFare
		}
	}

	// 低速運賃計算
	if lowSpeedRunningTime >= SlowFareAdditionTime {
		fare += AdditionalFare
		lowSpeedRunningTime -= SlowFareAdditionTime
		fare += (int(lowSpeedRunningTime) / SlowFareAdditionTime) * AdditionalFare
		if int(lowSpeedRunningTime)%SlowFareAdditionTime > 0.0 {
			fare += AdditionalFare
		}
	}

	return fare, nil
}

func toTime(str string) (time.Time, error) {
	day := 1

	// hh:mm:ss.fff
	hour, err := strconv.Atoi(str[0:2])
	if err != nil {
		return time.Time{}, err
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
