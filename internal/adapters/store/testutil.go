package adapters

import "os"

// tempfile returns a temporary file path.
func tempfile(name string) string {
	f, err := os.CreateTemp("", name+"-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}
