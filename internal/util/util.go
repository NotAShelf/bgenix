package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"bgenix/internal/config"
)

var (
	AgeBinary      = config.AgeBinary
	NixInstantiate = config.NixInstantiate
)

type Rule struct {
	PublicKeys []string `json:"publicKeys"`
}

func RekeyFiles(rules string) {
	files, err := GetFilesFromRules(rules)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting files from rules:", err)
		os.Exit(1)
	}

	for _, file := range files {
		fmt.Printf("Rekeying %s...\n", file)
		EditFile(file, "", rules)
	}
}

func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// TODO: agenix copies the cleartext file to cleartext.before
// and compares the cleartext file with the cleartext.before file
// to check if the file was modified. We should do the same.
func EditFile(file, privateKeyPath, rules string) {
	keys, err := GetKeysForFile(file, rules)
	if err != nil {
		fmt.Fprintf(os.Stderr, "There is no rule for %s in %s\n", file, rules)
		os.Exit(1)
	}

	cleartextDir, err := os.MkdirTemp("", "agenix-decrypted")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := os.RemoveAll(cleartextDir); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove temp directory: %v\n", err)
		}
	}()

	cleartextFile := filepath.Join(cleartextDir, filepath.Base(file))
	err = DecryptFile(file, privateKeyPath, cleartextFile, keys...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decrypt file: %v\n", err)
		os.Exit(1)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" || !IsStdinInteractive() {
		editor = "cat"
	}

	cmd := exec.Command(editor, cleartextFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to edit file: %v\n", err)
		os.Exit(1)
	}

	// Read the edited content from the temporary file
	editedContent, err := os.ReadFile(cleartextFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read edited content: %v\n", err)
		os.Exit(1)
	}

	// Re-encrypt the edited content and overwrite the original file
	err = EncryptFile(string(editedContent), file, privateKeyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encrypt edited file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("File successfully edited and saved.")
}

func DecryptFile(file, privateKeyPath, output string, keys ...string) error {
	args := []string{"--decrypt"}
	if privateKeyPath != "" {
		args = append(args, "--identity", privateKeyPath)
	}
	args = append(args, "-o", output, file)

	cmd := exec.Command(AgeBinary, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("failed to decrypt file: %s", stderr.String())
	}

	return nil
}

func EncryptFile(input, output string, privateKeyPath string) error {
	fmt.Printf("Encrypting %s to %s with private key %s\n", input, output, privateKeyPath)
	args := []string{"--encrypt"}

	// Use the private key for encryption
	// XXX: can we try using public keys? Would that make sense
	// for secrets but not so secrets? No? Okay.
	args = append(args, "--identity", privateKeyPath)

	args = append(args, "-o", output)
	cmd := exec.Command(AgeBinary, args...)
	cmd.Stdin = strings.NewReader(input)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("failed to encrypt file: %s", stderr.String())
	}

	return nil
}

func GetKeysForFile(file, rules string) ([]string, error) {
	absRulesPath, err := filepath.Abs(rules)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path of rules file: %w", err)
	}

	// XXX: Extremely hit-or-miss.
	cmd := exec.Command(NixInstantiate, "--json", "--eval", "--strict", "-E", fmt.Sprintf("(let rules = import %q; in rules.\"%s\".publicKeys)", absRulesPath, file))

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()

	if err != nil {
		fmt.Printf("Command failed with error: %s\n", err)
		fmt.Printf("Standard Error:\n%s\n", stderr.String())
		return nil, fmt.Errorf("failed to get keys for file: %w", err)
	}

	var keys []string
	err = json.Unmarshal(out.Bytes(), &keys)
	if err != nil {
		return nil, fmt.Errorf("failed to parse keys: %w", err)
	}

	return keys, nil
}

func GetFilesFromRules(rules string) ([]string, error) {
	cmd := exec.Command(NixInstantiate, "--json", "--eval", "-E", fmt.Sprintf("(let rules = import \"%q\"; in builtins.attrNames rules)", rules))

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to get files from rules: %w", err)
	}

	var files []string
	err = json.Unmarshal(out.Bytes(), &files)
	if err != nil {
		return nil, fmt.Errorf("failed to parse files: %w", err)
	}

	fmt.Print("Files:", files)

	return files, nil
}

func IsStdinInteractive() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&fs.ModeCharDevice != 0
}
