package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Store manage the configuration IO for the Clox CLI app.
//
// Store should be created by calling NewStore.
type Store struct {
	// The path to the .clox directory. Path will always be the path to the users directory
	// with /.clox appended at the end.
	Path string
}

// NewStore creates a Store and sets the Path to the users home directory joined with ".clox".
// If it cannot get the users home directory an error is returned.
func NewStore() (*Store, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed getting home directory: %w", err)
	}

	return &Store{
		Path: filepath.Join(homeDir, ".clox"),
	}, nil
}

// DirExists checks if the ".clox" directory exists on the file system. The path to the
// ".clox" directory is the value of this Store's Path value.
func (s *Store) DirExists() (bool, error) {
	fi, err := os.Stat(s.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	if fi.IsDir() {
		return true, nil
	}

	return false, errors.New(".clox already exists as a file in home directory")
}

// The parameters when writing the config.json file.
type WriteFileParams struct {
	Password string `json:"password"`
	APIToken string `json:"api_token"`
}

// Write will write the parameters to the config.json file. The config.json file will be
// stored within the Path of this Store on the file system.
func (s *Store) Write(p WriteFileParams) error {
	if err := os.Mkdir(s.Path, 0700); err != nil {
		return fmt.Errorf("failed creating directory %s: %w", s.Path, err)
	}

	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed marshalling data to json: %w", err)
	}

	filePath := filepath.Join(s.Path, "config.json")
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed writing file %s: %w", filePath, err)
	}

	return nil
}
