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
	defer os.RemoveAll(cleartextDir)

	cleartextFile := filepath.Join(cleartextDir, filepath.Base(file))
	err = DecryptFile(file, privateKeyPath, cleartextFile, keys...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decrypt file: %v\n", err)
		os.Exit(1)
	}

	// Create a backup of the cleartext file
	backupFile := cleartextFile + ".before"
	err = CopyFile(cleartextFile, backupFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create backup file: %v\n", err)
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

	// Check if the cleartext file was created/modified
	if _, err := os.Stat(cleartextFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: %s wasn't created.\n", file)
		return
	}

	// Compare original and edited files
	if _, err := os.Stat(backupFile); err == nil {
		diffCmd := exec.Command("diff", "-q", backupFile, cleartextFile)
		if err := diffCmd.Run(); err == nil {
			fmt.Printf("Warning: %s wasn't changed, skipping re-encryption.\n", file)
			return
		}
	}

	// Prepare to re-encrypt the edited content
	encryptionArgs := []string{"--encrypt"}
	for _, key := range keys {
		if key != "" {
			encryptionArgs = append(encryptionArgs, "--recipient", key)
		}
	}

	reEncryptedDir, err := os.MkdirTemp("", "agenix-reencrypted")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp directory for re-encryption: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(reEncryptedDir)

	reEncryptedFile := filepath.Join(reEncryptedDir, filepath.Base(file))
	encryptionArgs = append(encryptionArgs, "-o", reEncryptedFile)

	cmd = exec.Command(AgeBinary, encryptionArgs...)
	cmd.Stdin = strings.NewReader(cleartextFile)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encrypt file: %s\n", stderr.String())
		os.Exit(1)
	}

	// Move the re-encrypted file to the original location
	err = os.MkdirAll(filepath.Dir(file), 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create directory for original file: %v\n", err)
		os.Exit(1)
	}

	err = os.Rename(reEncryptedFile, file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to move re-encrypted file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("File successfully edited and saved.")
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

func cleanup() {
	// Placeholder but not really. This lets me handle cleanup separately
	// if I really need to.
	fmt.Println("Cleanup completed.")
}

func DecryptFile(file, privateKeyPath, output string, keys ...string) error {
	args := []string{"--decrypt"}
	args = append(args, "-o", output, file)

	for _, key := range keys {
		args = append(args, "--identity", key)
	}

	cmd := exec.Command(AgeBinary, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to decrypt file: %s, stderr: %s", err, stderr.String())
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
	// Use DefaultRulesFile if rules is empty
	if rules == "" {
		rules = config.DefaultRulesFile
	}

	// Convert rules path to absolute paths to satisfy Nix eval
	absRulesPath, err := filepath.Abs(rules)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path of rules file: %w", err)
	}

	cmd := exec.Command(NixInstantiate, "--json", "--eval", "-E", fmt.Sprintf("(let rules = import %q; in builtins.attrNames rules)", absRulesPath))
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to get files from rules: %w, stderr: %s", err, stderr.String())
	}

	var files []string
	err = json.Unmarshal(out.Bytes(), &files)
	if err != nil {
		return nil, fmt.Errorf("failed to parse files: %w", err)
	}

	fmt.Print("Files: ", files)
	return files, nil
}

func IsStdinInteractive() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&fs.ModeCharDevice != 0
}
