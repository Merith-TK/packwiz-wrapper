// Package modlist provides modlist generation functionality
package modlist

import (
	"fmt"
	"os"

	"github.com/Merith-TK/packwiz-wrapper/internal/modformat"
	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
)

// WriteSection writes a mod section to console and file
func WriteSection(header string, mods []packwiz.ModToml, f *os.File, opts modformat.ModlistOptions, authorCache map[int]modformat.ModAuthorInfo, offset int) {
	if len(mods) == 0 {
		return
	}

	// Write header to console and file
	fmt.Print(header)
	if f != nil {
		f.WriteString(header)
	}

	// Write each mod
	for i, mod := range mods {
		// Get author from cache if available, otherwise empty string
		author := ""
		if opts.ShowAuthors && authorCache != nil {
			if authorInfo, ok := authorCache[offset+i]; ok {
				author = authorInfo.Author
			}
		}

		line := modformat.FormatModLine(mod, opts, author)

		fmt.Print(line)
		if f != nil {
			f.WriteString(line)
		}
	}

	// Write newline separator
	fmt.Println()
	if f != nil {
		f.WriteString("\n")
	}
}
