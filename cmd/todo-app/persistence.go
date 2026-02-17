package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	t "github.com/darrenburns/terma"
)

const stateSchemaVersion = 1

var userHomeDir = os.UserHomeDir
var userConfigDir = os.UserConfigDir
var writeStateFile = saveState

type todoStateV1 struct {
	SchemaVersion int             `json:"schema_version"`
	Theme         string          `json:"theme"`
	ActiveListID  string          `json:"active_list_id"`
	Lists         []persistedList `json:"lists"`
}

type persistedList struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Tasks []persistedTask `json:"tasks"`
}

type persistedTask struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

func resolveStatePath() (string, error) {
	if xdgStateHome := strings.TrimSpace(os.Getenv("XDG_STATE_HOME")); xdgStateHome != "" {
		return filepath.Join(xdgStateHome, "github.com/darrenburns/terma", "todo-app", "state.json"), nil
	}

	homeDir, err := userHomeDir()
	if err == nil && homeDir != "" {
		return filepath.Join(homeDir, ".local", "state", "github.com/darrenburns/terma", "todo-app", "state.json"), nil
	}

	configDir, configErr := userConfigDir()
	if configErr != nil || configDir == "" {
		if err != nil {
			return "", fmt.Errorf("resolve state path: home=%w config=%w", err, configErr)
		}
		return "", fmt.Errorf("resolve state path: user config dir unavailable")
	}

	return filepath.Join(configDir, "github.com/darrenburns/terma", "todo-app", "state.json"), nil
}

func loadState(path string) (*todoStateV1, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read state: %w", err)
	}

	var state todoStateV1
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode state: %w", err)
	}

	if err := normalizeStateV1(&state); err != nil {
		return nil, err
	}

	return &state, nil
}

func saveState(path string, state *todoStateV1) error {
	if state == nil {
		return errors.New("save state: nil state")
	}
	if path == "" {
		return errors.New("save state: empty path")
	}

	state.SchemaVersion = stateSchemaVersion

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	payload = append(payload, '\n')

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, payload, 0o600); err != nil {
		return fmt.Errorf("write temp state: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(path)
		if errRetry := os.Rename(tempPath, path); errRetry != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("replace state file: %w", errRetry)
		}
	}

	return nil
}

func normalizeStateV1(state *todoStateV1) error {
	if state == nil {
		return errors.New("state is nil")
	}
	if state.SchemaVersion == 0 {
		return errors.New("missing schema_version")
	}
	if state.SchemaVersion != stateSchemaVersion {
		return fmt.Errorf("unsupported schema_version %d", state.SchemaVersion)
	}
	if len(state.Lists) == 0 {
		return errors.New("state has no lists")
	}

	listIDs := make(map[string]struct{}, len(state.Lists))
	for i := range state.Lists {
		list := &state.Lists[i]
		if strings.TrimSpace(list.ID) == "" {
			list.ID = fmt.Sprintf("list-%d", i+1)
		}
		if _, exists := listIDs[list.ID]; exists {
			return fmt.Errorf("duplicate list id %q", list.ID)
		}
		listIDs[list.ID] = struct{}{}

		if strings.TrimSpace(list.Name) == "" {
			list.Name = list.ID
		}

		taskIDs := make(map[string]struct{}, len(list.Tasks))
		for j := range list.Tasks {
			task := &list.Tasks[j]
			if strings.TrimSpace(task.ID) == "" {
				task.ID = fmt.Sprintf("%s-task-%d", list.ID, j+1)
			}
			if _, exists := taskIDs[task.ID]; exists {
				return fmt.Errorf("duplicate task id %q in list %q", task.ID, list.ID)
			}
			taskIDs[task.ID] = struct{}{}

			if task.CreatedAt.IsZero() {
				task.CreatedAt = time.Now().UTC()
			}
		}
	}

	if state.ActiveListID == "" {
		state.ActiveListID = state.Lists[0].ID
	}
	if _, ok := listIDs[state.ActiveListID]; !ok {
		state.ActiveListID = state.Lists[0].ID
	}

	if state.Theme == "" {
		state.Theme = t.ThemeNameKanagawa
	} else if _, ok := t.GetTheme(state.Theme); !ok {
		state.Theme = t.ThemeNameKanagawa
	}

	return nil
}

func (a *TodoApp) initializePersistence() {
	t.SetTheme(t.ThemeNameKanagawa)

	path, err := resolveStatePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "todo-app: could not resolve state path: %v\n", err)
		return
	}
	a.statePath = path

	state, err := loadState(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "todo-app: could not load state from %s: %v\n", path, err)
		return
	}
	if state == nil {
		return
	}

	a.applyPersistentState(state)
	t.SetTheme(state.Theme)
}

func (a *TodoApp) toPersistentState() *todoStateV1 {
	activeListID := ""
	if len(a.taskLists) > 0 {
		activeListID = a.activeList().ID
	}

	state := &todoStateV1{
		SchemaVersion: stateSchemaVersion,
		Theme:         t.CurrentThemeName(),
		ActiveListID:  activeListID,
		Lists:         make([]persistedList, 0, len(a.taskLists)),
	}

	for _, list := range a.taskLists {
		tasks := list.Tasks.GetItems()
		persistedTasks := make([]persistedTask, 0, len(tasks))
		for _, task := range tasks {
			persistedTasks = append(persistedTasks, persistedTask(task))
		}

		state.Lists = append(state.Lists, persistedList{
			ID:    list.ID,
			Name:  list.Name,
			Tasks: persistedTasks,
		})
	}

	return state
}

func (a *TodoApp) applyPersistentState(state *todoStateV1) {
	if state == nil || len(state.Lists) == 0 {
		return
	}

	taskLists := make([]*TaskList, 0, len(state.Lists))
	activeListIdx := 0

	for i, list := range state.Lists {
		tasks := make([]Task, 0, len(list.Tasks))
		for _, task := range list.Tasks {
			tasks = append(tasks, Task(task))
		}

		taskLists = append(taskLists, &TaskList{
			ID:          list.ID,
			Name:        list.Name,
			Tasks:       t.NewListState(tasks),
			ScrollState: t.NewScrollState(),
		})

		if list.ID == state.ActiveListID {
			activeListIdx = i
		}
	}

	a.taskLists = taskLists
	a.activeListIdx.Set(activeListIdx)
	a.refreshTagSuggestions()
	a.refreshFilteredTasks()
}

func (a *TodoApp) scheduleSave() {
	if a.statePath == "" {
		return
	}

	snapshot := a.toPersistentState()

	a.saveMu.Lock()
	a.pendingSave = snapshot
	if a.saveTimer == nil {
		a.saveTimer = time.AfterFunc(a.saveDelay, a.flushPendingSave)
	} else {
		a.saveTimer.Reset(a.saveDelay)
	}
	a.saveMu.Unlock()
}

func (a *TodoApp) flushPendingSave() {
	a.saveMu.Lock()
	defer a.saveMu.Unlock()

	if a.pendingSave == nil || a.statePath == "" {
		return
	}

	if err := writeStateFile(a.statePath, a.pendingSave); err != nil {
		fmt.Fprintf(os.Stderr, "todo-app: failed to save state: %v\n", err)
		return
	}

	a.pendingSave = nil
}

func (a *TodoApp) flushSave() {
	if a.statePath == "" {
		return
	}

	a.saveMu.Lock()
	if a.saveTimer != nil {
		a.saveTimer.Stop()
	}
	a.pendingSave = a.toPersistentState()
	a.saveMu.Unlock()

	a.flushPendingSave()
}
