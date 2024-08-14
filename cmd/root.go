package cmd

import (
	"fmt"
	"os"

	"bgenix/internal/config"
	"bgenix/internal/help"
	"bgenix/internal/util"
)

func Execute() {
	var (
		rekey          bool
		decryptOnly    bool
		fileToEdit     string
		privateKeyPath string
	)

	// Least scuffed argument parser be like:
	args := os.Args[1:]
	for len(args) > 0 {
		switch args[0] {
		case "-h", "--help":
			help.ShowHelp()
			os.Exit(0)

		// -e or --edit takes a file as an argument
		// and decrypts the file to be edited with
		// $EDITOR or cat. After editing, the file
		// is re-encrypted with the private key
		// and put back in place.
		case "-e", "--edit":
			args = args[1:]
			if len(args) == 0 {
				fmt.Println("no FILE specified")
				os.Exit(1)
			}
			fileToEdit = args[0]
			args = args[1:]

		// -i or --identity takes a private key as an argument
		// and uses it to decrypt the file. Only to be used
		// with other flags.
		case "-i", "--identity":
			args = args[1:]
			if len(args) == 0 {
				fmt.Println("no PRIVATE_KEY specified")
				os.Exit(1)
			}
			privateKeyPath = args[0]
			args = args[1:]

		// Rekey takes no arguments and rekeys the files
		// specified in the rules file.
		case "-r", "--rekey":
			rekey = true
			args = args[1:]

		// -d or --decrypt takes a file as an argument
		// and decrypts it using the given private key.
		case "-d", "--decrypt":
			decryptOnly = true
			args = args[1:]
			if len(args) == 0 {
				fmt.Println("no FILE specified")
				os.Exit(1)
			}
			fileToEdit = args[0]
			args = args[1:]

		// No args are provided. Display the help menu and exit.
		default:
			help.ShowHelp()
			os.Exit(1)
		}
	}

	// If no args are provided, display the help menu
	if len(args) == 0 && !rekey && !decryptOnly && fileToEdit == "" {
		help.ShowHelp()
		os.Exit(1)
	}

	// Look for the $RULES variable
	// if not set, try ./secrets.nix (default)
	rules := os.Getenv("RULES")
	if rules == "" {
		rules = config.DefaultRulesFile
	}

	// If rekey is requested, rekey and exit
	if rekey {
		util.RekeyFiles(rules)
		os.Exit(0)
	}

	// If we're performing a decryption operation only
	// perform that and exit
	if decryptOnly {
		util.DecryptFile(fileToEdit, privateKeyPath, rules)
		os.Exit(0)
	}

	// Edit the given file with private key path and rules.
	util.EditFile(fileToEdit, privateKeyPath, rules)
}
