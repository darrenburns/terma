package terma

import (
	"context"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

// snapshotFocusRequest exercises requests made by a widget during rendering,
// including requests emitted again during the final focused pass.
type snapshotFocusRequest struct {
	child Widget
}

func (w snapshotFocusRequest) Build(ctx BuildContext) Widget {
	ctx.RequestFocus("second")
	return w.child
}

func snapshotFocusCheckboxes() Widget {
	return Column{Children: []Widget{
		&Checkbox{ID: "first", Label: "First", State: NewCheckboxState(false)},
		&Checkbox{ID: "second", Label: "Second", State: NewCheckboxState(false)},
	}}
}

func isolateSnapshotFocus(t *testing.T) {
	t.Helper()
	oldPending := pendingFocusID
	t.Cleanup(func() { pendingFocusID = oldPending })
	pendingFocusID = ""
}

func requireSnapshotCheckboxFocus(t *testing.T, buf *uv.Buffer, row int) {
	t.Helper()
	require.Equal(t, getTheme().ActiveCursor.Hex(), FromANSI(buf.CellAt(0, row).Style.Bg).Hex(), "checkbox on row %d must be visibly focused", row)
	require.Equal(t, getTheme().Surface.Hex(), FromANSI(buf.CellAt(0, 1-row).Style.Bg).Hex(), "other checkbox must be unfocused")
}

func TestSnapshot_PreRenderFocusRequest(t *testing.T) {
	isolateSnapshotFocus(t)
	widget := snapshotFocusCheckboxes()
	RequestFocus("second")
	buf := RenderToBuffer(widget, 14, 2)
	assertBufferSnapshot(t, t.Name(), buf, 14, 2, DefaultSVGOptions(), "Second checkbox is visibly focused by a request before rendering; first checkbox is unfocused")
	requireSnapshotCheckboxFocus(t, buf, 1)
	require.Empty(t, pendingFocusID)
	// The same IDs in another render must not inherit the previous focus.
	requireSnapshotCheckboxFocus(t, RenderToBuffer(widget, 14, 2), 0)
}

func TestSnapshot_InvalidFocusRequest(t *testing.T) {
	isolateSnapshotFocus(t)
	for _, id := range []string{"", "missing", "label", "disabled"} {
		t.Run("id="+id, func(t *testing.T) {
			widget := Column{Children: []Widget{
				snapshotFocusCheckboxes(),
				Text{ID: "label", Content: "Not focusable"},
				DisabledWhen(true, &Checkbox{ID: "disabled", State: NewCheckboxState(false)}),
			}}
			RequestFocus(id)
			requireSnapshotCheckboxFocus(t, RenderToBuffer(widget, 14, 4), 0)
			require.Empty(t, pendingFocusID)
		})
	}
}

func TestSnapshot_RenderFocusRequestDoesNotLeak(t *testing.T) {
	isolateSnapshotFocus(t)
	widget := snapshotFocusCheckboxes()
	RequestFocus("first")
	requireSnapshotCheckboxFocus(t, RenderToBuffer(snapshotFocusRequest{child: widget}, 14, 2), 1)
	require.Empty(t, pendingFocusID, "final render pass must not leave a global focus request")
	requireSnapshotCheckboxFocus(t, RenderToBuffer(widget, 14, 2), 0)
}

func TestSnapshot_PreservesRunningAppFocusRequest(t *testing.T) {
	isolateSnapshotFocus(t)
	appRuntimeMu.RLock()
	oldContext, oldQueue := appLifecycleCtx, appDispatchQueue
	appRuntimeMu.RUnlock()
	ctx, cancel := context.WithCancel(context.Background())
	setAppRuntimeState(ctx, newDispatchQueue())
	t.Cleanup(func() {
		cancel()
		setAppRuntimeState(oldContext, oldQueue)
	})

	RequestFocus("second")
	widget := snapshotFocusCheckboxes()
	requireSnapshotCheckboxFocus(t, RenderToBuffer(widget, 14, 2), 0)
	require.Equal(t, "second", pendingFocusID, "a snapshot must not consume the running app's request")

	RequestFocus("app-input")
	requireSnapshotCheckboxFocus(t, RenderToBuffer(snapshotFocusRequest{child: widget}, 14, 2), 1)
	require.Equal(t, "app-input", pendingFocusID, "render-time focus requests must not overwrite the running app's request")
}

func TestSnapshot_ModalFocusRequest(t *testing.T) {
	isolateSnapshotFocus(t)
	widget := Column{Children: []Widget{
		Button{ID: "base", Label: "Base"},
		Floating{
			Visible: true,
			Config:  FloatConfig{Modal: true, Position: FloatPositionBottomLeft},
			Child:   snapshotFocusCheckboxes(),
		},
	}}
	// The modal occupies the bottom two rows; its first checkbox is focused
	// automatically, and a request for its second checkbox takes precedence.
	for _, id := range []string{"", "second"} {
		RequestFocus(id)
		buf := RenderToBuffer(widget, 14, 4)
		row := 2
		if id == "second" {
			row = 3
		}
		require.Equal(t, getTheme().ActiveCursor.Hex(), FromANSI(buf.CellAt(0, row).Style.Bg).Hex())
		require.Empty(t, pendingFocusID)
	}
}
