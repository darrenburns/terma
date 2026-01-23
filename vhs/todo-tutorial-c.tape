# TODO Tutorial C - Empty State Handling
# Run with: vhs todo-tutorial-c.tape

Output docs/assets/todo-tutorial-c.mp4

Set Shell "bash"
Set FontSize 25
Set Width 1024
Set Height 720
Set Padding 0
Set TypingSpeed 150ms

# Start the empty state todo list example (hidden)
Hide
Type "go run ./cmd/tutorial/05c-todo-empty-state"
Enter
Sleep 0.5s
Show

# Initial state - show empty message
Sleep 1.5s

# Add some tasks
Type "a"
Sleep 400ms
Type "a"
Sleep 400ms
Type "a"
Sleep 400ms
Type "a"
Sleep 400ms

# Navigate through them
Down
Sleep 300ms
Down
Sleep 300ms

# Delete a task
Type "d"
Sleep 400ms

# Add more tasks
Type "a"
Sleep 400ms
Type "a"
Sleep 400ms

# Navigate to top
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms

# Clear all tasks to show empty state again
Type "c"
Sleep 1.5s

# Add a task to show the transition back
Type "a"
Sleep 500ms

# Final pause
Sleep 1s
