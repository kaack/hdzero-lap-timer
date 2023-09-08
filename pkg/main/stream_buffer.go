// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

import (
	"errors"
	"fmt"
)

type StreamBuffer struct {
	data                []float64
	_headPos            int
	_tailPos            int
	_activationValue    float64
	_activationFrames   int
	_inactivationFrames int
	_lastData           float64
}

func NewStreamBuffer(capacity int) *StreamBuffer {
	buffer := StreamBuffer{
		data:              make([]float64, capacity),
		_headPos:          -1,
		_tailPos:          -1,
		_activationValue:  0,
		_activationFrames: 0,
		_lastData:         0,
	}

	return &buffer
}
func (s *StreamBuffer) Push(data float64, minActivationValue float64, minActivationFrames int, minInactivationFrames int) (activation *Activation) {
	if s._tailPos == -1 && s._headPos == -1 {
		s._tailPos = 0
		s._headPos = 0
		s.data[s._headPos] = data
		return
	}

	s._headPos = (s._headPos + 1) % len(s.data)
	s.data[s._headPos] = data

	if data <= 0 {
		s._inactivationFrames += 1
	} else {
		s._inactivationFrames = 0
	}

	if data > 0 && data > s._lastData {
		s._activationFrames += 1
		s._activationValue += data
		s._lastData = data
	}

	if s._inactivationFrames > minInactivationFrames {
		if s._activationValue >= minActivationValue && s._activationFrames >= minActivationFrames {
			activation = &Activation{
				Frames: s._activationFrames,
				Value:  s._activationValue,
			}
			fmt.Printf("activation(value: %d, frames: %d), inactivation(frames: %d)\n", int(s._activationValue), s._activationFrames, int(s._inactivationFrames))
		}

		s._activationFrames = 0
		s._activationValue = 0
		s._lastData = 0
	}

	if s._tailPos == s._headPos {
		s._tailPos = (s._headPos + 1) % len(s.data)
	}

	return activation
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
