# TabBar Widget Demo
# Run with: vhs tabs.tape

Output docs/assets/tabs-demo.mp4

Set Shell "bash"
Set FontSize 25
Set Width 1024
Set Height 720
Set Padding 0
Set TypingSpeed 150ms

# Start the tab example (hidden)
Hide
Type "go run ./cmd/tab-example"
Enter
Sleep 0.5s
Show

# Initial view - Home tab is active
Sleep 1s

# Navigate to next tab with l
Type "l"
Sleep 500ms

# Navigate to next tab again
Type "l"
Sleep 500ms

# Navigate to List tab
Type "l"
Sleep 500ms

# Jump to first tab with number 1
Type "1"
Sleep 500ms

# Jump to third tab with number 3
Type "3"
Sleep 500ms

# Reorder tab - move left with ctrl+h
Ctrl+h
Sleep 500ms

# Move left again
Ctrl+h
Sleep 500ms

# Move right with ctrl+l
Ctrl+l
Sleep 500ms

# Navigate to last tab
Type "4"
Sleep 500ms

# Close tab with ctrl+w
Ctrl+w
Sleep 500ms

# Navigate through remaining tabs
Type "h"
Sleep 500ms
Type "h"
Sleep 500ms

# Final pause
Sleep 1s
