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
	dvrWindow := gocv.NewWindow("HDZero DVR")
	binaryWindow := gocv.NewWindow("Binary Image")

	width := 240
	height := 180

	img := gocv.NewMat()
	resized := gocv.NewMat()
	dvr.Read(&img)
	gocv.Resize(img, &resized, image.Pt(width, height), 0, 0, gocv.InterpolationLinear)

	var gates []*Gate

	for _, gateConfig := range config.Gates {
		gates = append(gates, NewGate(
			gateConfig.Name,
			resized,
			GateColor2Scalar(gateConfig.Color.LowerBoundHSV),
			GateColor2Scalar(gateConfig.Color.UpperBoundHSV),
			gateConfig.Detection.MinMillisBetweenActivations,
			gateConfig.Detection.MinActivationValue,
			gateConfig.Detection.MinActivationFrames,
			gateConfig.Detection.MinInactivationFrames))
	}

	detector := NewDetector(resized, config.FramesPerSec, config.PropellerMask.Width, config.PropellerMask.Height)
	timer := NewTimer()
	for index, gate := range gates {
		detector.AddGate(gate)
		timer.AddGate(index, gate)
	}

	lapsMsg := fmt.Sprintf("Lap: 0, Time: 0")
	transitionsMsg := fmt.Sprintf("Transition: ... -> ... , Time: 0")

	var frameStart time.Time
	var frameStop time.Time
	for {
		if ok := dvr.Read(&img); !ok {
			break
		}
		frameStart = time.Now()
		gocv.Resize(img, &resized, image.Pt(240, 180), 0, 0, gocv.InterpolationArea)

		if detection := detector.Detect(&resized, binaryWindow); detection != nil {
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

			fmt.Println(lapsMsg)
			fmt.Println(transitionsMsg)
			fmt.Printf("%v\n\n", detection)
		}

		frameStop = time.Now()

		duration := frameStop.Sub(frameStart)

		gocv.PutText(&img, lapsMsg, image.Pt(300, 100), gocv.FontHersheyDuplex, 1, color.RGBA{R: 255, G: 255, B: 255}, 1)
		gocv.PutText(&img, transitionsMsg, image.Pt(300, 150), gocv.FontHersheyDuplex, 1, color.RGBA{R: 255, G: 255, B: 255}, 1)
		gocv.PutText(&img, fmt.Sprintf("Frame latency: %v", duration), image.Pt(300, 200), gocv.FontHersheyDuplex, 1, color.RGBA{R: 255, G: 255, B: 255}, 1)

		dvrWindow.IMShow(img)
		dvrWindow.WaitKey(1)

	}

	if err := dvrWindow.Close(); err != nil {
		panic(fmt.Errorf("could not close DVR window. %s", err.Error()))
	}
}
