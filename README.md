<div align="center">

<h1>AI Business HQ</h1>

<h3>The All-in-One AI-Powered Business Operations Platform</h3>

<p>
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Arch-x86__64%2C%20ARM64%2C%20MIPS%2C%20RISC--V%2C%20LoongArch-blue" alt="Hardware">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
  <br>
  <a href="https://github.com/raynaythegreat/ai-business-hq"><img src="https://img.shields.io/badge/GitHub-Repository-black?style=flat&logo=github&logoColor=white" alt="GitHub"></a>
</p>

</div>

---

## What is AI Business HQ?

**AI Business HQ** is a comprehensive, AI-powered business operations platform designed to streamline and automate every aspect of your business. Built on a lightweight Go foundation, it brings enterprise-grade AI capabilities to businesses of all sizes.

### Core Capabilities

- **Multi-Channel Communications** - Unified messaging across Telegram, Slack, Discord, WhatsApp, WeChat, Feishu, Matrix, and more
- **AI Agent Orchestration** - Intelligent task routing, scheduling, and multi-agent collaboration
- **Business Process Automation** - Cron jobs, webhooks, and automated workflows
- **Skills Marketplace** - Extensible plugin system for custom business logic
- **Memory & Context Management** - Persistent conversations with intelligent context handling
- **Multi-LLM Support** - Works with OpenAI, Anthropic, Azure, Bedrock, Ollama, and 20+ providers
- **Voice & Vision** - Audio transcription and image understanding capabilities
- **Web Dashboard** - Modern React-based launcher and management interface

### Why AI Business HQ?

| Feature | AI Business HQ | Traditional Solutions |
|---------|---------------|----------------------|
| Memory Usage | <20MB RAM | 500MB+ |
| Boot Time | <1 second | 30+ seconds |
| Hardware | $10 devices supported | Expensive servers |
| Channels | 15+ integrations | 3-5 typical |
| AI Providers | 20+ providers | 1-3 typical |
| Deployment | Single binary | Complex setup |

---

## Quick Start

### Prerequisites

- Go 1.25+ (for building from source)
- Or download pre-built binaries from [Releases](https://github.com/raynaythegreat/ai-business-hq/releases)

### Installation

```bash
# Clone the repository
git clone https://github.com/raynaythegreat/ai-business-hq.git
cd ai-business-hq

# Build
go build -o aibhq ./cmd/aibhq

# Run
./aibhq
```

### Docker

```bash
docker-compose up -d
```

---

## Features

### Multi-Channel Support

Connect your AI assistant to any platform:

- **Messaging**: Telegram, Discord, Slack, WhatsApp, WeChat, Feishu, LINE, Matrix, IRC
- **Enterprise**: WeCom (WeChat Work), DingTalk, QQ
- **Voice**: Direct audio transcription support
- **Web**: Built-in web chat interface

### AI Provider Integration

Works with all major LLM providers:

- OpenAI (GPT-4, GPT-4o, etc.)
- Anthropic (Claude)
- Azure OpenAI
- AWS Bedrock
- Google AI / Gemini
- Ollama (local models)
- vLLM, LM Studio
- Kimi, Minimax, and more

### Business Automation

- **Cron Scheduling**: Schedule AI-powered tasks
- **Webhooks**: Integrate with external services
- **Skills System**: Extend functionality with custom skills
- **MCP Protocol**: Model Context Protocol support for advanced integrations

### Security

- Credential encryption at rest
- OAuth 2.0 flows for authentication
- Configurable access controls
- Sensitive data filtering

---

## Architecture

```
ai-business-hq/
├── cmd/
│   ├── aibhq/              # Main CLI application
│   └── aibhq-launcher/     # Web/TUI launcher
├── pkg/
│   ├── agent/              # AI agent core
│   ├── channels/           # Communication channels
│   ├── providers/          # LLM providers
│   ├── tools/              # Built-in tools
│   ├── memory/             # Context & memory
│   └── skills/             # Skills system
├── web/                    # Web UI (React + Go backend)
├── workspace/              # Agent workspace & skills
└── docs/                   # Documentation
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

## Roadmap

See [ROADMAP.md](ROADMAP.md) for our development roadmap.

### Coming Soon

- [ ] SaaS multi-tenant support
- [ ] Advanced analytics dashboard
- [ ] Team collaboration features
- [ ] API marketplace
- [ ] Mobile apps

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Acknowledgments

AI Business HQ is built on the foundation of [AI Business HQ](https://github.com/raynaythegreat/ai-business-hq) by [Sipeed](https://sipeed.com), reimagined as a comprehensive business operations platform.
