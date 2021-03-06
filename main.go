package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	api "github.com/Financial-Times/api-endpoint"
	"github.com/Financial-Times/go-ft-http/fthttp"
	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	logger "github.com/Financial-Times/go-logger/v2"
	httphandlers "github.com/Financial-Times/http-handlers-go/v2/httphandlers"
	status "github.com/Financial-Times/service-status-go/httphandlers"

	"github.com/gorilla/mux"
	cli "github.com/jawher/mow.cli"
	metrics "github.com/rcrowley/go-metrics"

	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/health"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
)

const (
	appDescription = "Provides suggestions for draft content."
	defaultAppName = "draft-content-suggestions"
)

func main() {
	app := cli.App(defaultAppName, appDescription)

	appSystemCode := app.String(cli.StringOpt{
		Name:   "app-system-code",
		Value:  "draft-content-suggestions",
		Desc:   "System Code of the application",
		EnvVar: "APP_SYSTEM_CODE",
	})
	appName := app.String(cli.StringOpt{
		Name:   "app-name",
		Value:  defaultAppName,
		Desc:   "Application name",
		EnvVar: "APP_NAME",
	})
	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})
	apiYml := app.String(cli.StringOpt{
		Name:   "api-yml",
		Value:  "./api.yml",
		Desc:   "Location of the OpenAPI YML file.",
		EnvVar: "API_YML",
	})
	draftContentEndpoint := app.String(cli.StringOpt{
		Name:   "draft-content-endpoint",
		Value:  "http://draft-content-public-read:8080/content",
		Desc:   "Endpoint for Draft Content API",
		EnvVar: "DRAFT_CONTENT_ENDPOINT",
	})
	draftContentGtgEndpoint := app.String(cli.StringOpt{
		Name:   "draft-content-gtg-endpoint",
		Value:  "http://draft-content-public-read:8080/__gtg",
		Desc:   "GTG Endpoint for Draft Content API",
		EnvVar: "DRAFT_CONTENT_GTG_ENDPOINT",
	})
	suggestionsEndpoint := app.String(cli.StringOpt{
		Name:   "suggestions-umbrella-endpoint",
		Value:  "http://test.api.ft.com/content/suggest",
		Desc:   "Endpoint for Suggestions Umbrella",
		EnvVar: "SUGGESTIONS_ENDPOINT",
	})
	suggestionsGtgEndpoint := app.String(cli.StringOpt{
		Name:   "suggestions-umbrella-gtg-endpoint",
		Value:  "http://test.api.ft.com/content/suggest/__gtg",
		Desc:   "Endpoint for Suggestions Umbrella",
		EnvVar: "SUGGESTIONS_GTG_ENDPOINT",
	})
	suggestionsAPIKey := app.String(cli.StringOpt{
		Name:   "suggestions-api-key",
		Value:  "",
		Desc:   "API key to access Suggestions Umbrella",
		EnvVar: "SUGGESTIONS_API_KEY",
	})
	logLevel := app.String(cli.StringOpt{
		Name:   "log-level",
		Value:  "info",
		Desc:   "Logging Level",
		EnvVar: "LOG_LEVEL",
	})

	log := logger.NewUPPLogger(*appSystemCode, *logLevel)

	app.Action = func() {
		log.Infof("[Startup] System code: %s, App Name: %s, Port: %s", *appSystemCode, *appName, *port)

		// We don't want logging for GTG requests in the middleware
		healthCl, err := fthttp.NewClient(
			fthttp.WithTimeout(10*time.Second),
			fthttp.WithSysInfo("PAC", *appSystemCode))
		if err != nil {
			log.WithError(err).Error("Error creating healthchecks HTTP client, exiting ...")
			return
		}
		loggingCl, err := fthttp.NewClient(
			fthttp.WithTimeout(10*time.Second),
			fthttp.WithSysInfo("PAC", *appSystemCode),
			fthttp.WithLogging(log))
		if err != nil {
			log.WithError(err).Error("Error creating logging HTTP client, exiting ...")
			return
		}

		contentAPI, err := draft.NewContentAPI(*draftContentEndpoint, *draftContentGtgEndpoint, loggingCl, healthCl)
		if err != nil {
			log.WithError(err).Error("Draft Content API error, exiting ...")
			return
		}

		umbrellaAPI, err := suggestions.NewUmbrellaAPI(*suggestionsEndpoint, *suggestionsGtgEndpoint, *suggestionsAPIKey, loggingCl, healthCl)
		if err != nil {
			log.WithError(err).Error("Suggestions Umbrella API error, exiting ...")
			return
		}

		serveEndpoints(*appSystemCode, *appName, *port, apiYml, requestHandler{contentAPI, umbrellaAPI, log}, log)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.WithError(err).Errorf("%s could not start!", defaultAppName)
		return
	}
}

func serveEndpoints(appSystemCode string, appName string, port string, apiYml *string, requestHandler requestHandler, log *logger.UPPLogger) {
	healthService := health.NewService(appSystemCode, appName, appDescription,
		requestHandler.dca, requestHandler.sua, log)

	serveMux := http.NewServeMux()

	serveMux.HandleFunc(health.DefaultHealthPath, http.HandlerFunc(fthealth.Handler(healthService.Health())))
	serveMux.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(healthService.GTG))
	serveMux.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)

	if apiYml != nil {
		apiEndpoint, err := api.NewAPIEndpointForFile(*apiYml)
		if err != nil {
			log.WithError(err).WithField("file", apiYml).Warn("Failed to serve the API Endpoint for this service. Please validate the file exists, and that it fits the OpenAPI specification.")
		} else {
			serveMux.HandleFunc(api.DefaultPath, apiEndpoint.ServeHTTP)
		}
	}

	servicesRouter := mux.NewRouter()
	servicesRouter.HandleFunc("/drafts/content/{uuid}/suggestions",
		requestHandler.draftContentSuggestionsRequest).Methods("GET")

	monitoringRouter := httphandlers.TransactionAwareRequestLoggingHandler(log, servicesRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	serveMux.Handle("/", monitoringRouter)

	server := &http.Server{Addr: ":" + port, Handler: serveMux}

	done := make(chan struct{})
	go func() {
		waitForSignal()

		log.Infof("[Shutdown] %s is shutting down", defaultAppName)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.WithError(err).Error("Could not gracefully shutdown HTTP server")
		}

		close(done)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.WithError(err).Fatal("Error starting or closing HTTP server")
	}

	<-done
}

func waitForSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
