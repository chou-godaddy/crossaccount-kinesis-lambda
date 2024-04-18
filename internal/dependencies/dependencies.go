package dependencies

import (
	"context"
	"fulfillment-entitlements-api/internal/config"
	"net/url"
	"time"

	httpclient "github.com/gdcorp-domains/fulfillment-golang-httpclient"
	logging "github.com/gdcorp-domains/fulfillment-golang-logging"
	sso "github.com/gdcorp-domains/fulfillment-golang-sso-auth/client"
	"github.com/gdcorp-domains/fulfillment-golang-sso-auth/modifiers"
	"github.com/gdcorp-domains/fulfillment-golang-sso-auth/token/expiry"
	gosecrets "github.com/gdcorp-domains/fulfillment-gosecrets"
)

// Dependencies encapsulates all dependencies to be carried through request processing.
type Dependencies struct {
	Config             *config.Config
	HTTPPlain          httpclient.Client
	HTTPWithClientCert httpclient.Client
	HTTPWithCertSSO    httpclient.Client
	HTTPWithIAMSSO     httpclient.Client
	Logger             logging.Logger
}

// Initialize takes the static dependencies and does any one-time setup
func (d *Dependencies) Initialize() error {
	if d.Config.HTTPConfig != nil {
		client, err := httpclient.NewClientBuilder(nil).HTTPClient(context.Background(), d.Config.HTTPConfig)
		if err != nil {
			return err
		}
		d.HTTPPlain = client
	}

	d.Logger = logging.New(d.Config.Logging)

	iamClient := sso.NewIAMClient(*d.Config.SSO.IAMConfig, d.HTTPPlain)
	// TODO: config refresh interval by envs
	iamCache := sso.NewIAMRefresher(iamClient, 15*time.Minute, d.Logger)
	d.HTTPWithIAMSSO = d.HTTPPlain.WithRequestModifier(modifiers.NewAddIAMJWTModifier(expiry.Impact(d.Config.SSO.ExpireLevel), iamCache))

	secretRetriever := gosecrets.NewSecretRetriever()
	if d.Config.CertConfig != nil {
		certLoader := httpclient.NewCertGetter(*d.Config.CertConfig, secretRetriever)
		clientWithCerts, err := httpclient.NewClientBuilder(certLoader).HTTPClient(context.Background(), d.Config.HTTPConfig)
		if err != nil {
			return err
		}

		ssoURL, err := url.Parse(d.Config.SSO.URL)
		if err != nil {
			return err
		}
		ssoJWTGetter := sso.New(ssoURL, clientWithCerts, d.Config.SSO.SSOPKCacheConfig)
		d.HTTPWithCertSSO = d.HTTPPlain.WithRequestModifier(modifiers.NewAddCertJWTModifier(expiry.Impact(d.Config.SSO.ExpireLevel), ssoJWTGetter))
	}

	return nil
}

// New constructs a new Dependencies, or panics on error
func New(config *config.Config) *Dependencies {
	return &Dependencies{
		Config: config,
	}
}

// GetConfig returns the config
func (dep Dependencies) GetConfig() *config.Config {
	return dep.Config
}

// GetLogger returns the logger
func (dep Dependencies) GetLogger() logging.Logger {
	return dep.Logger
}
