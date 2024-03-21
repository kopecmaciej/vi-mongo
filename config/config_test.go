package config

import (
	"testing"
)

func TestParseMongoDBURI(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		wantHost string
		wantPort string
		wantDb   string
		wantErr  bool
	}{
		{
			name:     "Valid standard URI",
			uri:      "mongodb://localhost:27017/mydatabase",
			wantHost: "localhost",
			wantPort: "27017",
			wantDb:   "mydatabase",
			wantErr:  false,
		},
		{
			name:     "Valid srv URI",
			uri:      "mongodb+srv://user:password@example.mongodb.net/mydatabase?retryWrites=true",
			wantHost: "example.mongodb.net",
			wantPort: "27017",
			wantDb:   "mydatabase",
			wantErr:  false,
		},
		{
			name:     "Valid URI without port",
			uri:      "mongodb://localhost/mydatabase",
			wantHost: "localhost",
			wantPort: "27017",
			wantDb:   "mydatabase",
			wantErr:  false,
		},
		{
			name:     "Invalid prefix",
			uri:      "incorrect://localhost:27017/mydatabase",
			wantHost: "",
			wantPort: "",
			wantDb:   "",
			wantErr:  true,
		},
		{
			name:     "URI with options",
			uri:      "mongodb://localhost:27017/mydatabase?retryWrites=true",
			wantHost: "localhost",
			wantPort: "27017",
			wantDb:   "mydatabase",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPort, gotDb, err := ParseMongoDBURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMongoDBURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHost != tt.wantHost {
				t.Errorf("ParseMongoDBURI() gotHost = %v, want %v", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("ParseMongoDBURI() gotPort = %v, want %v", gotPort, tt.wantPort)
			}
			if gotDb != tt.wantDb {
				t.Errorf("ParseMongoDBURI() gotDb = %v, want %v", gotDb, tt.wantDb)
			}
		})
	}
}
