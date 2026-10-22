package files

import "os"

// Read returns the contents of the file at path as text, ready to hand to a
// viewer.
func Read(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
