package visuals

import (
	"fmt"
	"strconv"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

const (
	hrSensorID   = 56482
	hrThreshLow  = 90
	hrThreshMid  = 120
	hrThreshHigh = 150
	hrThreshExt  = 170

	metricsTextY = 1355

	MSToMPH float64 = 2.2369
	MToMi   float64 = 0.000621
)

var (
	enabled      bool
	currentHR    int
	carsBack     int
	currentSpeed float64
	estDistance  float64
	prevDistance float64
	metricsFont  rl.Font
)

func InitMetrics() {
	metricsFont = rl.LoadFontEx("caskaydia.TTF", 72, nil, 0)
}

func DrawMetrics() {
	if !enabled {
		return
	}
	rl.DrawRectangle(0, 1340, 2560, 100, rl.Color{R: 0, G: 0, B: 0, A: 192})
	if currentSpeed > 2.0 && currentSpeed < 50.0 {
		rl.DrawTextEx(metricsFont, fmt.Sprintf("%.1fmph", currentSpeed), rl.Vector2{X: 1150, Y: metricsTextY}, 72, 0, rl.Blue)
	}
	rl.DrawTextEx(metricsFont, fmt.Sprintf("~%.2fmi", estDistance), rl.Vector2{X: 50, Y: metricsTextY}, 72, 0, rl.Blue)

	if currentHR != 0 {
		hrColor := rl.Blue
		switch {
		case currentHR >= hrThreshExt:
			hrColor = rl.Red
		case currentHR >= hrThreshHigh:
			hrColor = rl.Orange
		case currentHR >= hrThreshMid:
			hrColor = rl.Yellow
		case currentHR >= hrThreshLow:
			hrColor = rl.Green
		}
		rl.DrawTextEx(metricsFont, fmt.Sprintf("%dbpm", currentHR), rl.Vector2{X: 2290, Y: metricsTextY}, 72, 0, hrColor)
	}
}

func MetricsEnabled() bool {
	return enabled
}

func EnableMetrics(b bool) {
	enabled = b
}

func HandleMetricsMessage(args []string) {
	if len(args) < 2 {
		return
	}

	switch args[0] {
	case "hr":
		n, err := strconv.Atoi(args[1])
		if err != nil {
			break
		}
		currentHR = n
	case "cars":
		n, err := strconv.Atoi(args[1])
		if err != nil {
			break
		}
		carsBack = n
	case "speed":
		n, err := strconv.ParseFloat(args[1], 32)
		if err != nil {
			break
		}
		currentSpeed = n * MSToMPH
	case "distance":
		if args[1] == "reset" {
			prevDistance, estDistance = 0, 0
			break
		}
		n, err := strconv.ParseFloat(args[1], 32)
		if err != nil {
			break
		}
		d := n * MToMi
		if d < estDistance {
			prevDistance = estDistance
		}
		estDistance = prevDistance + d
	}
}
