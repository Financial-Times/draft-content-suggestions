package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

	"github.com/Financial-Times/draft-content-suggestions/config"
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
		Value:  "https://upp-staging-delivery-glb.upp.ft.com/content/suggest",
		Desc:   "Endpoint for Suggestions Umbrella",
		EnvVar: "SUGGESTIONS_ENDPOINT",
	})
	suggestionsGtgEndpoint := app.String(cli.StringOpt{
		Name:   "suggestions-umbrella-gtg-endpoint",
		Value:  "https://upp-staging-delivery-glb.upp.ft.com/content/suggest/__gtg",
		Desc:   "Endpoint for Suggestions Umbrella",
		EnvVar: "SUGGESTIONS_GTG_ENDPOINT",
	})
	deliveryBasicAuth := app.String(cli.StringOpt{
		Name:   "delivery-basic-auth",
		Value:  "username:password",
		Desc:   "Basic auth for access to the delivery UPP clusters",
		EnvVar: "DELIVERY_BASIC_AUTH",
	})
	validatorYml := app.String(cli.StringOpt{
		Name:   "validator-yml",
		Value:  "./config.yml",
		Desc:   "Location of the Validator configuration YML file.",
		EnvVar: "VALIDATOR_YML",
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

		validatorConfig, err := config.ReadConfig(*validatorYml)
		if err != nil {
			log.WithError(err).Fatal("unable to read r/w YAML configuration")
		}

		contentTypeMapping := draft.BuildContentTypeMapping(validatorConfig, loggingCl, log)
		resolver := draft.NewContentValidatorResolver(contentTypeMapping)

		contentAPI, err := draft.NewContentAPI(*draftContentEndpoint, *draftContentGtgEndpoint, loggingCl, healthCl, resolver)
		if err != nil {
			log.WithError(err).Error("Draft Content API error, exiting ...")
			return
		}

		basicAuthCredentials := strings.Split(*deliveryBasicAuth, ":")
		if len(basicAuthCredentials) != 2 {
			log.Fatal("error while resolving basic auth")
		}

		umbrellaAPI, err := suggestions.NewUmbrellaAPI(*suggestionsEndpoint, *suggestionsGtgEndpoint, basicAuthCredentials[0], basicAuthCredentials[1], loggingCl, healthCl)
		if err != nil {
			log.WithError(err).Error("Suggestions Umbrella API error, exiting ...")
			return
		}

		healthService, err := health.NewService(*appSystemCode, *appName, appDescription,
			contentAPI, umbrellaAPI, validatorConfig, extractServices(contentTypeMapping), log)
		if err != nil {
			log.WithError(err).Fatal("Unable to create health service")
		}

		serveEndpoints(*port, apiYml, requestHandler{contentAPI, umbrellaAPI, log}, healthService, log)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.WithError(err).Errorf("%s could not start!", defaultAppName)
		return
	}
}

func extractServices(dcm map[string]draft.ContentValidator) []health.ExternalService {
	result := make([]health.ExternalService, 0, len(dcm))

	for _, value := range dcm {
		result = append(result, value)
	}

	return result
}

func serveEndpoints(port string, apiYml *string, requestHandler requestHandler, healthService *health.Service, log *logger.UPPLogger) {
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
	servicesRouter.HandleFunc("/drafts/content/suggestions",
		requestHandler.getDraftSuggestionsForContent).Methods("POST")

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
