# docs/demo-find.tape
# Demo of snipt's floating search palette
Output docs/demo-find.gif

Set Shell "zsh"
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Catppuccin Mocha"
Set Padding 20

# Launch find palette
Type "snipt find"
Enter
Sleep 2s

# Type a search query
Type "http"
Sleep 1.5s

# Navigate results
Down
Sleep 1s
Down
Sleep 1s
Up
Sleep 1s

# Select and copy to clipboard
Enter
Sleep 1s
