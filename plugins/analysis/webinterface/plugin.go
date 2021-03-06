package webinterface

import (
	"errors"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gobuffalo/packr/v2"
	"github.com/iotaledger/goshimmer/packages/shutdown"
	"github.com/iotaledger/goshimmer/plugins/config"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
)

// PluginName is the name of the analysis server web interface plugin.
const PluginName = "Analysis-WebInterface"

var (
	// Plugin is the plugin instance of the analysis server web interface plugin.
	Plugin    = node.NewPlugin(PluginName, node.Disabled, configure, run)
	log       *logger.Logger
	engine    *echo.Echo
	assetsBox = packr.New("Analysis_WebInterface_Assets", "./static")
)

func configure(plugin *node.Plugin) {
	log = logger.NewLogger(PluginName)

	engine = echo.New()
	engine.HideBanner = true
	engine.HidePort = true

	// we only need this special flag, because we always keep a packed box in the same directory
	if config.Node.GetBool(CfgAnalysisWebInterfaceDev) {
		engine.Static("/static", "./plugins/analysis/webinterface/static")
		engine.File("/", "./plugins/analysis/webinterface/static/index.html")
	} else {
		for _, res := range assetsBox.List() {
			engine.GET(filepath.ToSlash(filepath.Join("/static", res)), echo.WrapHandler(http.StripPrefix("/static", http.FileServer(assetsBox))))
		}
		engine.GET("/", index)
	}

	engine.GET("/datastream", echo.WrapHandler(websocket.Handler(dataStream)))
	configureEventsRecording(plugin)
}

func run(_ *node.Plugin) {
	if err := daemon.BackgroundWorker(PluginName, worker, shutdown.PriorityAnalysis); err != nil {
		log.Errorf("Error starting as daemon: %s", err)
	}

	runEventsRecordManager()
}

func worker(shutdownSignal <-chan struct{}) {
	defer log.Infof("Stopping %s ... done", PluginName)

	stopped := make(chan struct{})
	bindAddr := config.Node.GetString(CfgAnalysisWebInterfaceBindAddress)
	go func() {
		log.Infof("Started %s: http://%s", PluginName, bindAddr)
		if err := engine.Start(bindAddr); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Errorf("Error serving: %s", err)
			}
			close(stopped)
		}
	}()

	// stop if we are shutting down or the server could not be started
	select {
	case <-shutdownSignal:
	case <-stopped:
	}

	log.Infof("Stopping %s ...", PluginName)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := engine.Shutdown(ctx); err != nil {
		log.Errorf("Error stopping: %s", err)
	}
}

func index(e echo.Context) error {
	indexHTML, err := assetsBox.Find("index.html")
	if err != nil {
		return err
	}
	return e.HTMLBlob(http.StatusOK, indexHTML)
}
