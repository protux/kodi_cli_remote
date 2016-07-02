package administration

import (
    "encoding/json"
    homedir "github.com/mitchellh/go-homedir"
    "io/ioutil"
    "os"
)

const (
    fileDirectory = `.config/kodiremote/`
    filePath = fileDirectory + `kodiremote.conf`
)
var fullPathCache string = ``

// Configuration represents all configurable options inside this tool.
type Configuration struct {
    Host string
    Port string    
}

func getFullConfigPath() (string, error) {
    if len(fullPathCache) == 0 {
        home, err := homedir.Dir()
        if err == nil {
            fullPathCache = home + `/` + filePath
        } else {
            return ``, err
        }
    }
    return fullPathCache, nil
}

func loadConfiguration() (Configuration, error) {
    var configuration Configuration
    path, err := getFullConfigPath()
    
    if err == nil {
        if jsonString, err := ioutil.ReadFile(path); err == nil {
            if err = json.Unmarshal([]byte(jsonString), &configuration); err == nil {
                return configuration, nil
            }
        } else {
            return configuration, err
        }
    }
    return configuration, err
}

// WriteConfiguration writes the configuration to the filesystem.
func WriteConfiguration(configuration Configuration) error {
    
    jsonConf, err := json.Marshal(configuration)
    if err == nil {
        if home, err := homedir.Dir(); err == nil {
            err = ioutil.WriteFile(home + `/` + filePath, jsonConf, 0700)
        }
    }
    return err
}

// CreateConfiguration checks if an configuration exists and if there
// exists one it is loaded and returned, otherwise an empty configuration
// will be created, saved and returned.
func CreateConfiguration() (Configuration, error) {
    homedir.DisableCache = false
    
    if configuration, err := loadConfiguration(); err != nil {
        if home, err := homedir.Dir(); err == nil {
            var initialConfig Configuration
            os.MkdirAll(home + `/` + fileDirectory, os.ModeDir | 0700)
            err = WriteConfiguration(initialConfig)
            return initialConfig, err
        } else {
            return configuration, err
        }
    } else {
        return configuration, nil
    }
}

