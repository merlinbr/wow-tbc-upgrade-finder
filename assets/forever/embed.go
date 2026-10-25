// Package foreverassets embeds the generated Forever catalog. Runtime code
// never fetches upstream.
package foreverassets

import "embed"

//go:embed catalog.json.gz manifest.json capabilities.json talents/*.json
var FS embed.FS
