# TODO Tutorial B - Scrollable List
# Run with: vhs todo-tutorial-b.tape

Output docs/assets/todo-tutorial-b.mp4

Set Shell "bash"
Set FontSize 25
Set Width 1024
Set Height 720
Set Padding 0
Set TypingSpeed 150ms

# Start the scrollable todo list example (hidden)
Hide
Type "go run ./cmd/tutorial/05b-todo-scrollable"
Enter
Sleep 0.5s
Show

# Initial state - show the 3 tasks
Sleep 1s

# Add many tasks to exceed viewport height
Type "a"
Sleep 250ms
Type "a"
Sleep 250ms
Type "a"
Sleep 250ms
Type "a"
Sleep 250ms
Type "a"
Sleep 250ms
Type "a"
Sleep 250ms
Type "a"
Sleep 250ms
Type "a"
Sleep 250ms
Type "a"
Sleep 250ms
Type "a"
Sleep 250ms

# Now we have 13 tasks - navigate down to show scrolling
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms

# Scroll back up
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms

# Final pause
Sleep 1s
