package common

import "os"

// tempfile returns a temporary file path.
func Tempfile(name string) (string, error) {
	f, err := os.CreateTemp("", name+"-")
	if err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	if err := os.Remove(f.Name()); err != nil {
		return "", err
	}
	return f.Name(), nil
}
