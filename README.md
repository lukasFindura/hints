# hints
CLI tool to execute commands stored in JSON config file.

```bash
go run main.go example.json
```

### config
See `example.json` for the structure.

Behavior if command is prefixed with:

`_` - doesn't print command being ran
`!` - path to a script to be run. E.g.:
```json
{
    "name": "script",
    "command": "!$HOME/.scripts/script.sh"
}
```

## `pick` extension

`hints`' dependency `gocliselect` can be used as hint as well.

`pick` prints menu to stderr, user's choice to stdout.

E.g. you can have:
```json
{
    "name": "pick file",
    "command": "choice=`ls | pick` ; echo $choice"
}
```

It can be used as a standalone utility too (not necessarily in `hints`).
```bash
alias example='choice=`ls | pick2` ; vi $choice'
```

### `pick` install
 ```bash
mkdir pick; cd pick
 ```
<details>
<summary><b>main.go</b></summary>

```go
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/buger/goterm"
	"github.com/lukasFindura/gocliselect"
)

func main() {

	gocliselect.Cursor.ItemPrompt = "❯"
	if prompt := os.Getenv("PICK_PROMPT"); prompt != "" {
		gocliselect.Cursor.ItemPrompt = prompt
	}
	envColor := os.Getenv("PICK_COLOR")

	// Variable to store the Goterm color
	var selectedColor int

	// Switch to map environment values to Goterm colors
	switch envColor {
	case "red":
		selectedColor = goterm.RED
	case "green":
		selectedColor = goterm.GREEN
	case "blue":
		selectedColor = goterm.BLUE
	case "yellow":
		selectedColor = goterm.YELLOW
	case "magenta":
		selectedColor = goterm.MAGENTA
	case "cyan":
		selectedColor = goterm.CYAN
	case "white":
		selectedColor = goterm.WHITE
	default:
		// Set a default color if the environment variable is invalid or not set
		selectedColor = gocliselect.Cursor.ItemColor
	}
	gocliselect.Cursor.ItemColor = selectedColor
	gocliselect.Cursor.Suffix = " "

	var verbose = false
	if verbose_ := os.Getenv("PICK_VERBOSE"); verbose_ != "" {
		verbose = true
	}

	// Check if input is coming from a pipe
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// No pipe input
		fmt.Fprintln(os.Stderr, "Usage: ... | pick\n\t'pick' expects input - each input line is an option to be listed.")
		os.Exit(1)
	}

	// Create a new menu
	menu := gocliselect.NewMenu("Select an item", 0)

	// Read from pipe
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			menu.AddItem(line, "")
		}
	}

	// Display the menu
	if _, choice := menu.Display(menu); choice != nil {
		if verbose {
			fmt.Printf("\nPicked: %s\n", choice.Text)
		}
		// Print user's choice
		fmt.Print(choice.Text)
	} else {
		// User canceled choice
		os.Exit(1)
	}
}
```

</details>

```go
go mod init github.com/lukasFindura/pick
go mod tidy
go install
```

#### optional configuration:
```bash
export PICK_PROMPT=...   # prompt char
export PICK_COLOR=...    # prompt color
export PICK_VERBOSE=...  # prints choice if set (default is false)
```

### `pick-less` extension
`pick-less` extension will add pagination to `pick` which would struggle with inputs higher than current terminal's height.

Example:
```bash
$ cat input | pick-less
❯ ...previous
  6
  7
  8
  9
  10
  ...next
```

Use `PICKLESS_SIZE` env var to set batch size (default is 20).

<details>
<summary><b>pick-less.sh</b></summary>

```bash
#!/usr/bin/env bash

# Function to clean up the current output
clear_stdout() {
  local lines_to_clear=$1
  for ((i = 0; i < lines_to_clear; i++)); do
    echo -en "\033[F\033[2K" >&2 # Move up and clear the line
  done
}

# Read the input file (or stdin) into a variable
input=$(cat)

# Split the input into 20-line chunks and process them
if [ -z "${PICKLESS_SIZE}" ]; then
  batch_size=20
else
  batch_size=$PICKLESS_SIZE
fi
next_marker="...next"
prev_marker="...previous"

# Split input into an array of lines
mapfile -t lines <<< "$input"

# Initialize variables
history=()  # Keeps track of all displayed chunks
current_index=-1  # Tracks the current position in history

# Main loop
while true; do
  # Determine the chunk to display
  if [[ $current_index -ge 0 && $current_index -lt ${#history[@]} ]]; then
    # Use the history for the current index
    chunk_text="${history[$current_index]}"
  else
    # Load a new chunk if there are remaining lines
    if [[ ${#lines[@]} -gt 0 ]]; then
      chunk=("${lines[@]:0:$batch_size}")
      lines=("${lines[@]:$batch_size}")
      chunk_text=$(printf "%s\n" "${chunk[@]}")
      history+=("$chunk_text")
      current_index=${#history[@]}-1
    else
      break
    fi
  fi

  # Add navigation markers
  [[ $current_index -gt 0 ]] && chunk_text="$prev_marker\n$chunk_text"
  [[ $current_index -lt $((${#history[@]} - 1)) || ${#lines[@]} -gt 0 ]] && chunk_text+="\n$next_marker"

  # Count the number of lines in the current chunk
  lines_in_chunk=$(echo -e "$chunk_text" | wc -l)

  # Run pick with the chunk, capture the choice
  choice=$(echo -e "$chunk_text" | pick) || exit 1

  # Handle navigation
  if [[ "$choice" == "$next_marker" ]]; then
    clear_stdout "$lines_in_chunk"
    if [[ $current_index -lt $((${#history[@]} - 1)) ]]; then
      ((current_index++))
    elif [[ ${#lines[@]} -gt 0 ]]; then
      # Load a new chunk when reaching the end of history
      chunk=("${lines[@]:0:$batch_size}")
      lines=("${lines[@]:$batch_size}")
      chunk_text=$(printf "%s\n" "${chunk[@]}")
      history+=("$chunk_text")
      current_index=${#history[@]}-1
    fi
  elif [[ "$choice" == "$prev_marker" ]]; then
    clear_stdout "$lines_in_chunk"
    if [[ $current_index -gt 0 ]]; then
      ((current_index--))
    fi
  else
    break
  fi
done

# Print choice on the output
echo $choice
```
</details>
