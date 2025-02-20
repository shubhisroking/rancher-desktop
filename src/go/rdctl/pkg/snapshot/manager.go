package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/rancher-sandbox/rancher-desktop/src/go/rdctl/pkg/paths"
)

const completeFileName = "complete.txt"
const completeFileContents = "The presence of this file indicates that this snapshot is complete and valid."
const maxNameLength = 250
const nameDisplayCutoffSize = 30

var ErrNameExists = errors.New("name already exists")
var ErrIncompleteSnapshot = errors.New("snapshot is not complete")

func writeMetadataFile(appPaths paths.Paths, snapshot Snapshot) error {
	snapshotDir := filepath.Join(appPaths.Snapshots, snapshot.ID)
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}
	metadataPath := filepath.Join(snapshotDir, "metadata.json")
	metadataFile, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer metadataFile.Close()
	encoder := json.NewEncoder(metadataFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(snapshot); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}
	return nil
}

// Manager handles all snapshot-related functionality.
type Manager struct {
	Paths       paths.Paths
	Snapshotter Snapshotter
}

func NewManager(paths paths.Paths) Manager {
	return Manager{
		Paths:       paths,
		Snapshotter: NewSnapshotterImpl(paths),
	}
}

func (manager Manager) GetSnapshotId(desiredName string) (string, error) {
	snapshots, err := manager.List(false)
	if err != nil {
		return "", fmt.Errorf("failed to list snapshots: %w", err)
	}
	for _, candidate := range snapshots {
		if desiredName == candidate.Name {
			return candidate.ID, nil
		}
	}
	return "", fmt.Errorf(`can't find snapshot %q`, desiredName)
}

// ValidateName - does syntactic validation on the name
func (manager Manager) ValidateName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("snapshot name must not be the empty string")
	}
	reportedName := name
	if len(reportedName) > nameDisplayCutoffSize {
		reportedName = reportedName[0:nameDisplayCutoffSize] + "…"
	}
	if len(name) > maxNameLength {
		return fmt.Errorf(`invalid name %q: max length is %d, %d were specified`, reportedName, maxNameLength, len(name))
	}
	if err := checkForInvalidCharacter(name); err != nil {
		return err
	}
	if unicode.IsSpace(rune(name[0])) {
		return fmt.Errorf(`invalid name %q: must not start with a white-space character`, reportedName)
	}
	if unicode.IsSpace(rune(name[len(name)-1])) {
		if len(name) > nameDisplayCutoffSize {
			reportedName = "…" + name[len(name)-nameDisplayCutoffSize:]
		}
		return fmt.Errorf(`invalid name %q: must not end with a white-space character`, reportedName)
	}
	currentSnapshots, err := manager.List(false)
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}
	for _, currentSnapshot := range currentSnapshots {
		if currentSnapshot.Name == name {
			return fmt.Errorf("invalid name %q: %w", name, ErrNameExists)
		}
	}
	return nil
}

// Create a new snapshot.
func (manager Manager) Create(name, description string) (*Snapshot, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID for snapshot: %w", err)
	}
	snapshot := Snapshot{
		Created:     time.Now(),
		Name:        name,
		ID:          id.String(),
		Description: description,
	}

	// do operations that can fail, rolling back if failure is encountered
	snapshotDir := filepath.Join(manager.Paths.Snapshots, snapshot.ID)
	if err := manager.Snapshotter.CreateFiles(snapshot); err != nil {
		if err2 := os.RemoveAll(snapshotDir); err2 != nil {
			err = errors.Join(err, fmt.Errorf("failed to delete created snapshot directory: %w", err2))
		}
		return nil, err
	}

	return &snapshot, nil
}

// List snapshots that are present on the system. If includeIncomplete is
// true, includes snapshots that are currently being created, are currently
// being deleted, or are otherwise incomplete and cannot be restored from.
func (manager Manager) List(includeIncomplete bool) ([]Snapshot, error) {
	dirEntries, err := os.ReadDir(manager.Paths.Snapshots)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return []Snapshot{}, fmt.Errorf("failed to read snapshots directory: %w", err)
	}
	snapshots := make([]Snapshot, 0, len(dirEntries))
	for _, dirEntry := range dirEntries {
		if _, err := uuid.Parse(dirEntry.Name()); err != nil {
			continue
		}
		snapshot := Snapshot{}
		metadataPath := filepath.Join(manager.Paths.Snapshots, dirEntry.Name(), "metadata.json")
		contents, err := os.ReadFile(metadataPath)
		if err != nil {
			return []Snapshot{}, fmt.Errorf("failed to read %q: %w", metadataPath, err)
		}
		if err := json.Unmarshal(contents, &snapshot); err != nil {
			return []Snapshot{}, fmt.Errorf("failed to unmarshal contents of %q: %w", metadataPath, err)
		}
		snapshot.Created = snapshot.Created.Local()

		completeFilePath := filepath.Join(manager.Paths.Snapshots, snapshot.ID, completeFileName)
		_, err = os.Stat(completeFilePath)
		completeFileExists := err == nil

		if !includeIncomplete && !completeFileExists {
			continue
		}

		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}

// Delete a snapshot.
func (manager Manager) Delete(id string) error {
	snapshotDir := filepath.Join(manager.Paths.Snapshots, id)
	// Remove complete.txt file. This must be done first because restoring
	// from a partially-deleted snapshot could result in errors.
	completeFilePath := filepath.Join(snapshotDir, completeFileName)
	if err := os.RemoveAll(completeFilePath); err != nil {
		return fmt.Errorf("failed to remove %q: %w", completeFileName, err)
	}
	if err := os.RemoveAll(snapshotDir); err != nil {
		return fmt.Errorf("failed to remove dir %q: %w", snapshotDir, err)
	}
	return nil
}

// Restore Rancher Desktop to the state saved in a snapshot.
func (manager Manager) Restore(id string) error {
	// Before doing anything, ensure that the snapshot is complete
	completeFilePath := filepath.Join(manager.Paths.Snapshots, id, completeFileName)
	if _, err := os.Stat(completeFilePath); err != nil {
		return fmt.Errorf("snapshot %q: %w", id, ErrIncompleteSnapshot)
	}

	// Get metadata about snapshot
	metadataPath := filepath.Join(manager.Paths.Snapshots, id, "metadata.json")
	contents, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata for snapshot %q: %w", id, err)
	}
	snapshot := Snapshot{}
	if err := json.Unmarshal(contents, &snapshot); err != nil {
		return fmt.Errorf("failed to unmarshal contents of %q: %w", metadataPath, err)
	}

	if err := manager.Snapshotter.RestoreFiles(snapshot); err != nil {
		return fmt.Errorf("failed to restore files: %w", err)
	}

	return nil
}

func checkForInvalidCharacter(name string) error {
	for idx, c := range name {
		if !unicode.IsPrint(rune(c)) {
			return fmt.Errorf("invalid character value %d at position %d in name: all characters must be printable or a space", c, idx)
		}
	}
	return nil
}
