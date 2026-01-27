# Menu Widget Demo
# Run with: vhs menu.tape

Output docs/assets/menu-demo.mp4

Set Shell "bash"
Set FontSize 25
Set Width 1024
Set Height 720
Set Padding 0
Set TypingSpeed 150ms

# Start the menu example (hidden)
Hide
Type "go run ./cmd/menu-example"
Enter
Sleep 0.3s
Show

Sleep 500ms

# Press the File button to open the menu
Enter
Sleep 500ms

# Navigate down to Settings (skips divider automatically)
# New -> Open -> Open Recent -> Settings
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms

# Open Settings submenu
Right
Sleep 500ms

# Navigate down from Editor to Theme
Down
Sleep 300ms

# Open Theme submenu
Right
Sleep 500ms

# Pause to show the fully expanded nested menus
Sleep 1s

# Close Theme submenu
Escape
Sleep 500ms

# Close Settings submenu
Escape
Sleep 500ms

# Close the entire menu
Escape
Sleep 1s
