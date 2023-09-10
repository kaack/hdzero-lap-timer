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

func main() {
	var videoPath string
	var config *Config
	var err error
	if videoPath, config, err = ProcessArgs(); err != nil {
		panic(err)
	}

	var dvrWindow *gocv.Window
	var detectionWindow *gocv.Window
	var plotWindow *gocv.Window

	if config.Windows.ShowDVR {
		dvrWindow = gocv.NewWindow("DVR")
	}

	if config.Windows.ShowDetection {
		detectionWindow = gocv.NewWindow("Detection")
	}

	if config.Windows.ShowPlot {
		plotWindow = gocv.NewWindow("Plot")
	}

	dvr, _ := gocv.OpenVideoCapture(videoPath)

	width := 480
	height := 360

	dvrImage := gocv.NewMat()
	resizedImg := gocv.NewMat()
	detectionImage := gocv.NewMat()
	dvr.Read(&dvrImage)
	gocv.Resize(dvrImage, &resizedImg, image.Pt(width, height), 0, 0, gocv.InterpolationLinear)

	var gates []*Gate

	for _, gateConfig := range config.Gates {
		gates = append(gates, NewGate(gateConfig.Name, resizedImg, gateConfig))
	}

	detector := NewDetector(resizedImg, config)
	timer := NewTimer()
	for index, gate := range gates {
		detector.AddGate(gate)
		timer.AddGate(index, gate)
	}

	lapsMsg := fmt.Sprintf("Lap: 0, Time: 0")
	transitionsMsg := fmt.Sprintf("Transition: ... -> ... , Time: 0")

	var frameStart time.Time
	var frameStop time.Time
	var announcedLap *Lap

	frameCount := uint64(0)
	var peakDetected *Detection

	gMat := gocv.NewMat()
	dataChan := make(chan []Point, 100)
	matChan := make(chan *gocv.Mat, 100)
	go graph(dataChan, matChan, &gMat, gates)

Loop:
	for {
		select {
		case nMat := <-matChan:
			if plotWindow != nil {
				plotWindow.IMShow(*nMat)
			}

		default:

			if ok := dvr.Read(&dvrImage); !ok {
				break Loop
			}

			frameCount += 1
			frameStart = time.Now()
			gocv.Resize(dvrImage, &resizedImg, image.Pt(width, height), 0, 0, gocv.InterpolationArea)

			_, peakDetected = detector.Detect(&resizedImg, &detectionImage, frameCount)

			if peakDetected != nil {
				lastDetection := timer.LastDetectionByGate(peakDetected.Activation.Gate)
				if lastDetection != nil {
					framesBetweenPeaks := peakDetected.FrameOffset - lastDetection.FrameOffset
					if framesBetweenPeaks < peakDetected.Activation.Gate.Config.Detection.MinFramesBetweenPeaks {
						// discard detection
						peakDetected = nil
					}
				}
			}

			if peakDetected != nil {
				fmt.Printf("%v\n", peakDetected)

				timer.AddDetection(peakDetected)
				if lastLap := timer.LastLap(); lastLap != nil {
					lastLapTime := time.Duration(int64(detector.MillisPerFrame()*float64(lastLap.Frames()))) * time.Millisecond
					lapsMsg = fmt.Sprintf("Lap: %d, Time: %v, Gate: %s", timer.LapsCount(), lastLapTime, lastLap.Gate().Name)

					if config.Announcements.SayLaps && lastLap != announcedLap {
						//goland:noinspection GoUnhandledErrorResult
						go ttsSay(fmt.Sprintf("%s %s", config.Announcements.PilotName, durationAsTTS(lastLapTime)))
						announcedLap = lastLap
					}

				}

				if lastTransition := timer.LastTransition(); lastTransition != nil {
					lastTransitionTime := time.Duration(detector.MillisPerFrame()*float64(lastTransition.Frames())) * time.Millisecond
					transitionsMsg = fmt.Sprintf("Transition: %s -> %s , Time: %v", lastTransition.start.Activation.Gate.Name, lastTransition.stop.Activation.Gate.Name, lastTransitionTime)

					if lastTransition.stop.Activation.Gate.Config.Announcements.SayTransitions {
						//goland:noinspection GoUnhandledErrorResult
						go ttsSay(fmt.Sprintf("split %s", durationAsTTS(lastTransitionTime)))
					}

				} else if lastDetection := timer.LastDetection(); lastDetection != nil {
					transitionsMsg = fmt.Sprintf("Transition: %s -> ... , Time: 0", lastDetection.Activation.Gate.Name)
				}

				fmt.Println(lapsMsg)
				fmt.Println(transitionsMsg)
				fmt.Println()
				fmt.Println()

			}

			var points []Point
			for _, gate := range gates {
				points = append(points, Point{float64(frameCount), gate._sum4.First()})
			}
			dataChan <- points

			frameStop = time.Now()
			duration := frameStop.Sub(frameStart)

			gocv.PutText(&dvrImage, lapsMsg, image.Pt(300, 100), gocv.FontHersheyDuplex, 1, color.RGBA{R: 255, G: 255, B: 255}, 1)
			gocv.PutText(&dvrImage, transitionsMsg, image.Pt(300, 150), gocv.FontHersheyDuplex, 1, color.RGBA{R: 255, G: 255, B: 255}, 1)
			gocv.PutText(&dvrImage, fmt.Sprintf("Frame latency: %v", duration), image.Pt(300, 200), gocv.FontHersheyDuplex, 1, color.RGBA{R: 255, G: 255, B: 255}, 1)

			if dvrWindow != nil {
				dvrWindow.IMShow(dvrImage)
			}

			if detectionWindow != nil {
				detectionWindow.IMShow(detectionImage)
			}

			if dvrWindow != nil || detectionWindow != nil {
				dvrWindow.WaitKey(1)
			} else {
				time.Sleep(1)
			}

		}
	}

	if err := dvrWindow.Close(); err != nil {
		panic(fmt.Errorf("could not close DVR window. %s", err.Error()))
	}
}
