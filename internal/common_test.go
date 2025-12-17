package internal

import (
	"os"
	"testing"
)

func TestCreateDirIfNotExists(t *testing.T) {

	tests := []struct {
		name         string
		path         string
		expectErr    bool
		expectExists bool
		expectMkdir  bool
	}{
		{"DirectoryExists", "existing_dir", false, true, false},
		{"DirectoryDoesNotExist", "new_dir", false, false, true},
		{"DirectoryCreationError", "error_dir", true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mkdirCalled := false

			fileExists := func(path string) bool {
				return tt.expectExists
			}

			mkdir := func(path string, perm os.FileMode) error {
				mkdirCalled = true
				if tt.expectErr {
					return os.ErrPermission
				}
				return nil
			}

			err := createDirIfNotExistsSilent(tt.path, fileExists, mkdir)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error for %v: %v, got: %v", tt.name, tt.expectErr, err)
			}
			if mkdirCalled != tt.expectMkdir {
				t.Errorf("Expected mkdir called for %v: %v, got: %v", tt.name, tt.expectMkdir, mkdirCalled)
			}
		})
	}
}
