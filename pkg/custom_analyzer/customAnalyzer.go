package custom_analyzer

import (
	"fmt"
	"reflect"
	"regexp"
)

type CustomAnalyzerConfiguration struct {
	Name       string     `mapstructure:"name"`
	Connection Connection `mapstructure:"connection"`
}

type Connection struct {
	Url  string `mapstructure:"url"`
	Port int    `mapstructure:"port"`
}

type CustomAnalyzer struct{}

func NewCustomAnalyzer() *CustomAnalyzer {
	return &CustomAnalyzer{}
}

func (*CustomAnalyzer) Check(actualConfig []CustomAnalyzerConfiguration, name, url string, port int) error {
	validNameRegex := `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	validName := regexp.MustCompile(validNameRegex)
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid name format. Must match %s", validNameRegex)
	}

	for _, analyzer := range actualConfig {
		if analyzer.Name == name {
			return fmt.Errorf("custom analyzer with the name '%s' already exists. Please use a different name", name)
		}

		if reflect.DeepEqual(analyzer.Connection, Connection{
			Url:  url,
			Port: port,
		}) {
			return fmt.Errorf("custom analyzer with the same connection configuration (URL: '%s', Port: %d) already exists. Please use a different URL or port", url, port)
		}
	}

	return nil
}
