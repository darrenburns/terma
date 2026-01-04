# Terma Unit Test Plan

This document outlines the unit tests to write for Terma, prioritized by criticality.

## Current State

- **Existing tests:** Only `markup_test.go` (35+ tests for `ParseMarkup()`)
- **Core code without tests:** ~3,300 lines across 11 critical files

## Priority Tiers

### Tier 1: Foundation (CRITICAL - Test First)

These are the building blocks everything else depends on.

#### 1. `dimension_test.go` - Dimension System

The dimension type is fundamental to all layout calculations.

| Test Case | Description |
|-----------|-------------|
| `TestCells_ReturnsFixedDimension` | `Cells(10)` creates fixed dimension |
| `TestCells_ZeroValue` | `Cells(0)` is valid |
| `TestFr_ReturnsFractionalDimension` | `Fr(1)` creates fractional dimension |
| `TestFr_ZeroValue` | `Fr(0)` behavior |
| `TestFr_LargeValue` | `Fr(100)` works correctly |
| `TestAuto_IsAutoTrue` | `Auto.IsAuto()` returns true |
| `TestCells_IsAutoFalse` | `Cells(n).IsAuto()` returns false |
| `TestCells_IsCellsTrue` | `Cells(n).IsCells()` returns true |
| `TestFr_IsFrTrue` | `Fr(n).IsFr()` returns true |
| `TestCells_CellsValue` | `Cells(10).CellsValue()` returns 10 |
| `TestFr_FrValue` | `Fr(2).FrValue()` returns 2 |
| `TestDimension_ZeroValueIsAuto` | Default zero value behaves as Auto |

#### 2. `signal_test.go` - Reactive State System

Signals are the core of the reactive architecture.

| Test Case | Description |
|-----------|-------------|
| `TestNewSignal_InitialValue` | Signal created with initial value |
| `TestSignal_Get_ReturnsValue` | `Get()` returns current value |
| `TestSignal_Set_UpdatesValue` | `Set()` changes value |
| `TestSignal_Set_SameValue_NoRebuild` | Setting same value doesn't trigger rebuild |
| `TestSignal_Update_FunctionalUpdate` | `Update(fn)` applies function |
| `TestSignal_Peek_NoSubscription` | `Peek()` reads without subscribing |
| `TestSignal_Get_DuringBuild_Subscribes` | `Get()` during build creates subscription |
| `TestSignal_Set_TriggersSubscriberRebuild` | Subscribed widgets marked dirty |
| `TestAnySignal_NonComparableTypes` | Works with slices, maps, etc. |
| `TestAnySignal_Set_AlwaysDirties` | Always marks dirty (can't compare) |
| `TestSignal_MultipleSubscribers` | Multiple widgets subscribe correctly |
| `TestSignal_Unsubscribe_OnWidgetRemoval` | Cleanup when widget removed |

#### 3. `keybind_test.go` - Keybind Matching

The keybind system is used throughout for input handling.

| Test Case | Description |
|-----------|-------------|
| `TestKeybind_MatchSimpleKey` | `"enter"` matches enter key |
| `TestKeybind_MatchLetter` | `"a"` matches 'a' key |
| `TestKeybind_MatchWithModifier` | `"ctrl+c"` matches correctly |
| `TestKeybind_NoMatchDifferentKey` | Wrong key doesn't match |
| `TestKeybind_ActionExecuted` | Action callback runs on match |
| `TestKeybind_HiddenNotDisplayed` | Hidden keybinds work but aren't shown |
| `TestKeybind_MultipleBindings` | First matching binding wins |
| `TestKeybind_CaseInsensitive` | `"Enter"` matches `"enter"` |

### Tier 2: Core Functionality (HIGH - Test Second)

#### 4. `layout_test.go` - Layout Engine

The two-pass layout algorithm is complex and critical.

**Column Layout Tests:**

| Test Case | Description |
|-----------|-------------|
| `TestColumn_SingleChild_Auto` | Single Auto child gets content size |
| `TestColumn_SingleChild_Cells` | Single fixed child respects size |
| `TestColumn_SingleChild_Fr` | Single Fr child fills available space |
| `TestColumn_MultipleChildren_AllAuto` | Multiple Auto children stack |
| `TestColumn_MultipleChildren_AllCells` | Multiple fixed children stack |
| `TestColumn_MultipleChildren_AllFr` | Fr children divide space proportionally |
| `TestColumn_MixedDimensions` | Auto + Cells + Fr together |
| `TestColumn_Spacing` | Spacing adds gaps between children |
| `TestColumn_MainAlign_Start` | Children at top |
| `TestColumn_MainAlign_Center` | Children centered vertically |
| `TestColumn_MainAlign_End` | Children at bottom |
| `TestColumn_CrossAlign_Start` | Children left-aligned |
| `TestColumn_CrossAlign_Center` | Children centered horizontally |
| `TestColumn_CrossAlign_End` | Children right-aligned |
| `TestColumn_FrDistribution_1_1` | `Fr(1), Fr(1)` = 50/50 split |
| `TestColumn_FrDistribution_1_2` | `Fr(1), Fr(2)` = 33/67 split |
| `TestColumn_FrDistribution_1_1_1` | `Fr(1), Fr(1), Fr(1)` = thirds |
| `TestColumn_ChildWithPadding` | Padding affects layout |
| `TestColumn_ChildWithMargin` | Margin affects layout |
| `TestColumn_EmptyChildren` | No children = zero size |
| `TestColumn_NestedColumns` | Columns inside columns |

**Row Layout Tests:**

| Test Case | Description |
|-----------|-------------|
| `TestRow_SingleChild_Auto` | Single Auto child gets content size |
| `TestRow_SingleChild_Cells` | Single fixed child respects size |
| `TestRow_SingleChild_Fr` | Single Fr child fills available space |
| `TestRow_MultipleChildren_AllAuto` | Multiple Auto children in row |
| `TestRow_MixedDimensions` | Auto + Cells + Fr horizontally |
| `TestRow_Spacing` | Spacing adds horizontal gaps |
| `TestRow_MainAlign_Start` | Children at left |
| `TestRow_MainAlign_Center` | Children centered horizontally |
| `TestRow_MainAlign_End` | Children at right |
| `TestRow_CrossAlign_Start` | Children top-aligned |
| `TestRow_CrossAlign_Center` | Children centered vertically |
| `TestRow_CrossAlign_End` | Children bottom-aligned |
| `TestRow_FrDistribution` | Fr children divide horizontal space |

#### 5. `context_test.go` - Build Context

| Test Case | Description |
|-----------|-------------|
| `TestBuildContext_AutoID_Root` | Root widget gets `_auto:0` |
| `TestBuildContext_AutoID_FirstChild` | First child gets `_auto:0.0` |
| `TestBuildContext_AutoID_SecondChild` | Second child gets `_auto:0.1` |
| `TestBuildContext_AutoID_DeepNesting` | Deep path like `_auto:0.1.2.3` |
| `TestBuildContext_PushChild_UpdatesPath` | Path updates correctly |
| `TestBuildContext_IsFocused_True` | Returns true when widget focused |
| `TestBuildContext_IsFocused_False` | Returns false when not focused |
| `TestBuildContext_Theme_Available` | Theme accessible in context |

#### 6. `focus_test.go` - Focus Management

| Test Case | Description |
|-----------|-------------|
| `TestFocusManager_InitialFocus` | First focusable gets focus |
| `TestFocusManager_FocusNext` | Tab moves to next focusable |
| `TestFocusManager_FocusPrevious` | Shift+Tab moves to previous |
| `TestFocusManager_FocusNext_Wraps` | Tab wraps from last to first |
| `TestFocusManager_FocusPrevious_Wraps` | Shift+Tab wraps from first to last |
| `TestFocusManager_FocusByID` | Focus specific widget by ID |
| `TestFocusManager_HandleKey_RouteToFocused` | Key event goes to focused widget |
| `TestFocusManager_HandleKey_Bubbles` | Unhandled events bubble up |
| `TestFocusManager_ActiveKeybinds` | Collects keybinds from focus chain |
| `TestFocusManager_SkipsNonFocusable` | Non-focusable widgets skipped |
| `TestFocusManager_SingleFocusable` | Works with only one focusable |
| `TestFocusManager_NoFocusables` | Handles empty focusable list |
| `TestFocusManager_InvalidFocus_AutoCorrects` | Removes focus from removed widgets |

### Tier 3: Complex Widgets (MEDIUM - Test Third)

#### 7. `scroll_test.go` - Scrolling Container

| Test Case | Description |
|-----------|-------------|
| `TestScrollState_InitialOffset_Zero` | Starts at offset 0 |
| `TestScrollState_ScrollDown` | Offset increases |
| `TestScrollState_ScrollUp` | Offset decreases |
| `TestScrollState_ScrollUp_ClampToZero` | Can't scroll above 0 |
| `TestScrollState_ScrollDown_ClampToMax` | Can't scroll past content |
| `TestScrollState_ScrollToView_AlreadyVisible` | No scroll if visible |
| `TestScrollState_ScrollToView_Below` | Scrolls down to reveal |
| `TestScrollState_ScrollToView_Above` | Scrolls up to reveal |
| `TestScrollable_Layout_SetsViewportSize` | Viewport dimensions set |
| `TestScrollable_ContentTallerThanViewport` | Scrollbar appears |
| `TestScrollable_ContentShorterThanViewport` | No scrollbar needed |
| `TestScrollbar_ThumbPosition_AtTop` | Thumb at top when offset 0 |
| `TestScrollbar_ThumbPosition_AtBottom` | Thumb at bottom when scrolled max |
| `TestScrollbar_ThumbPosition_Middle` | Thumb in middle proportionally |
| `TestScrollbar_SubCellPrecision` | Uses partial block characters |

#### 8. `list_test.go` - List Widget

| Test Case | Description |
|-----------|-------------|
| `TestListState_InitialCursor_Zero` | Cursor starts at 0 |
| `TestListState_SelectNext` | Cursor moves down |
| `TestListState_SelectPrevious` | Cursor moves up |
| `TestListState_SelectNext_AtEnd_NoChange` | Cursor stays at last item |
| `TestListState_SelectPrevious_AtStart_NoChange` | Cursor stays at first item |
| `TestListState_SelectFirst` | Jumps to first item |
| `TestListState_SelectLast` | Jumps to last item |
| `TestListState_JumpTo_ValidIndex` | Jumps to specific index |
| `TestListState_JumpTo_InvalidIndex_Clamps` | Invalid index clamped |
| `TestListState_EmptyList` | Handles empty list gracefully |
| `TestListState_SetItems_ResetsCursor` | New items reset cursor |
| `TestListState_RemoveAt_Middle` | Removes item, cursor adjusts |
| `TestListState_RemoveAt_AtCursor` | Removes current item correctly |
| `TestListState_RemoveAt_LastItem` | Cursor adjusts when last removed |
| `TestListState_Append` | Adds item to end |
| `TestListState_Prepend` | Adds item to start |
| `TestListState_MultiSelect_Toggle` | Toggle selection on/off |
| `TestListState_MultiSelect_Range` | Shift+move selects range |
| `TestListState_MultiSelect_Anchor` | Anchor point management |
| `TestList_Build_RendersAllItems` | All items rendered |
| `TestList_Build_CurrentItemStyled` | Current item highlighted |
| `TestList_OnSelect_Called` | Callback fires on Enter |
| `TestList_ScrollIntegration` | Scrolls to keep cursor visible |

#### 9. `button_test.go` - Button Widget

| Test Case | Description |
|-----------|-------------|
| `TestButton_Build_ReturnsText` | Button renders label |
| `TestButton_Keybinds_Enter` | Enter triggers OnPress |
| `TestButton_Keybinds_Space` | Space triggers OnPress |
| `TestButton_Focused_Style` | Different style when focused |
| `TestButton_NotFocused_Style` | Default style when not focused |
| `TestButton_OnClick_Handler` | Click handler called |
| `TestButton_RequiresID` | ID required for focus |

### Tier 4: Conditional & Utility Widgets (LOW - Test Fourth)

#### 10. `conditional_test.go` - Visibility Wrappers

| Test Case | Description |
|-----------|-------------|
| `TestShowWhen_True_ShowsChild` | Shows child when condition true |
| `TestShowWhen_False_ReturnsNil` | Returns nil when condition false |
| `TestHideWhen_True_ReturnsNil` | Returns nil when condition true |
| `TestHideWhen_False_ShowsChild` | Shows child when condition false |
| `TestVisibleWhen_True_ShowsChild` | Shows child when condition true |
| `TestVisibleWhen_False_ReservesSpace` | Reserves space when false |
| `TestInvisibleWhen_True_ReservesSpace` | Reserves space when true |
| `TestInvisibleWhen_False_ShowsChild` | Shows child when condition false |

#### 11. `switcher_test.go` - Content Switcher

| Test Case | Description |
|-----------|-------------|
| `TestSwitcher_ShowsActiveChild` | Active key child rendered |
| `TestSwitcher_InvalidKey_ReturnsNil` | Unknown key returns nil |
| `TestSwitcher_SwitchesContent` | Changing Active switches display |
| `TestSwitcher_PreservesState` | State persists across switches |

## Test Infrastructure Needed

### Test Utilities to Create

```go
// test_helpers.go

// MockWidget - Simple widget for testing layout
type MockWidget struct {
    Width, Height Dimension
    PreferredSize Size // For Auto dimension
}

// MockFocusable - Focusable widget for testing focus
type MockFocusable struct {
    ID       string
    Keybinds []Keybind
}

// MockBuildContext - Creates BuildContext for testing
func MockBuildContext(focusedID string) BuildContext

// AssertSize - Helper for comparing sizes
func AssertSize(t *testing.T, got, want Size)

// AssertDimension - Helper for comparing dimensions
func AssertDimension(t *testing.T, got, want Dimension)
```

## Implementation Order

1. **Week 1:** dimension_test.go, signal_test.go, keybind_test.go
2. **Week 2:** layout_test.go (Column focus first, then Row)
3. **Week 3:** context_test.go, focus_test.go
4. **Week 4:** scroll_test.go, list_test.go
5. **Week 5:** button_test.go, conditional_test.go, switcher_test.go

## Success Metrics

- **Tier 1:** 100% coverage of public API
- **Tier 2:** 90%+ coverage with edge cases
- **Tier 3:** 80%+ coverage of state management
- **Tier 4:** Basic happy path coverage

## Notes

- Follow existing pattern from `markup_test.go`: standard `testing.T`, clear naming
- Use table-driven tests for parameterized cases
- Test edge cases: zero values, empty inputs, boundary conditions
- Keep tests isolated - no shared mutable state between tests
