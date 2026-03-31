<div align="center">

<img src="web/frontend/public/logo_with_text.png" alt="OctAi Logo" width="280" />

<h3>The All-in-One AI-Powered Business Operations Platform</h3>

<p>
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react&logoColor=black" alt="React">
  <img src="https://img.shields.io/badge/Linux%20%7C%20macOS%20%7C%20Windows%20%7C%20FreeBSD%20%7C%20NetBSD-blue" alt="Platforms">
  <img src="https://img.shields.io/badge/x86__64%20%7C%20ARM64%20%7C%20ARM%20%7C%20RISC--V%20%7C%20LoongArch%20%7C%20MIPS-blue" alt="Architectures">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
  <br>
  <a href="https://github.com/raynaythegreat/OctAI"><img src="https://img.shields.io/badge/GitHub-Repository-black?style=flat&logo=github&logoColor=white" alt="GitHub"></a>
</p>

**One-line install:**
```bash
curl -fsSL https://raw.githubusercontent.com/raynaythegreat/OctAI/master/install.sh | bash
```

</div>

---

## What is OctAi?

**OctAi** is a comprehensive, AI-powered business operations platform. A single lightweight binary (<20MB RAM) gives you a multi-agent AI system with 20+ LLM providers, 16 messaging channels, 25+ built-in tools, a modern web dashboard, a full TUI, cron scheduling, MCP protocol support, and extensible skills — all running on hardware from a $10 SBC to a cloud server.

### At a Glance

| | OctAi |
|---|---|
| **Memory** | <20MB RAM |
| **Boot** | <1 second |
| **Platforms** | Linux, macOS, Windows, FreeBSD, NetBSD |
| **Architectures** | x86_64, ARM64, ARM, RISC-V, LoongArch, MIPS, s390x |
| **LLM Providers** | 20+ providers, 60+ models |
| **Channels** | Telegram, Discord, Slack, WhatsApp, WeChat, Feishu, Matrix, IRC, LINE, QQ, DingTalk, WeCom, and more |
| **Tools** | 25+ built-in (shell, filesystem, web search, memory, knowledge base, subagents, cron, I2C/SPI) |
| **Skills** | 14+ built-in + installable from registries |
| **Deployment** | Single binary, Docker, or `curl | bash` |

---

## Quick Start

### One-Line Install (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/raynaythegreat/OctAI/master/install.sh | bash
```

The script detects your OS/architecture, downloads the correct binary, and installs to `~/.local/bin`. Run `octai onboard` to configure your first provider.

<details>
<summary>Windows (PowerShell)</summary>

```powershell
iwr -useb https://raw.githubusercontent.com/raynaythegreat/OctAI/master/install.ps1 | iex
```

</details>

<details>
<summary>Docker</summary>

```bash
# Web console + gateway
docker run -d -p 18800:18800 -v octai-data:/root/.octai ghcr.io/raynaythegreat/octai:latest

# Long-running gateway (all channels)
docker run -d -v octai-data:/root/.octai ghcr.io/raynaythegreat/octai:latest gateway
```

</details>

<details>
<summary>Build from Source</summary>

```bash
git clone https://github.com/raynaythegreat/OctAI.git
cd OctAI

# Build (requires Go 1.25+)
make build

# Or build for all platforms
make build-all
```

</details>

### First Run

```bash
# Interactive setup wizard — pick your AI provider, configure channels
octai onboard

# Start the web dashboard and gateway (http://localhost:18800)
octai serve

# Or jump straight into the TUI
octai tui

# Or start a single agent session
octai agent
```

---

## Features

### AI Providers (20+ providers, 60+ models)

| Provider | Models | Auth |
|----------|--------|------|
| **OpenAI** | GPT-5.4, GPT-5, o3, o4-mini | OAuth, Device Code, API Key |
| **Anthropic** | Claude Opus 4.6, Claude Sonnet 4.6, Claude Haiku 4.5 | Browser OAuth, Setup Token, API Key |
| **Google Gemini** | Gemini 2.5 Pro, 2.5 Flash, 2.5 Flash Lite | Antigravity OAuth, API Key |
| **xAI Grok** | Grok 4 (Reasoning, Non-Reasoning, Fast) | API Key |
| **DeepSeek** | DeepSeek Chat, DeepSeek R1 | API Key |
| **OpenRouter** | 100+ models via single key | API Key |
| **GitHub Copilot** | GPT-5.4 via Copilot subscription | OAuth (no extra cost) |
| **Azure OpenAI** | Custom deployments | API Key |
| **AWS Bedrock** | Claude, Titan, etc. | API Key |
| **Groq** | Llama 4, Qwen3, Mixtral | API Key |
| **Cerebras** | GPT-OSS 120B, Llama 3.1 8B | API Key |
| **Mistral AI** | Mistral Large, Codestral, Devstral | API Key |
| **Perplexity** | Sonar, Sonar Pro, Sonar Reasoning | API Key |
| **Together AI** | Llama 4, Qwen3, DeepSeek R1, Kimi K2 | API Key |
| **Ollama** | Llama, Qwen, Mistral, Phi, Gemma (local) | None (local) |
| **vLLM / LiteLLM** | Any model via local server | API Key |
| **Kimi / Moonshot** | Kimi K2.5, K2 Thinking, K2 Turbo | API Key |
| **Minimax** | MiniMax M2.5 | API Key |
| **Qwen (Alibaba)** | Qwen Plus | API Key |
| **Zhipu AI (GLM)** | GLM 4.7 | API Key |
| **Avian** | DeepSeek V3.2, Kimi K2.5 | API Key |
| **NVIDIA** | Nemotron 4 340B | API Key |

**Provider features:** automatic failover with error classification, per-model cooldown/backoff, multi-key load balancing (round-robin), streaming, extended thinking, native web search.

### Messaging Channels (16 integrations)

| Channel | Protocol | Features |
|---------|----------|----------|
| **Telegram** | Bot API | Typing indicators, streaming, Markdown v2, command registration |
| **Discord** | Gateway | Typing, placeholders, mention-only mode, group triggers |
| **Slack** | Socket Mode | Bot token + app token auth |
| **WhatsApp** | Baileys / Native | Bridge mode + native protocol |
| **WeChat (Weixin)** | Web/API | OAuth, media handling |
| **WeCom** | WebSocket | Enterprise WeChat, request dedup |
| **Feishu / Lark** | Open API | Encryption, random reactions |
| **DingTalk** | Open API | Client ID/Secret auth |
| **QQ** | BotGo SDK | Audio detection, max message length |
| **Matrix** | Client-Server API | E2EE crypto, join-on-invite |
| **LINE** | Messaging API | Webhook receiver |
| **IRC** | IRC protocol | NickServ, SASL auth |
| **OneBot** | WebSocket | Auto-reconnect |
| **Pico** | Custom | Lightweight embedded protocol |
| **MaixCam** | Custom | Sipeed MaixCam embedded device |
| **Web** | Built-in | Full web chat interface |

### Built-in Tools (25+)

**Core Agent Tools:**
- `exec` — Shell command execution (PTY support, workspace sandboxing, dangerous command blocking)
- `read_file` / `write_file` / `append_file` / `edit_file` / `list_dir` — Full filesystem access
- `web_search` — Web search via 7 providers (DuckDuckGo, Brave, Tavily, Perplexity, SearXNG, GLM, Baidu)
- `web_fetch` — Fetch and parse web pages
- `send_file` / `message` — Send files and messages to any channel

**Agent Orchestration:**
- `spawn` — Spawn async background subagents
- `spawn_status` — Check spawned subagent status
- `subagent` — Synchronous subagent execution
- `team` — Delegate tasks to specialist team members
- `background_agent` / `check_background` — Submit and monitor background tasks

**Memory & Knowledge:**
- `save_memory` / `recall_memory` — Persistent cross-session memory (user, project, feedback, reference, fact)
- `knowledge_search` / `knowledge_add` — BM25 full-text search knowledge base (SQLite + FTS5)
- `search_references` — Search saved reference URL library

**Automation:**
- `cron` — Schedule/manage cron jobs (add, remove, list, enable, disable)

**Discovery:**
- `find_skills` / `install_skill` — Search and install skills from registries

**Hardware (Linux):**
- `i2c` — I2C bus interaction (detect, scan, read, write)
- `spi` — SPI bus interaction (list, transfer, read)

### Multi-Agent System

- **8 built-in roles:** Orchestrator, Sales, Support, Research, Content, Analytics, Admin, Custom
- **Team configuration:** Orchestrator + specialist agents with per-agent model, skills, hooks, budget
- **Agent bindings:** Route specific channels/users to specific agents
- **SubTurn system:** Nested agent execution with depth limits, concurrency caps, token budgets
- **Auto-mode:** Safety classifier (strict/balanced/permissive) for autonomous operation
- **Lifecycle hooks:** Observer, interceptor, approval hooks with HTTP webhook support
- **Extended thinking:** Anthropic-style thinking support
- **Intelligent routing:** Structural complexity scoring routes simple tasks to lighter models

### Skills System

14+ built-in skills accessible via `/skillname` slash commands in chat, TUI, and Telegram:

`brainstorming` · `dispatching-parallel-agents` · `executing-plans` · `finishing-a-development-branch` · `receiving-code-review` · `requesting-code-review` · `subagent-driven-development` · `systematic-debugging` · `test-driven-development` · `using-git-worktrees` · `using-superpowers` · `verification-before-completion` · `writing-plans` · `writing-skills`

Install community skills from registries: `octai skills search <query>` → `octai skills install <name>`

### MCP (Model Context Protocol)

- Full MCP server lifecycle management (start, stop, list tools)
- Per-server configuration: command, args, env vars, headers
- `.env` file loading per MCP server
- Tool name sanitization for provider compatibility
- Discovery mode with BM25 + regex search over MCP tools

### Web Dashboard

Modern React 19 UI at `http://localhost:18800`:

- Chat interface with WebSocket streaming
- Provider/credential management with OAuth flows
- Channel configuration per integration
- Agent and team management
- Skills marketplace (browse, install, configure)
- MCP server management
- Cron job and loop scheduling
- Workflow editor
- Model configuration with status indicators
- Session history
- Log viewer with ANSI support
- Dark mode, i18n (English, Chinese, Spanish)
- Plan/Build mode toggle

### TUI Interface

Full-featured terminal UI with:
- Interactive chat with streaming responses
- Slash commands (`/skills`, `/models`, `/clear`, `/help`)
- Session history
- Plan/Build mode
- Configuration dashboard

### Image & Video Generation

- **Image models:** DALL-E 3, Google Imagen 4, Stability AI SDXL, FLUX.1 Schnell
- **Video models:** Runway Gen-4, Kling AI V2, Google Veo 2, MiniMax Video

### Business Automation

- **Cron scheduling** — Time-based AI task execution
- **Loop system** — Recurring agent tasks with configurable intervals
- **Webhooks** — Integrate with external services
- **Workflows** — DAG-based workflow execution with agent dispatch
- **AI URL Scanner** — Point at any URL to auto-discover MCP servers, skills, and tools
- **Reference URL Library** — AI-categorized link collection for agent research

### Security

- Credential encryption at rest (0600 file permissions)
- Separate `.security.yml` config (never in main config)
- OAuth 2.0 flows (Anthropic, OpenAI, Google, WeChat, WeCom)
- Automatic sensitive data filtering from LLM context
- Per-channel `allow_from` access control lists
- Workspace filesystem sandboxing with symlink escape prevention
- Dangerous shell command blocking (rm -rf, format, dd, block device writes)
- Structured audit logging with PII redaction
- Compliance checker framework

### Observability

- Real-time cost tracking per agent/team
- Token usage metrics (input/output)
- Cache savings estimation
- Agent health monitoring with alert system
- Built-in pricing table for popular models

---

## CLI Commands

```
octai onboard              Interactive setup wizard
octai agent                Start AI chat session
octai serve                Start web dashboard and gateway (port 18800)
octai tui                  Start terminal UI
octai gateway              Start long-running gateway server
octai auth login           OAuth login (--provider openai|anthropic|google-antigravity)
octai auth status          Show authentication state
octai auth logout          Clear credentials
octai models               List available models
octai cron                 Manage scheduled tasks
octai loop                 Create recurring agent tasks
octai skills list          List installed skills
octai skills search <q>    Search skill registries
octai skills install <n>   Install a skill
octai version              Show version info
octai status               Show system status
```

---

## Architecture

```
OctAi/
├── cmd/
│   ├── aibhq/                  # Main CLI binary
│   └── aibhq-launcher/         # Web/TUI launcher binary
├── pkg/
│   ├── agent/                  # Multi-agent system (roles, teams, hooks, routing)
│   ├── channels/               # 16 messaging channel integrations
│   ├── providers/              # 20+ LLM providers with fallback & load balancing
│   ├── tools/                  # 25+ built-in tools
│   ├── memory/                 # Session memory + persistent agent store (SQLite)
│   ├── knowledge/              # Knowledge base with BM25/FTS5 search
│   ├── skills/                 # Skill loader, registry, installer
│   ├── mcp/                    # Model Context Protocol manager
│   ├── workflow/               # DAG workflow engine
│   ├── gateway/                # Long-running multi-channel server
│   ├── audit/                  # Audit logging with PII redaction
│   ├── compliance/             # Compliance checker framework
│   ├── observability/          # Cost tracking, metrics, alerts
│   ├── batch/                  # Batch API (OpenAI, Anthropic)
│   ├── integrations/           # Pluggable integration framework (HubSpot, etc.)
│   ├── marketplace/            # Skill marketplace
│   └── config/                 # Configuration with migration & security
├── web/
│   ├── frontend/               # React 19 + TypeScript UI
│   └── backend/                # Go REST API + WebSocket server
├── docker/                     # Dockerfiles + docker-compose
├── skills/                     # Built-in skills (SKILL.md)
└── workspace/                  # Agent workspace
```

---

## Documentation

- [Configuration Guide](docs/configuration.md)
- [Provider Setup](docs/providers.md)
- [Channel Configuration](docs/channels/)
- [Tools & Skills](docs/tools_configuration.md)
- [Docker Deployment](docs/docker.md)
- [Troubleshooting](docs/troubleshooting.md)

---

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

## Acknowledgments

OctAi is built on the foundation of [OctAi](https://github.com/sipeed/picoclaw) by [Sipeed](https://sipeed.com), reimagined as a comprehensive AI-powered business operations platform.
