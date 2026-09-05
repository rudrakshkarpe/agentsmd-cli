package config

import (
	"encoding/json"
	"os"
)

type Values struct {
	Region string `json:"region"`
}

type LookupEnv func(string) (string, bool)

func Load(path string, lookup LookupEnv) (Values, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Values{}, err
	}
	var values Values
	if err := json.Unmarshal(data, &values); err != nil {
		return Values{}, err
	}
	if region, _ := lookup("APP_REGION"); region != "" {
		values.Region = region
	}
	return values, nil
}
