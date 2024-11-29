# hints
CLI tool to execute commands stored in JSON config file.

```bash
go run main.go example.json
```

### config
See `example.json` for the structure.
Note that command prefixed with `!` is a path to a script to be run. E.g.:
```json
{
    "name": "script",
    "command": "!$HOME/.scripts/script.sh"
}
```

## `picker` extension

`hints`' dependency `gocliselect` can be used as hint as well.
E.g. you can have:
```json
{
    "name": "pick file",
    "command": "ls | picker"
}
```
To capture the selected option we need a workaround since common tools like `grep`, `awk` or `tee` won't work given the way `gocliselect` is implemented (clearing and redrawing console prompt and waiting for user's input).

Workaround (works on MacOS):
```json
{
    "name": "pick file",
    "command": "script -q /tmp/picker bash -c \"'ls' | picker\" && grep 'Picked:' /tmp/picker | awk '{print $2}'"
}
```
`script` captures the output of the command
 to a temporary file.
 Wrapper `bash -c` is usually needed only if the command is actually a pipe - more commands chained together. 
 Alternatively, `picker`'s code can be changed to propagate the choice in a different way.

 ```bash
mkdir picker; cd picker
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
	gocliselect.Cursor.ItemColor = goterm.YELLOW
	gocliselect.Cursor.Suffix = " "

	// Check if input is coming from a pipe
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// No pipe input
		fmt.Fprintln(os.Stderr, "Usage: ... | picker\n\tPicker expects input - each input line is an option in Picker.")
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
		fmt.Printf("Picked: %s\n", choice.Text)
	}
}
```

</details>

```go
go mod init github.com/lukasFindura/picker
go mod tidy
go install
```
