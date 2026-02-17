package main

import (
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResolveStatePath_UsesXDGStateHome(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/tmp/custom-state")

	path, err := resolveStatePath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/tmp/custom-state", "github.com/darrenburns/terma", "todo-app", "state.json"), path)
}

func TestResolveStatePath_UsesHomeLocalState(t *testing.T) {
	origHome := userHomeDir
	origConfig := userConfigDir
	t.Cleanup(func() {
		userHomeDir = origHome
		userConfigDir = origConfig
	})

	t.Setenv("XDG_STATE_HOME", "")
	userHomeDir = func() (string, error) {
		return "/tmp/fake-home", nil
	}
	userConfigDir = func() (string, error) {
		return "", errors.New("should not be called")
	}

	path, err := resolveStatePath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/tmp/fake-home", ".local", "state", "github.com/darrenburns/terma", "todo-app", "state.json"), path)
}

func TestResolveStatePath_FallsBackToConfigDir(t *testing.T) {
	origHome := userHomeDir
	origConfig := userConfigDir
	t.Cleanup(func() {
		userHomeDir = origHome
		userConfigDir = origConfig
	})

	t.Setenv("XDG_STATE_HOME", "")
	userHomeDir = func() (string, error) {
		return "", errors.New("home unavailable")
	}
	userConfigDir = func() (string, error) {
		return "/tmp/fake-config", nil
	}

	path, err := resolveStatePath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/tmp/fake-config", "github.com/darrenburns/terma", "todo-app", "state.json"), path)
}

func TestLoadState_MissingFileReturnsNil(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	state, err := loadState(path)
	require.NoError(t, err)
	require.Nil(t, state)
}

func TestLoadState_CorruptFileReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	require.NoError(t, os.WriteFile(path, []byte("{not-json"), 0o600))

	state, err := loadState(path)
	require.Error(t, err)
	require.Nil(t, state)
}

func TestSaveAndLoadState_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	in := &todoStateV1{
		SchemaVersion: stateSchemaVersion,
		Theme:         "kanagawa",
		ActiveListID:  "inbox",
		Lists: []persistedList{
			{
				ID:   "today",
				Name: "Today",
				Tasks: []persistedTask{
					{
						ID:        "task-aaaaaa000001",
						Title:     "one",
						Completed: false,
						CreatedAt: time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC),
					},
				},
			},
			{
				ID:   "inbox",
				Name: "Inbox",
				Tasks: []persistedTask{
					{
						ID:        "task-aaaaaa000002",
						Title:     "two",
						Completed: true,
						CreatedAt: time.Date(2026, 2, 1, 11, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	}

	require.NoError(t, saveState(path, in))

	out, err := loadState(path)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, in.SchemaVersion, out.SchemaVersion)
	require.Equal(t, in.Theme, out.Theme)
	require.Equal(t, in.ActiveListID, out.ActiveListID)
	require.Equal(t, in.Lists, out.Lists)
}

func TestScheduleSave_DebouncesBurstWrites(t *testing.T) {
	origWrite := writeStateFile
	t.Cleanup(func() {
		writeStateFile = origWrite
	})

	var writes atomic.Int32
	writeStateFile = func(path string, state *todoStateV1) error {
		writes.Add(1)
		return saveState(path, state)
	}

	app := NewTodoApp()
	app.statePath = filepath.Join(t.TempDir(), "state.json")
	app.saveDelay = 40 * time.Millisecond

	app.scheduleSave()
	app.scheduleSave()
	app.scheduleSave()

	time.Sleep(180 * time.Millisecond)
	require.Equal(t, int32(1), writes.Load())
}

func TestFlushSave_PersistsPendingStateImmediately(t *testing.T) {
	app := NewTodoApp()
	app.statePath = filepath.Join(t.TempDir(), "state.json")
	app.saveDelay = 5 * time.Second

	app.addTask("persist me")
	app.flushSave()

	state, err := loadState(app.statePath)
	require.NoError(t, err)
	require.NotNil(t, state)

	found := false
	for _, list := range state.Lists {
		for _, task := range list.Tasks {
			if task.Title == "persist me" {
				found = true
				break
			}
		}
	}
	require.True(t, found)
}
