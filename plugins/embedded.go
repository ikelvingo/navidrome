package plugins

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed builtin/netease.ndp
var neteasePluginData []byte

var builtinPlugins = []struct {
	Name string
	Data []byte
}{
	{Name: "netease", Data: neteasePluginData},
}

// extractBuiltinPlugins copies embedded plugins to the plugins folder.
// Returns the list of plugin names that were newly extracted.
func extractBuiltinPlugins(folder string) []string {
	var extracted []string
	for _, bp := range builtinPlugins {
		target := filepath.Join(folder, bp.Name+PackageExtension)
		if _, err := os.Stat(target); err == nil {
			continue
		}
		if err := os.WriteFile(target, bp.Data, 0644); err != nil {
			fmt.Printf("Failed to extract builtin plugin %s: %v\n", bp.Name, err)
			continue
		}
		extracted = append(extracted, bp.Name)
	}
	return extracted
}
