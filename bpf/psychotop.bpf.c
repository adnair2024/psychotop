// +build ignore

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

char __license[] SEC("license") = "Dual MIT/GPL";

struct syscall_info {
	__u64 count;
	char comm[16];
};

struct {
	__uint(type, BPF_MAP_TYPE_LRU_HASH);
	__uint(max_entries, 10240);
	__type(key, __u32);   // PID
	__type(value, struct syscall_info);
} syscall_counts SEC(".maps");

SEC("tracepoint/raw_syscalls/sys_enter")
int count_syscalls(void *ctx) {
	__u32 pid = bpf_get_current_pid_tgid() >> 32;
	struct syscall_info *info;

	info = bpf_map_lookup_elem(&syscall_counts, &pid);
	if (info) {
		__sync_fetch_and_add(&info->count, 1);
		// Update comm in case it changed (rare for same PID but possible)
		bpf_get_current_comm(&info->comm, sizeof(info->comm));
	} else {
		struct syscall_info initial = { .count = 1 };
		bpf_get_current_comm(&initial.comm, sizeof(initial.comm));
		bpf_map_update_elem(&syscall_counts, &pid, &initial, BPF_NOEXIST);
	}

	return 0;
}
