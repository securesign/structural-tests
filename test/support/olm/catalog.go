package olm

import (
	"encoding/json"
	"fmt"
	"io"
)

func ParseCatalogJSON(reader io.Reader) ([]interface{}, error) {
	var err error

	// Read the JSON file and decode multiple documents
	decoder := json.NewDecoder(reader)

	result := make([]interface{}, 0)

	for decoder.More() {
		var raw map[string]interface{}
		if err := decoder.Decode(&raw); err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}

		// Dispatch based on "schema" field
		switch raw["schema"] {
		case "olm.package":
			var pkg Package
			err = json.Unmarshal(rawJSON(raw), &pkg)
			if err != nil {
				return nil, fmt.Errorf("failed to parse OLM package: %w", err)
			}
			result = append(result, pkg)

		case "olm.channel":
			var channel Channel
			err = json.Unmarshal(rawJSON(raw), &channel)
			if err != nil {
				return nil, fmt.Errorf("failed to parse OLM channel: %w", err)
			}
			result = append(result, channel)

		case "olm.bundle":
			var bundle Bundle
			err = json.Unmarshal(rawJSON(raw), &bundle)
			if err != nil {
				return nil, fmt.Errorf("failed to parse OLM bundle: %w", err)
			}
			result = append(result, bundle)
		
		case "olm.deprecations":
			var deprecation Deprecation
			err = json.Unmarshal(rawJSON(raw), &deprecation)
			if err != nil {
				return nil, fmt.Errorf("failed to parse OLM deprecation: %w", err)
			}
			result = append(result, deprecation)

		default:
			return nil, fmt.Errorf("unknown schema: %v", raw["schema"])
		}
	}

	return result, nil
}

// Helper to convert map to JSON for unmarshalling.
func rawJSON(raw map[string]interface{}) []byte {
	rawData, err := json.Marshal(raw)
	if err != nil {
		panic(err)
	}
	return rawData
}
