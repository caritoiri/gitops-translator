# 🔮 Go GitOps Values Translator

A hyper-performance, single-binary translation engine that automatically adapts legacy standalone GitOps values configuration schemas into the modern unified `charts/webapp` Helm template structure.

This is the **Go (Golang)** equivalent of the Python/Streamlit tool, built to deliver sub-millisecond execution, zero dependencies, and an ultra-responsive, beautiful glassmorphism web interface.

---

## ⚡ Features & Performance

- **Sub-Millisecond Speed**: Translates inputs and serves API calls instantly (under 2ms) without Python startup overhead.
- **Embedded Frontend**: Serves a highly styled Vanilla HTML5/CSS/JS glassmorphic frontend directly from the compiled binary.
- **Dynamic JavaScript Clipboard Copying**: Interactive visual components that copy paths and outputs to clipboard with one click.
- **Real-Time Ajax Debouncing**: Translates as you paste or type—no need to press `Ctrl+Enter` or click outside.
- **Optional JVM Memory Tuning**: Sidebar checkbox toggle to declare whether an app runs on the JVM or other runtimes (Node.js, Go, Python).
- **Layout Collapse Immunity**: Enforced CSS Grid `minmax` configurations and `word-break`/`overflow-wrap` safety boundaries, ensuring paste events with massive base64 contents never deform or collapse the UI/UX panels.
- **Single-Binary Zero-Dependency Portability**: Compiles everything (code, templates, logic) into a single ~15MB binary that runs anywhere without needing python or package installation.

---

## 📁 Directory Structure & File Manifest

The codebase is organized cleanly as follows:

```text
├── main.go                     # Web server entrypoint, serves API routes, embeds static/index.html, loads env configs
├── go.mod                      # Go Module definition (defines module name and dependencies)
├── go.sum                      # Dependency checksums to lock versions
├── Dockerfile                  # Multi-stage lightweight Docker build instructions (~20MB alpine image)
├── compose.yml                 # Orchestration setup for running containerized translator on port 8080
├── README.md                   # Comprehensive project architecture, configuration, and execution instructions
├── static/
│   └── index.html              # Beautiful Vanilla HTML5/CSS/JS frontend (Embedded into the Go binary at build time)
└── translator/
    ├── translator.go           # Core translation engine (handles YAML parsing, validation, re-mapping, and line unwrapping)
    └── translator_test.go      # Go unit test suite verifying various legacy yaml schemas and edge cases
```

---

## ⚙️ Environment Configuration

The application implements a smart, dynamic configuration system. It loads configuration directly from standard system environment variables, removing the need for local `.env` files.

### Default Fallback Values

In case environment variables are not defined in the host system or the container shell, the application code embeds the following robust defaults:

| Variable | Description | Built-In Go Fallback Default |
| :--- | :--- | :--- |
| `PORT` | Listening port for the Go HTTP server | `8080` |
| `DEFAULT_CLUSTER` | Fallback destination cluster suggested path | `on-premise` |
| `JAVA_TOOL_OPTIONS` | Fallback default memory tuning parameters for JVM apps | `-Xms256m -Xmx768m -XX:+UseG1GC` |

---

## 🚀 Execution & Deployment Guide

### Local Native Development

#### 1. Compile & Build
To build the high-performance, single-binary executable:
```bash
go build -o gitops-translator main.go
```

#### 2. Run Locally
To run using system variables or system defaults:
```bash
./gitops-translator
```
To run with custom values defined inline:
```bash
PORT=9090 DEFAULT_CLUSTER=cloud-aws ./gitops-translator
```
Once started, navigate to:
👉 **[http://localhost:8080](http://localhost:8080)** (or your specified port)

#### 3. Run the Test Suite
Validate the translation logic, indentations, and sealed secrets unwrappers:
```bash
go test -v ./...
```

---

### Containerized Deployment (Docker & Compose)

Deploy the application as a lightweight containerized microservice using `compose.yml`.

#### Exposing & Personalizing Environment Variables
The application's variables are declared directly in `compose.yml`. You can easily adjust them:

```yaml
services:
  translator:
    ports:
      - "8080:8080" # Maps local port 8080 to container port 8080
    environment:
      - PORT=8080
      - DEFAULT_CLUSTER=on-premise
      - JAVA_TOOL_OPTIONS=-Xms256m -Xmx768m -XX:+UseG1GC
```

#### Launching the Service
To build and launch the containerized application:
```bash
docker compose up --build -d
```
To view logs:
```bash
docker compose logs -f
```
To shut down the service:
```bash
docker compose down
```
