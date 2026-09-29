package terma

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunExternal_WithoutAppRunsCommand(t *testing.T) {
	var out bytes.Buffer
	cmd := exec.Command("sh", "-c", "echo external")
	cmd.Stdout = &out

	require.NoError(t, RunExternal(cmd))
	assert.Equal(t, "external\n", out.String())
	assert.Equal(t, os.Stdin, cmd.Stdin, "unset stdin is connected to the terminal")
	assert.Equal(t, os.Stderr, cmd.Stderr, "unset stderr is connected to the terminal")
}

func TestRunExternal_RunsInsideSuspendedUI(t *testing.T) {
	var steps []string
	setSuspender(func(fn func() error) error {
		steps = append(steps, "suspend")
		err := fn()
		steps = append(steps, "resume")
		return err
	})
	t.Cleanup(func() { setSuspender(nil) })

	var out bytes.Buffer
	cmd := exec.Command("sh", "-c", "exit 3")
	cmd.Stdout = &out
	err := RunExternal(cmd)

	var exitErr *exec.ExitError
	require.True(t, errors.As(err, &exitErr), "the command's error is returned")
	assert.Equal(t, 3, exitErr.ExitCode())
	assert.Equal(t, []string{"suspend", "resume"}, steps)
}
