package help

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"bgenix/internal/config"
)

var (
	PackageName = config.PackageName
	Version     = config.Version
	AgeBinary   = config.AgeBinary
)

func ShowHelp() {
	fmt.Println(PackageName, "- edit and rekey age secret files")
	fmt.Println("")
	fmt.Println(PackageName, "-e FILE [-i PRIVATE_KEY]")
	fmt.Println(PackageName, "-r [-i PRIVATE_KEY]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("-h, --help                show help")
	fmt.Println("-e, --edit FILE           edits FILE using $EDITOR")
	fmt.Println("-r, --rekey               re-encrypts all secrets with specified recipients")
	fmt.Println("-d, --decrypt FILE        decrypts FILE to STDOUT")
	fmt.Println("-i, --identity            identity to use when decrypting")
	fmt.Println("-v, --verbose             verbose output")
	fmt.Println("")
	fmt.Println("Notes:")
	fmt.Println("FILE an age-encrypted file")
	fmt.Println("PRIVATE_KEY a path to a private SSH key used to decrypt file")
	fmt.Println("EDITOR environment variable of editor to use when editing FILE")
	fmt.Println("")
	fmt.Println("If STDIN is not interactive, EDITOR will be set to \"cp /dev/stdin\"")
	fmt.Println("")
	fmt.Println("RULES environment variable with path to Nix file specifying recipient public keys.")
	fmt.Println("Defaults to './secrets.nix'")
	fmt.Println("")
	fmt.Printf("%s version: %s\n", PackageName, Version)
	fmt.Printf("age binary path: %s\n", AgeBinary)

	// get age version
	cmd := exec.Command(AgeBinary, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		fmt.Println("Failed to get age version:", err)
		return
	}

	fmt.Printf("age version: %s\n", strings.TrimSpace(out.String()))
}
