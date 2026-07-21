package modal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathAutocompleteEntries(t *testing.T) {
	directory := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(directory, "nested"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "existing.json"), []byte("[]"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "ignored.txt"), []byte("ignored"), 0o600))

	entries := pathAutocompleteEntries(directory + string(filepath.Separator))
	require.Len(t, entries, 2)
	require.Equal(t, filepath.Join(directory, "nested")+string(filepath.Separator), entries[0].Main)
	require.Equal(t, "Directory", entries[0].Secondary)
	require.Equal(t, filepath.Join(directory, "existing.json"), entries[1].Main)
	require.Equal(t, "JSON file", entries[1].Secondary)
}

func TestPathAutocompleteEntriesFiltersPrefixCaseInsensitively(t *testing.T) {
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, "Report.JSON"), []byte("[]"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "other.json"), []byte("[]"), 0o600))

	entries := pathAutocompleteEntries(filepath.Join(directory, "rep"))
	require.Len(t, entries, 1)
	require.Equal(t, filepath.Join(directory, "Report.JSON"), entries[0].Main)
}
