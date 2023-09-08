// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

import (
	"errors"
	"fmt"
)

type StreamBuffer struct {
	data           []float64
	_headPos       int
	_tailPos       int
	_activeHeight  float64
	_activeWidth   int
	_inactiveWidth int
	_lastData      float64
}

func NewStreamBuffer(capacity int) *StreamBuffer {
	buffer := StreamBuffer{
		data:          make([]float64, capacity),
		_headPos:      -1,
		_tailPos:      -1,
		_activeHeight: 0,
		_activeWidth:  0,
		_lastData:     0,
	}

	return &buffer
}
func (s *StreamBuffer) Push(data float64, minActiveHeight float64, minActiveWidth int, minInactiveWidth int) (isPeak bool) {
	if s._tailPos == -1 && s._headPos == -1 {
		s._tailPos = 0
		s._headPos = 0
		s.data[s._headPos] = data
		return
	}

	s._headPos = (s._headPos + 1) % len(s.data)
	s.data[s._headPos] = data

	if data <= 0 {
		s._inactiveWidth += 1
	} else {
		s._inactiveWidth = 0
	}

	if data > 0 && data > s._lastData {
		s._activeWidth += 1
		s._activeHeight += data
		s._lastData = data
	}

	if s._inactiveWidth > minInactiveWidth {
		if s._activeHeight >= minActiveHeight && s._activeWidth >= minActiveWidth {
			isPeak = true
			fmt.Printf("peak: %v, aw: %d, ah: %d, iw: %d, value: %d\n", isPeak, s._activeWidth, int(s._activeHeight), s._inactiveWidth, int(data))
		}

		s._activeWidth = 0
		s._activeHeight = 0
		s._lastData = 0
	}

	if s._tailPos == s._headPos {
		s._tailPos = (s._headPos + 1) % len(s.data)
	}

	return isPeak
}

func (s *StreamBuffer) Len() int {
	if s._tailPos == -1 && s._headPos == -1 {
		return 0
	} else if s._tailPos == s._headPos {
		return 1
	} else if s._headPos > s._tailPos {
		return s._headPos - s._tailPos + 1
	} else {
		return (s._headPos + 1) + (len(s.data) - s._tailPos)
	}
}

func (s *StreamBuffer) Reset() {
	s._tailPos = -1
	s._headPos = -1
}

func (s *StreamBuffer) At(index int) (float64, error) {
	if s._headPos == -1 && s._tailPos == -1 {
		return 0, errors.New("index out of bounds")
	} else if s._tailPos == s._headPos {
		if index != 0 {
			return 0, errors.New("index out of bounds")
		}
		return s.data[s._headPos], nil
	} else if s._headPos > s._tailPos {
		pos := s._headPos - index
		if pos < s._tailPos {
			return 0, errors.New("index out of bounds")
		}
		return s.data[pos], nil
	} else if s._headPos < s._tailPos {
		length := (s._headPos + 1) + (len(s.data) - s._tailPos)
		if index > length {
			return 0, errors.New("index out of bounds")
		}

		if index <= s._headPos {
			return s.data[s._headPos-index], nil
		} else {
			return s.data[len(s.data)-(index-s._headPos)], nil
		}
	}
	return 0, errors.New("index out of bounds")
}
