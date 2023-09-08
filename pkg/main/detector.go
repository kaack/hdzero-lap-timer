// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

import (
	"gocv.io/x/gocv"
	"image"
	"image/color"
)

type Detector struct {
	gates []*Gate

	_millisPerFrame int

	_buff         *StreamBuffer
	_lastSeenGate *Gate

	_frameCount uint64

	_markersMask   gocv.Mat
	_img           gocv.Mat
	_hsvImg        gocv.Mat
	_grayImg       gocv.Mat
	_binaryImg     gocv.Mat
	_leftPropPoly  gocv.PointsVector
	_rightPropPoly gocv.PointsVector
	_kernel        gocv.Mat
}

func NewDetector(img gocv.Mat,
	framesPerSec int,
	propWidth int,
	propHeight int) Detector {

	detectionWindowInMillis := 3000
	detectionWindowInFrames := detectionWindowInMillis / 1000 * framesPerSec

	detector := Detector{
		_millisPerFrame: 1000 / framesPerSec,
		_buff:           NewStreamBuffer(detectionWindowInFrames),
		_frameCount:     0,
		_lastSeenGate:   nil,

		_markersMask: gocv.NewMat(),
		_hsvImg:      gocv.NewMat(),
		_img:         gocv.NewMat(),
		_grayImg:     gocv.NewMat(),
		_binaryImg:   gocv.NewMat(),

		_leftPropPoly:  gocv.NewPointsVectorFromPoints([][]image.Point{{image.Pt(0, img.Rows()), image.Pt(0, img.Rows()-propHeight), image.Pt(propWidth, img.Rows())}}),
		_rightPropPoly: gocv.NewPointsVectorFromPoints([][]image.Point{{image.Pt(img.Cols(), img.Rows()), image.Pt(img.Cols(), img.Rows()-propHeight), image.Pt(img.Cols()-propWidth, img.Rows())}}),

		_kernel: gocv.GetStructuringElement(gocv.MorphRect, image.Pt(12, 12)),
	}

	return detector
}

func (t *Detector) Detect(img *gocv.Mat, _ *gocv.Window) *Detection {
	t._frameCount += 1

	frame := *img
	gocv.CvtColor(frame, &t._hsvImg, gocv.ColorBGRToHSV)

	t._hsvImg.CopyTo(&t._img)

	for i := 0; i < len(t.gates); i++ {
		gocv.InRange(t._img, t.gates[i]._markerLowerBoundHSV, t.gates[i]._markerUpperBoundHSV, &t.gates[i]._markerMask)
		gocv.Merge([]gocv.Mat{t.gates[i]._markerMask, t.gates[i]._markerMask, t.gates[i]._markerMask}, &t.gates[i]._markerMask)

		if i == 0 {
			gocv.BitwiseOr(t.gates[i]._markerMask, t.gates[i]._markerMask, &t._markersMask)
		} else {
			gocv.BitwiseOr(t.gates[i]._markerMask, t._markersMask, &t._markersMask)
		}
	}

	gocv.BitwiseAnd(t._img, t._markersMask, &t._img)
	gocv.CvtColor(t._img, &t._grayImg, gocv.ColorBGRToGray)
	gocv.Threshold(t._grayImg, &t._binaryImg, 100, 255, gocv.ThresholdBinary)
	gocv.FillPoly(&t._binaryImg, t._leftPropPoly, color.RGBA{})
	gocv.FillPoly(&t._binaryImg, t._rightPropPoly, color.RGBA{})
	gocv.Erode(t._binaryImg, &t._binaryImg, t._kernel)
	gocv.Dilate(t._binaryImg, &t._binaryImg, t._kernel)

	contours := gocv.FindContours(t._binaryImg, gocv.RetrievalTree, gocv.ChainApproxSimple)
	totalArea := float64(0)
	largestContourIndex := -1
	largestContourArea := float64(-1)
	for i := 0; i < contours.Size(); i++ {
		contourArea := gocv.ContourArea(contours.At(i))
		if contourArea > largestContourArea {
			largestContourIndex = i
			largestContourArea = contourArea
		}

		totalArea += contourArea
		gocv.DrawContours(&frame, contours, i, color.RGBA{R: 255, G: 255, B: 255}, 5)
	}

	if largestContourIndex != -1 {
		center := gocv.MinAreaRect(contours.At(largestContourIndex)).Center
		pixel := t._hsvImg.GetVecbAt(center.Y, center.X)

		for i := 0; i < len(t.gates); i++ {
			if t.gates[i].IsSameHue(pixel) {
				t._lastSeenGate = t.gates[i]
			}
		}
	}

	if t._lastSeenGate == nil {
		//no point in looking for peaks
		return nil
	}

	if !t._buff.Push(totalArea, t._lastSeenGate.minActivationValue, t._lastSeenGate.minActivationFrames, t._lastSeenGate.minInactivationFrames) {
		//no peak
		return nil
	}

	//peak detected
	detection := Detection{
		Gate:        t._lastSeenGate,
		FrameOffset: t._frameCount,
	}

	if t._lastSeenGate.lastDetection == nil {
		return &detection
	}

	millisSinceLastDetection := detection.Diff(t._lastSeenGate.lastDetection) * int64(t._millisPerFrame)
	if int(millisSinceLastDetection) < detection.Gate.minMillisBetweenActivations {
		// ignore detection
		return nil
	}

	return &detection
}

func (t *Detector) AddGate(gate *Gate) {
	t.gates = append(t.gates, gate)
}

func (t *Detector) MillisPerFrame() int {
	return t._millisPerFrame
}
