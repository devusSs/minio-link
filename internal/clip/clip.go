package clip

import (
	"github.com/atotto/clipboard"
)

// CopyToClipboard copies the input to the clipboard
func CopyToClipboard(input string) error {
	err := clipboard.WriteAll(input)
	return err
}
