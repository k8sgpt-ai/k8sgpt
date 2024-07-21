package customanalyzer

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/k8sgpt-ai/k8sgpt/pkg/customAnalyzer/docker"
)

type CustomAnalyzerConfiguration struct {
	Name        string     `mapstructure:"name"`
	Connection  Connection `mapstructure:"connection"`
	InstallType string     `mapstructure:"installtype,omitempty"`
}

type Connection struct {
	Url  string `mapstructure:"url"`
	Port int    `mapstructure:"port"`
}

type ICustomAnalyzer interface {
	Deploy(packageUrl, name, url string, port int) error
	UnDeploy(name string) error
}

type CustomAnalyzer struct{}

var customAnalyzerType = map[string]ICustomAnalyzer{
	"docker": docker.NewDocker(),
}

func NewCustomAnalyzer() *CustomAnalyzer {
	return &CustomAnalyzer{}
}

func (*CustomAnalyzer) GetInstallType(name string) (ICustomAnalyzer, error) {
	if _, ok := customAnalyzerType[name]; !ok {
		return nil, errors.New("integration not found")
	}
	return customAnalyzerType[name], nil
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

func (ca *CustomAnalyzer) UnDeploy(analyzer CustomAnalyzerConfiguration) error {
	if analyzer.InstallType != "" {
		// Try to undeploy if install-type is set
		install, err := ca.GetInstallType(analyzer.InstallType)
		if err != nil {
			return err
		}

		err = install.UnDeploy(analyzer.Name)
		if err != nil {
			return err
		}
	}

	return nil

}
