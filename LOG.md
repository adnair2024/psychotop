# PSYCHOTOP IMPLEMENTATION LOG

## [2026-04-19] - Initial Setup & Rebranding
- **Action:** Rebranded project from `entropy-top` to `psychotop`.
- **Action:** Initialized Go module `github.com/ashwinnair/psychotop`.
- **Action:** Created directory structure: `cmd/`, `bpf/`, `pkg/ui/`, `pkg/monitor/`, `pkg/integrity/`.
- **Action:** Implemented basic Bubble Tea CLI entry point in `cmd/psychotop/main.go`.

## [2026-04-19] - Feature: User-Friendly UI Overhaul
- **Action:** Implemented tabbed interface (Entropy, Processes, Integrity).
- **Action:** Added fuzzy filtering for PIDs in the Process List.
- **Action:** Integrated interactive controls: pause/resume (`p`), resolution adjustment (`+/-`), and directory switching (`d`).
- **Action:** Added a contextual help menu (`?`) using `bubbles/help`.
- **Action:** Refined visual style with Lip Gloss borders and "Digital Noir" aesthetics.
- **Status:** Completed.
