// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

type Transition struct {
	start *Detection
	stop  *Detection
}

func NewTransition(start *Detection, stop *Detection) *Transition {
	return &Transition{
		start: start,
		stop:  stop,
	}
}

func (t *Transition) Frames() int {
	return int(t.stop.FrameOffset - t.start.FrameOffset)
}

func (t *Transition) Gate() *Gate {
	return t.start.Gate
}
