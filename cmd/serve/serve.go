/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package serve

import (
	"os"
	"strconv"

	k8sgptserver "github.com/k8sgpt-ai/k8sgpt/pkg/server"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	defaultTemperature float32 = 0.7
	defaultTopP        float32 = 1.0
	defaultTopK        int32   = 50
)

var (
	port        string
	metricsPort string
	backend     string
	enableHttp  bool
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Runs k8sgpt as a server",
	Long:  `Runs k8sgpt as a server to allow for easy integration with other applications.`,
	Run: func(cmd *cobra.Command, args []string) {

		var configAI ai.AIConfiguration
		err := viper.UnmarshalKey("ai", &configAI)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		var aiProvider *ai.AIProvider
		if len(configAI.Providers) == 0 {
			// we validate and set temperature for our backend
			temperature := func() float32 {
				env := os.Getenv("K8SGPT_TEMPERATURE")
				if env == "" {
					return defaultTemperature
				}
				temperature, err := strconv.ParseFloat(env, 32)
				if err != nil {
					color.Red("Unable to convert Temperature value: %v", err)
					os.Exit(1)
				}
				if temperature > 1.0 || temperature < 0.0 {
					color.Red("Error: temperature ranges from 0 to 1.")
					os.Exit(1)
				}
				return float32(temperature)
			}
			topP := func() float32 {
				env := os.Getenv("K8SGPT_TOP_P")
				if env == "" {
					return defaultTopP
				}
				topP, err := strconv.ParseFloat(env, 32)
				if err != nil {
					color.Red("Unable to convert topP value: %v", err)
					os.Exit(1)
				}
				if topP > 1.0 || topP < 0.0 {
					color.Red("Error: topP ranges from 0 to 1.")
					os.Exit(1)
				}
				return float32(topP)
			}
			topK := func() int32 {
				env := os.Getenv("K8SGPT_TOP_K")
				if env == "" {
					return defaultTopK
				}
				topK, err := strconv.ParseFloat(env, 32)
				if err != nil {
					color.Red("Unable to convert topK value: %v", err)
					os.Exit(1)
				}
				if topK < 10 || topK > 100 {
					color.Red("Error: topK ranges from 1 to 100.")
					os.Exit(1)
				}
				return int32(topK)
			}
			// Check for env injection
			backend = os.Getenv("K8SGPT_BACKEND")
			password := os.Getenv("K8SGPT_PASSWORD")
			model := os.Getenv("K8SGPT_MODEL")
			baseURL := os.Getenv("K8SGPT_BASEURL")
			engine := os.Getenv("K8SGPT_ENGINE")
			proxyEndpoint := os.Getenv("K8SGPT_PROXY_ENDPOINT")
			providerId := os.Getenv("K8SGPT_PROVIDER_ID")
			// If the envs are set, allocate in place to the aiProvider
			// else exit with error
			envIsSet := backend != "" || password != "" || model != ""
			if envIsSet {
				aiProvider = &ai.AIProvider{
					Name:          backend,
					Password:      password,
					Model:         model,
					BaseURL:       baseURL,
					Engine:        engine,
					ProxyEndpoint: proxyEndpoint,
					ProviderId:    providerId,
					Temperature:   temperature(),
					TopP:          topP(),
					TopK:          topK(),
				}

				configAI.Providers = append(configAI.Providers, *aiProvider)

				viper.Set("ai", configAI)
				if err := viper.WriteConfig(); err != nil {
					color.Red("Error writing config file: %s", err.Error())
					os.Exit(1)
				}
			} else {
				color.Red("Error: AI provider not specified in configuration. Please run k8sgpt auth")
				os.Exit(1)
			}
		}

		// Backend string will have high priority than a default provider
		// Hence, use the default provider only if the backend is not specified by the user.
		if configAI.DefaultProvider != "" && backend == "" {
			backend = configAI.DefaultProvider
		}
		// Using default provider openai
		if backend == "" {
			backend = "openai"
		}

		if aiProvider == nil {
			for _, provider := range configAI.Providers {
				if backend == provider.Name {
					// the pointer to the range variable is not really an issue here, as there
					// is a break right after, but to prevent potential future issues, a temp
					// variable is assigned
					p := provider
					aiProvider = &p
					break
				}
			}
		}

		if aiProvider == nil || aiProvider.Name == "" {
			color.Red("Error: AI provider %s not specified in configuration. Please run k8sgpt auth", backend)
			os.Exit(1)
		}

		logger, err := zap.NewProduction()
		if err != nil {
			color.Red("failed to create logger: %v", err)
			os.Exit(1)
		}
		defer func() {
			if err := logger.Sync(); err != nil {
				color.Red("failed to sync logger: %v", err)
				os.Exit(1)
			}
		}()

		server := k8sgptserver.Config{
			Backend:     aiProvider.Name,
			Port:        port,
			MetricsPort: metricsPort,
			EnableHttp:  enableHttp,
			Token:       aiProvider.Password,
			Logger:      logger,
		}
		go func() {
			if err := server.ServeMetrics(); err != nil {
				color.Red("Error: %v", err)
				os.Exit(1)
			}
		}()

		go func() {
			if err := server.Serve(); err != nil {
				color.Red("Error: %v", err)
				os.Exit(1)
			}
		}()

		// Wait for both servers to exit
		select {}
	},
}

func init() {
	// add flag for backend
	ServeCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
	ServeCmd.Flags().StringVarP(&metricsPort, "metrics-port", "", "8081", "Port to run the metrics-server on")
	ServeCmd.Flags().StringVarP(&backend, "backend", "b", "", "Backend AI provider")
	ServeCmd.Flags().BoolVarP(&enableHttp, "http", "", false, "Enable REST/http using gppc-gateway")
}
