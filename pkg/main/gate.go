// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

import (
	"fmt"
	"gocv.io/x/gocv"
)

type Gate struct {
	Name   string
	Config *GateConfig
	//minFramesBetweenPeaks int
	//minActiveValue          float64
	//minActiveFrames         int
	//minInactiveFrames       int

	LastDetection *Detection

	_markerLowerBoundHSV gocv.Mat
	_markerUpperBoundHSV gocv.Mat
	_markerMask          gocv.Mat

	_lastArea       int
	_activeArea     int
	_activeValue    int
	_activeFrames   int
	_inactiveFrames int

	_sum0 *SumBuffer
	_sum1 *SumBuffer
	_sum2 *SumBuffer
	_sum3 *SumBuffer
	_sum4 *SumBuffer

	_peaks []*Detection
}

func NewGate(name string, img gocv.Mat,
	config *GateConfig,
) *Gate {
	return &Gate{
		Name:                 name,
		Config:               config,
		_markerLowerBoundHSV: gocv.NewMatWithSizeFromScalar(GateColor2Scalar(config.Color.LowerBoundHSV), img.Rows(), img.Cols(), gocv.MatTypeCV8UC3),
		_markerUpperBoundHSV: gocv.NewMatWithSizeFromScalar(GateColor2Scalar(config.Color.UpperBoundHSV), img.Rows(), img.Cols(), gocv.MatTypeCV8UC3),
		_markerMask:          gocv.NewMatWithSize(img.Cols(), img.Rows(), gocv.MatTypeCV8UC3),

		_sum0: NewSumBuffer(10),
		_sum1: NewSumBuffer(10),
		_sum2: NewSumBuffer(5),
		_sum3: NewSumBuffer(3),
		_sum4: NewSumBuffer(3),

		LastDetection: nil,
	}
}

func (g *Gate) IsSameHue(pixel gocv.Vecb) bool {
	lowerPixel := g._markerLowerBoundHSV.GetVecbAt(0, 0)
	upperPixel := g._markerUpperBoundHSV.GetVecbAt(0, 0)
	return pixel[0] >= lowerPixel[0] && pixel[0] <= upperPixel[0]
}

func GateColor2Scalar(hsv []int) gocv.Scalar {
	return gocv.NewScalar(float64(hsv[0]), float64(hsv[1]), float64(hsv[2]), 0.0)
}

func (g *Gate) State() string {
	return fmt.Sprintf("name: %s, av: %v, af: %v, if: %v", g.Name, g._activeValue, g._activeFrames, g._inactiveFrames)
}
