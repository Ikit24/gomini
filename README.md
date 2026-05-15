```text
  ____                 _       _ 
 / ___| ___  _ __ ___ (_)_ __ (_)
| |  _ / _ \| '_ ` _ \| | '_ \| |
| |_| | (_) | | | | | | | | | | |
 \____|\___/|_| |_| |_|_|_| |_|_|
```

Gomini is a production-ready, lightweight Go backend API that orchestrates conversation sessions using SQLite for persistence and the Gemini AI live client for intelligent responses.

## 🚀 Features

- **AI Chat Orchestration:** Manages multi-turn conversations, persisting user queries and streaming/saving Gemini AI responses.
- **Session Management:** Full CRUD capabilities for tracking separate chat sessions, history, and metadata.
- **Robust Health Monitoring:** Active server and database pulse checking via custom SQLite pool delegation.
- **Graceful Shutdown:** Cleans up active database connections and handles server lifecycles safely on termination signals.

## 📁 Project Structure

The project follows the standard Go enterprise layout to prevent circular dependencies:
- `cmd/gomini/` - The main application entry point.
- `internal/database/` - Isolated SQLite connection pool and low-level data persistence handlers.
- `internal/gemini/` - Wrapper client for managing Google AI configurations.
- `internal/handlers/` - HTTP routing, server logic, and request orchestration.

## 🛣️ API Endpoints

| Method | Endpoint | Description | Status |
| :--- | :--- | :--- | :--- |
| `GET` | `/healthz` | Checks API and Database pool health | Complete |
| `GET` | `/api/sessions/{session_id}` | Retrieves a specific chat session | Complete |
| `PATCH` | `/api/sessions/{session_id}` | Updates session titles and metadata | Complete |
| `GET` | `/api/sessions/{session_id}/messages` | Fetches full message history for a session | Complete |
| `POST` | `/api/sessions/{session_id}/messages` | Sends a message to the database and fetches Gemini AI reply | Complete |

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
