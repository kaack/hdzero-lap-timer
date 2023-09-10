// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

import (
	"errors"
)

type ChartBuffer struct {
	dataX    []float64
	dataY    []float64
	_headPos int
	_tailPos int
}

func NewChartBuffer(capacity int) *ChartBuffer {
	buffer := ChartBuffer{
		dataY:    make([]float64, capacity),
		dataX:    make([]float64, capacity),
		_headPos: -1,
		_tailPos: -1,
	}

	return &buffer
}
func (s *ChartBuffer) Push(x float64, y float64) {
	if s._tailPos == -1 && s._headPos == -1 {
		s._tailPos = 0
		s._headPos = 0
		s.dataX[s._headPos] = x
		s.dataY[s._headPos] = y
		return
	}

	s._headPos = (s._headPos + 1) % len(s.dataX)
	s.dataX[s._headPos] = x
	s.dataY[s._headPos] = y

	if s._tailPos == s._headPos {
		s._tailPos = (s._headPos + 1) % len(s.dataX)
	}
}

func (s *ChartBuffer) Len() int {
	if s._tailPos == -1 && s._headPos == -1 {
		return 0
	} else if s._tailPos == s._headPos {
		return 1
	} else if s._headPos > s._tailPos {
		return s._headPos - s._tailPos + 1
	} else {
		return (s._headPos + 1) + (len(s.dataX) - s._tailPos)
	}
}

func (s *ChartBuffer) Reset() {
	s._tailPos = -1
	s._headPos = -1
}

func (s *ChartBuffer) Data() (seriesX []float64, seriesY []float64) {
	if s._headPos == -1 && s._tailPos == -1 {
		return nil, nil
	} else if s._tailPos == s._headPos {
		return s.dataX[s._tailPos:1], s.dataY[s._tailPos:1]
	} else if s._headPos > s._tailPos {
		return s.dataX[s._tailPos : s._headPos+1], s.dataY[s._tailPos : s._headPos+1]
	} else if s._headPos < s._tailPos {
		xChunk1 := s.dataX[s._tailPos:len(s.dataX)]
		xChunk2 := s.dataX[0 : s._headPos+1]
		seriesX = make([]float64, len(xChunk1)+len(xChunk2))
		copy(seriesX[0:len(xChunk1)], xChunk1)
		copy(seriesX[len(xChunk1):], xChunk2)

		yChunk1 := s.dataY[s._tailPos:len(s.dataY)]
		yChunk2 := s.dataY[0 : s._headPos+1]
		seriesY = make([]float64, len(yChunk1)+len(yChunk2))
		copy(seriesY[0:len(yChunk1)], yChunk1)
		copy(seriesY[len(yChunk1):], yChunk2)

		return seriesX, seriesY
	}

	return nil, nil
}

func (s *ChartBuffer) At(index int) (x float64, y float64, err error) {
	if s._headPos == -1 && s._tailPos == -1 {
		return 0, 0, errors.New("index out of bounds")
	} else if s._tailPos == s._headPos {
		if index != 0 {
			return 0, 0, errors.New("index out of bounds")
		}
		return s.dataX[s._headPos], s.dataY[s._headPos], nil
	} else if s._headPos > s._tailPos {
		pos := s._headPos - index
		if pos < s._tailPos {
			return 0, 0, errors.New("index out of bounds")
		}
		return s.dataX[pos], s.dataY[pos], nil
	} else if s._headPos < s._tailPos {
		length := (s._headPos + 1) + (len(s.dataX) - s._tailPos)
		if index > length {
			return 0, 0, errors.New("index out of bounds")
		}

		if index <= s._headPos {
			pos := s._headPos - index
			return s.dataX[pos], s.dataY[pos], nil
		} else {
			pos := len(s.dataX) - (index - s._headPos)
			return s.dataX[pos], s.dataY[pos], nil
		}
	}
	return 0, 0, errors.New("index out of bounds")
}
