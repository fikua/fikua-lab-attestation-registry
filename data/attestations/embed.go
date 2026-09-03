// Package attestations embeds the bundled attestation definition JSON files
// so they ship inside the compiled binary.
package attestations

import "embed"

//go:embed *.json
var FS embed.FS
