package templates

import "embed"

//go:embed all:base all:cmd all:api all:mcp all:ui all:db all:cache all:docker all:nomad all:tests all:add all:ci all:otel
var FS embed.FS
