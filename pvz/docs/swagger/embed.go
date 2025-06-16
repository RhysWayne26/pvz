package swagger

import "embed"

// FS contains embedded Swagger UI static files for serving via HTTP.
//
//go:embed *
var FS embed.FS
