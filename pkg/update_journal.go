package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	UpdateJournalTomlName   = "openscadgen_updates.toml"
	UpdateJournalLogName    = "openscadgen_updates.log"
	UpdateJournalMaxLogSize = 256 * 1024
	UpdateJournalTrimSize   = 64 * 1024
)

type UpdateJournalEntry struct {
	At      string
	Name    string
	From    string
	To      string
	Success bool
	Details string
}

type UpdateJournalState struct {
	UpdatedAt string
	Current   map[string]string
	History   []UpdateJournalEntry
}

func RecordUpdateJournal(rootDir, name, fromState, toState, details string, success bool) error {
	rootDir = filepath.Clean(rootDir)
	entry := UpdateJournalEntry{
		At:      time.Now().Format(time.RFC3339),
		Name:    name,
		From:    fromState,
		To:      toState,
		Success: success,
		Details: details,
	}

	statePath := filepath.Join(rootDir, UpdateJournalTomlName)
	logPath := filepath.Join(rootDir, UpdateJournalLogName)

	state, err := loadUpdateJournalState(statePath)
	if err != nil {
		return err
	}
	if state.Current == nil {
		state.Current = map[string]string{}
	}
	state.UpdatedAt = entry.At
	state.Current[name] = toState
	if success {
		state.History = append(state.History, entry)
	}

	if err := writeUpdateJournalState(statePath, state); err != nil {
		return err
	}
	if err := appendUpdateJournalLog(logPath, entry, success); err != nil {
		return err
	}

	if success {
		LogStagef("update", "%s updated: %s -> %s", name, fromState, toState)
	} else {
		LogStagef("update", "%s update failed: %s -> %s", name, fromState, toState)
	}
	if strings.TrimSpace(details) != "" {
		LogInfof("update details: %s", details)
	}
	return nil
}

func loadUpdateJournalState(path string) (UpdateJournalState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return UpdateJournalState{Current: map[string]string{}}, nil
		}
		return UpdateJournalState{}, err
	}
	return parseUpdateJournalState(string(data)), nil
}

func writeUpdateJournalState(path string, state UpdateJournalState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("[meta]\n")
	fmt.Fprintf(&b, "updated_at = %q\n\n", state.UpdatedAt)

	b.WriteString("[current]\n")
	keys := make([]string, 0, len(state.Current))
	for k := range state.Current {
		keys = append(keys, k)
	}
	sortStrings(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "%q = %q\n", k, state.Current[k])
	}
	b.WriteString("\n")
	for _, entry := range state.History {
		b.WriteString("[[history]]\n")
		fmt.Fprintf(&b, "at = %q\n", entry.At)
		fmt.Fprintf(&b, "name = %q\n", entry.Name)
		fmt.Fprintf(&b, "from = %q\n", entry.From)
		fmt.Fprintf(&b, "to = %q\n", entry.To)
		fmt.Fprintf(&b, "success = %t\n", entry.Success)
		fmt.Fprintf(&b, "details = %q\n\n", entry.Details)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func appendUpdateJournalLog(path string, entry UpdateJournalEntry, success bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	line := fmt.Sprintf("%s [%s] %s: %s -> %s | %s\n",
		entry.At,
		map[bool]string{true: "success", false: "failed"}[success],
		entry.Name,
		entry.From,
		entry.To,
		entry.Details,
	)
	if err := truncateIfNeeded(path, UpdateJournalMaxLogSize, UpdateJournalTrimSize); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line)
	return err
}

func truncateIfNeeded(path string, maxSize int64, trimSize int64) error {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	if info.Size() <= maxSize {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if int64(len(data)) <= trimSize {
		return nil
	}
	keep := data[len(data)-int(trimSize):]
	return os.WriteFile(path, keep, 0o644)
}

func parseUpdateJournalState(data string) UpdateJournalState {
	state := UpdateJournalState{Current: map[string]string{}}
	lines := strings.Split(data, "\n")
	section := ""
	var currentEntry *UpdateJournalEntry
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		switch line {
		case "[meta]":
			section = "meta"
			continue
		case "[current]":
			section = "current"
			continue
		case "[[history]]":
			section = "history"
			state.History = append(state.History, UpdateJournalEntry{})
			currentEntry = &state.History[len(state.History)-1]
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		switch section {
		case "meta":
			if key == "updated_at" {
				state.UpdatedAt = value
			}
		case "current":
			state.Current[strings.Trim(key, `"`)] = value
		case "history":
			if currentEntry == nil {
				continue
			}
			switch key {
			case "at":
				currentEntry.At = value
			case "name":
				currentEntry.Name = value
			case "from":
				currentEntry.From = value
			case "to":
				currentEntry.To = value
			case "details":
				currentEntry.Details = value
			case "success":
				currentEntry.Success = value == "true"
			}
		}
	}
	return state
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
