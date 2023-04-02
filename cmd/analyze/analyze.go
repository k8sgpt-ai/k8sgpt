package analyze

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analysis"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	explain   bool
	backend   string
	output    string
	filters   []string
	language  string
	nocache   bool
	namespace string
)

// AnalyzeCmd represents the problems command
var AnalyzeCmd = &cobra.Command{
	Use:     "analyze",
	Aliases: []string{"analyse"},
	Short:   "This command will find problems within your Kubernetes cluster",
	Long: `This command will find problems within your Kubernetes cluster and
	provide you with a list of issues that need to be resolved`,
	Run: func(cmd *cobra.Command, args []string) {

		// get backend from file
		backendType := viper.GetString("backend_type")
		if backendType == "" {
			color.Red("No backend set. Please run k8sgpt auth")
			os.Exit(1)
		}
		// override the default backend if a flag is provided
		if backend != "" {
			backendType = backend
		}
		// get the token with viper
		token := viper.GetString(fmt.Sprintf("%s_key", backendType))
		// check if nil
		if token == "" {
			color.Red("No %s key set. Please run k8sgpt auth", backendType)
			os.Exit(1)
		}

		var aiClient ai.IAI
		switch backendType {
		case "openai":
			aiClient = &ai.OpenAIClient{}
			if err := aiClient.Configure(token, language); err != nil {
				color.Red("Error: %v", err)
				os.Exit(1)
			}
		default:
			color.Red("Backend not supported")
			os.Exit(1)
		}

		ctx := context.Background()
		// Get kubernetes client from viper
		client := viper.Get("kubernetesClient").(*kubernetes.Client)
		// AnalysisResult configuration

		analysis := &analysis.Analysis{
			Context:   ctx,
			Namespace: namespace,
			NoCache:   nocache,
			Explain:   explain,
			AIClient:  aiClient,
			Filters:   filters,
			Client:    client,
		}

		// Run analysis
		_ = analysis.RunAnalysis()

		analysis.PrintAnalysisResult()
	},
}

func init() {
	// namespace flag
	AnalyzeCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace to analyze")
	// no cache flag
	AnalyzeCmd.Flags().BoolVarP(&nocache, "no-cache", "c", false, "Do not use cached data")
	// array of strings flag
	AnalyzeCmd.Flags().StringSliceVarP(&filters, "filter", "f", []string{}, "Filter for these analyzers (e.g. Pod, PersistentVolumeClaim, Service, ReplicaSet)")
	// explain flag
	AnalyzeCmd.Flags().BoolVarP(&explain, "explain", "e", false, "Explain the problem to me")
	// add flag for backend
	AnalyzeCmd.Flags().StringVarP(&backend, "backend", "b", "openai", "Backend AI provider")
	// output as json
	AnalyzeCmd.Flags().StringVarP(&output, "output", "o", "text", "Output format (text, json)")
	// add language options for output
	AnalyzeCmd.Flags().StringVarP(&language, "language", "l", "english", "Languages to use for AI (e.g. 'English', 'Spanish', 'French', 'German', 'Italian', 'Portuguese', 'Dutch', 'Russian', 'Chinese', 'Japanese', 'Korean')")
}
