// Package db exposes the SQL migrations as an embedded filesystem so the
// migrate binary carries them and needs no files at runtime.
package db

import "embed"

// Migrations holds every forward migration, applied in filename order.
//
//go:embed migrations/*.sql
var Migrations embed.FS
