package tritonhttp

import (
	"bufio"
	"os"
	"strings"
)

/**
	Load and parse the mime.types file
**/
func ParseMIME(MIMEPath string) (MIMEMap map[string]string, err error) {
	mimeMap := make(map[string]string)

	// Open mime.types
	file, err := os.Open(MIMEPath)
	if err != nil {
		return nil, err
	}

	// Read line by line and push to mimeMap
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		keyValuePair := strings.Split(scanner.Text(), " ")
		mimeMap[keyValuePair[0]] = keyValuePair[1]
	}
	file.Close()

	return mimeMap, nil
}
