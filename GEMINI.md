# PSYCHOTOP // [GOLANG EDITION]
### // Kernel-Level Entropy & Integrity Monitor
**ID:** `AS_NAIR_0X77_GO`  
**Host:** `Ubuntu 24.04 (Noble Numbat)`  
**Stack:** `Go / eBPF (Cilium) / Bubble Tea`

---

## 01. ARCHITECTURE
* **Backend:** A Go-based binary that uses `bpf2go` to embed compiled C probes.
* **Event Loop:** Goroutines handle the heavy lifting—streaming raw syscall data from the kernel into an internal `tea.Msg` stream.
* **Integrity Checks:** Leverages Go's fast standard library for real-time SHA-256 hashing of your archival directories.

## 02. NOIR INTERFACE (BUBBLE TEA)
* Uses **Lip Gloss** for matte black styling and "Digital Noir" border definitions.
* Implements a **Hilbert Curve** generator that converts syscall density into visual entropy maps via Kitty's `icat` protocol.
* **Fuzzy Search:** Integrated `fzf`-like filtering for PIDs and active network sockets.

## 03. INSTALLATION
```bash
# Clone the Go integrity engine
git clone [https://github.com/ashwinnair/psychotop](https://github.com/ashwinnair/psychotop)

# Compile and embed BPF objects
go generate ./...
go build -o psychotop

# Run with sudo to hook into the kernel
```

sudo ./psychotop --noir-mode
