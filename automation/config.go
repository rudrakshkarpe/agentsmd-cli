// Package automation processes captured runs through reflection and evaluation.
package automation

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
)

type Config struct {
	ReflectCommand  []string `json:"reflect_command"`
	EvaluateCommand []string `json:"evaluate_command"`
	AutoPromote     bool     `json:"auto_promote"`
	MinConfidence   float64  `json:"min_confidence"`
}

func DefaultConfig() Config {
	return Config{ReflectCommand: []string{}, EvaluateCommand: []string{}, MinConfidence: 0.8}
}

func Load(p *project.Project) (Config, error) {
	data, err := os.ReadFile(p.AutomationPath())
	if errors.Is(err, os.ErrNotExist) {
		return DefaultConfig(), nil
	}
	if err != nil {
		return Config{}, err
	}
	value := DefaultConfig()
	if err := json.Unmarshal(data, &value); err != nil {
		return Config{}, fmt.Errorf("decode automation config: %w", err)
	}
	return value, Validate(value)
}

func Save(p *project.Project, value Config) error {
	if err := Validate(value); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return project.AtomicWrite(p.AutomationPath(), append(data, '\n'), 0o644)
}

func Validate(value Config) error {
	if value.MinConfidence < 0 || value.MinConfidence > 1 {
		return fmt.Errorf("minimum confidence must be between 0 and 1")
	}
	if value.AutoPromote && len(value.EvaluateCommand) == 0 {
		return fmt.Errorf("automatic promotion requires an evaluation command")
	}
	if value.AutoPromote && len(value.ReflectCommand) == 0 {
		return fmt.Errorf("automatic promotion requires a reflection command")
	}
	return nil
}
