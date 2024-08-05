package customanalyzer

import (
	"fmt"
	"reflect"
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
