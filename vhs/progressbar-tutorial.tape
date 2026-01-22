# Progress Bar Tutorial Demo
# Run with: vhs progressbar-tutorial.tape

Output docs/assets/progressbar-tutorial.mp4

Set Shell "bash"
Set FontSize 25
Set Width 1024
Set Height 720
Set Padding 0
Set TypingSpeed 150ms

# Start the progress bar tutorial example (hidden)
Hide
Type "go run ./cmd/tutorial/03-progress-bar"
Enter
Sleep 0.5s
Show

# Increase progress with Up key
Up
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 400ms

# Decrease once
Down
Sleep 400ms

# Fill to 10/10 (need 5 more presses from 4)
Up
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 400ms
Up
Sleep 400ms

# Final pause to show the green success color
Sleep 1.5s
