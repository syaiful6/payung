package config

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/viper"
	"github.com/syaiful6/payung/logger"
)

var (
	// Exist Is config file exist
	Exist bool
	// Models configs
	Models []ModelConfig
	// HomeDir of user
	HomeDir = os.Getenv("HOME")
)

// ModelConfig for special case
type ModelConfig struct {
	Name              string
	TempPath          string
	DumpPath          string
	CompressWith      SubConfig
	EncryptWith       SubConfig
	StoreWith         SubConfig
	Archive           *viper.Viper
	SplitIntoChunksOf int
	Databases         []SubConfig
	Storages          []SubConfig
	Viper             *viper.Viper
}

// SubConfig sub config info
type SubConfig struct {
	Name  string
	Type  string
	Viper *viper.Viper
}

// loadConfig from:
// - ./gobackup.yml
// - ~/.gobackup/gobackup.yml
// - /etc/gobackup/gobackup.yml
func Init(configFile string, dumpPath string) {
	viper.SetConfigType("yaml")

	// set config file directly
	if len(configFile) > 0 {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("gobackup")

		// ./gobackup.yml
		viper.AddConfigPath(".")
		// ~/.gobackup/gobackup.yml
		viper.AddConfigPath("$HOME/.gobackup") // call multiple times to add many search paths
		// /etc/gobackup/gobackup.yml
		viper.AddConfigPath("/etc/gobackup/") // path to look for the config file in
	}

	err := viper.ReadInConfig()
	if err != nil {
		logger.Error("Load gobackup config faild", err)
		return
	}

	Exist = true
	Models = []ModelConfig{}
	for key := range viper.GetStringMap("models") {
		Models = append(Models, loadModel(key, dumpPath))
	}
}

func loadModel(key string, dumpPath string) (model ModelConfig) {
	model.Name = key
	if dumpPath != "" {
		model.TempPath = path.Join(dumpPath, fmt.Sprintf("%d", time.Now().UnixNano()))
	} else {
		model.TempPath = path.Join(os.TempDir(), "gobackup", fmt.Sprintf("%d", time.Now().UnixNano()))
	}

	model.DumpPath = path.Join(model.TempPath, key)
	model.Viper = viper.Sub("models." + key)

	model.CompressWith = SubConfig{
		Type:  model.Viper.GetString("compress_with.type"),
		Viper: model.Viper.Sub("compress_with"),
	}

	model.EncryptWith = SubConfig{
		Type:  model.Viper.GetString("encrypt_with.type"),
		Viper: model.Viper.Sub("encrypt_with"),
	}

	model.StoreWith = SubConfig{
		Type:  model.Viper.GetString("store_with.type"),
		Viper: model.Viper.Sub("store_with"),
	}

	model.Archive = model.Viper.Sub("archive")
	model.SplitIntoChunksOf = model.Viper.GetInt("split_into_chunks_of")

	loadDatabasesConfig(&model)
	loadStoragesConfig(&model)

	return
}

func loadDatabasesConfig(model *ModelConfig) {
	subViper := model.Viper.Sub("databases")
	for key := range model.Viper.GetStringMap("databases") {
		dbViper := subViper.Sub(key)
		model.Databases = append(model.Databases, SubConfig{
			Name:  key,
			Type:  dbViper.GetString("type"),
			Viper: dbViper,
		})
	}
}

func loadStoragesConfig(model *ModelConfig) {
	subViper := model.Viper.Sub("storages")
	for key := range model.Viper.GetStringMap("storages") {
		dbViper := subViper.Sub(key)
		model.Storages = append(model.Storages, SubConfig{
			Name:  key,
			Type:  dbViper.GetString("type"),
			Viper: dbViper,
		})
	}
}

// GetModelByName get model by name
func GetModelByName(name string) (model *ModelConfig) {
	for _, m := range Models {
		if m.Name == name {
			model = &m
			return
		}
	}
	return
}

// GetDatabaseByName get database config by name
func (model *ModelConfig) GetDatabaseByName(name string) (subConfig *SubConfig) {
	for _, m := range model.Databases {
		if m.Name == name {
			subConfig = &m
			return
		}
	}
	return
}
