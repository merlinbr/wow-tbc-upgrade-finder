package enchants

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed descriptions.json
var descriptionsJSON []byte

// Descriptions maps enchant effect IDs to the effect text shown on the
// wowsims item display (e.g. "2673": "Mongoose").
func Descriptions() map[int32]string {
	result := make(map[int32]string)
	if err := json.Unmarshal(descriptionsJSON, &result); err != nil {
		panic(fmt.Errorf("unmarshal enchant descriptions: %w", err))
	}
	return result
}
