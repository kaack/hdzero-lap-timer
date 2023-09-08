// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

type Timer struct {
	DetectionsInOrder    []*Detection
	DetectionsByGateName map[string][]*Detection
	GatesByPosition      map[int]*Gate
	GatesByName          map[string]*Gate
	Laps                 []*Lap
	Transitions          []*Transition
}

func NewTimer() *Timer {
	return &Timer{
		DetectionsInOrder:    []*Detection{},
		DetectionsByGateName: map[string][]*Detection{},
		GatesByPosition:      map[int]*Gate{},
		GatesByName:          map[string]*Gate{},
		Laps:                 []*Lap{},
		Transitions:          []*Transition{},
	}
}

func (t *Timer) AddDetection(detection *Detection) {

	lastDetection := t.LastDetection()

	if lastDetection != nil {
		//process laps
		if startGate := t.StartGate(); startGate != nil {
			if startGate.Name == detection.Gate.Name {
				if startGateDetections, ok := t.DetectionsByGateName[startGate.Name]; ok && len(startGateDetections) > 0 {
					t.Laps = append(t.Laps, NewLap(startGateDetections[len(startGateDetections)-1], detection))
				}
			}
		}

		//process transitions
		t.Transitions = append(t.Transitions, NewTransition(lastDetection, detection))
	}

	// add detection to overall list
	t.DetectionsInOrder = append(t.DetectionsInOrder, detection)

	//track detections by gate name
	if _, ok := t.DetectionsByGateName[detection.Gate.Name]; !ok {
		t.DetectionsByGateName[detection.Gate.Name] = []*Detection{}
	}

	t.DetectionsByGateName[detection.Gate.Name] = append(t.DetectionsByGateName[detection.Gate.Name], detection)
}

func (t *Timer) StartGate() *Gate {
	if startGate, ok := t.GatesByPosition[0]; ok {
		return startGate
	}
	return nil
}

func (t *Timer) LastDetection() *Detection {

	detectionsLength := len(t.DetectionsInOrder)
	if detectionsLength == 0 {
		return nil
	}
	return t.DetectionsInOrder[detectionsLength-1]

}
func (t *Timer) LastLap() *Lap {
	if len(t.Laps) == 0 {
		return nil
	}

	return t.Laps[len(t.Laps)-1]
}

func (t *Timer) LastTransition() *Transition {
	if len(t.Transitions) == 0 {
		return nil
	}

	return t.Transitions[len(t.Transitions)-1]
}

func (t *Timer) LastDetectedGate() *Gate {
	detection := t.LastDetection()
	if detection == nil {
		return nil
	}
	return detection.Gate
}

func (t *Timer) LapsCount() int {
	return len(t.Laps)
}

func (t *Timer) AddGate(index int, gate *Gate) {

	t.GatesByName[gate.Name] = gate
	t.GatesByPosition[index] = gate
}
