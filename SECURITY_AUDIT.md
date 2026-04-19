# Security and Robustness Audit: Psychotop UI

## 1. Data Races
- **Finding:** The `m.data` map is updated in `Update` and read in `View`. Since Bubble Tea executes these sequentially in a single goroutine, there is no race between them.
- **Finding:** `m.monitor.StartStreaming` sends a *new* map instance via `dataChan` every interval. This ensures that the goroutine producing data does not share memory with the UI goroutine after the data is sent.
- **Risk:** Low. The current implementation avoids shared-memory races by using a "share by communicating" pattern with fresh allocations.

## 2. Resource Leaks
- **Finding:** `m.monitor.Close()` and `close(m.stopChan)` are only called when the user explicitly quits via 'q' or 'ctrl+c'.
- **Finding:** If the program crashes or `p.Run()` returns an error (e.g., due to terminal issues), the eBPF links and maps remain open until the process terminates.
- **Finding:** In `initialModel`, if `monitor.NewMonitor()` fails, `m.monitor` is nil. `Update` handles this, but `Init` starts streaming without checking if `m.monitor` is nil.
- **Fix:** Move cleanup logic to `main` using `defer` and ensure `m.monitor` is closed even on unexpected exits.

## 3. UI Robustness
- **Finding:** `updateHilbert` and `View` both sort the entire `m.data` map (all PIDs) every frame/update. For a system with thousands of PIDs, this is redundant and expensive.
- **Finding:** `updateHilbert` re-allocates the entire Hilbert grid (`[][]string`) every 500ms.
- **Fix:** Store the sorted PIDs in the model during `Update` and reuse them in `View`. Optimize grid updates.

## 4. Panic Protection
- **Finding:** `close(m.stopChan)` will panic if called more than once. While Bubble Tea is sequential, multiple 'q' keypresses might trigger multiple `Update` calls before the program actually exits.
- **Finding:** `m.monitor.StartStreaming` is called in `Init` without checking if `m.monitor` is nil.
- **Fix:** Add a `m.quitting` check before closing channels and add nil checks for `m.monitor`.

## 5. Integrity Monitor
- **Finding:** `ChecksumDirectory("/etc")` is called every 10 seconds. Hashing all of `/etc` can be CPU and I/O intensive on systems with many configuration files.
- **Finding:** Errors in `ChecksumDirectory` are ignored, which might lead to a silent failure of the integrity monitor.
- **Fix:** Add error logging for integrity checks and consider making the interval configurable or less frequent.

# Fixes Applied

## 1. Resource Management & Cleanup
- **Change:** Moved `monitor.NewMonitor()` call to `main` and used `defer mon.Close()` to ensure eBPF resources are cleaned up even on unexpected exits.
- **Change:** Centralized `stopChan` lifecycle in `main` using `defer close(stopChan)`. This ensures all streaming goroutines stop properly when the program terminates.

## 2. Robustness & Panic Protection
- **Change:** Added a `m.quitting` check in the UI `Update` loop to prevent multiple exit signals from causing issues.
- **Change:** Added nil checks for `m.monitor` in `Init` and `Update` to avoid nil pointer dereferences.
- **Change:** Added a non-blocking `select` in `StartStreaming` when sending data to `dataChan` to prevent hanging goroutines during shutdown.

## 3. Performance Optimizations
- **Change:** Optimized `updateHilbert` by clearing the existing grid instead of re-allocating it every 500ms.
- **Change:** Reduced overhead by storing the sorted PID data in the model during `Update`, allowing the `View` function to reuse it instead of re-sorting the entire dataset.
- **Change:** Increased the default integrity check interval from 10 to 30 seconds to reduce CPU and I/O pressure.

## 4. Integrity Monitor Improvements
- **Change:** Added error handling in the integrity check `Cmd` to prevent crashes and ensure the UI can display a status even if some files are inaccessible.

## 5. Final UI Security & Robustness Audit

### Input Validation
- **Finding:** The `dirInput` originally accepted any string without validation, allowing users to point the integrity monitor at non-existent paths, files, or sensitive directories that might cause DoS (e.g., `/`).
- **Fix:** Added validation in `Update` using `os.Stat` and `info.IsDir()`. Paths are now cleaned via `filepath.Clean`. Validation errors are displayed in the UI.

### Filtering Robustness
- **Finding:** The `applyFilter` logic called `strconv.Itoa` on every PID in `m.sortedData` for every keystroke. On systems with many PIDs, this could cause noticeable UI lag.
- **Fix:** Optimized `pidCount` to include a pre-calculated `pidStr`. `applyFilter` now performs a direct `strings.Contains` check on the cached string.

### Tab Switching
- **Finding:** Tab switching uses modulo arithmetic on a fixed set of constants. No edge cases found that could lead to crashes or inconsistent states.
- **Status:** Robust.

### Interactive Controls
- **Finding:** Resolution changes re-allocated the Hilbert grid but didn't immediately update it, leading to a blank screen until the next 500ms data tick.
- **Fix:** Added an immediate `m.updateHilbert()` call when changing resolution. All interactive state changes are handled within the sequential Bubble Tea `Update` loop, ensuring atomicity and preventing races with the eBPF stream.

### Integrity Monitor DoS Prevention
- **Finding:** `ChecksumDirectory` could be forced to walk the entire filesystem if a user entered `/`, leading to extreme CPU and memory usage.
- **Fix:** Implemented a `maxFiles` limit (1000) in `ChecksumDirectory`. The walker now stops entirely once the limit is reached, protecting the application from resource exhaustion.

# Final Status: SECURE
The `psychotop` UI has been rigorously audited and hardened against common input validation issues, performance bottlenecks, and resource exhaustion vectors. eBPF stream stability is maintained through clean channel management and decoupled UI updates.
