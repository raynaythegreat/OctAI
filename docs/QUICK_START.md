# Quick Start Guide

> Get AI Business HQ running in 5 minutes

## Prerequisites

Before starting, ensure you have:

| Requirement | Description |
|-------------|-------------|
| **Go 1.25+** | Required only if building from source |
| **API Key** | At least one LLM provider API key |
| **4MB RAM** | Minimum memory requirement |
| **10MB Disk** | For installation and workspace |

### Get Your API Key

Choose one provider to start:

| Provider | Free Tier | Get Key |
|----------|-----------|---------|
| OpenRouter | 200K tokens/month | [openrouter.ai/keys](https://openrouter.ai/keys) |
| Zhipu (GLM) | 200K tokens/month | [bigmodel.cn](https://bigmodel.cn) |
| DeepSeek | Limited free tier | [platform.deepseek.com](https://platform.deepseek.com) |
| OpenAI | Pay-as-you-go | [platform.openai.com](https://platform.openai.com) |
| Anthropic | Pay-as-you-go | [console.anthropic.com](https://console.anthropic.com) |
| Groq | Free tier available | [console.groq.com](https://console.groq.com) |

---

## Installation

### Option 1: Download Binary (Recommended)

```bash
# Linux/macOS
curl -fsSL https://github.com/raynaythegreat/ai-business-hq/releases/latest/download/aibhq-$(uname -s)-$(uname -m) -o aibhq
chmod +x aibhq
sudo mv aibhq /usr/local/bin/

# Verify
aibhq --version
```

### Option 2: Build from Source

```bash
git clone https://github.com/raynaythegreat/ai-business-hq.git
cd ai-business-hq
go build -o aibhq ./cmd/aibhq
```

### Option 3: Docker

```bash
docker pull ghcr.io/raynaythegreat/aibhq:latest
```

---

## First-Time Setup

### Step 1: Initialize Configuration

```bash
aibhq onboard
```

This creates `~/.aibhq/config.json` with sensible defaults.

### Step 2: Add Your API Key

Edit `~/.aibhq/config.json`:

```json
{
  "model_list": [
    {
      "model_name": "gpt-4o-mini",
      "model": "openai/gpt-4o-mini",
      "api_key": "sk-your-api-key-here"
    }
  ],
  "agents": {
    "defaults": {
      "model_name": "gpt-4o-mini"
    }
  }
}
```

### Step 3: Test Your Setup

```bash
aibhq agent -m "Hello! What can you help me with?"
```

---

## First Tasks

### 1. Start the Web Dashboard

```bash
aibhq launcher
```

Open http://localhost:18800 in your browser.

### 2. Connect a Chat Channel

Add a Telegram bot (easiest option):

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "YOUR_BOT_TOKEN_FROM_BOTFATHER",
      "allow_from": ["YOUR_USER_ID"]
    }
  }
}
```

Then start the gateway:

```bash
aibhq gateway
```

### 3. Enable Web Search (Optional)

```json
{
  "tools": {
    "web": {
      "duckduckgo": {
        "enabled": true,
        "max_results": 5
      }
    }
  }
}
```

---

## Common Commands

| Command | Description |
|---------|-------------|
| `aibhq agent -m "message"` | Single query |
| `aibhq agent` | Interactive chat |
| `aibhq gateway` | Start bot gateway |
| `aibhq launcher` | Web dashboard |
| `aibhq onboard` | Initial setup |
| `aibhq auth login --provider anthropic` | OAuth login |

---

## Next Steps

- [Configuration Guide](configuration.md) - Full configuration options
- [Channels Guide](CHANNELS_GUIDE.md) - Connect to Telegram, Slack, Discord
- [Deployment Guide](DEPLOYMENT.md) - Production deployment
- [API Reference](API_REFERENCE.md) - REST API documentation

---

## Troubleshooting

### "No default model configured"

Add a model to `model_list` and set `agents.defaults.model_name`:

```json
{
  "model_list": [
    {
      "model_name": "my-model",
      "model": "openai/gpt-4o-mini",
      "api_key": "sk-..."
    }
  ],
  "agents": {
    "defaults": {
      "model_name": "my-model"
    }
  }
}
```

### "API key invalid"

1. Verify your API key is correct
2. Check the key has not expired
3. Ensure you have credits/quota available

### "Connection refused"

1. Check your internet connection
2. Verify the API endpoint is accessible
3. Check if you're behind a proxy

---

## Getting Help

- [GitHub Issues](https://github.com/raynaythegreat/ai-business-hq/issues)
- [Troubleshooting Guide](troubleshooting.md)
