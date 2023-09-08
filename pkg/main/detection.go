// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

import "fmt"

type Detection struct {
	Gate        *Gate
	FrameOffset uint64
}

func (d *Detection) Diff(detection *Detection) int64 {
	return int64(d.FrameOffset - detection.FrameOffset)
}

func (d *Detection) String() string {
	return fmt.Sprintf("gate: %s, frame: %v", d.Gate.Name, d.FrameOffset)
}
