# TODO Tutorial D - Multi-Select for Bulk Delete
# Run with: vhs todo-tutorial-d.tape

Output docs/assets/todo-tutorial-d.mp4

Set Shell "bash"
Set FontSize 25
Set Width 1024
Set Height 720
Set Padding 0
Set TypingSpeed 150ms

# Start the multi-select todo list example (hidden)
Hide
Type "go run ./cmd/tutorial/05d-todo-multiselect"
Enter
Sleep 0.5s
Show

# Initial state
Sleep 1s

# Select multiple items with Shift+J (extend selection while moving down)
Type "J"
Sleep 300ms
Type "J"
Sleep 300ms
Type "J"
Sleep 500ms

# Delete all selected items at once
Type "d"
Sleep 1s

# Show remaining items
Sleep 500ms

# Add a few more items
Type "a"
Sleep 300ms
Type "a"
Sleep 300ms

# Final pause
Sleep 1s
