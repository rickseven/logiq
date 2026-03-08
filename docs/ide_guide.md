# LogIQ: IDE Integration & Token Optimization Guide

LogIQ is designed to be the "Context Gatekeeper" for your AI Coding Assistant. This guide explains how to integrate LogIQ into various IDEs and why it is essential for saving LLM tokens and staying within usage limits.

---

## 🚀 Why Use LogIQ? (The Token Math)

AI Models (LLMs) like Claude or Gemini process text in **Tokens**. Most IDEs send the entire terminal output to the AI, which can quickly consume your daily limits and clutter the conversation context.

### The LogIQ Advantage

- **20+ Intelligent Parsers:** Support for almost any stack: .NET, Python, Go, Flutter, Node.js, Webpack, Nuxt, Next.js, and more.
- **Git Context Awareness:** AI can now understand `git status` and `diff` summaries without reading the entire raw diff.
- **Database Intelligence:** Integrated parsers for Prisma, Drizzle, and other migration tools.
- **Unicode Excellence:** Standardized icons across all platforms for professional status reporting.

### Token Comparison: Running Test Suite

| Metric          | Without LogIQ (Raw Logs) | With LogIQ (Optimized) | Savings            |
| :-------------- | :----------------------- | :--------------------- | :----------------- |
| **Data Volume** | ~5,000 - 20,000 chars    | ~250 chars             | **98% Reduction**  |
| **Token Cost**  | ~2,500 Tokens            | ~60 Tokens             | **40x Lower**      |
| **AI Focus**    | Low (Lost in Noise)      | High (Focus on Signal) | **Better Results** |

---

## 🛠 Integrating with MCP-Native IDEs

IDEs like **Antigravity**, **Cursor**, and **Windsurf** support the Model Context Protocol (MCP) natively via `stdio`.

### How to Connect:

1. Open your IDE's MCP Server settings.
2. Add a new MCP Server (Type: `command`).
3. Set the command to the path of your LogIQ binary:
   - **Command:** `C:\path\to\logiq.exe`
   - **Arguments:** `mcp`
4. The IDE will automatically detect 4 tools: `run_command`, `analyze_logs`, `explain`, and `doctor`.

---

## 🤖 Using with Claude Desktop

Claude Desktop can use LogIQ to run local commands and analyze logs.

1. Open your `claude_desktop_config.json`:
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
2. Add LogIQ to the `mcpServers` section:

```json
{
  "mcpServers": {
    "logiq": {
      "command": "C:\\path\\to\\logiq.exe",
      "args": ["mcp"]
    }
  }
}
```

3. Restart Claude Desktop.

## 🐱 GitHub Copilot & VS Code Integration

GitHub Copilot (via the VS Code Chat environment) supports MCP servers using a `.vscode/mcp.json` configuration file (supported by extensions like "MCP Client" or native updates).

### Setup Instructions:

1. Create a `.vscode` folder in your project root if it doesn't exist.
2. Create a file named `mcp.json` inside that folder.
3. Add the following configuration (use absolute paths):

```json
{
  "servers": {
    "logiq": {
      "command": "C:\\path\\to\\logiq.exe",
      "args": ["mcp"]
    }
  }
}
```

4. Copilot Chat will automatically detect LogIQ. You can verify this by asking: _"Test apakah mcp logiq berjalan dengan benar disini."_
5. Copilot will use **LogIQ/doctor** to diagnose your system or **LogIQ/run_command** to execute tasks efficiently.

---

## 🔍 Smart Context: Signal over Noise

LogIQ ensures that your AI Assistant's "Brain" (Context Window) isn't cluttered with repetitive success messages or progress bars. By keeping the context clean, your AI remains:

- **Fast:** Less text to process means faster responses.
- **Cheap:** You won't hit token limits as frequently.
- **Accurate:** The AI focuses on the code and the errors, not the logs.

---
