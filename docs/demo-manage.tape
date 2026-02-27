# docs/demo-manage.tape
# Demo of snipt's full-screen manage TUI
Output docs/demo-manage.gif

Set Shell "zsh"
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Catppuccin Mocha"
Set Padding 20

# Launch manage (default command)
Type "./snipt"
Enter
Sleep 2s

# Navigate through snippets
Sleep 500ms
Type "j"
Sleep 1s
Type "j"
Sleep 1s
Type "j"
Sleep 1s

# Go back up
Type "k"
Sleep 500ms
Type "k"
Sleep 1s

# Fuzzy search by title
Type "/"
Sleep 500ms
Type "http"
Sleep 1.5s

# Clear search
Escape
Sleep 1s

# Tag search with # prefix
Type "/"
Sleep 500ms
Type "#server"
Sleep 1.5s
Escape
Sleep 1s

# Copy a snippet to clipboard
Enter
Sleep 1.5s

# Toggle pin
Type "p"
Sleep 1.5s

# Quit
Type "q"
Sleep 1500ms
