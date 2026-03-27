package internal

import (
	"os"
	"path/filepath"

	"github.com/raynaythegreat/ai-business-hq/pkg"
	"github.com/raynaythegreat/ai-business-hq/pkg/config"
	"github.com/raynaythegreat/ai-business-hq/pkg/logger"
)

const Logo = pkg.Logo

// GetPicoclawHome returns the octai home directory.
// Priority: $OCTAI_HOME > ~/.octai
func GetPicoclawHome() string {
	if home := os.Getenv(config.EnvHome); home != "" {
		return home
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, pkg.DefaultAIBusinessHQHome)
}

func GetConfigPath() string {
	if configPath := os.Getenv(config.EnvConfig); configPath != "" {
		return configPath
	}
	return filepath.Join(GetPicoclawHome(), "config.json")
}

func LoadConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig(GetConfigPath())
	if err != nil {
		return nil, err
	}
	logger.SetLevelFromString(cfg.Gateway.LogLevel)
	return cfg, nil
}

// FormatVersion returns the version string with optional git commit
// Deprecated: Use pkg/config.FormatVersion instead
func FormatVersion() string {
	return config.FormatVersion()
}

// FormatBuildInfo returns build time and go version info
// Deprecated: Use pkg/config.FormatBuildInfo instead
func FormatBuildInfo() (string, string) {
	return config.FormatBuildInfo()
}

// GetVersion returns the version string
// Deprecated: Use pkg/config.GetVersion instead
func GetVersion() string {
	return config.GetVersion()
}
