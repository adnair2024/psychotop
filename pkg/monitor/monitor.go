package monitor

import (
	"fmt"
	"log"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

// SyscallEntry represents a single process's monitoring data.
type SyscallEntry struct {
	Count uint64
	Comm  string
}

// SyscallData represents the counts of syscalls per PID.
type SyscallData map[uint32]SyscallEntry

// Monitor handles the lifecycle of the eBPF program.
type Monitor struct {
	objs bpfObjects
	link link.Link
}

// NewMonitor initializes and loads the eBPF program.
func NewMonitor() (*Monitor, error) {
	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("failed to remove memlock: %v", err)
	}

	// Load pre-compiled programs and maps into the kernel.
	var objs bpfObjects
	if err := loadBpfObjects(&objs, nil); err != nil {
		return nil, fmt.Errorf("failed to load bpf objects: %v", err)
	}

	// Attach the program to a tracepoint.
	tp, err := link.Tracepoint("raw_syscalls", "sys_enter", objs.CountSyscalls, nil)
	if err != nil {
		objs.Close()
		return nil, fmt.Errorf("failed to attach tracepoint: %v", err)
	}

	return &Monitor{
		objs: objs,
		link: tp,
	}, nil
}

// Close cleans up eBPF resources.
func (m *Monitor) Close() error {
	var errs []error

	if err := m.link.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close link: %v", err))
	}

	if err := m.objs.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close bpf objects: %v", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}
	return nil
}

// GetCounts retrieves the current syscall counts from the BPF map.
func (m *Monitor) GetCounts() (SyscallData, error) {
	var (
		key    uint32
		value  bpfSyscallInfo
		counts = make(SyscallData)
	)

	iter := m.objs.SyscallCounts.Iterate()
	for iter.Next(&key, &value) {
		counts[key] = SyscallEntry{
			Count: value.Count,
			Comm:  int8ToString(value.Comm),
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate map: %v", err)
	}

	return counts, nil
}

func int8ToString(arr [16]int8) string {
	b := make([]byte, 0, len(arr))
	for _, v := range arr {
		if v == 0 {
			break
		}
		b = append(b, byte(v))
	}
	return string(b)
}

// StartStreaming sends syscall data updates over a channel at the specified interval.
func (m *Monitor) StartStreaming(interval time.Duration, dataChan chan<- SyscallData, stopChan <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			counts, err := m.GetCounts()
			if err != nil {
				log.Printf("Error getting syscall counts: %v", err)
				continue
			}
			select {
			case dataChan <- counts:
			case <-stopChan:
				return
			}
		case <-stopChan:
			return
		}
	}
}
