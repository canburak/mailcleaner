package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Accounts []Account `yaml:"accounts"`
	Rules    []Rule    `yaml:"rules"`
	Schedule Schedule  `yaml:"schedule"`
}

// Account represents an email account configuration
type Account struct {
	Name        string `yaml:"name"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	PasswordEnv string `yaml:"password_env"`
	TLS         bool   `yaml:"tls"`
}

// GetPassword returns the password, reading from environment if password_env is set
func (a *Account) GetPassword() string {
	if a.PasswordEnv != "" {
		return os.Getenv(a.PasswordEnv)
	}
	return a.Password
}

// Rule represents an email rule configuration
type Rule struct {
	Name       string     `yaml:"name"`
	Account    string     `yaml:"account"`
	Folder     string     `yaml:"folder"`
	Conditions Conditions `yaml:"conditions"`
	Action     Action     `yaml:"action"`
}

// Conditions represents the conditions for matching emails
type Conditions struct {
	From           string `yaml:"from"`
	FromContains   string `yaml:"from_contains"`
	To             string `yaml:"to"`
	ToContains     string `yaml:"to_contains"`
	Subject        string `yaml:"subject"`
	SubjectContains string `yaml:"subject_contains"`
	OlderThanDays  int    `yaml:"older_than_days"`
	NewerThanDays  int    `yaml:"newer_than_days"`
	IsRead         *bool  `yaml:"is_read"`
	IsUnread       *bool  `yaml:"is_unread"`
	HasAttachment  *bool  `yaml:"has_attachment"`
	SizeLargerThan string `yaml:"size_larger_than"`
}

// Action represents the action to take on matched emails
type Action struct {
	Type       string `yaml:"type"`
	MoveTo     string `yaml:"move_to"`
	AddFlag    string `yaml:"add_flag"`
	RemoveFlag string `yaml:"remove_flag"`
}

// Schedule represents the scheduler configuration
type Schedule struct {
	IntervalMinutes int    `yaml:"interval_minutes"`
	Cron            string `yaml:"cron"`
}

// GetInterval returns the schedule interval as a duration
func (s *Schedule) GetInterval() time.Duration {
	if s.IntervalMinutes > 0 {
		return time.Duration(s.IntervalMinutes) * time.Minute
	}
	return 60 * time.Minute // default to 1 hour
}

// Load reads and parses a configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Set defaults
	for i := range cfg.Accounts {
		if cfg.Accounts[i].Port == 0 {
			if cfg.Accounts[i].TLS {
				cfg.Accounts[i].Port = 993
			} else {
				cfg.Accounts[i].Port = 143
			}
		}
	}

	return &cfg, nil
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	if len(c.Accounts) == 0 {
		return fmt.Errorf("at least one account is required")
	}

	accountNames := make(map[string]bool)
	for i, acc := range c.Accounts {
		if acc.Name == "" {
			return fmt.Errorf("account %d: name is required", i)
		}
		if accountNames[acc.Name] {
			return fmt.Errorf("duplicate account name: %s", acc.Name)
		}
		accountNames[acc.Name] = true

		if acc.Host == "" {
			return fmt.Errorf("account %s: host is required", acc.Name)
		}
		if acc.Username == "" {
			return fmt.Errorf("account %s: username is required", acc.Name)
		}
		if acc.Password == "" && acc.PasswordEnv == "" {
			return fmt.Errorf("account %s: password or password_env is required", acc.Name)
		}
	}

	for i, rule := range c.Rules {
		if rule.Name == "" {
			return fmt.Errorf("rule %d: name is required", i)
		}
		if rule.Account == "" {
			return fmt.Errorf("rule %s: account is required", rule.Name)
		}
		if !accountNames[rule.Account] {
			return fmt.Errorf("rule %s: unknown account %s", rule.Name, rule.Account)
		}
		if rule.Folder == "" {
			return fmt.Errorf("rule %s: folder is required", rule.Name)
		}
		if rule.Action.Type == "" {
			return fmt.Errorf("rule %s: action type is required", rule.Name)
		}
	}

	return nil
}

// GetAccount returns an account by name
func (c *Config) GetAccount(name string) *Account {
	for i := range c.Accounts {
		if c.Accounts[i].Name == name {
			return &c.Accounts[i]
		}
	}
	return nil
}
