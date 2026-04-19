# PSYCHOTOP

**A Kernel-Level Integrity Monitor & Entropy Visualizer for Linux.**

Built with **Go**, **eBPF**, and **Bubble Tea**, Psychotop provides a "Digital Noir" terminal interface for real-time forensic auditing. It maps system-wide syscall density onto a Hilbert Curve to visualize behavioral entropy and performs continuous integrity checks on critical system directories.

---

## 🚀 Features

- **Real-time eBPF Tracing:** High-performance syscall monitoring using kernel tracepoints.
- **Entropy Mapping:** Visualizes process activity patterns using a Hilbert Curve.
- **Integrity Engine:** Continuous SHA-256 verification of critical directories (e.g., `/etc`) with symlink and DoS protection.
- **Noir Interface:** A tabbed TUI with fuzzy search, resolution controls, and a contextual help system.
- **Smart Filtering:** Instant filtering of the process list by PID or command name (e.g., "python", "systemd").

---

## 🛠 Installation

### 1. Prerequisites

Ensure you are running **Ubuntu 24.04** (or a similar modern Linux kernel) and have the necessary build tools:

```bash
sudo apt-get update
sudo apt-get install -y clang llvm libbpf-dev linux-tools-$(uname -r) linux-headers-$(uname -r)
```

### 2. Build from Source

```bash
# Clone the repository
git clone https://github.com/ashwinnair/psychotop
cd psychotop

# Generate BPF bindings and build
go generate ./...
go build -o psychotop ./cmd/psychotop
```

---

## 🕹 Usage

Run with `sudo` to allow the eBPF probes to hook into the kernel:

```bash
sudo ./psychotop
```

### Keybindings

| Key | Action |
|-----|--------|
| `← / →` | Switch Tabs (Entropy Map, Process List, Integrity Logs) |
| `/` | Focus search bar (in Process List) |
| `p` | Pause/Resume kernel monitoring |
| `+ / -` | Increase/Decrease Hilbert Map resolution |
| `d` | Change monitored directory for integrity checks |
| `?` | Toggle help menu |
| `q` | Quit |

---

## 🏗 Architecture

- **Backend:** Go-based event loop handling eBPF map data via `cilium/ebpf`.
- **Probes:** C-based BPF programs utilizing `LRU_HASH` maps for thread-safe syscall counting.
- **UI:** Interactive terminal rendering powered by `charmbracelet/bubbletea` and `lipgloss`.

---

## 🔒 Security

Psychotop is built with security as a priority. All core features undergo automated security audits, including checks for:
- Integer overflows and OOM protection in visualization logic.
- Atomic race conditions in BPF maps.
- Symlink and directory traversal protection in the integrity engine.

See `SECURITY_AUDIT.md` for detailed findings.
