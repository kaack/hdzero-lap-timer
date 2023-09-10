// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

import (
	"errors"
)

type SumBuffer struct {
	_values []float64
	Sum     float64
	Avg     float64

	_headPos int
	_tailPos int
}

func NewSumBuffer(capacity int) *SumBuffer {
	buffer := SumBuffer{
		_values:  make([]float64, capacity),
		Sum:      0,
		_headPos: -1,
		_tailPos: -1,
	}

	return &buffer
}
func (s *SumBuffer) Push(x float64) {
	if s._tailPos == -1 && s._headPos == -1 {
		s._tailPos = 0
		s._headPos = 0
		s._values[s._headPos] = x
		s.Sum += x
		return
	}

	removed := float64(0)
	s._headPos = (s._headPos + 1) % len(s._values)
	if s._tailPos == s._headPos {
		s._tailPos = (s._headPos + 1) % len(s._values)
		removed = s._values[s._tailPos]
	}

	s._values[s._headPos] = x
	s.Sum += x - removed
	s.Avg = s.Sum / float64(s.Len())

}

func (s *SumBuffer) Len() int {
	if s._tailPos == -1 && s._headPos == -1 {
		return 0
	} else if s._tailPos == s._headPos {
		return 1
	} else if s._headPos > s._tailPos {
		return s._headPos - s._tailPos + 1
	} else {
		return (s._headPos + 1) + (len(s._values) - s._tailPos)
	}
}

func (s *SumBuffer) Reset() {
	s._tailPos = -1
	s._headPos = -1
}

func (s *SumBuffer) Data() (seriesX []float64) {
	if s._headPos == -1 && s._tailPos == -1 {
		return nil
	} else if s._tailPos == s._headPos {
		return s._values[s._tailPos:1]
	} else if s._headPos > s._tailPos {
		return s._values[s._tailPos : s._headPos+1]
	} else if s._headPos < s._tailPos {
		xChunk1 := s._values[s._tailPos:len(s._values)]
		xChunk2 := s._values[0 : s._headPos+1]
		seriesX = make([]float64, len(xChunk1)+len(xChunk2))
		copy(seriesX[0:len(xChunk1)], xChunk1)
		copy(seriesX[len(xChunk1):], xChunk2)

		return seriesX
	}

	return nil
}

func (s *SumBuffer) Peak() float64 {
	if s.Len() < 3 {
		return 0
	}

	a, _ := s.At(2)
	b, _ := s.At(1)
	c, _ := s.At(0)

	if a < b && b > c {
		return b
	}

	return 0
}
func (s *SumBuffer) At(index int) (x float64, err error) {
	if s._headPos == -1 && s._tailPos == -1 {
		return 0, errors.New("index out of bounds")
	} else if s._tailPos == s._headPos {
		if index != 0 {
			return 0, errors.New("index out of bounds")
		}
		return s._values[s._headPos], nil
	} else if s._headPos > s._tailPos {
		pos := s._headPos - index
		if pos < s._tailPos {
			return 0, errors.New("index out of bounds")
		}
		return s._values[pos], nil
	} else if s._headPos < s._tailPos {
		length := (s._headPos + 1) + (len(s._values) - s._tailPos)
		if index > length {
			return 0, errors.New("index out of bounds")
		}

		if index <= s._headPos {
			pos := s._headPos - index
			return s._values[pos], nil
		} else {
			pos := len(s._values) - (index - s._headPos)
			return s._values[pos], nil
		}
	}
	return 0, errors.New("index out of bounds")
}

func (s *SumBuffer) First() (x float64) {
	first, _ := s.At(0)
	return first
}
