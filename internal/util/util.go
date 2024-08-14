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
		fmt.Printf("rekeying %s...\n", file)
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

// FIXME: the file is cleared and never rewritten. We need to ensure that
// this never happens in production.
func EditFile(file, privateKeyPath, rules string) {
	keys, err := GetKeysForFile(file, rules)
	if err != nil {
		fmt.Fprintf(os.Stderr, "There is no rule for %s in %s\n", file, rules)
		os.Exit(1)
	}

	cleartextDir, err := os.MkdirTemp("", "agenix-cleartext")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(cleartextDir)

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

	err = os.WriteFile(file, []byte{}, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to clear original file: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(file, []byte{}, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write edited content to original file: %v\n", err)
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

func EncryptFile(input, output string, keys ...string) error {
	fmt.Printf("Encrypting %s to %s with keys %v\n", input, output, keys)
	args := []string{"--encrypt"}

	for _, key := range keys {
		args = append(args, "--recipient", key)
	}

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
