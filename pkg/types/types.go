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
	Device  string `json:"device"`
	Circuit string `json:"circuit"`
	Topic   string `json:"topic"`
}

type Config struct {
	Mappings []struct{ mapping } `yaml:"mappings"`
}
