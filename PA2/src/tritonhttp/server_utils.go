package tritonhttp

import (
	"io/ioutil"
	"strings"
)

/**
	Load and parse the mime.types file
**/
func ParseMIME(MIMEPath string) (MIMEMap map[string]string, err error) {
	mimeMap := make(map[string]string)

	// Open mime.types, read line by line and push to mimeMap
	data, err := ioutil.ReadFile(MIMEPath)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		split := strings.Split(line, " ")
		mimeMap[split[0]] = split[1]
	}

	return mimeMap, nil
}
