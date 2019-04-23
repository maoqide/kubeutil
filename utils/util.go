package utils

import (
	"io/ioutil"
	"log"
	"os"
)

// ReadFile read bytes from file
func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	return b, err
}
