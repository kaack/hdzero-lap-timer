// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

import "fmt"

type Detection struct {
	FrameOffset uint64
	Activation  *Activation
}

func (d *Detection) Diff(detection *Detection) int64 {
	return int64(d.FrameOffset - detection.FrameOffset)
}

func (d *Detection) String() string {
	return fmt.Sprintf("detection(gate: %s, frame: %v) activation(value: %d, frames: %d)", d.Activation.Gate.Name, d.FrameOffset, int(d.Activation.Value), d.Activation.Frames)
}
