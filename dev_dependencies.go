//go:build dev
// +build dev

package main

import (
	_ "github.com/fergusstrange/embedded-postgres"
	_ "github.com/peterldowns/pgtestdb"
)
