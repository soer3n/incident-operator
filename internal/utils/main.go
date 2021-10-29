package utils

import (
	"fmt"
	"os"
)

// Contains represents func for checking if a string is in a list of strings
func Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// WriteFile represents func for writing content to a local file
func WriteFile(name, path string, content []byte) error {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	f, err := os.Create(path + "/" + name)

	if err != nil {
		return err
	}

	l, err := f.Write(content)

	if err != nil {
		return err
	}

	fmt.Println(l, "bytes written successfully")
	return nil
}
