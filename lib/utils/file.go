package utils

import "os"

func GetEazyHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	eazyDir := homeDir + "/.eazy"

	return eazyDir, nil
}
