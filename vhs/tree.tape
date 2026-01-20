# Simple Tree Widget Demo
# Run with: vhs tree.tape

Output docs/assets/tree-demo.mp4

Set Shell "bash"
Set FontSize 25
Set Width 1024
Set Height 720
Set Padding 0
Set TypingSpeed 150ms

# Start the simple tree example (hidden)
Hide
Type "go run ./cmd/simple-tree-example"
Enter
Sleep 0.3s
Show

# Focus the tree (filter input is focused by default)
Tab
Sleep 300ms

# Collapse Fruits
Space
Sleep 400ms

# Expand Fruits again
Space
Sleep 400ms

# Navigate into Fruits children
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms

# Multi-select with Shift+j (capital J)
Type "J"
Sleep 300ms
Type "J"
Sleep 300ms
Type "J"
Sleep 500ms
Enter
Sleep 700ms

# Cycle through 3 themes
Type "t"
Sleep 1000ms
Type "t"
Sleep 1000ms
Type "t"
Sleep 1000ms
Type "t"
Sleep 1000ms
Type "t"
Sleep 1000ms

Down
Sleep 300ms

# Focus the filter input
Tab
Sleep 300ms

# Type filter query "Cher"
Type "Cher"
Sleep 2s

# Clear filter
Backspace
Sleep 300ms
Backspace
Sleep 300ms
Backspace
Sleep 300ms
Backspace

# Final pause
Sleep 1s
