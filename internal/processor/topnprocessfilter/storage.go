package topnprocessfilter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type ProcessStateStorage interface {
	Load() (map[int]*TrackedProcess, error)

	Save(map[int]*TrackedProcess) error

	Close() error
}
type FileStorage struct {
	filePath string
	mu       sync.Mutex
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	return &FileStorage{
		filePath: filePath,
	}, nil
}

func (s *FileStorage) Load() (map[int]*TrackedProcess, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		// Return empty map if file doesn't exist yet
		return make(map[int]*TrackedProcess), nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	var processes map[int]*TrackedProcess
	if err := json.Unmarshal(data, &processes); err != nil {
		return nil, err
	}

	return processes, nil
}

func (s *FileStorage) Save(processes map[int]*TrackedProcess) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.Marshal(processes)
	if err != nil {
		return err
	}

	tempFile := s.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tempFile, s.filePath)
}

func (s *FileStorage) Close() error {
	return nil
}

func createDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}
