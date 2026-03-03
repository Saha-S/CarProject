package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadData(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading data file: %w", err)
	}
	if err := json.Unmarshal(data, &db); err != nil {
		return fmt.Errorf("parsing data file: %w", err)
	}
	return nil
}
