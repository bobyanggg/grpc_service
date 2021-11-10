package model

import (
	"encoding/json"
	"os"
)

func OpenJson(filePath string) (map[string]interface{}, error) {
	// Open the file.
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	// Parse
	var data map[string]interface{}
	if err = json.NewDecoder(jsonFile).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}
