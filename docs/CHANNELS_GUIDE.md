# Channels Configuration Guide

> Complete guide to configuring communication channels in AI Business HQ

## Overview

AI Business HQ supports 15+ communication channels. This guide covers setup for each channel.

---

## Quick Reference

| Channel | Difficulty | Auth Method | Public IP Required |
|---------|------------|-------------|-------------------|
| Telegram | Easy | Bot Token | No |
| Discord | Easy | Bot Token | No |
| Slack | Easy | Bot Token + App Token | No |
| WhatsApp (Native) | Easy | QR Scan | No |
| WeChat | Easy | QR Scan | No |
| Matrix | Medium | Access Token | No |
| QQ | Medium | App ID + Secret | No |
| DingTalk | Medium | Client ID + Secret | No |
| Feishu | Advanced | App ID + Secret | No |
| WeCom | Advanced | Bot ID + Secret | No |
| LINE | Advanced | Channel Secret + Token | Yes (HTTPS) |
| IRC | Medium | Password (optional) | No |
| OneBot | Medium | WebSocket URL | No |

---

## Telegram (Recommended)
The easiest channel to set up. Recommended for beginners.

### Step 1: Create a Bot

1. Open Telegram and search for `@BotFather`
2. Send `/newbot`
3. Follow the prompts to name your bot
4. Copy the provided token (format: `123456789:ABCdefGHI...`)

### Step 2: Get Your User ID

1. Search for `@userinfobot`
2. Start it and copy your numeric user ID

### Step 3: Configure

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "123456789:ABCdefGHI...",
      "allow_from": ["YOUR_USER_ID"],
      "use_markdown_v2": false
    }
  }
}
```

### Step 4: Run

```bash
aibhq gateway
```

### Telegram Commands

| Command | Description |
|---------|-------------|
| `/start` | Start the bot |
| `/help` | Show help message |
| `/show` | Show current configuration |
| `/list skills` | List available skills |
| `/use <skill>` | Use a specific skill |

### Advanced: MarkdownV2

Enable enhanced formatting:

```json
{
  "channels": {
    "telegram": {
      "use_markdown_v2": true
    }
  }
}
```

---

## Discord

### Step 1: Create a Bot Application

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application" and name it
3. Navigate to "Bot" section and click "Add Bot"
4. Copy the bot token

### Step 2: Enable Intents

In Bot settings, enable:
- **MESSAGE CONTENT INTENT** (required)
- **SERVER MEMBERS INTENT** (optional, for member-based allow lists)

### Step 3: Get Your User ID

1. Go to Discord Settings → Advanced
2. Enable **Developer Mode**
3. Right-click your avatar → **Copy User ID**

### Step 4: Configure

```json
{
  "channels": {
    "discord": {
      "enabled": true,
      "token": "MTk4NjI...YourBotToken",
      "allow_from": ["YOUR_USER_ID"]
    }
  }
}
```

### Step 5: Invite the Bot

1. In Developer Portal, go to OAuth2 → URL Generator
2. Select scopes: `bot`, `applications.commands`
3. Bot Permissions: `Send Messages`, `Read Message History`
4. Open the generated URL to add the bot to your server

### Group Trigger Options

```json
{
  "channels": {
    "discord": {
      "group_trigger": {
        "mention_only": true
      }
    }
  }
}
```

Or trigger by prefix:

```json
{
  "group_trigger": {
    "prefixes": ["!bot", "!ai"]
  }
}
```

---

## Slack

### Step 1: Create a Slack App

1. Go to [Slack API](https://api.slack.com/apps)
2. Click "Create New App"
3. Choose "From scratch" and name your app

### Step 2: Configure OAuth & Permissions

Add these Bot Token Scopes:
- `chat:write`
- `app_mentions:read`
- `im:history`
- `im:read`
- `im:write`

### Step 3: Enable Socket Mode

1. Go to "Socket Mode" section
2. Enable Socket Mode
3. Copy the **App-Level Token** (`xapp-...`)

### Step 4: Install the App

1. Go to "Install App" section
2. Install to your workspace
3. Copy the **Bot Token** (`xoxb-...`)

### Step 5: Configure

```json
{
  "channels": {
    "slack": {
      "enabled": true,
      "bot_token": "xoxb-YOUR-BOT-TOKEN",
      "app_token": "xapp-YOUR-APP-TOKEN",
      "allow_from": []
    }
  }
}
```

---

## WhatsApp

### Option 1: Native Mode (Recommended)

Uses whatsmeow library directly. No external bridge needed.

**Prerequisites:**
- Build with `whatsapp_native` tag: `go build -tags whatsapp_native ./cmd/...`

**Configuration:**

```json
{
  "channels": {
    "whatsapp": {
      "enabled": true,
      "use_native": true,
      "session_store_path": "",
      "allow_from": []
    }
  }
}
```

**First Run:**
1. Start the gateway
2. Scan the QR code displayed in the terminal
3. Use WhatsApp → Linked Devices on your phone

### Option 2: Bridge Mode

Connect to an external WhatsApp Web bridge.

```json
{
  "channels": {
    "whatsapp": {
      "enabled": true,
      "use_native": false,
      "bridge_url": "ws://localhost:3001",
      "allow_from": []
    }
  }
}
```

---

## WeChat (Personal)

### Step 1: Login

```bash
aibhq auth weixin
```

### Step 2: Scan QR Code

A QR code will appear in the terminal. Scan it with:
- WeChat mobile app → Me → Settings → Devices → Scan

### Step 3: Configure (Optional)

```json
{
  "channels": {
    "weixin": {
      "enabled": true,
      "token": "AUTO_CONFIGURED",
      "allow_from": ["YOUR_WECHAT_ID"]
    }
  }
}
```

---

## WeCom (Enterprise WeChat)

### Quick Setup

```bash
aibhq auth wecom
```

Scan the QR code with WeCom to authenticate.

### Manual Configuration

```json
{
  "channels": {
    "wecom": {
      "enabled": true,
      "bot_id": "YOUR_BOT_ID",
      "secret": "YOUR_SECRET",
      "websocket_url": "wss://openws.work.weixin.qq.com",
      "send_thinking_message": true,
      "allow_from": [],
      "reasoning_channel_id": ""
    }
  }
}
```

---

## Matrix

### Step 1: Create a Bot Account

1. Register on a homeserver (e.g., [matrix.org](https://matrix.org))
2. Create a dedicated user account for the bot
3. Get an access token from Element or API

### Step 2: Configure

```json
{
  "channels": {
    "matrix": {
      "enabled": true,
      "homeserver": "https://matrix.org",
      "user_id": "@your-bot:matrix.org",
      "access_token": "YOUR_ACCESS_TOKEN",
      "device_id": "DEVICE_ID",
      "allow_from": []
    }
  }
}
```

### Optional Settings

```json
{
  "join_on_invite": true,
  "group_trigger": {
    "mention_only": true
  },
  "placeholder": "Thinking..."
}
```

---

## QQ

### Quick Setup (Recommended)

1. Open [QQ Bot Quick Start](https://q.qq.com/qqbot/openclaw/index.html)
2. Scan QR code to log in
3. A bot is created automatically
4. Copy **App ID** and **App Secret**

### Configure

```json
{
  "channels": {
    "qq": {
      "enabled": true,
      "app_id": "YOUR_APP_ID",
      "app_secret": "YOUR_APP_SECRET",
      "allow_from": []
    }
  }
}
```

> **Note:** The App Secret is shown only once. Save it immediately!

---

## Feishu (Lark)

### Step 1: Create an App

1. Go to [Feishu Open Platform](https://open.feishu.cn/)
2. Create an application
3. Enable the **Bot** capability
4. Create a version and publish

### Step 2: Configure

```json
{
  "channels": {
    "feishu": {
      "enabled": true,
      "app_id": "cli_xxx",
      "app_secret": "YOUR_APP_SECRET",
      "encrypt_key": "",
      "verification_token": "",
      "allow_from": []
    }
  }
}
```

### Group Settings

```json
{
  "group_trigger": {
    "mention_only": true
  }
}
```

---

## DingTalk

### Step 1: Create an App

1. Go to [DingTalk Open Platform](https://open.dingtalk.com/)
2. Create an internal app
3. Copy **Client ID** and **Client Secret**

### Step 2: Configure

```json
{
  "channels": {
    "dingtalk": {
      "enabled": true,
      "client_id": "YOUR_CLIENT_ID",
      "client_secret": "YOUR_CLIENT_SECRET",
      "allow_from": []
    }
  }
}
```

---

## LINE

### Step 1: Create a LINE Official Account

1. Go to [LINE Developers Console](https://developers.line.biz/)
2. Create a provider
3. Create a Messaging API channel
4. Copy **Channel Secret** and **Channel Access Token**

### Step 2: Configure

```json
{
  "channels": {
    "line": {
      "enabled": true,
      "channel_secret": "YOUR_CHANNEL_SECRET",
      "channel_access_token": "YOUR_CHANNEL_ACCESS_TOKEN",
      "webhook_path": "/webhook/line",
      "allow_from": []
    }
  }
}
```

### Step 3: Set Up Webhook

LINE requires HTTPS webhooks. Use a tunnel:

```bash
ngrok http 18790
```

Set webhook URL in LINE Console: `https://your-ngrok-url.ngrok.io/webhook/line`

---

## IRC

### Basic Configuration

```json
{
  "channels": {
    "irc": {
      "enabled": true,
      "server": "irc.libera.chat:6697",
      "tls": true,
      "nick": "aibhq-bot",
      "channels": ["#your-channel"],
      "password": "",
      "allow_from": []
    }
  }
}
```

### Advanced Authentication

```json
{
  "nickserv_password": "YOUR_NICKSERV_PASSWORD",
  "sasl_user": "username",
  "sasl_password": "sasl_password"
}
```

---

## OneBot (QQ Protocol)

### Step 1: Set Up OneBot Implementation

Install a OneBot v11 compatible framework:
- [Lagrange](https://github.com/LagrangeDev/Lagrange.Core)
- [NapCat](https://github.com/NapNeko/NapCatQQ)

### Step 2: Enable WebSocket

Configure your OneBot implementation to expose a WebSocket server.

### Step 3: Configure

```json
{
  "channels": {
    "onebot": {
      "enabled": true,
      "ws_url": "ws://127.0.0.1:8080",
      "access_token": "",
      "reconnect_interval": 5,
      "allow_from": []
    }
  }
}
```

---

## Troubleshooting

### Common Issues

#### Telegram: "Unauthorized"

**Problem:** Bot doesn't respond to messages.

**Solution:** Add your user ID to `allow_from`:
```json
"allow_from": ["123456789"]
```

#### Discord: "Missing Intent"

**Problem:** Bot can't read messages.

**Solution:** Enable **MESSAGE CONTENT INTENT** in Discord Developer Portal.

#### Slack: "Socket Mode Failed"

**Problem:** Connection fails.

**Solution:** 
1. Verify Socket Mode is enabled
2. Check that `app_token` starts with `xapp-`
3. Ensure `bot_token` starts with `xoxb-`

#### WhatsApp: "QR Code Timeout"

**Problem:** QR code expires before scanning.

**Solution:** Restart the gateway to generate a new QR code.

#### Matrix: "Failed to Join Room"

**Problem:** Bot can't join invited rooms.

**Solution:** Enable `join_on_invite`:
```json
"join_on_invite": true
```

#### LINE: "Webhook Verification Failed"

**Problem:** LINE can't reach webhook.

**Solution:**
1. Verify ngrok or reverse proxy is running
2. Check HTTPS certificate
3. Confirm webhook URL in LINE Console

### Debug Mode

Enable verbose logging:

```bash
aibhq gateway --debug
```

Check logs for channel-specific errors.

---

## Security Best Practices

### 1. Always Use Allow Lists

```json
"allow_from": ["YOUR_USER_ID"]
```

### 2. Use Security File for Tokens

Store sensitive tokens in `~/.aibhq/.security.yml`:

```yaml
channels:
  telegram:
    token: "YOUR_BOT_TOKEN"
  discord:
    token: "YOUR_BOT_TOKEN"
```

### 3. Restrict Group Responses

```json
"group_trigger": {
  "mention_only": true
}
```

### 4. Regular Token Rotation

Periodically rotate bot tokens and update configuration.
