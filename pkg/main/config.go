// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type PropellerMaskConfig struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type GateDetectionConfig struct {
	MinMillisBetweenActivations int     `json:"minMillisBetweenActivations"`
	MinActivationValue          float64 `json:"minActivationValue"`
	MinActivationFrames         int     `json:"minActivationFrames"`
	MinInactivationFrames       int     `json:"minInactivationFrames"`
}

type GateColorConfig struct {
	LowerBoundHSV []int `json:"lowerBoundHSV"`
	UpperBoundHSV []int `json:"upperBoundHSV"`
}

type GateConfig struct {
	Name      string              `json:"name"`
	Detection GateDetectionConfig `json:"detection"`
	Color     GateColorConfig     `json:"color"`
}

type Config struct {
	FramesPerSec  int                 `json:"framesPerSec"`
	PropellerMask PropellerMaskConfig `json:"propellerMask"`
	Gates         []GateConfig        `json:"gates"`
}

func YAMLtoJSON(r io.Reader) ([]byte, error) {
	var v interface{}
	var err error
	var res []byte

	dec := yaml.NewDecoder(r)
	if err = dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("file is not valid yaml. %s", err)
	}

	if res, err = json.Marshal(v); err != nil {
		return nil, fmt.Errorf("file could not be converted to JSON. %s", err)
	}

	return res, nil
}

func NewConfig(path string) (*Config, error) {

	var err error
	var configYaml []byte

	if configYaml, err = os.ReadFile(path); err != nil {
		return nil, fmt.Errorf("could not read config file. %s", err.Error())
	}

	var configJson []byte
	if configJson, err = YAMLtoJSON(bytes.NewReader(configYaml)); err != nil {
		return nil, fmt.Errorf("could not parse config file. %s", err.Error())
	}

	config := Config{}
	if err = json.Unmarshal(configJson, &config); err != nil {
		return nil, fmt.Errorf("could not parse config file. %s", err.Error())
	}

	return &config, nil
}
