# TextInput Widget Demo
# Run with: vhs textinput.tape

Output docs/assets/textinput-demo.mp4

Set Shell "bash"
Set FontSize 25
Set Width 1024
Set Height 720
Set Padding 0
Set TypingSpeed 80ms

# Start the textinput example (hidden)
Hide
Type "go run ./cmd/text-input-example"
Enter
Sleep 0.5s
Show

# Focus starts on Name field
# Type a name
Type "Jane Doe"
Sleep 500ms

# Tab to Email field
Tab
Sleep 300ms

# Type an email
Type "jane@example.com"
Sleep 500ms

# Tab to Message field
Tab
Sleep 300ms

# Type a message (character count updates)
Type "Hello, Terma!"
Sleep 500ms

# Tab to Submit button
Tab
Sleep 300ms

# Press Enter to submit
Enter
Sleep 800ms

# Final pause to show submitted result
Sleep 1.5s
