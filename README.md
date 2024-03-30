# gopatch

A custom patch and file merge tool for Go that patches files based on line numbers.

## Context

I needed a custom patch file implementation for my [other project](https://github.com/SKevo18/mhmods) that I can use to merge and edit multiple files together, as if all edits are applied at once.

Chances are, you probably do not need this tool, as it is very simple and likely not production-ready. I made bare working version in 1 day, and will likely push occassional fixes to suit my needs (regarding the other project above). If it somehow interests you and you see a potential for improvements, PRs are welcome (thanks!).

### Usage

You can get the latest CLI version from the [releases page](https://github.com/SKevo18/gopatch/releases/latest).

```bash
gopatch <input folder> <output folder> <patch file>
```

- **input folder**: The folder containing the files to be patched.
- **output folder**: The folder where the patched files will be saved.
- **patch file**: The patch file to be applied to the input files (see syntax below). The only requirement is that the file has valid syntax (it can have any file extension).

## Syntax

The patch file syntax is different from the standard `diff` tool. When creating a `gopatch` file, you should think of it as if you were editing the original file. You do not need to calculate new ine numbers if you insert a line - the tool was made so that it seems as if all edits were applied at once.

Here is an example of a `gopatch` file:

```patch
# Comments are allowed
# and should start with a hashtag
@ + test.txt 8 0 true
{
    "ape": "monkey",
    "monkey": "ape"
}

@ - test.txt 5 6 true

```

- `@`: The symbol to start a new patch (header)
- `+`: Indicates addition of a new line
- `-`: Indicates removal of line range
- `test.txt`: The file to be patched - relative to the input folder
- `8`: Line number (from) - this is the starting line number where edits will be applied. Lines are 1-indexed.
- `0`: Line number (to):
  - in removal mode, lines are removed until this line (if 0, only the line at the starting line number is removed)
  - in addition mode, this has no effect and is ignored
- `true`: Indicates overwrite mode:
  - in addition mode, if `true`, the content will overwrite lines in place
  - in addition mode, if `false`, the content will be inserted after the line at the starting line number
  - in removal mode, if `true`, the line ranges will be removed and the surrounding lines will collapse
  - in removal mode, if `false`, the line ranges will be replaced with a blank line

All arguments must be present at all times in the correct order, surrounded by spaces, for the tool to recognize them properly. Header Go format: `@ %c %s %d %d %t`.

The lines after the header represent the lines (content) to be added. Content is ignored in removal mode.
If a content line starts with `\` (backslash), it will be stripped from the output (e. g.: if you want your line to begin with `@`, you can escape it with `\@`). If a line contains only backslash, the tool treats it as `\n` (newline/empty line).

All lines starting with `#` are ignored. If you want to begin your lines with a hashtag, escape it like this: `\#`.

Note: escaping only works at the beginning of the line (there is no reason to escape characters elsewhere, as it doesn't ruin the patch file format - it only makes sense to escape characters at the beginning of a line).

You can see the [fixtures](/fixtures/) directory for more examples and sample output.

## Use in Go (directly)

You can also use the `gopatch` package directly in your Go code. Here is an example:

<!-- markdownlint-disable MD010 -->
```go
package main

import (
	"fmt"
	"os"

	"github.com/SKevo18/gopatch"
)

func main() {
	if len(os.Args) != 4 {
		os.Exit(1)
	}
	originalDir := os.Args[1]
	outputDir := os.Args[2]
	patchFile := os.Args[3]

	patchLines, err := gopatch.ReadPatchFile(patchFile) // or `gopatch.ReadPatchFiles([]string{patchFile, ...})` to join multiple patch files together
	if err != nil {
		fmt.Printf("Failed to read patch file %s: %v", patchFile, err)
		os.Exit(1)
	}
	if err := gopatch.PatchDir(originalDir, outputDir, patchLines); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

```
<!-- markdownlint-enable MD010 -->

## License

This project is under minimal maintenance and licensed under the MIT License - see the [LICENSE.md](/LICENSE.md) file for details.
