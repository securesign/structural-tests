package olm

// Package schema.
type Package struct {
	Schema         string `json:"schema"`
	Name           string `json:"name"`
	DefaultChannel string `json:"defaultChannel"`
}

// Channel schema.
type Channel struct {
	Schema  string         `json:"schema"`
	Name    string         `json:"name"`
	Package string         `json:"package"`
	Entries []ChannelEntry `json:"entries"`
}

type ChannelEntry struct {
	Name     string   `json:"name"`
	Replaces string   `json:"replaces,omitempty"`
	Skips    []string `json:"skips,omitempty"`
}

// Bundle schema.
type Bundle struct {
	Schema     string     `json:"schema"`
	Name       string     `json:"name"`
	Package    string     `json:"package"`
	Image      string     `json:"image"`
	Properties []Property `json:"properties"`
}

type Property struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}
