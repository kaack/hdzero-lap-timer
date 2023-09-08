// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

import "gocv.io/x/gocv"

type Gate struct {
	Name                        string
	minMillisBetweenActivations int
	minActivationValue          float64
	minActivationFrames         int
	minInactivationFrames       int

	lastDetection *Detection

	_markerLowerBoundHSV gocv.Mat
	_markerUpperBoundHSV gocv.Mat
	_markerMask          gocv.Mat
}

func NewGate(name string, img gocv.Mat,
	markerLowerBoundHSV gocv.Scalar,
	markerUpperBoundHSV gocv.Scalar,
	minMillisBetweenActivations int,
	minActivationValue float64,
	minActivationFrames int,
	minInactivationFrames int) *Gate {
	return &Gate{
		Name:                        name,
		minMillisBetweenActivations: minMillisBetweenActivations,
		minActivationValue:          minActivationValue,
		minActivationFrames:         minActivationFrames,
		minInactivationFrames:       minInactivationFrames,
		_markerLowerBoundHSV:        gocv.NewMatWithSizeFromScalar(markerLowerBoundHSV, img.Rows(), img.Cols(), gocv.MatTypeCV8UC3),
		_markerUpperBoundHSV:        gocv.NewMatWithSizeFromScalar(markerUpperBoundHSV, img.Rows(), img.Cols(), gocv.MatTypeCV8UC3),
		_markerMask:                 gocv.NewMatWithSize(img.Cols(), img.Rows(), gocv.MatTypeCV8UC3),
		lastDetection:               nil,
	}
}

func (g *Gate) IsSameHue(pixel gocv.Vecb) bool {
	lowerPixel := g._markerLowerBoundHSV.GetVecbAt(0, 0)
	upperPixel := g._markerUpperBoundHSV.GetVecbAt(0, 0)
	return pixel[0] >= lowerPixel[0] && pixel[0] <= upperPixel[0]
}
