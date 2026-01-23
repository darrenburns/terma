# Animation Tutorial Demo
# Run with: vhs animation-tutorial.tape

Output docs/assets/animation-tutorial.mp4

Set Shell "bash"
Set FontSize 25
Set Width 1024
Set Height 720
Set Padding 0
Set TypingSpeed 150ms

# Start the animation tutorial example (hidden)
Hide
Type "go run ./cmd/tutorial/04-animation"
Enter
Sleep 0.5s
Show

# Increase progress with Up key (each press adds 20)
Up
Sleep 600ms
Up
Sleep 600ms
Up
Sleep 600ms

# Decrease once
Down
Sleep 600ms

# Fill to 100% (need 4 more presses from 40)
Up
Sleep 600ms
Up
Sleep 600ms
Up
Sleep 600ms

# Final pause to show the green success color
Sleep 1.5s
