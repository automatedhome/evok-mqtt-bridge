package types

import (
	"encoding/json"
)

type Message struct {
	Command string      `json:"cmd,omitempty"`
	Circuit string      `json:"circuit"`
	Device  string      `json:"dev"`
	Value   json.Number `json:"value"`
}

type mapping struct {
	Device  string  `json:"device"`
	Circuit string  `json:"circuit"`
	Topic   string  `json:"topic"`
	Offset  float64 `json:"offset,omitempty"`
}

type Config struct {
	Interval int       `yaml:"sync_interval"`
	Mappings []mapping `yaml:"mappings"`
}

type GPIOStates struct {
	Status string `json:"status"`
	Data   []struct {
		Value   float64 `json:"value"`
		Circuit string  `json:"circuit"`
		Dev     string  `json:"dev"`
	} `json:"data"`
}
