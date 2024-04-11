package config

import (
	_ "embed"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"path/filepath"
)

//go:embed configs.toml
var defaultConfigs string

const (
	White = lipgloss.Color("#FFF")

	NamadaYellow = lipgloss.Color("#FFFF00")
	DarkGray     = lipgloss.Color("#767676")
)

type configs struct {
	Namada struct {
		Node           string `toml:"node" env:"EZSHIELD_NAMADA_NODE"`
		ChainId        string `toml:"chain-id" env:"EZSHIELD_NAMADA_CHAIN_ID"`
		ChannelOsmosis struct {
			ChannelId string `toml:"channel-id" env:"EZSHIELD_NAMADA_OSMOSIS_CHANNEL_ID"`
		} `toml:"osmosis"`
	} `toml:"namada"`

	Osmosis struct {
		Node          string `toml:"node" env:"EZSHIELD_OSMOSIS_NODE"`
		ChainId       string `toml:"chain-id" env:"EZSHIELD_OSMOSIS_CHAIN_ID"`
		ChannelNamada struct {
			ChannelId string `toml:"channel-id" env:"EZSHIELD_OSMOSIS_NAMADA_CHANNEL_ID"`
		} `toml:"namada"`
	} `toml:"osmosis"`
}

var Cfg = configs{}

func InitConfigs() {
	configFilePath := getConfigFilePath()

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if err := createDefaultConfig(configFilePath); err != nil {
			fmt.Println("Error creating default config")
			panic(err)
		}
	}
	err := cleanenv.ReadConfig(configFilePath, &Cfg)
	if err != nil {
		panic(err)
	}
}
func CreateAssetsDirectories() {
	tempDirPath := GetTempDirPath()
	if _, err := os.Stat(tempDirPath); os.IsNotExist(err) {
		err := os.MkdirAll(tempDirPath, 0755)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}
}

func GetTempDirPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homeDir, ".ezshield", "temp")
}

func getConfigFilePath() string {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".ezshield")
	configFilePath := filepath.Join(configDir, "configs.toml")
	return configFilePath
}

func createDefaultConfig(configFilePath string) error {
	configDir := filepath.Dir(configFilePath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	file, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(defaultConfigs))
	if err != nil {
		return err
	}

	return nil
}
