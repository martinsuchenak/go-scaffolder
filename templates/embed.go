package templates

import "embed"

//go:embed all:base all:cmd all:api all:mcp all:ui all:db all:cache all:resolve all:docker all:nomad all:tests
var FS embed.FS
