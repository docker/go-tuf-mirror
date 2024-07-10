package embed

import (
	_ "embed"
)

//go:embed embedded-roots/1.root-dev.json
var DevRoot []byte

//go:embed embedded-roots/1.root-staging.json
var StagingRoot []byte

//go:embed embedded-roots/1.root.json
var ProdRoot []byte

var DefaultRoot = ProdRoot
