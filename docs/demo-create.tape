# docs/demo-create.tape
# Demo of creating a new snippet in snipt manage
Output docs/demo-create.gif

Set Shell "zsh"
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 50ms

# Launch manage
Type "./snipt"
Enter
Sleep 2s

# Press n for new snippet (opens $EDITOR with frontmatter template)
Type "n"
Sleep 2s

# The editor opens with the YAML frontmatter template.
# This tape assumes vi/nvim as $EDITOR.
#
# Template looks like:
# ---
# title:
# language: text
# tags: []
# pinned: false
# ---
#
# Cursor starts at top of file. Navigate to fill in fields.

# Go to title line and fill it in
Type "jA"
Sleep 300ms
Type " Docker Compose Template"
Escape
Sleep 500ms

# Go to language line and change "text" to "yaml"
Type "j0ftcw"
Sleep 200ms
Type "yaml"
Escape
Sleep 500ms

# Go to tags line and add tags
Type "j0f[a"
Sleep 200ms
Type "docker, devops"
Escape
Sleep 500ms

# Move below the closing --- and add content
Type "jjo"
Sleep 300ms
Type "version: '3.8'"
Enter
Type "services:"
Enter
Type "  app:"
Enter
Type "    image: node:20-alpine"
Enter
Type "    ports:"
Enter
Type "      - '3000:3000'"
Enter
Type "    volumes:"
Enter
Type "      - .:/app"
Escape
Sleep 1s

# Save and quit editor
Type ":wq"
Enter
Sleep 2s

# New snippet should appear in the list
Sleep 2s

# Quit
Type "q"
Sleep 500ms
