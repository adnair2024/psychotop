package monitor

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target amd64 bpf ../../bpf/psychotop.bpf.c -- -I/usr/include/bpf -I/usr/include -I/usr/include/x86_64-linux-gnu
