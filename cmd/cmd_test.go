package cmd

import (
	"strings"
	"testing"
)

func TestValidateDirectNavigateFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid format with hyphens",
			input:   "my-db/my-collection",
			wantErr: false,
		},
		{
			name:    "invalid - missing collection",
			input:   "mydb/",
			wantErr: true,
			errMsg:  "both db-name and collection-name are required",
		},
		{
			name:    "invalid - missing database",
			input:   "/mycollection",
			wantErr: true,
			errMsg:  "both db-name and collection-name are required",
		},
		{
			name:    "invalid - too many slashes",
			input:   "mydb/mycollection/extra",
			wantErr: true,
			errMsg:  "format should be db-name/collection-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDirectNavigateFormat(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateDirectNavigateFormat() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateDirectNavigateFormat() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateDirectNavigateFormat() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
