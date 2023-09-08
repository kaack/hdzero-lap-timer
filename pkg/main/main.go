// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

import "C"
import (
	"flag"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"time"
)

func ProcessArgs() (string, *Config, error) {
	self, _ := os.Executable()
	self = filepath.Base(self)

	var videoPath string
	var configPath string

	flag.StringVar(&videoPath, "video", "", "path to mp4, ts, or rtsp stream")
	flag.StringVar(&configPath, "config", "", "path to config file")

	flag.Parse()

	if videoPath == "" {
		fmt.Printf("%s: error: video argument is required\n", self)
		os.Exit(1)
	}

	if configPath == "" {
		fmt.Printf("%s: error: config argument is required\n", self)
		os.Exit(1)
	}

	var config *Config
	var err error
	if config, err = NewConfig(configPath); err != nil {
		return "", nil, err
	}

	return videoPath, config, nil
}

func GateColor2Scalar(hsv []int) gocv.Scalar {
	return gocv.NewScalar(float64(hsv[0]), float64(hsv[1]), float64(hsv[2]), 0.0)
}

func main() {

	var videoPath string
	var config *Config
	var err error
	if videoPath, config, err = ProcessArgs(); err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", config)

	dvr, _ := gocv.OpenVideoCapture(videoPath)
	window := gocv.NewWindow("HDZero DVR")

	img := gocv.NewMat()
	dvr.Read(&img)

	var gates []*Gate

	for _, gateConfig := range config.Gates {
		gates = append(gates, NewGate(
			gateConfig.Name,
			img,
			GateColor2Scalar(gateConfig.Color.LowerBoundHSV),
			GateColor2Scalar(gateConfig.Color.UpperBoundHSV),
			gateConfig.Detection.MinMillisBetweenActivations,
			gateConfig.Detection.MinActivationValue,
			gateConfig.Detection.MinActivationFrames,
			gateConfig.Detection.MinInactivationFrames))
	}

	detector := NewDetector(img, config.FramesPerSec, config.PropellerMask.Width, config.PropellerMask.Height)
	timer := NewTimer()
	for index, gate := range gates {
		detector.AddGate(gate)
		timer.AddGate(index, gate)
	}

	lapsMsg := fmt.Sprintf("Lap: 0, Time: 0")
	transitionsMsg := fmt.Sprintf("Transition: ... -> ... , Time: 0")

	for {
		if ok := dvr.Read(&img); !ok {
			break
		}

		if detection := detector.Detect(&img, window); detection != nil {
			timer.AddDetection(detection)
			if lastLap := timer.LastLap(); lastLap != nil {
				lastLapTime := time.Duration(detector.MillisPerFrame()*lastLap.Frames()) * time.Millisecond
				lapsMsg = fmt.Sprintf("Lap: %d, Time: %v, Gate: %s", timer.LapsCount(), lastLapTime, lastLap.Gate().Name)
			}

			if lastTransition := timer.LastTransition(); lastTransition != nil {
				lastTransitionTime := time.Duration(detector.MillisPerFrame()*lastTransition.Frames()) * time.Millisecond
				transitionsMsg = fmt.Sprintf("Transition: %s -> %s , Time: %v", lastTransition.start.Gate.Name, lastTransition.stop.Gate.Name, lastTransitionTime)
			} else if lastDetection := timer.LastDetection(); lastDetection != nil {
				transitionsMsg = fmt.Sprintf("Transition: %s -> ... , Time: 0", lastDetection.Gate.Name)
			}

			fmt.Printf("%v\n", detection)
		}

		gocv.PutText(&img, lapsMsg, image.Pt(300, 100), gocv.FontHersheyDuplex, 1, color.RGBA{R: 255, G: 255, B: 255}, 2)
		gocv.PutText(&img, transitionsMsg, image.Pt(300, 150), gocv.FontHersheyDuplex, 1, color.RGBA{R: 255, G: 255, B: 255}, 2)

		window.IMShow(img)
		window.WaitKey(1)
	}

	if err := window.Close(); err != nil {
		panic(fmt.Errorf("could not close DVR window. %s", err.Error()))
	}
}
