package embed

import (
	_ "embed"
	"fmt"
)

//go:embed embedded-roots/1.root-dev.json
var DevRoot []byte

//go:embed embedded-roots/1.root-staging.json
var StagingRoot []byte

//go:embed embedded-roots/1.root.json
var ProdRoot []byte

var DefaultRoot = ProdRoot

func GetRootBytes(root string) ([]byte, error) {
	switch root {
	case "dev":
		return DevRoot, nil
	case "staging":
		return StagingRoot, nil
	case "prod":
		return ProdRoot, nil
	case "":
		return DefaultRoot, nil
	default:
		return nil, fmt.Errorf("invalid tuf root: %s", root)
	}
}
