package health

import (
	"context"
	"fmt"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/service-status-go/gtg"

	"github.com/Financial-Times/draft-content-suggestions/config"
	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
)

const DefaultHealthPath = "/__health"

type Service struct {
	config       *appConfig
	healthChecks []fthealth.Check
	gtgChecks    []gtg.StatusChecker
	contentAPI   draft.ContentAPI
	umbrellaAPI  suggestions.UmbrellaAPI
	log          *logger.UPPLogger
}

type appConfig struct {
	appSystemCode  string
	appName        string
	appDescription string
}

type ExternalService interface {
	Endpoint() string
	GTG() error
}

func NewService(appSystemCode string, appName string,
	appDescription string, contentAPI draft.ContentAPI,
	umbrellaAPI suggestions.UmbrellaAPI, hcConfig *config.Config, services []ExternalService, log *logger.UPPLogger) (*Service, error) {
	hc := &Service{
		config: &appConfig{
			appSystemCode:  appSystemCode,
			appName:        appName,
			appDescription: appDescription,
		},
		contentAPI:  contentAPI,
		umbrellaAPI: umbrellaAPI,
		log:         log,
	}

	hc.healthChecks = []fthealth.Check{hc.draftContentCheck(), hc.suggestionsCheck()}

	draftContentCheck := func() gtg.Status {
		return gtgCheck(hc.draftContentChecker)
	}
	suggestionsCheck := func() gtg.Status {
		return gtgCheck(hc.suggestionsChecker)
	}

	gtgChecks := append(hc.gtgChecks, draftContentCheck, suggestionsCheck)
	hc.gtgChecks = gtgChecks

	for endpoint, cfg := range hcConfig.HealthChecks {
		externalService, err := findService(endpoint, services)
		if err != nil {
			return nil, err
		}

		c := fthealth.Check{
			ID:               cfg.ID,
			BusinessImpact:   cfg.BusinessImpact,
			Name:             cfg.Name,
			PanicGuide:       cfg.PanicGuide,
			Severity:         cfg.Severity,
			TechnicalSummary: fmt.Sprintf(cfg.TechnicalSummary, endpoint),
			Checker:          externalServiceChecker(externalService, cfg.CheckerName),
		}
		hc.healthChecks = append(hc.healthChecks, c)
	}
	return hc, nil
}

func findService(endpoint string, services []ExternalService) (ExternalService, error) {
	for _, s := range services {
		if s.Endpoint() == endpoint {
			return s, nil
		}
	}

	return nil, fmt.Errorf("unable to find service with endpoint %v", endpoint)
}

func externalServiceChecker(s ExternalService, serviceName string) func() (string, error) {
	return func() (string, error) {
		if err := s.GTG(); err != nil {
			return fmt.Sprintf("%s is not good-to-go", serviceName), err
		}
		return fmt.Sprintf("%s is good-to-go", serviceName), nil
	}
}

func (s *Service) Health() fthealth.HC {
	return &fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{
			SystemCode:  s.config.appSystemCode,
			Name:        s.config.appName,
			Description: s.config.appDescription,
			Checks:      s.healthChecks,
		},
		Timeout: 10 * time.Second,
	}
}

func (s *Service) GTG() gtg.Status {
	return gtg.FailFastParallelCheck(s.gtgChecks)()
}

func (s *Service) draftContentCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Unable to provide suggestions to editorial for tagging content",
		Name:             "Draft Content Service Health Check",
		PanicGuide:       "https://runbooks.ftops.tech/draft-content-suggestions",
		Severity:         1,
		TechnicalSummary: "Checks whether the health endpoint of draft-content-api returns successful responses",
		Checker:          s.draftContentChecker,
	}
}

func (s *Service) suggestionsCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Unable to provide suggestions to editorial for tagging content",
		Name:             "Suggestions Umbrella Service Health Check",
		PanicGuide:       "https://runbooks.ftops.tech/draft-content-suggestions",
		Severity:         1,
		TechnicalSummary: "Checks whether the suggestions umbrella endpoint is accessible and returns responses",
		Checker:          s.suggestionsChecker,
	}
}

func (s *Service) draftContentChecker() (string, error) {
	res, err := s.contentAPI.IsGTG(context.Background())
	if err != nil {
		s.log.WithField("healthEndpoint", s.contentAPI.Endpoint()).WithError(err).Error("Draft Content GTG check failed")
	}

	return res, err
}

func (s *Service) suggestionsChecker() (string, error) {
	res, err := s.umbrellaAPI.IsGTG(context.Background())
	if err != nil {
		s.log.WithField("healthEndpoint", s.umbrellaAPI.Endpoint()).WithError(err).Error("UPP Suggestions API GTG check failed")
	}

	return res, err
}

func gtgCheck(handler func() (string, error)) gtg.Status {
	if _, err := handler(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}
