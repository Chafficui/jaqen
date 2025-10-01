package internal

const (
	DefaultPreserve       = false
	DefaultXMLPath        = "./config.xml"
	DefaultRTFPath        = "./newgen.rtf"
	DefaultImagesPath     = "./"
	DefaultFMVersion      = "2024"
	DefaultAllowDuplicate = false
)

// GetDefaultConfigPath returns the default config path, preferring user config directory
// Falls back to local directory if user config directory cannot be accessed
func GetDefaultConfigPath() string {
	if userConfigPath, err := GetUserConfigPath(); err == nil {
		return userConfigPath
	}
	// Fallback to local directory
	return "./jaqen.toml"
}
