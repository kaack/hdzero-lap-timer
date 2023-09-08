// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

type Lap struct {
	start *Detection
	stop  *Detection
}

func NewLap(start *Detection, stop *Detection) *Lap {
	return &Lap{
		start: start,
		stop:  stop,
	}
}

func (l *Lap) Frames() int {
	return int(l.stop.FrameOffset - l.start.FrameOffset)
}

func (l *Lap) Gate() *Gate {
	return l.start.Gate
}
