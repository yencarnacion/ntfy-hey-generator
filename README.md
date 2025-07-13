# ntfy-hey-generator

> Play a custom **MP3** every time a new [ntfy](https://docs.ntfy.sh/) message lands on one or more topics.

This tiny cross‑platform CLI (Ubuntu Linux & macOS) subscribes to ntfy topics over **WebSockets** and triggers an audio notification for each *new* message. It ignores any messages that existed before the program started and exits cleanly on **Ctrl‑C**.

---

## ✨ Features

* ⏺  Subscribe to **multiple** ntfy topics at once
* 🔊  Mix overlapping messages so no notification is lost
* 🚦  Graceful shutdown & automatic reconnects
* ⚙️  Fully configurable via `.env`
* 🐧  Works on Linux (ALSA) and macOS (CoreAudio)
* 📜  MIT‑licensed

---

## 📦 Folder layout

```
ntfy-hey-generator/
├── main.go          # Program source
├── go.mod / go.sum  # Go modules manifest
├── sounds/
│   └── hey.mp3      # Default sample sound (replace with your own)
├── env.example     # Copy → .env and edit (see below)
└── go.sh            # Helper script: `./go.sh`
```

---

## 🚀 Quick‑start

### 1  Install Go

| OS                            | Command    |
| ----------------------------- | ---------- |
| **Ubuntu 22.04 +**            | \`\`\`bash |
| sudo apt update && \\         |            |
| sudo apt install -y golang-go |            |

````|
| **macOS (Homebrew)** | ```bash
brew install go
``` |

> **Minimum version**: Go 1.22⁺ (the repo is declared `go 1.23` so any future‑stable 1.23 toolchain is perfect).

Check:

```bash
go version  # go1.23.x
````

### 2  Clone & configure

```bash
git clone https://github.com/yencarnacion/ntfy-hey-generator.git
cd ntfy-hey-generator

cp env.example .env   # then edit with your favourite editor
```

`.env` keys

| Variable          | Example              | Description                                    |
| ----------------- | -------------------- | ---------------------------------------------- |
| `NTFY_SERVER_URL` | `ntfy.sh`            | Host or IP where ntfy is running               |
| `NTFY_PORT`       | `80`                 | Port (usually 80 / 443)                        |
| `NTFY_TOPICS`     | `doorbell,frontgate` | Comma‑separated list of topics to monitor      |
| `MP3_FILE`        | `sounds/hey.mp3`     | Relative/absolute path to the MP3 to be played |

### 3  Download Go dependencies *(first run only)*

The simplest way is to let the helper script do it for you:

```bash
./go.sh   # runs `go run main.go` and `go` will auto‑download modules
```

If you prefer manual control:

```bash
go mod tidy   # resolves & caches all dependencies
```

### 4  Run

```bash
# EITHER
./go.sh          # convenient wrapper

# OR
export $(grep -v '^#' .env | xargs)   # optional – if you don’t want to keep a .env

go run .         # same as go run main.go
```

Press **Ctrl‑C** and the process exits instantly after closing all sockets & the audio device.

---

## 🛠  Development tips

* **Change the sound** — replace `sounds/hey.mp3` with any low‑latency clip.
* **Recompile to a static binary**:

  ```bash
  go build -o ntfy-audio-subscriber
  ./ntfy-audio-subscriber
  ```
* **Linux audio headers** — building on fresh servers may require:

  ```bash
  sudo apt install -y build-essential pkg-config libasound2-dev
  ```

---

## Contributing

Pull requests are welcome!  Please run `go vet ./...` and `go test ./...` (if tests exist) before submitting.

---

## License

Released under the **MIT License**.  See [`LICENSE`](LICENSE) for details.
