package viper

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/abhissng/neuron/adapters/vault"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/spf13/viper"
)

// Viper struct holds the configuration for the Viper client
type Viper struct {
	configName string
	configType string
	configPath string // it should only contain the absolute path for the folder rest other details will be added by sdk
}

// NewViper creates the viper configuration using the RunMode environment.
func NewViper(configName, configType, configPath string) *Viper {
	env := helpers.GetEnvironment()
	if helpers.IsEmpty(env) {
		env = "dev" // default enviroment
	}
	// Remove the trailing slash if it exists
	configPath = strings.TrimSuffix(configPath, "/")

	return &Viper{
		configName: configName,
		configType: configType,
		configPath: configPath + "/" + env + "/",
	}
}

// InitialiseViper initialises the viper client
func (v *Viper) InitialiseViper() error {
	viper.SetConfigName(v.configName) // Name of configuration file
	viper.SetConfigType(v.configType) // Configuration file type
	viper.AddConfigPath(v.configPath) // Look for configuration file in the given directory

	// Enable Viper to read environment variables
	viper.AutomaticEnv()

	// Attempt to read configuration file
	if err := viper.ReadInConfig(); err != nil {
		err = fmt.Errorf("error reading configuration file: %s", err)
		return err
	}

	return nil
}

// LoadDynamicConfig loads the configuration and replaces placeholders with values fetched from Vault
func (v *Viper) LoadDynamicConfig(vault *vault.Vault) error {
	if vault == nil {
		return errors.New("vault cannot be nil in case of loading the dynamic configuration")
	}

	return loadAndReplaceConfig(vault)
}

// Function to load configuration and replace placeholders with values fetched from Vault
func loadAndReplaceConfig(vlt *vault.Vault) error {

	//  Iterate through all settings and replace placeholders
	for key, value := range viper.AllSettings() {
		// Check if the value is a string and contains placeholders like {{.ENV.DBPASSWORD}}
		if strValue, ok := value.(string); ok {
			// Replace placeholders in the string
			updatedValue, err := getSecretsFromVault(strValue, vlt)
			if err != nil {
				helpers.Println(constant.ERROR, "Error fetching secret from Vault for key ", key, ": ", err)
				continue
			} else {
				// Update the Viper configuration with the new value
				viper.Set(key, updatedValue)
			}
		}
	}
	return nil
}

// Function to get secrets from vault
func getSecretsFromVault(configContent string, vaultManager *vault.Vault) (string, error) {
	// Define the regular expression to match placeholders like {{.DBPASSWORD}}
	re := regexp.MustCompile(`{{\s*\.[^}]+\s*}}`)

	// Replace the placeholders in the config string with the corresponding secret from Vault
	updatedConfig := re.ReplaceAllStringFunc(configContent, func(placeholder string) string {
		// Trim the {{. and }} from the placeholder to get the key directly
		key := placeholder[3 : len(placeholder)-2]

		// Fetch the value from Vault (you can fetch other secrets depending on your Vault setup)
		value, err := vaultManager.FetchVaultValue(key)
		if err != nil {
			helpers.Println(constant.ERROR, "Error fetching secret ", key, " from Vault: ", err)
			return "" // Return empty string if fetching fails
		}

		value, err = vaultManager.DecryptVaultValues(key, value)
		if err != nil {
			helpers.Println(constant.ERROR, "Error decrypting secret ", key, " from Vault: ", err)
			return "" // Return empty string if fetching fails
		}

		return value
	})

	return updatedConfig, nil
}

// UnmarshalConfig unmarshals the entire Viper configuration into the provided struct reference.
// It helps you avoid calling viper.GetString / viper.GetInt repeatedly by binding
// configuration values directly into a typed struct.
//
// Example:
//
//	type AppConfig struct {
//	    Server struct {
//	        Host string `mapstructure:"host"`
//	        Port int    `mapstructure:"port"`
//	    } `mapstructure:"server"`
//
//	    Database struct {
//	        User     string `mapstructure:"user,squash"`
//	        Password string `mapstructure:"password"`
//	        Host     string `mapstructure:"host"`
//	        Port     int    `mapstructure:"port"`
//	    } `mapstructure:"database"`
//
//	    LogLevel string `mapstructure:"log_level"`
//	}
func UnmarshalConfig[T any](target *T) error {
	if target == nil {
		return fmt.Errorf("target struct cannot be nil")
	}

	if err := viper.Unmarshal(target); err != nil {
		return fmt.Errorf("failed to unmarshal viper config: %w", err)
	}

	return nil
}
