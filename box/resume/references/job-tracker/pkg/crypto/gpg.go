package crypto

import (
	"fmt"
	"os"
	"os/exec"
)

type GPGCrypto struct {
	recipient string
}

func NewGPGCrypto(recipient string) *GPGCrypto {
	return &GPGCrypto{
		recipient: recipient,
	}
}

func (g *GPGCrypto) Encrypt(inputFile, outputFile string) error {
	cmd := exec.Command("gpg", "--encrypt", "--trust-model", "always", "--yes", "--quiet", "--recipient", g.recipient, "--output", outputFile, inputFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to encrypt: %w", err)
	}
	return nil
}

func (g *GPGCrypto) Decrypt(inputFile string) ([]byte, error) {
	cmd := exec.Command("gpg", "--decrypt", "--quiet", inputFile)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}
	return output, nil
}

func (g *GPGCrypto) EncryptData(data []byte, outputFile string) error {
	tempFile := outputFile + ".tmp"
	
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}
	defer os.Remove(tempFile)
	
	return g.Encrypt(tempFile, outputFile)
}