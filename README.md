# Hostlog 🪵📡

**Hostlog** is a lightweight, high-performance syslog server designed to bring clarity to network chaos. Whether you're monitoring a fleet of hosts, routers, access points, or IoT devices, Hostlog provides the visibility you need to keep your network working smoothly.

## 📖 The Backstory

Born in the summer of 2025, Hostlog started as an exploration project to work with logs from my home network routers and access points. It was a tool built for personal troubleshooting and to see what was happening across my local infrastructure at a glance.

By December 2025, the project took a leap forward. I integrated **Model Context Protocol (MCP)** support, transforming Hostlog from a passive log viewer into a core component of automated AI workflows. Now, your logs aren't just sitting in a database; they're accessible to AI agents that can help you troubleshoot and analyze your network quickly and in real-time.

## ✨ Features

- **🚀 Dual-Protocol Syslog**: Listen for logs over TCP and UDP (default port 514).
- **🧠 Smart Visibility Scoring**: Hostlog uses a sophisticated algorithm to score host activity:
  - `Visibility = α * e^(-λ * T) + β * V + γ * S`
  - **Recency (T)**: Newer logs have higher impact.
  - **Volume (V)**: Busy hosts rise to the top.
  - **Severity (S)**: Critical errors are prioritized.
- **🤖 MCP Integration**: Seamlessly connect your logs to AI tools like Claude for automated troubleshooting and analysis.
- **🌐 Web Interface**: A clean, Bulma-powered dashboard to visualize logs and host health in real-time.
- **📦 OpenWrt Ready**: Native support for building as an OpenWrt package, making it perfect for custom firmware routers.

## 🛠️ Usage

### Syslog Server & Web UI
To start the syslog and web server:
```bash
go run .
```
The web interface will be available at `http://localhost:8080`.

### MCP Server
To run as an MCP server (via stdio):
```bash
go run . -mcp
```

#### MCP Tools
- `list_hosts`: List all hosts that have sent logs.
- `get_logs`: Get recent logs, optionally filtered by host IPs and page number.
- `get_host_scores`: Get visibility scores for all hosts.

## ⚙️ Configuration

### Jetbrains AI
Modify the hostlog host and add the following to your MCP server settings:

```json
{
  "mcpServers": {
    "hostlog": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "http://<hostlog host>:8080/mcp/sse",
        "--transport",
        "sse-only"
      ]
    }
  }
}
```

### Claude Desktop
Add the following to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "hostlog": {
      "command": "/path/to/hostlog",
      "args": ["-mcp"],
      "env": {
        "HOSTLOG_DB_PATH": "/path/to/your/logs.db"
      }
    }
  }
}
```

---
*Created with ❤️ by a network enthusiast.*
