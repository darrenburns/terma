# TODO Tutorial A - Basic List
# Run with: vhs todo-tutorial-a.tape

Output docs/assets/todo-tutorial-a.mp4

Set Shell "bash"
Set FontSize 25
Set Width 1024
Set Height 720
Set Padding 0
Set TypingSpeed 150ms

# Start the basic todo list example (hidden)
Hide
Type "go run ./cmd/tutorial/05a-todo-list"
Enter
Sleep 0.5s
Show

# Initial state - show the 3 tasks
Sleep 1s

# Navigate down through the list
Down
Sleep 400ms
Down
Sleep 400ms

# Navigate back up
Up
Sleep 400ms

# Add a new task
Type "a"
Sleep 400ms

# Add another task
Type "a"
Sleep 400ms

# Navigate to see the new tasks
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms

# Delete the current task
Type "d"
Sleep 500ms

# Navigate up and delete another
Up
Sleep 300ms
Type "d"
Sleep 500ms

# Final pause
Sleep 1s
