package environment

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/auth"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/config"
)

type Snapshot struct {
	ConfigPath        string `json:"configPath"`
	ConfigPresent     bool   `json:"configPresent"`
	DefaultAppID      string `json:"defaultAppId,omitempty"`
	KeychainAvailable bool   `json:"keychainAvailable"`
	KeychainBypassed  bool   `json:"keychainBypassed"`
	WorkflowPath      string `json:"workflowPath"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Snapshot() (Snapshot, error) {
	configPath, err := config.Path()
	if err != nil {
		return Snapshot{}, err
	}

	cfg, cfgErr := config.Load()
	keychainAvailable, keychainErr := auth.KeychainAvailable()
	if keychainErr != nil {
		return Snapshot{}, keychainErr
	}
	if cfgErr != nil && !errors.Is(cfgErr, config.ErrNotFound) {
		return Snapshot{}, cfgErr
	}

	workflowPath := ""
	if cwd, err := os.Getwd(); err == nil {
		workflowPath = filepath.Join(cwd, ".asc", "workflow.json")
	}

	snapshot := Snapshot{
		ConfigPath:        configPath,
		ConfigPresent:     cfgErr == nil,
		KeychainAvailable: keychainAvailable,
		KeychainBypassed:  auth.ShouldBypassKeychain(),
		WorkflowPath:      workflowPath,
	}
	if cfg != nil {
		snapshot.DefaultAppID = cfg.AppID
	}
	return snapshot, nil
}
