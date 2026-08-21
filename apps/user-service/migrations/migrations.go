// Package migrations embeds this service's forward-only SQL migration files
// so they ship inside the compiled binary instead of depending on a file
// path at deploy time.
package migrations

import "embed"

//go:embed *.sql
var Files embed.FS
