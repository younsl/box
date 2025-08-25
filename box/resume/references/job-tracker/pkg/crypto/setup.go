package crypto

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type GPGKey struct {
	ID    string
	Email string
	Name  string
}

func ValidateAndSetupGPG(recipient string) (string, error) {
	if recipient == "" {
		return promptForGPGRecipient()
	}

	if err := validateGPGKey(recipient); err != nil {
		return "", fmt.Errorf("GPG key validation failed for %s: %w", recipient, err)
	}

	// Test encryption to ensure everything works
	if err := testGPGEncryption(recipient); err != nil {
		return "", fmt.Errorf("GPG encryption test failed for %s: %w", recipient, err)
	}

	fmt.Printf("✓ GPG validated: %s\n", recipient)
	return recipient, nil
}

func promptForGPGRecipient() (string, error) {
	keys, err := listGPGKeys()
	if err != nil {
		return "", fmt.Errorf("failed to list GPG keys: %w", err)
	}

	if len(keys) == 0 {
		fmt.Println("No GPG keys found. Please generate a GPG key first:")
		fmt.Println("  gpg --full-generate-key")
		return "", fmt.Errorf("no GPG keys available")
	}

	fmt.Println("\n=== GPG Key Selection ===")
	fmt.Println("Available GPG keys:")
	
	validKeys := []GPGKey{}
	for _, key := range keys {
		if key.Email != "" && !strings.Contains(key.Name, "expired") {
			validKeys = append(validKeys, key)
			fmt.Printf("%d. %s (%s)\n", len(validKeys), key.Email, key.Name)
		}
	}

	if len(validKeys) == 0 {
		return "", fmt.Errorf("no valid GPG keys found")
	}

	fmt.Printf("Select a key (1-%d): ", len(validKeys))
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	var selection int
	if _, err := fmt.Sscanf(strings.TrimSpace(input), "%d", &selection); err != nil {
		return "", fmt.Errorf("invalid selection")
	}

	if selection < 1 || selection > len(validKeys) {
		return "", fmt.Errorf("selection out of range")
	}

	selectedKey := validKeys[selection-1]
	
	// Test encryption with selected key
	fmt.Printf("Testing encryption with %s...\n", selectedKey.Email)
	if err := testGPGEncryption(selectedKey.Email); err != nil {
		return "", fmt.Errorf("encryption test failed: %w", err)
	}

	fmt.Printf("✓ Selected GPG key: %s\n", selectedKey.Email)
	fmt.Printf("💡 Tip: Set environment variable to skip this prompt:\n")
	fmt.Printf("   export GPG_RECIPIENT=\"%s\"\n\n", selectedKey.Email)
	
	return selectedKey.Email, nil
}

func listGPGKeys() ([]GPGKey, error) {
	cmd := exec.Command("gpg", "--list-secret-keys", "--keyid-format", "LONG")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseGPGKeys(string(output)), nil
}

func parseGPGKeys(output string) []GPGKey {
	var keys []GPGKey
	lines := strings.Split(output, "\n")
	
	uidRegex := regexp.MustCompile(`uid\s+\[\s*(\w+)\s*\]\s*(.+)\s*<(.+)>`)
	
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "uid") {
			matches := uidRegex.FindStringSubmatch(line)
			if len(matches) >= 4 {
				status := matches[1]
				name := strings.TrimSpace(matches[2])
				email := strings.TrimSpace(matches[3])
				
				keys = append(keys, GPGKey{
					ID:    status,
					Name:  name,
					Email: email,
				})
			}
		}
	}
	
	return keys
}

func validateGPGKey(recipient string) error {
	cmd := exec.Command("gpg", "--list-keys", recipient)
	return cmd.Run()
}

func testGPGEncryption(recipient string) error {
	tempData := []byte("test data")
	tempFile := "/tmp/gpg_test_" + strings.Replace(recipient, "@", "_", -1)
	
	if err := os.WriteFile(tempFile, tempData, 0644); err != nil {
		return err
	}
	defer os.Remove(tempFile)
	
	encFile := tempFile + ".gpg"
	defer os.Remove(encFile)
	
	// Test encryption
	cmd := exec.Command("gpg", "--encrypt", "--trust-model", "always", "--yes", "--recipient", recipient, "--output", encFile, tempFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}
	
	// Test decryption
	cmd = exec.Command("gpg", "--decrypt", "--quiet", encFile)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}
	
	if string(output) != string(tempData) {
		return fmt.Errorf("decryption result mismatch")
	}
	
	return nil
}