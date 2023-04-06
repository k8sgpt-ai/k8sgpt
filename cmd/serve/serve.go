package serve

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	port string
)

// generateCmd represents the auth command
var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		rootCmd := viper.Get("rootCmd").(*cobra.Command)
		// Start http server
		router := httprouter.New()
		router.GET("/:command", func(w http.ResponseWriter,
			r *http.Request, p httprouter.Params) {

			// find the command
			command := p.ByName("command")
			cmd, string, err := rootCmd.Find([]string{command})
			if err != nil {
				w.Write([]byte(err.Error()))
			}
			old := os.Stdout // keep backup of the real stdout
			rd, d, _ := os.Pipe()
			os.Stdout = d
			cmd.Run(cmd, string)
			d.Close()
			out, _ := ioutil.ReadAll(rd)
			os.Stdout = old // restoring the real stdout
			w.Write(out)
		})

		log.Fatal(http.ListenAndServe(":8080", router))
	},
}

func init() {
	// add flag for backend
	ServeCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to serve on")
}
