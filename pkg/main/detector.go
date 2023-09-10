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

	_millisPerFrame float64

	_lastSeenGate *Gate

	_markersMask   gocv.Mat
	_img           gocv.Mat
	_hsvImg        gocv.Mat
	_grayImg       gocv.Mat
	_binaryImg     gocv.Mat
	_nonZeroPixels gocv.Mat

	_pixelsByGate map[*Gate]int

	_leftPropPoly  gocv.PointsVector
	_rightPropPoly gocv.PointsVector
	_kernel        gocv.Mat
}

func NewDetector(img gocv.Mat,
	config *Config) Detector {

	detector := Detector{
		_millisPerFrame: 1000 / config.FramesPerSec,
		_lastSeenGate:   nil,

		_markersMask: gocv.NewMat(),
		_hsvImg:      gocv.NewMat(),
		_img:         gocv.NewMat(),
		_grayImg:     gocv.NewMat(),
		_binaryImg:   gocv.NewMat(),

		_nonZeroPixels: gocv.NewMat(),

		_pixelsByGate: map[*Gate]int{},

		_leftPropPoly:  gocv.NewPointsVectorFromPoints([][]image.Point{{image.Pt(0, img.Rows()), image.Pt(0, img.Rows()-config.PropellerMask.Height), image.Pt(config.PropellerMask.Width, img.Rows())}}),
		_rightPropPoly: gocv.NewPointsVectorFromPoints([][]image.Point{{image.Pt(img.Cols(), img.Rows()), image.Pt(img.Cols(), img.Rows()-config.PropellerMask.Height), image.Pt(img.Cols()-config.PropellerMask.Width, img.Rows())}}),

		_kernel: gocv.GetStructuringElement(gocv.MorphRect, image.Pt(config.Detection.Erode, config.Detection.Erode)),
	}

	return detector
}

//goland:noinspection GoUnusedParameter
func (t *Detector) Detect(img *gocv.Mat, detectionMat *gocv.Mat, frameCount uint64) (*Gate, *Detection) {

	// convert the image to HSV format so that we can easily isolate the markers by color ranges (mainly Hue)
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

	// convert the color-isolated image to grayscale, and apply threshold so that we can end up with binary image
	gocv.CvtColor(t._img, &t._grayImg, gocv.ColorBGRToGray)
	gocv.Threshold(t._grayImg, &t._binaryImg, 100, 255, gocv.ThresholdBinary)

	// draw black triangles on the bottom left, and bottom right of the image to hide the props (if they are in view)
	// this is necessary because some props have same color as that of markers
	gocv.FillPoly(&t._binaryImg, t._leftPropPoly, color.RGBA{})
	gocv.FillPoly(&t._binaryImg, t._rightPropPoly, color.RGBA{})

	//erode and dilate the binary image to remove any last bits of noise
	gocv.Erode(t._binaryImg, &t._binaryImg, t._kernel)
	gocv.Dilate(t._binaryImg, &t._binaryImg, t._kernel)

	t._binaryImg.CopyTo(detectionMat)

	// Check if there is any gate marker in view, and identify which one (by color)
	//
	// First we just count how many white (non-zero) pixels the binary image  has
	// this tells us the total area for all gate markers visible, but we don't know which one
	// is closes to the drone.
	//
	// To identify which gate marker is most prominent, we sample
	// the pixel colors form the original HSV image and determine which gate marker has the most
	// pixels in view.
	//
	// We do not go through all pixels of the original image, instead we go
	// through a subset of the pixel locations found earlier with the binary image

	gocv.FindNonZero(t._binaryImg, &t._nonZeroPixels)
	var detectedGate *Gate
	var detectedArea int
	totalArea := t._nonZeroPixels.Total()
	if totalArea > 0 {
		// initialize histogram to zeroes
		for i := 0; i < len(t.gates); i++ {
			t._pixelsByGate[t.gates[i]] = 0
		}

		// sample every 10 pixels
		for i := 0; i < totalArea; i += 1 {
			location := t._nonZeroPixels.GetVeciAt(0, i)

			pixel := t._hsvImg.GetVecbAt(int(location[1]), int(location[0]))
			for i := 0; i < len(t.gates); i++ {
				if t.gates[i].IsSameHue(pixel) {
					t._pixelsByGate[t.gates[i]] += 1
				}
			}
		}

		detectedArea = 0
		for gate, pixelCount := range t._pixelsByGate {
			if pixelCount > detectedArea {
				detectedArea = pixelCount
				detectedGate = gate
			}
		}

		if float64(detectedArea) > float64(detectedGate._lastArea)*1.1 {
			detectedGate._activeFrames += 1
			detectedGate._activeValue += detectedArea
			detectedGate._lastArea += detectedArea
			detectedGate._inactiveFrames = 0
		}

		detectedGate._activeArea = detectedArea
		detectedGate._lastArea = detectedArea

		//fmt.Printf("*** Active: %s, area: %v, value: %v\n", detectedGate.Name, detectedArea, detectedGate._activeValue)

	}

	var peakDetected *Detection
	for _, gate := range t.gates {
		if gate != detectedGate {
			gate._inactiveFrames += 1
		}

		//fmt.Println(gate.State())

		if gate._inactiveFrames >= gate.Config.Detection.MinInactiveFrames {

			if peakDetected == nil {
				if gate._activeFrames >= gate.Config.Detection.MinActiveFrames &&
					float64(gate._activeValue) >= gate.Config.Detection.MinActiveValue {
					//(gate.LastDetection == nil ||
					//	(t._frameCount-gate.LastDetection.FrameOffset) > uint64(gate.Config.Detection.MinFramesBetweenPeaks))

					peakDetected = &Detection{
						FrameOffset: frameCount,
						Activation: &Activation{
							Gate:   gate,
							Frames: gate._activeFrames,
							Value:  float64(gate._activeValue),
						},
					}

					gate.LastDetection = peakDetected
				} else if gate._activeFrames > 0 {
					//fmt.Printf("*** False Peak: %s, value: %v, frames: %v\n\n\n\n", gate.Name, gate._activeValue, gate._activeFrames)
				}
			}

			gate._activeFrames = 0
			gate._activeValue = 0
			gate._lastArea = 0
		}
	}

	return detectedGate, peakDetected

}

func (t *Detector) AddGate(gate *Gate) {
	t.gates = append(t.gates, gate)
}

func (t *Detector) MillisPerFrame() float64 {
	return t._millisPerFrame
}
