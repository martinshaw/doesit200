package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	URL                    string   `json:"url"`
	SleepAmount            uint16   `json:"sleep"`
	MaxDepth               uint8    `json:"depth"`
	IncludeDomainWildcards []string `json:"include_domain_wildcards"`
}

func GetConfigFromEnvironment() (*Config, error) {
	var url string
	var sleep uint16
	var depth uint8
	var includeDomainWildcards []string

	fmt.Print("Enter select the crawler config (see config.json): ")

	// read from config.json
	var configFile string
	if _, err := fmt.Scanln(&configFile); err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	// open config.json
	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close()

	// decode config.json as an array of configurations
	var configs []struct {
		URL                    string   `json:"url"`
		SleepAmount            uint16   `json:"sleep"`
		MaxDepth               uint8    `json:"depth"`
		IncludeDomainWildcards []string `json:"include_domain_wildcards"`
	}
	if err := json.NewDecoder(file).Decode(&configs); err != nil {
		return nil, fmt.Errorf("could not decode config file: %w", err)
	}

	// List available configurations with index which can be selected by input
	for i, config := range configs {
		fmt.Printf("%d: URL: %s, Sleep: %d seconds, Depth: %d\n", i+1, config.URL, config.SleepAmount, config.MaxDepth)
	}
	fmt.Print("Select a configuration by number: ")
	var choice int
	if _, err := fmt.Scanln(&choice); err != nil {
		return nil, fmt.Errorf("could not read choice: %w", err)
	}

	if choice < 1 || choice > len(configs) {
		return nil, fmt.Errorf("invalid choice: %d", choice)
	}

	// Get the selected configuration
	selectedConfig := configs[choice-1]
	url = selectedConfig.URL
	sleep = selectedConfig.SleepAmount
	depth = selectedConfig.MaxDepth
	includeDomainWildcards = selectedConfig.IncludeDomainWildcards

	return &Config{
		URL:                    url,
		SleepAmount:            sleep,
		MaxDepth:               depth,
		IncludeDomainWildcards: includeDomainWildcards,
	}, nil
}

func GetConfigFromInputs() (*Config, error) {
	var url string
	var sleep uint16
	var depth uint8
	var includeDomainWildcards []string

	fmt.Print("Enter URL to be scanned: ")
	if _, err := fmt.Scanln(&url); err != nil {
		return nil, fmt.Errorf("could not read URL: %w", err)
	}

	fmt.Print("Enter sleep duration between navigation in seconds (default 15): ")
	if _, err := fmt.Scanln(&sleep); err != nil {
		sleep = 15 // default sleep duration
	}

	fmt.Print("Enter page depth (default 1): ")
	if _, err := fmt.Scanln(&depth); err != nil {
		depth = 1 // default depth
	}

	return &Config{
		URL:                    url,
		SleepAmount:            sleep,
		MaxDepth:               depth,
		IncludeDomainWildcards: includeDomainWildcards,
	}, nil
}
