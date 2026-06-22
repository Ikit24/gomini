```
____                 _       _ 
 / ___| ___  _ __ ___ (_)_ __ (_)
| |  _ / _ \| '_ ` _ \| | '_ \| |
| |_| | (_) | | | | | | | | | | |
 \____|\___/|_| |_| |_|_|_| |_|_|
```

Gomini is a production-ready, terminal-based AI chat application written in Go. It provides a seamless, keyboard-driven interface for conversing with the Gemini AI, with all chat history and sessions persisted locally using SQLite.

## 🚀 Features

- **Terminal User Interface (TUI):** Fully interactive, state-driven command-line interface.
- **Local Session Management:** Automatically saves and organizes all past chats in a local SQLite database.
- **Interactive History Browser:** Scroll through past sessions and resume conversations instantly.
- **Live AI Streaming:** Real-time streaming of Gemini AI responses directly to the terminal viewport.

## 📁 Project Structure

The project follows a modular Go layout to separate concerns:
- `cmd/gomini/` - The main application entry point and TUI initialization.
- `internal/database/` - Isolated SQLite connection pool and CRUD operations for sessions and messages.
- `internal/gemini/` - Wrapper client for managing Google AI configurations and stream processing.
- `internal/tui/` - State machine routing, keyboard event handling, and viewport rendering.

## ⌨️ Keyboard Controls

| Context | Key | Action |
| :--- | :--- | :--- |
| **Global** | `ctrl+c` | Quit application |
| **Welcome** | `ctrl+n` | Start a new chat |
| **Welcome** | `ctrl+b` | Browse past chat sessions |
| **Browse** | `up` | Move cursor up |
| **Browse** | `down` | Move cursor down |
| **Browse** | `enter` | Load selected chat |
| **Browse** | `esc` | Return to Welcome screen |
| **Chat** | `up` / `pgup` | Scroll chat history up |
| **Chat** | `down` / `pgdn` | Scroll chat history down |
| **Chat** | `enter` | Submit message |
| **Chat** | `ctrl+b` | Browse past chat sessions |
| **Chat** | `ctrl+n` | Start new chat |

## 🛠️ Getting Started

### Installation
```bash
go install [github.com/Ikit24/gomini@latest](https://github.com/Ikit24/gomini@latest)
```

Running Locally

Ensure you have your Gemini API key set in your environment variables, then run:
```bash
export GEMINI_API_KEY="your_api_key_here"
go run ./cmd/gomini/main.go
```
