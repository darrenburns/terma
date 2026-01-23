# Todo App Demo
# Run with: vhs todo.tape

Output docs/assets/todo-demo.mp4

Set Shell "bash"
Set FontSize 22
Set Width 1024
Set Height 720
Set Padding 0
Set TypingSpeed 100ms

# Start the todo app (hidden)
Hide
Type "go run ./cmd/todo-app"
Enter
Sleep 0.5s
Show

# Initial view
Sleep 1s

# Tab to focus the todo list
Tab
Sleep 500ms

# Navigate down a couple of items
Down
Sleep 300ms
Down
Sleep 300ms

# Cycle through some themes
Type "t"
Sleep 800ms
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms
Enter
Sleep 800ms

# Multi-select using Shift+Down (J)
Type "J"
Sleep 300ms
Type "J"
Sleep 300ms
Type "J"
Sleep 500ms

# Toggle all 3 selected items at once with space
Space
Sleep 800ms

# Move selected items to the bottom with ctrl+j (multiple times)
Ctrl+j
Sleep 400ms
Ctrl+j
Sleep 400ms
Ctrl+j
Sleep 400ms
Ctrl+j
Sleep 400ms
Ctrl+j
Sleep 400ms
Ctrl+j
Sleep 800ms

# Clear selection and go back up
Escape
Sleep 300ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 500ms

# Enter filter mode with /
Type "/"
Sleep 500ms

# Type filter query
Type "cat"
Sleep 1s

# Exit filter mode
Escape
Sleep 500ms

# Edit the current task with e
Type "e"
Sleep 500ms

# Delete existing content and type new value
Backspace 28
Sleep 300ms
Type "Teach the dog to pay bills"
Sleep 500ms

# Submit the edit
Enter
Sleep 1s

# Final pause
Sleep 1s
