// OctAi - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 OctAi contributors

package config

// Runtime environment variable keys for the aibhq process.
// These control the location of files and binaries at runtime and are read
// directly via os.Getenv / os.LookupEnv. All aibhq-specific keys use the
// OCTAI_ prefix. Reference these constants instead of inline string
// literals to keep all supported knobs visible in one place and to prevent
// typos.
const (
	// EnvHome overrides the base directory for all aibhq data
	// (config, workspace, skills, auth store, …).
	// Default: ~/.aibhq
	EnvHome = "OCTAI_HOME"

	// EnvConfig overrides the full path to the JSON config file.
	// Default: $OCTAI_HOME/config.json
	EnvConfig = "OCTAI_CONFIG"

	// EnvBuiltinSkills overrides the directory from which built-in
	// skills are loaded.
	// Default: <cwd>/skills
	EnvBuiltinSkills = "OCTAI_BUILTIN_SKILLS"

	// EnvBinary overrides the path to the aibhq executable.
	// Used by the web launcher when spawning the gateway subprocess.
	// Default: resolved from the same directory as the current executable.
	EnvBinary = "OCTAI_BINARY"

	// EnvGatewayHost overrides the host address for the gateway server.
	// Default: "127.0.0.1"
	EnvGatewayHost = "OCTAI_GATEWAY_HOST"
)
