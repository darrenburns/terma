package terma

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// After Quit, the event loop may still be painting its last frame, so a
// dispatch from another goroutine must not run there and then.
func TestDispatchIsDroppedWhileAppShutsDown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	setAppRuntimeState(ctx, newDispatchQueue())
	t.Cleanup(clearAppRuntimeState)
	cancel()

	ran := false
	Dispatch(func() { ran = true })
	require.False(t, ran, "dispatch ran during shutdown")

	clearAppRuntimeState()
	Dispatch(func() { ran = true })
	require.True(t, ran, "dispatch after Run returned runs immediately")
}
