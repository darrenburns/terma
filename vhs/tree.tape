# Tree Widget Demo
# Run with: vhs tree.tape

Output docs/assets/tree-demo.mp4

Set Shell "bash"
Set FontSize 18
Set Width 800
Set Height 500
Set Padding 10
Set TypingSpeed 50ms

# Start the tree example (hidden)
Hide
Type "go run ./cmd/tree-example"
Enter
Sleep 1.5s
Show

# Focus the tree (filter is focused by default)
Tab
Sleep 300ms

# Navigate down through the tree
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms

# Expand a directory (cmd folder - it's lazy loaded)
Up
Up
Up
Sleep 300ms
Right
Sleep 800ms  # Wait for lazy load

# Navigate into expanded content
Down
Sleep 300ms
Down
Sleep 300ms

# Collapse the directory
Left
Sleep 500ms

# Expand again and show multi-select with Shift+Down
Right
Sleep 500ms
Down
Sleep 300ms

# Multi-select demonstration (Shift+j)
Type "J"
Sleep 300ms
Type "J"
Sleep 300ms
Type "J"
Sleep 500ms

# Clear selection by moving without shift
Down
Sleep 500ms

# Show filtering - focus the filter input
Ctrl+f
Sleep 300ms

# Type a filter query
Type "main"
Sleep 1s

# Show filtered results
Sleep 500ms

# Clear filter
Escape
Sleep 500ms

# Final pause to show the result
Sleep 1s
