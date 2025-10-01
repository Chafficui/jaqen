package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Profile represents a saved configuration profile
type Profile struct {
	Name      string      `json:"name"`
	GamePath  string      `json:"game_path"` // Path to FM installation
	Config    JaqenConfig `json:"config"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// ProfileManager manages profiles
type ProfileManager struct {
	profilesDir string
	profiles    map[string]*Profile
}

// NewProfileManager creates a new profile manager
func NewProfileManager() (*ProfileManager, error) {
	configDir, err := GetUserConfigDir()
	if err != nil {
		return nil, err
	}

	profilesDir := filepath.Join(configDir, "profiles")

	// Create profiles directory if it doesn't exist
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create profiles directory: %w", err)
	}

	pm := &ProfileManager{
		profilesDir: profilesDir,
		profiles:    make(map[string]*Profile),
	}

	// Load existing profiles
	if err := pm.loadProfiles(); err != nil {
		return nil, err
	}

	return pm, nil
}

// loadProfiles loads all profiles from disk
func (pm *ProfileManager) loadProfiles() error {
	entries, err := os.ReadDir(pm.profilesDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		profilePath := filepath.Join(pm.profilesDir, entry.Name())
		profile, err := pm.loadProfile(profilePath)
		if err != nil {
			// Log error but continue loading other profiles
			fmt.Printf("Warning: Failed to load profile %s: %v\n", entry.Name(), err)
			continue
		}

		pm.profiles[profile.Name] = profile
	}

	return nil
}

// loadProfile loads a single profile from disk
func (pm *ProfileManager) loadProfile(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// SaveProfile saves a profile to disk
func (pm *ProfileManager) SaveProfile(profile *Profile) error {
	profile.UpdatedAt = time.Now()

	profilePath := filepath.Join(pm.profilesDir, profile.Name+".json")

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		return err
	}

	pm.profiles[profile.Name] = profile
	return nil
}

// GetProfile returns a profile by name
func (pm *ProfileManager) GetProfile(name string) (*Profile, error) {
	profile, exists := pm.profiles[name]
	if !exists {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}
	return profile, nil
}

// ListProfiles returns a sorted list of profile names
func (pm *ProfileManager) ListProfiles() []string {
	names := make([]string, 0, len(pm.profiles))
	for name := range pm.profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// CreateProfile creates a new profile based on an existing one
func (pm *ProfileManager) CreateProfile(name string, basedOn *Profile) (*Profile, error) {
	if _, exists := pm.profiles[name]; exists {
		return nil, fmt.Errorf("profile '%s' already exists", name)
	}

	newProfile := &Profile{
		Name:      name,
		GamePath:  basedOn.GamePath,
		Config:    basedOn.Config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := pm.SaveProfile(newProfile); err != nil {
		return nil, err
	}

	return newProfile, nil
}

// CreateProfileForGame creates a new profile for a specific FM installation
func (pm *ProfileManager) CreateProfileForGame(name string, gamePath string) (*Profile, error) {
	if _, exists := pm.profiles[name]; exists {
		return nil, fmt.Errorf("profile '%s' already exists", name)
	}

	// Create profile with game-specific settings
	newProfile := &Profile{
		Name:     name,
		GamePath: gamePath,
		Config: JaqenConfig{
			Preserve:        &[]bool{true}[0],
			XMLPath:         &[]string{""}[0], // Will be set when user selects image folder
			RTFPath:         &[]string{""}[0],
			IMGPath:         &[]string{""}[0],
			FMVersion:       &[]string{DefaultFMVersion}[0],
			AllowDuplicate:  &[]bool{true}[0],
			MappingOverride: &map[string]string{},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := pm.SaveProfile(newProfile); err != nil {
		return nil, err
	}

	return newProfile, nil
}

// DeleteProfile deletes a profile
func (pm *ProfileManager) DeleteProfile(name string) error {
	if _, exists := pm.profiles[name]; !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	profilePath := filepath.Join(pm.profilesDir, name+".json")
	if err := os.Remove(profilePath); err != nil {
		return err
	}

	delete(pm.profiles, name)
	return nil
}

// GetFirstProfile returns the first available profile (alphabetically)
func (pm *ProfileManager) GetFirstProfile() (*Profile, error) {
	profiles := pm.ListProfiles()
	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profiles available")
	}

	return pm.GetProfile(profiles[0])
}
