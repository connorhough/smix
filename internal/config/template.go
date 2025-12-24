package config

const configTemplate = `# smix configuration file
# Provider settings control which LLM provider to use

# Global default provider (claude or gemini)
provider: claude

# Global default model (optional, uses provider default if omitted)
# model: sonnet

# Provider-specific settings
providers:
  claude:
    # Path to claude CLI if not in PATH (optional)
    # cli_path: /usr/local/bin/claude
  gemini:
    # API key (prefer GEMINI_API_KEY environment variable)
    # api_key: ${GEMINI_API_KEY}

# Per-command overrides (optional)
# Uncomment and customize as needed
#commands:
#  ask:
#    provider: gemini
#    model: gemini-1.5-flash
#  do:
#    provider: gemini
#    model: gemini-1.5-flash
#  pr:
#    provider: claude
#    model: sonnet

# Observability settings
log_level: info  # debug, info, warn, error
`
