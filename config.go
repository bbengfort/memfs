// Implements the reading and writing to and from a JSON config file.

package memfs

import (
	"encoding/json"
	"io/ioutil"
)

//===========================================================================
// Configuration Structs
//===========================================================================

// Replica implements the definition for a remote replica connections.
type Replica struct {
	PID  uint   `json:"pid"`  // Precedence ID for the replica
	Name string `json:"name"` // Name of the replica
	Host string `json:"host"` // IP address or hostname of the replica
	Port int    `json:"port"` // Port the replica is listening on
}

// Config implements the local configuration directives.
type Config struct {
	Name     string     `json:"name"`     // Identifier for replica lists
	Replicas []*Replica `json:"replicas"` // List of remote replicas in system
	Path     string     `json:"-"`        // Path the config was loaded from
}

//===========================================================================
// Config Methods
//===========================================================================

// Load a configuration from a path on disk by deserializing the JSON data.
func (conf *Config) Load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// Unmarshal the JSON data
	if err := json.Unmarshal(data, &conf); err != nil {
		return err
	}

	// Save the loaded path
	conf.Path = path
	return nil
}

// Dump a configuration to JSON to the path on disk. If dump is an empty
// string then will dump the config to the path it was loaded from.
func (conf *Config) Dump(path string) error {
	if path == "" {
		path = conf.Path
	}

	// Marshal the JSON configuration data
	data, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	// Write the data to disk
	return ioutil.WriteFile(path, data, 0644)
}
