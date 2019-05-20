package domain

import (
	"github.com/BurntSushi/toml"
	"github.com/adverax/echo/database/sql"
	"os"
	"path/filepath"
	"regexp"
)

const (
	configFileName = "config.toml"
)

type DatabaseOptions struct {
	DbId      sql.DbId   `toml:"-"`         // Type of reactor
	Nodes     []*sql.DSN `toml:"node"`      // Database options
	Heartbeat int        `toml:"heartbeat"` // Database heartbeat (seconds)
}

func (options DatabaseOptions) DSC() sql.DSC {
	return sql.DSC{
		Driver: "mysql",
		DbId:   options.DbId,
		DSN:    options.Nodes,
	}
}

type BrokerOptions struct {
	Server string `toml:"server"` // Url of the NATS server
}

// Primary service configuration
type Configuration struct {
	WorkDir  string          `toml:"-"`        // Work directory
	Broker   BrokerOptions   `toml:"broker"`   // Broker options
	Database DatabaseOptions `toml:"database"` // Database options
}

var (
	Config = Configuration{
		Database: DatabaseOptions{
			Heartbeat: 60,
			DbId:      1,
		},
		Broker: BrokerOptions{
			Server: "nats://localhost:4222",
		},
	}
	tmpDirRe = regexp.MustCompile("^/tmp/")
)

func DecodeConfig(source string, config interface{}) {
	_, err := toml.Decode(source, config)
	if err != nil {
		panic(err)
	}
}

func DecodeConfigFile(fileName string, config interface{}) {
	_, err := toml.DecodeFile(fileName, config)
	if err != nil {
		panic(err)
	}
}

func findWorkDir() (string, error) {
	// Try use directory with current executable file
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(ex)

	if tmpDirRe.MatchString(ex) {
		// If debugging than using current work directory
		dir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	} else {
		err = os.Chdir(dir)
		if err != nil {
			return "", err
		}
	}

	// Walk from current directory to root (work dir must contains config file)
	for dir != "." {
		if _, err := os.Stat(dir + "/" + configFileName); err == nil {
			break
		}
		dir = filepath.Dir(dir)
	}

	return dir, nil
}

func init() {
	workDir, err := findWorkDir()
	if err != nil {
		panic(err)
	}
	Config.WorkDir = workDir
	DecodeConfigFile(workDir+"/"+configFileName, &Config)
}
