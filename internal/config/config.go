package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Kubeconfig  string
	Context     string
	Namespace   string
	Theme       ThemeConfig       `mapstructure:"theme"`
	Keybindings KeybindingsConfig `mapstructure:"keybindings"`
	Defaults    DefaultsConfig    `mapstructure:"defaults"`
	Panels      PanelsConfig      `mapstructure:"panels"`
}

type ThemeConfig struct {
	PrimaryColor    string `mapstructure:"primaryColor"`
	SecondaryColor  string `mapstructure:"secondaryColor"`
	ErrorColor      string `mapstructure:"errorColor"`
	WarningColor    string `mapstructure:"warningColor"`
	BackgroundColor string `mapstructure:"backgroundColor"`
	TextColor       string `mapstructure:"textColor"`
	BorderColor     string `mapstructure:"borderColor"`
}

type KeybindingsConfig struct {
	Quit        []string `mapstructure:"quit"`
	Help        []string `mapstructure:"help"`
	NextPanel   []string `mapstructure:"nextPanel"`
	PrevPanel   []string `mapstructure:"prevPanel"`
	Up          []string `mapstructure:"up"`
	Down        []string `mapstructure:"down"`
	Search      []string `mapstructure:"search"`
	Delete      []string `mapstructure:"delete"`
	Describe    []string `mapstructure:"describe"`
	Yaml        []string `mapstructure:"yaml"`
	Logs        []string `mapstructure:"logs"`
	Exec        []string `mapstructure:"exec"`
	PortForward []string `mapstructure:"portForward"`
	Scale       []string `mapstructure:"scale"`
	Restart     []string `mapstructure:"restart"`
	Refresh     []string `mapstructure:"refresh"`
}

type DefaultsConfig struct {
	Namespace       string `mapstructure:"namespace"`
	LogLines        int    `mapstructure:"logLines"`
	FollowLogs      bool   `mapstructure:"followLogs"`
	RefreshInterval int    `mapstructure:"refreshInterval"`
}

type PanelsConfig struct {
	Visible []string `mapstructure:"visible"`
	Layout  string   `mapstructure:"layout"`
}

func Load() (*Config, error) {
	cfg := &Config{
		Theme: ThemeConfig{
			PrimaryColor:    "#7aa2f7",
			SecondaryColor:  "#9ece6a",
			ErrorColor:      "#f7768e",
			WarningColor:    "#e0af68",
			BackgroundColor: "#1a1b26",
			TextColor:       "#c0caf5",
			BorderColor:     "#3b4261",
		},
		Keybindings: KeybindingsConfig{
			Quit:        []string{"q", "ctrl+c"},
			Help:        []string{"?"},
			NextPanel:   []string{"tab"},
			PrevPanel:   []string{"shift+tab"},
			Up:          []string{"k", "up"},
			Down:        []string{"j", "down"},
			Search:      []string{"/"},
			Delete:      []string{"D"},
			Describe:    []string{"d"},
			Yaml:        []string{"y"},
			Logs:        []string{"l"},
			Exec:        []string{"x"},
			PortForward: []string{"p"},
			Scale:       []string{"s"},
			Restart:     []string{"r"},
			Refresh:     []string{"ctrl+r"},
		},
		Defaults: DefaultsConfig{
			Namespace:       "default",
			LogLines:        100,
			FollowLogs:      true,
			RefreshInterval: 5,
		},
		Panels: PanelsConfig{
			Visible: []string{"namespaces", "pods", "deployments", "services"},
			Layout:  "vertical",
		},
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if configDir, err := os.UserConfigDir(); err == nil {
		viper.AddConfigPath(filepath.Join(configDir, "lazy-k8s"))
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(filepath.Join(homeDir, ".config", "lazy-k8s"))
		viper.AddConfigPath(filepath.Join(homeDir, ".lazy-k8s"))
	}

	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			// Missing config file is expected (user may rely on defaults); only fail on actual read/parse errors
			return nil, err
		}
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	if cfg.Namespace == "" {
		cfg.Namespace = cfg.Defaults.Namespace
	}

	return cfg, nil
}
