package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ashwinnair/psychotop/pkg/integrity"
	"github.com/ashwinnair/psychotop/pkg/monitor"
	"github.com/ashwinnair/psychotop/pkg/ui"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Tab types
type tab int

const (
	tabEntropy tab = iota
	tabProcesses
	tabIntegrity
)

// KeyMap defines the application's keybindings.
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Help      key.Binding
	Quit      key.Binding
	Pause     key.Binding
	IncOrder  key.Binding
	DecOrder  key.Binding
	Directory key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Up, k.Down},
		{k.Pause, k.IncOrder, k.DecOrder, k.Directory},
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "prev tab"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next tab"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Pause: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "pause/resume"),
	),
	IncOrder: key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "inc resolution"),
	),
	DecOrder: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "dec resolution"),
	),
	Directory: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "change dir"),
	),
}

// Styles
var (
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 2).
			MarginRight(1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Padding(0, 2).
				MarginRight(1)

	windowStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(1).
			MarginTop(1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	integrityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
)

type syscallMsg monitor.SyscallData
type integrityMsg []integrity.FileHash

type pidCount struct {
	pid    uint32
	pidStr string
	comm   string
	cnt    uint64
}

type model struct {
	monitor      *monitor.Monitor
	data         monitor.SyscallData
	sortedData   []pidCount
	filteredData []pidCount
	integrity    []integrity.FileHash
	activeTab    tab
	paused       bool
	quitting     bool
	help         help.Model
	filterInput  textinput.Model
	dirInput     textinput.Model
	enteringDir  bool
	monitoredDir string
	dataChan     chan monitor.SyscallData
	stopChan     chan struct{}
	hilbertMap   [][]string
	order        int
	showHelp     bool
	err          error
}

func initialModel(mon *monitor.Monitor, stopChan chan struct{}) model {
	ti := textinput.New()
	ti.Placeholder = "Filter (e.g. \"python\" or \"1234\")"
	ti.CharLimit = 30
	ti.Width = 30

	di := textinput.New()
	di.Placeholder = "Enter directory path..."
	di.Width = 50

	order := 4
	return model{
		monitor:      mon,
		data:         make(monitor.SyscallData),
		activeTab:    tabEntropy,
		help:         help.New(),
		filterInput:  ti,
		dirInput:     di,
		monitoredDir: "/etc",
		dataChan:     make(chan monitor.SyscallData),
		stopChan:     stopChan,
		order:        order,
		hilbertMap:   makeHilbertGrid(order),
	}
}

func makeHilbertGrid(order int) [][]string {
	n := 1 << order
	grid := make([][]string, n)
	for i := range grid {
		grid[i] = make([]string, n)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}
	return grid
}

func (m model) Init() tea.Cmd {
	if m.monitor == nil {
		return tea.Quit
	}
	go m.monitor.StartStreaming(500*time.Millisecond, m.dataChan, m.stopChan)
	return tea.Batch(
		waitForSyscalls(m.dataChan),
		checkIntegrity(m.monitoredDir),
	)
}

func waitForSyscalls(sub chan monitor.SyscallData) tea.Cmd {
	return func() tea.Msg {
		return syscallMsg(<-sub)
	}
}

func checkIntegrity(dir string) tea.Cmd {
	return func() tea.Msg {
		hashes, err := integrity.ChecksumDirectory(dir)
		if err != nil {
			return integrityMsg(nil)
		}
		return integrityMsg(hashes)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.enteringDir {
			switch msg.String() {
			case "enter":
				path := filepath.Clean(m.dirInput.Value())
				info, err := os.Stat(path)
				if err != nil {
					m.err = fmt.Errorf("invalid path: %v", err)
					return m, nil
				}
				if !info.IsDir() {
					m.err = fmt.Errorf("not a directory: %s", path)
					return m, nil
				}
				m.monitoredDir = path
				m.enteringDir = false
				m.err = nil
				m.dirInput.Blur()
				return m, checkIntegrity(m.monitoredDir)
			case "esc":
				m.enteringDir = false
				m.err = nil
				m.dirInput.Blur()
				return m, nil
			}
			m.dirInput, _ = m.dirInput.Update(msg)
			return m, nil
		}

		if m.activeTab == tabProcesses && m.filterInput.Focused() {
			if msg.String() == "esc" || msg.String() == "enter" {
				m.filterInput.Blur()
				return m, nil
			}
			m.filterInput, _ = m.filterInput.Update(msg)
			m.applyFilter()
			return m, nil
		}

		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, keys.Left):
			m.activeTab = (m.activeTab - 1 + 3) % 3

		case key.Matches(msg, keys.Right):
			m.activeTab = (m.activeTab + 1) % 3

		case key.Matches(msg, keys.Help):
			m.showHelp = !m.showHelp

		case key.Matches(msg, keys.Pause):
			m.paused = !m.paused

		case key.Matches(msg, keys.IncOrder):
			if m.order < 6 {
				m.order++
				m.hilbertMap = makeHilbertGrid(m.order)
				m.updateHilbert()
			}

		case key.Matches(msg, keys.DecOrder):
			if m.order > 2 {
				m.order--
				m.hilbertMap = makeHilbertGrid(m.order)
				m.updateHilbert()
			}

		case key.Matches(msg, keys.Directory):
			m.enteringDir = true
			m.dirInput.Focus()
			m.dirInput.SetValue(m.monitoredDir)
			m.err = nil

		case msg.String() == "/" && m.activeTab == tabProcesses:
			m.filterInput.Focus()
			return m, textinput.Blink
		}

	case syscallMsg:
		if !m.paused {
			m.data = monitor.SyscallData(msg)
			m.updateHilbert()
			m.applyFilter()
		}
		return m, waitForSyscalls(m.dataChan)

	case integrityMsg:
		m.integrity = []integrity.FileHash(msg)
		return m, tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
			return checkIntegrity(m.monitoredDir)()
		})
	}

	return m, tea.Batch(cmds...)
}

func (m *model) applyFilter() {
	filter := strings.ToLower(m.filterInput.Value())
	if filter == "" {
		m.filteredData = m.sortedData
		return
	}
	m.filteredData = nil
	for _, d := range m.sortedData {
		if strings.Contains(d.pidStr, filter) || strings.Contains(strings.ToLower(d.comm), filter) {
			m.filteredData = append(m.filteredData, d)
		}
	}
}

func (m *model) updateHilbert() {
	n := 1 << m.order
	for y := range m.hilbertMap {
		for x := range m.hilbertMap[y] {
			m.hilbertMap[y][x] = " "
		}
	}

	m.sortedData = make([]pidCount, 0, len(m.data))
	for p, entry := range m.data {
		m.sortedData = append(m.sortedData, pidCount{
			pid:    p,
			pidStr: strconv.Itoa(int(p)),
			comm:   entry.Comm,
			cnt:    entry.Count,
		})
	}
	sort.Slice(m.sortedData, func(i, j int) bool {
		return m.sortedData[i].cnt > m.sortedData[j].cnt
	})

	for i, pc := range m.sortedData {
		if i >= n*n {
			break
		}
		p := ui.Hilbert(n, i)
		char := "."
		if pc.cnt > 1000 {
			char = "█"
		} else if pc.cnt > 500 {
			char = "▓"
		} else if pc.cnt > 100 {
			char = "▒"
		} else if pc.cnt > 10 {
			char = "░"
		}
		m.hilbertMap[p.Y][p.X] = char
	}
}

func (m model) View() string {
	if m.quitting {
		return "  Shutting down PSYCHOTOP...\n"
	}

	var s strings.Builder

	// Header
	s.WriteString(titleStyle.Render(" PSYCHOTOP // "))
	if m.paused {
		s.WriteString(errorStyle.Render(" PAUSED "))
	} else {
		s.WriteString(integrityStyle.Render(" MONITORING "))
	}
	s.WriteString("\n\n")

	// Tabs
	tabs := []string{"ENTROPY MAP", "PROCESS LIST", "INTEGRITY LOGS"}
	for i, t := range tabs {
		style := inactiveTabStyle
		if m.activeTab == tab(i) {
			style = activeTabStyle
		}
		s.WriteString(style.Render(t))
	}
	s.WriteString("\n")

	// Main Window
	var content string
	switch m.activeTab {
	case tabEntropy:
		content = m.renderEntropy()
	case tabProcesses:
		content = m.renderProcesses()
	case tabIntegrity:
		content = m.renderIntegrity()
	}

	s.WriteString(windowStyle.Width(70).Render(content))
	s.WriteString("\n")

	// Input Overlay
	if m.enteringDir {
		s.WriteString("\n" + m.dirInput.View() + "\n")
		if m.err != nil {
			s.WriteString(errorStyle.Render(m.err.Error()) + "\n")
		}
	}

	// Help
	if m.showHelp {
		s.WriteString("\n" + m.help.View(keys) + "\n")
	} else {
		s.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("Press ? for help") + "\n")
	}

	return s.String()
}

func (m model) renderEntropy() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("Resolution: %dx%d (order %d)\n\n", 1<<m.order, 1<<m.order, m.order))
	for _, row := range m.hilbertMap {
		s.WriteString(strings.Join(row, ""))
		s.WriteString("\n")
	}
	return s.String()
}

func (m model) renderProcesses() string {
	var s strings.Builder
	s.WriteString(m.filterInput.View() + " (Press / to focus)\n\n")
	s.WriteString(fmt.Sprintf("%-10s %-15s %-15s\n", "PID", "COMMAND", "SYSCALLS"))
	s.WriteString(strings.Repeat("-", 42) + "\n")
	for i, pc := range m.filteredData {
		if i >= 15 {
			break
		}
		comm := pc.comm
		if len(comm) > 14 {
			comm = comm[:11] + "..."
		}
		s.WriteString(fmt.Sprintf("%-10d %-15s %-15d\n", pc.pid, comm, pc.cnt))
	}
	return s.String()
}

func (m model) renderIntegrity() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("Directory: %s\n", m.monitoredDir))
	s.WriteString(fmt.Sprintf("Status: %d files verified\n\n", len(m.integrity)))
	for i, f := range m.integrity {
		if i >= 15 {
			s.WriteString("...\n")
			break
		}
		s.WriteString(fmt.Sprintf("✔ %s\n", f.Path))
	}
	return s.String()
}

func main() {
	mon, err := monitor.NewMonitor()
	if err != nil {
		fmt.Printf("Error initializing monitor: %v\nTry running with sudo.\n", err)
		os.Exit(1)
	}
	defer mon.Close()

	stopChan := make(chan struct{})
	defer func() {
		defer func() { recover() }()
		close(stopChan)
	}()

	p := tea.NewProgram(initialModel(mon, stopChan))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting psychotop: %v", err)
		os.Exit(1)
	}
}
