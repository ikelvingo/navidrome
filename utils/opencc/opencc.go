package opencc

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"github.com/navidrome/navidrome/log"
	"github.com/yanmingcao/opencc-go"
)

var (
	s2tConverter *opencc.SimpleConverter // simplified-to-traditional converter
	t2sConverter *opencc.SimpleConverter // traditional-to-simplified converter
	initError    error
	initOnce     sync.Once
)

// initConverters lazily initializes OpenCC converters on first use
func initConverters() {
	initOnce.Do(func() {
		log.Debug("OpenCC initialization starting...")

		// Try multiple possible paths for OpenCC config files
		configPaths := []string{
			getOpenCCConfigPathFromDocker(),     // Docker-specific paths first
			getOpenCCConfigPathFromExecutable(), // then executable-relative paths
			getOpenCCConfigPathFromGOPATH(),
			getOpenCCConfigPathFromGoMod(),
			getOpenCCConfigPathFromModuleCache(),
			getOpenCCConfigPathFromEmbedded(),
		}

		log.Debug("OpenCC initialization started", "configPaths", configPaths, "count", len(configPaths))

		// Initialize converters
		for i, configPath := range configPaths {
			log.Debug("Trying config path", "index", i, "path", configPath)

			if configPath == "" {
				log.Debug("OpenCC config path is empty, skipping")
				continue
			}

			// Check if this is a real path or an embedded virtual path
			if configPath == "embedded" {
				log.Debug("Trying embedded config path")
				// Try using embedded configuration files
				if initEmbeddedConverters() {
					log.Debug("OpenCC converters initialized successfully from embedded resources")
					return
				}
				continue
			}

			s2tPath := filepath.Join(configPath, "s2t.json")
			t2sPath := filepath.Join(configPath, "t2s.json")

			log.Debug("Checking OpenCC config files", "path", configPath, "s2tPath", s2tPath, "t2sPath", t2sPath)

			// Verify config files exist
			if _, err := os.Stat(s2tPath); err != nil {
				log.Debug("OpenCC s2t.json not found", "path", s2tPath, "error", err)
				continue
			}
			if _, err := os.Stat(t2sPath); err != nil {
				log.Debug("OpenCC t2s.json not found", "path", t2sPath, "error", err)
				continue
			}

			log.Debug("OpenCC config files found, creating converters", "configPath", configPath)

			s2tConverter, initError = opencc.NewSimpleConverter(s2tPath)
			if initError != nil {
				log.Warn("Failed to create s2t converter", "path", s2tPath, "error", initError)
				continue
			}

			t2sConverter, initError = opencc.NewSimpleConverter(t2sPath)
			if initError != nil {
				log.Warn("Failed to create t2s converter", "path", t2sPath, "error", initError)
				s2tConverter = nil
				continue
			}

			log.Debug("OpenCC converters initialized successfully", "configPath", configPath)
			return
		}

		// Log a warning if all paths failed, but don't panic -- conversion degrades gracefully
		log.Warn("OpenCC converters failed to initialize. Chinese simplified/traditional conversion will be disabled.")
		log.Debug("OpenCC initialization completed with failure")
	})
}

// getOpenCCConfigPathFromGOPATH looks for OpenCC config files in the GOPATH module cache
func getOpenCCConfigPathFromGOPATH() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Debug("OpenCC: Cannot get user home directory", "error", err)
			return ""
		}
		gopath = filepath.Join(home, "go")
	}

	log.Debug("OpenCC: GOPATH", "gopath", gopath)

	// Try multiple possible path patterns
	possiblePaths := []string{
		filepath.Join(gopath, "pkg", "mod", "github.com", "yanmingcao", "opencc-go@v1.0.0", "data", "config"),
		filepath.Join(gopath, "pkg", "mod", "github.com", "yanmingcao", "opencc-go@*", "data", "config"),
	}

	for _, configPath := range possiblePaths {
		log.Debug("OpenCC: Trying GOPATH path", "path", configPath)

		// Expand wildcards in path
		if strings.Contains(configPath, "*") {
			matches, err := filepath.Glob(configPath)
			if err == nil && len(matches) > 0 {
				// Use the first matching directory
				configPath = matches[0]
				log.Debug("OpenCC: Found matching path via glob", "original", configPath, "matched", configPath)
			}
		}

		s2tPath := filepath.Join(configPath, "s2t.json")
		if _, err := os.Stat(s2tPath); err == nil {
			log.Debug("OpenCC: Found config files in GOPATH", "path", configPath)
			return configPath
		} else {
			log.Debug("OpenCC: s2t.json not found in GOPATH", "path", s2tPath, "error", err)
		}
	}

	log.Debug("OpenCC: No config files found in GOPATH")
	return ""
}

// getOpenCCConfigPathFromGoMod looks for OpenCC config files in the go.mod vendor directory
func getOpenCCConfigPathFromGoMod() string {
	// Start from current directory and look upward for go.mod
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Check project-local vendor or cache paths
	configPath := filepath.Join(dir, "vendor", "github.com", "yanmingcao", "opencc-go", "data", "config")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	return ""
}

// getOpenCCConfigPathFromModuleCache looks for OpenCC config files in the Go module cache
func getOpenCCConfigPathFromModuleCache() string {
	// Try GOMODCACHE environment variable first (may be set in production)
	if gomodcache := os.Getenv("GOMODCACHE"); gomodcache != "" {
		pattern := filepath.Join(gomodcache, "github.com", "yanmingcao", "opencc-go@*", "data", "config")
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return matches[0]
		}
	}

	// Try using go env GOMODCACHE (development environment)
	cmd := exec.Command("go", "env", "GOMODCACHE")
	output, err := cmd.Output()
	if err == nil {
		gomodcache := strings.TrimSpace(string(output))
		if gomodcache != "" {
			pattern := filepath.Join(gomodcache, "github.com", "yanmingcao", "opencc-go@*", "data", "config")
			matches, err := filepath.Glob(pattern)
			if err == nil && len(matches) > 0 {
				return matches[0]
			}
		}
	}

	// Fall back to the default module cache path
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Go 1.15+ default module cache path
	defaultCache := filepath.Join(home, "go", "pkg", "mod")
	pattern := filepath.Join(defaultCache, "github.com", "yanmingcao", "opencc-go@*", "data", "config")
	matches, err := filepath.Glob(pattern)
	if err == nil && len(matches) > 0 {
		return matches[0]
	}

	return ""
}

// getOpenCCConfigPathFromExecutable looks for OpenCC config files relative to the executable path
func getOpenCCConfigPathFromExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Debug("OpenCC: Cannot get executable path", "error", err)
		return ""
	}

	exeDir := filepath.Dir(exePath)
	log.Debug("OpenCC: Executable directory", "exeDir", exeDir, "exePath", exePath)

	// Try searching for config files near the executable directory
	possiblePaths := []string{
		filepath.Join(exeDir, "opencc", "config"),
		filepath.Join(exeDir, "config", "opencc"),
		filepath.Join(exeDir, "..", "share", "opencc"),
		// Common Docker paths
		filepath.Join(exeDir, "..", "opencc", "config"),
		filepath.Join("/usr", "share", "opencc"),
		filepath.Join("/usr", "local", "share", "opencc"),
	}

	log.Debug("OpenCC: Checking executable-based paths", "paths", possiblePaths)

	for _, configPath := range possiblePaths {
		s2tPath := filepath.Join(configPath, "s2t.json")
		log.Debug("OpenCC: Checking path", "configPath", configPath, "s2tPath", s2tPath)

		if _, err := os.Stat(s2tPath); err == nil {
			log.Debug("OpenCC: Found config files via executable path", "path", configPath)
			return configPath
		} else {
			log.Debug("OpenCC: s2t.json not found in executable path", "path", s2tPath, "error", err)
		}
	}

	log.Debug("OpenCC: No config files found via executable path")
	return ""
}

// getOpenCCConfigPathFromDocker looks for OpenCC config files in Docker container paths
func getOpenCCConfigPathFromDocker() string {
	// Common paths found in Docker containers
	dockerPaths := []string{
		// Standard system paths
		"/usr/share/opencc",
		"/usr/local/share/opencc",
		// Alpine Linux paths
		"/usr/lib/opencc",
		"/usr/lib64/opencc",
		// Go module locations (if installed via apk)
		"/usr/lib/go/pkg/mod/github.com/yanmingcao/opencc-go@v1.0.0/data/config",
		"/usr/local/lib/go/pkg/mod/github.com/yanmingcao/opencc-go@v1.0.0/data/config",
		// Executable-relative paths
		"/app/opencc/config",
		"/app/config/opencc",
		// Home directory paths
		"/root/.local/share/opencc",
		// Other common locations
		"/opt/opencc/config",
		"/var/lib/opencc",
	}

	log.Debug("OpenCC: Checking Docker-specific paths", "paths", dockerPaths)

	for _, configPath := range dockerPaths {
		s2tPath := filepath.Join(configPath, "s2t.json")
		log.Debug("OpenCC: Checking Docker path", "configPath", configPath, "s2tPath", s2tPath)

		if _, err := os.Stat(s2tPath); err == nil {
			log.Debug("OpenCC: Found config files in Docker path", "path", configPath)
			return configPath
		} else {
			log.Debug("OpenCC: s2t.json not found in Docker path", "path", s2tPath, "error", err)
		}
	}

	log.Debug("OpenCC: No config files found in Docker paths")
	return ""
}

// getOpenCCConfigPathFromEmbedded returns a sentinel value indicating embedded config should be used
func getOpenCCConfigPathFromEmbedded() string {
	return "embedded"
}

// initEmbeddedConverters tries to initialize converters using embedded configuration files
func initEmbeddedConverters() bool {
	// Embedded config files are not yet implemented
	return false
}

// ContainsChinese returns true if the string contains any Chinese (Han) characters
func ContainsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// ConvertBoth converts the input text into both simplified and traditional Chinese forms.
// Returns (simplified, traditional). If no Chinese characters are detected, returns original text.
func ConvertBoth(text string) (string, string) {
	if !ContainsChinese(text) {
		return text, text
	}

	// Lazily initialize converters
	initConverters()

	// Return original text if converters are not available
	if s2tConverter == nil || t2sConverter == nil {
		log.Debug("OpenCC converters are nil, returning original text", "text", text)
		return text, text
	}

	// Convert simplified to traditional
	traditional := s2tConverter.Convert(text)

	// Convert traditional to simplified
	simplified := t2sConverter.Convert(text)

	log.Debug("OpenCC conversion result", "original", text, "simplified", simplified, "traditional", traditional)
	return simplified, traditional
}

// GetSearchQueries returns a deduplicated list of search query variants.
// If the input contains Chinese characters, returns simplified and traditional variants.
// If no Chinese characters, returns a single-element slice with the original text.
func GetSearchQueries(text string) []string {
	if !ContainsChinese(text) {
		return []string{text}
	}

	simplified, traditional := ConvertBoth(text)

	// Deduplicate variants
	queries := []string{text}
	if simplified != text {
		queries = append(queries, simplified)
	}
	if traditional != text && traditional != simplified {
		queries = append(queries, traditional)
	}

	// Ensure both simplified and traditional forms are present
	hasSimplified := false
	hasTraditional := false
	for _, q := range queries {
		if q == simplified {
			hasSimplified = true
		}
		if q == traditional {
			hasTraditional = true
		}
	}
	if !hasSimplified {
		queries = append(queries, simplified)
	}
	if !hasTraditional && traditional != simplified {
		queries = append(queries, traditional)
	}

	return queries
}

// NormalizeQuery trims whitespace and trailing wildcards from a query string
func NormalizeQuery(q string) string {
	q = strings.TrimSpace(q)
	q = strings.TrimSuffix(q, "*")
	return q
}
