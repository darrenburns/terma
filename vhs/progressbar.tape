# ProgressBar Widget Demo
# Run with: vhs progressbar.tape

Output docs/assets/progressbar-demo.mp4

Set Shell "bash"
Set FontSize 25
Set Width 1024
Set Height 720
Set Padding 0
Set TypingSpeed 150ms

# Start the progressbar example (hidden)
Hide
Type "go run ./cmd/progressbar-example"
Enter
Sleep 0.5s
Show

# Focus starts on -10% button
# Move to +10% button
Tab
Sleep 300ms

# Add 10% a few times
Enter
Sleep 500ms
Enter
Sleep 500ms
Enter
Sleep 500ms
Enter
Sleep 500ms

# Move back to -10% and remove once
Shift+Tab
Sleep 300ms
Enter
Sleep 500ms

# Move to Fill button (Tab twice: +10% -> Reset -> Fill)
Tab
Sleep 300ms
Tab
Sleep 300ms
Tab
Sleep 300ms

# Press Fill
Enter
Sleep 800ms

# Move to Reset button (Shift+Tab once)
Shift+Tab
Sleep 300ms

# Press Reset
Enter
Sleep 800ms

# Final pause
Sleep 1s
