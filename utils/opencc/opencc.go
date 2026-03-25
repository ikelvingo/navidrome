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
	s2tConverter *opencc.SimpleConverter // 简体转繁体
	t2sConverter *opencc.SimpleConverter // 繁体转简体
	initError    error
	initOnce     sync.Once
)

// initConverters 延迟初始化转换器
func initConverters() {
	initOnce.Do(func() {
		log.Debug("OpenCC initialization starting...")

		// 尝试多个可能的路径
		configPaths := []string{
			getOpenCCConfigPathFromDocker(),     // 首先尝试Docker专用路径
			getOpenCCConfigPathFromExecutable(), // 然后尝试可执行文件路径
			getOpenCCConfigPathFromGOPATH(),
			getOpenCCConfigPathFromGoMod(),
			getOpenCCConfigPathFromModuleCache(),
			getOpenCCConfigPathFromEmbedded(),
		}

		log.Debug("OpenCC initialization started", "configPaths", configPaths, "count", len(configPaths))

		// 初始化转换器
		for i, configPath := range configPaths {
			log.Debug("Trying config path", "index", i, "path", configPath)

			if configPath == "" {
				log.Debug("OpenCC config path is empty, skipping")
				continue
			}

			// 检查是真实路径还是嵌入的虚拟路径
			if configPath == "embedded" {
				log.Debug("Trying embedded config path")
				// 尝试使用嵌入的配置文件
				if initEmbeddedConverters() {
					log.Debug("OpenCC converters initialized successfully from embedded resources")
					return
				}
				continue
			}

			s2tPath := filepath.Join(configPath, "s2t.json")
			t2sPath := filepath.Join(configPath, "t2s.json")

			log.Debug("Checking OpenCC config files", "path", configPath, "s2tPath", s2tPath, "t2sPath", t2sPath)

			// 检查配置文件是否存在
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

		// 如果所有路径都失败，记录错误但不panic
		log.Warn("OpenCC converters failed to initialize. Chinese simplified/traditional conversion will be disabled.")
		log.Debug("OpenCC initialization completed with failure")
	})
}

// getOpenCCConfigPathFromGOPATH 从GOPATH获取opencc-go库的配置文件路径
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

	// 尝试多个可能的路径模式
	possiblePaths := []string{
		filepath.Join(gopath, "pkg", "mod", "github.com", "yanmingcao", "opencc-go@v1.0.0", "data", "config"),
		filepath.Join(gopath, "pkg", "mod", "github.com", "yanmingcao", "opencc-go@*", "data", "config"),
	}

	for _, configPath := range possiblePaths {
		log.Debug("OpenCC: Trying GOPATH path", "path", configPath)

		// 对于通配符路径，需要查找匹配的目录
		if strings.Contains(configPath, "*") {
			matches, err := filepath.Glob(configPath)
			if err == nil && len(matches) > 0 {
				// 使用第一个匹配的目录
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

// getOpenCCConfigPathFromGoMod 尝试从当前项目的go.mod缓存获取
func getOpenCCConfigPathFromGoMod() string {
	// 尝试从当前目录向上查找go.mod
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// 检查项目本地的vendor或缓存路径
	configPath := filepath.Join(dir, "vendor", "github.com", "yanmingcao", "opencc-go", "data", "config")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	return ""
}

// getOpenCCConfigPathFromModuleCache 从Go模块缓存获取配置文件路径
func getOpenCCConfigPathFromModuleCache() string {
	// 在生产环境（Docker）中，可能没有go命令，所以先尝试环境变量
	if gomodcache := os.Getenv("GOMODCACHE"); gomodcache != "" {
		pattern := filepath.Join(gomodcache, "github.com", "yanmingcao", "opencc-go@*", "data", "config")
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return matches[0]
		}
	}

	// 尝试使用go env GOMODCACHE（开发环境）
	cmd := exec.Command("go", "env", "GOMODCACHE")
	output, err := cmd.Output()
	if err == nil {
		gomodcache := strings.TrimSpace(string(output))
		if gomodcache != "" {
			// 尝试查找opencc-go的配置文件
			pattern := filepath.Join(gomodcache, "github.com", "yanmingcao", "opencc-go@*", "data", "config")
			matches, err := filepath.Glob(pattern)
			if err == nil && len(matches) > 0 {
				return matches[0]
			}
		}
	}

	// 如果GOMODCACHE失败，尝试使用默认的模块缓存路径
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Go 1.15+ 默认模块缓存路径
	defaultCache := filepath.Join(home, "go", "pkg", "mod")
	pattern := filepath.Join(defaultCache, "github.com", "yanmingcao", "opencc-go@*", "data", "config")
	matches, err := filepath.Glob(pattern)
	if err == nil && len(matches) > 0 {
		return matches[0]
	}

	return ""
}

// getOpenCCConfigPathFromExecutable 从可执行文件所在目录获取配置文件路径
func getOpenCCConfigPathFromExecutable() string {
	// 获取可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		log.Debug("OpenCC: Cannot get executable path", "error", err)
		return ""
	}

	exeDir := filepath.Dir(exePath)
	log.Debug("OpenCC: Executable directory", "exeDir", exeDir, "exePath", exePath)

	// 尝试在可执行文件目录下查找配置文件
	possiblePaths := []string{
		filepath.Join(exeDir, "opencc", "config"),
		filepath.Join(exeDir, "config", "opencc"),
		filepath.Join(exeDir, "..", "share", "opencc"),
		// Docker容器中的常见路径
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

// getOpenCCConfigPathFromDocker 专门为Docker容器设计的配置文件路径
func getOpenCCConfigPathFromDocker() string {
	// Docker容器中的可能路径
	// 参考其他Go库和系统库的安装位置
	dockerPaths := []string{
		// 标准系统路径
		"/usr/share/opencc",
		"/usr/local/share/opencc",
		// Alpine Linux中的常见路径
		"/usr/lib/opencc",
		"/usr/lib64/opencc",
		// Go模块可能的位置（如果通过apk安装）
		"/usr/lib/go/pkg/mod/github.com/yanmingcao/opencc-go@v1.0.0/data/config",
		"/usr/local/lib/go/pkg/mod/github.com/yanmingcao/opencc-go@v1.0.0/data/config",
		// 可执行文件相关路径
		"/app/opencc/config",
		"/app/config/opencc",
		// 用户主目录路径
		"/root/.local/share/opencc",
		// 其他可能路径
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

// getOpenCCConfigPathFromEmbedded 返回嵌入配置的标识符
func getOpenCCConfigPathFromEmbedded() string {
	// 这是一个虚拟路径，用于触发initEmbeddedConverters
	return "embedded"
}

// initEmbeddedConverters 尝试使用嵌入的配置文件初始化转换器
func initEmbeddedConverters() bool {
	// 目前没有嵌入的配置文件，返回false
	// 未来可以考虑将配置文件嵌入到二进制中
	return false
}

// ContainsChinese 检测字符串是否包含中文字符
func ContainsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// ConvertBoth 将查询字符串转换为简体和繁体两种形式
// 返回：(简体形式, 繁体形式)
func ConvertBoth(text string) (string, string) {
	if !ContainsChinese(text) {
		return text, text
	}

	// 延迟初始化转换器
	initConverters()

	// 如果转换器未初始化，返回原始文本
	if s2tConverter == nil || t2sConverter == nil {
		log.Debug("OpenCC converters are nil, returning original text", "text", text)
		return text, text
	}

	// 简体转繁体
	traditional := s2tConverter.Convert(text)

	// 繁体转简体
	simplified := t2sConverter.Convert(text)

	log.Debug("OpenCC conversion result", "original", text, "simplified", simplified, "traditional", traditional)
	return simplified, traditional
}

// GetSearchQueries 获取搜索查询变体列表
// 如果输入包含中文，返回简体和繁体两种形式（去重）
// 如果不包含中文，返回原始查询
func GetSearchQueries(text string) []string {
	if !ContainsChinese(text) {
		return []string{text}
	}

	simplified, traditional := ConvertBoth(text)

	// 去重
	queries := []string{text}
	if simplified != text {
		queries = append(queries, simplified)
	}
	if traditional != text && traditional != simplified {
		queries = append(queries, traditional)
	}

	// 确保返回的查询列表包含简体和繁体两种形式
	// 如果输入是简体，确保包含繁体
	// 如果输入是繁体，确保包含简体
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

// NormalizeQuery 规范化查询字符串
// 去除首尾空格和尾部通配符
func NormalizeQuery(q string) string {
	q = strings.TrimSpace(q)
	q = strings.TrimSuffix(q, "*")
	return q
}
