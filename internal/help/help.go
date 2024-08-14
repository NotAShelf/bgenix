package help

import (
	"fmt"

	"bgenix/internal/config"
)

var (
	PackageName = config.PackageName
	Version     = config.Version
	AgeBinary   = config.AgeBinary
	AgeVersion  = config.AgeVersion
)

func ShowHelp() {
	helpText := fmt.Sprintf(`
%s - edit and rekey age secret files

Options:
-h, --help                show help
-e, --edit FILE           edits FILE using $EDITOR
-r, --rekey               re-encrypts all secrets with specified recipients
-d, --decrypt FILE        decrypts FILE to STDOUT
-i, --identity            identity to use when decrypting
-v, --verbose             verbose output

Notes:
FILE an age-encrypted file
PRIVATE_KEY a path to a private SSH key used to decrypt file
EDITOR environment variable of editor to use when editing FILE

If STDIN is not interactive, EDITOR will be set to "cp /dev/stdin"

RULES environment variable with path to Nix file specifying recipient public keys.
Defaults to './secrets.nix'

Program information:
%s version: %s
age binary path: %s
age version: %s
`, PackageName, PackageName, Version, AgeBinary, AgeVersion)

	fmt.Println(helpText)
}
