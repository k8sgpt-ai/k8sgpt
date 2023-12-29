package server

import (
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"os"
	"sync"
	"testing"
)

func TestServerInit(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		color.Red("failed to create logger: %v", err)
		os.Exit(1)
	}
	defer logger.Sync()
	server_config := Config{
		Backend:     "openai",
		Port:        "0",
		MetricsPort: "0",
		Token:       "none",
		Logger:      logger,
	}
	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		err := server_config.Serve()
		if err != nil {
			assert.Fail(t, "serve: %s", err.Error())
		}
		server_config.Shutdown()
		if err != nil {
			assert.Fail(t, "shutdown: %s", err.Error())
		}
		wg.Done()
	}()
	wg.Wait()
}
