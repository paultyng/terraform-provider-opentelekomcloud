package cfg

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsCredentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/jinzhu/copier"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3/credentials"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/obs"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/helper/pathorcontents"
)

const (
	osPrefix = "OS_"
)

type Config struct {
	AccessKey           string
	SecretKey           string
	CACertFile          string
	AllowReauth         bool
	ClientCertFile      string
	ClientKeyFile       string
	Cloud               string
	DomainID            string
	DomainName          string
	EndpointType        string
	IdentityEndpoint    string
	Insecure            bool
	Password            string
	Passcode            string
	Region              string
	Swauth              bool
	TenantID            string
	TenantName          string
	Token               string
	SecurityToken       string
	Username            string
	UserID              string
	AgencyName          string
	AgencyDomainName    string
	DelegatedProject    string
	MaxRetries          int
	MaxBackoffRetries   int
	BackoffRetryTimeout int

	UserAgent string

	HwClient *golangsdk.ProviderClient
	s3sess   *session.Session

	DomainClient *golangsdk.ProviderClient

	environment *openstack.Env
}

func (c *Config) LoadAndValidate() error {
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries should be a positive value")
	}

	if err := c.Load(); err != nil {
		return err
	}

	if c.IdentityEndpoint == "" {
		return fmt.Errorf("'auth_url' must be specified")
	}

	if err := c.validateEndpoint(); err != nil {
		return err
	}

	if err := c.validateProject(); err != nil {
		return err
	}

	var err error
	switch {
	case c.Token != "":
		err = buildClientByToken(c)
	case c.AccessKey != "" && c.SecretKey != "":
		err = buildClientByAKSK(c)
	case c.Password != "" && (c.Username != "" || c.UserID != ""):
		err = buildClientByPassword(c)
	default:
		err = errors.New(
			"no auth means provided. Token, AK/SK or username/password are required for authentication")
	}
	if err != nil {
		return fmt.Errorf("failed to authenticate:\n%s", err)
	}

	var osDebug bool
	if os.Getenv("OS_DEBUG") != "" {
		osDebug = true
	}
	return c.newS3Session(osDebug)
}

// setIfEmpty set non-empty `loaded` value to empty `target` variable
func setIfEmpty(target *string, loaded string) {
	if *target == "" && loaded != "" {
		*target = loaded
	}
}

// Load - load existing configuration from config files (`clouds.yaml`, etc.) and env variables
func (c *Config) Load() error {
	if c.environment == nil {
		c.environment = openstack.NewEnv(osPrefix)
	}
	cloud, err := c.environment.Cloud(c.Cloud)
	if err != nil {
		return err
	}
	// Auth data
	setIfEmpty(&c.Username, cloud.AuthInfo.Username)
	setIfEmpty(&c.UserID, cloud.AuthInfo.UserID)

	if c.UserID != "" {
		c.Username = ""
	}

	setIfEmpty(&c.TenantName, cloud.AuthInfo.ProjectName)
	setIfEmpty(&c.TenantID, cloud.AuthInfo.ProjectID)
	setIfEmpty(&c.DomainName, cloud.AuthInfo.DomainName)
	setIfEmpty(&c.DomainID, cloud.AuthInfo.DomainID)

	// project scope
	setIfEmpty(&c.DomainName, cloud.AuthInfo.ProjectDomainName)
	setIfEmpty(&c.DomainID, cloud.AuthInfo.ProjectDomainID)

	// user scope
	setIfEmpty(&c.DomainName, cloud.AuthInfo.UserDomainName)
	setIfEmpty(&c.DomainID, cloud.AuthInfo.UserDomainID)

	// default domain
	setIfEmpty(&c.DomainID, cloud.AuthInfo.DefaultDomain)

	setIfEmpty(&c.IdentityEndpoint, cloud.AuthInfo.AuthURL)
	setIfEmpty(&c.Token, cloud.AuthInfo.Token)
	setIfEmpty(&c.Password, cloud.AuthInfo.Password)

	// General cloud info
	setIfEmpty(&c.Region, cloud.RegionName)
	setIfEmpty(&c.CACertFile, cloud.CACertFile)
	setIfEmpty(&c.ClientCertFile, cloud.ClientCertFile)
	setIfEmpty(&c.ClientKeyFile, cloud.ClientKeyFile)
	if cloud.Verify != nil {
		c.Insecure = !*cloud.Verify
	}
	return nil
}

func (c *Config) generateTLSConfig() (*tls.Config, error) {
	config := &tls.Config{}
	if c.CACertFile != "" {
		caCert, _, err := pathorcontents.Read(c.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("error reading CA Cert: %s", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caCert))
		config.RootCAs = caCertPool
	}

	if c.Insecure {
		config.InsecureSkipVerify = true
	}

	if c.ClientCertFile != "" && c.ClientKeyFile != "" {
		clientCert, _, err := pathorcontents.Read(c.ClientCertFile)
		if err != nil {
			return nil, fmt.Errorf("error reading Client Cert: %s", err)
		}
		clientKey, _, err := pathorcontents.Read(c.ClientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("error reading Client Key: %s", err)
		}

		cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
		if err != nil {
			return nil, err
		}

		config.Certificates = []tls.Certificate{cert}

		config.BuildNameToCertificate()
	}

	return config, nil
}

// GetCredentials This function is responsible for reading credentials from the
// environment in the case that they're not explicitly specified
// in the Terraform configuration.
func (c *Config) GetCredentials() (*awsCredentials.Credentials, error) {
	// build a chain provider, lazy-evaluated by aws-sdk
	providers := []awsCredentials.Provider{
		&awsCredentials.StaticProvider{Value: awsCredentials.Value{
			AccessKeyID:     c.AccessKey,
			SecretAccessKey: c.SecretKey,
			SessionToken:    c.SecurityToken,
		}},
		&awsCredentials.EnvProvider{},
		&awsCredentials.SharedCredentialsProvider{
			Filename: "",
			Profile:  "",
		},
	}

	// Build isolated HTTP client to avoid issues with globally-shared settings
	client := cleanhttp.DefaultClient()

	// Keep the default timeout (100ms) low as we don't want to wait in non-EC2 environments
	client.Timeout = 100 * time.Millisecond
	const userTimeoutEnvVar = "AWS_METADATA_TIMEOUT"
	userTimeout := os.Getenv(userTimeoutEnvVar)
	if userTimeout != "" {
		newTimeout, err := time.ParseDuration(userTimeout)
		if err == nil {
			if newTimeout.Nanoseconds() > 0 {
				client.Timeout = newTimeout
			} else {
				log.Printf("[WARN] Non-positive value of %s (%s) is meaningless, ignoring", userTimeoutEnvVar, newTimeout.String())
			}
		} else {
			log.Printf("[WARN] Error converting %s to time.Duration: %s", userTimeoutEnvVar, err)
		}
	}

	log.Printf("[INFO] Setting AWS metadata API timeout to %s", client.Timeout.String())
	config := &aws.Config{
		HTTPClient: client,
	}
	usedEndpoint := SetOptionalEndpoint(config)

	// Add the default AWS provider for ECS Task Roles if the relevant env variable is set
	if uri := os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI"); len(uri) > 0 {
		providers = append(providers, defaults.RemoteCredProvider(*config, defaults.Handlers()))
		log.Print("[INFO] ECS container credentials detected, RemoteCredProvider added to auth chain")
	}

	// Real AWS should reply to a simple metadata request.
	// We check it actually does to ensure something else didn't just
	// happen to be listening on the same IP:Port
	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	metadataClient := ec2metadata.New(sess)
	if metadataClient.Available() {
		providers = append(providers, &ec2rolecreds.EC2RoleProvider{
			Client: metadataClient,
		})
		log.Print("[INFO] AWS EC2 instance detected via default metadata" +
			" API endpoint, EC2RoleProvider added to the auth chain")
	} else {
		if usedEndpoint == "" {
			usedEndpoint = "default location"
		}
		log.Printf("[INFO] Ignoring AWS metadata API endpoint at %s "+
			"as it doesn't return any instance-id", usedEndpoint)
	}

	return awsCredentials.NewChainCredentials(providers), nil
}

func (c *Config) newS3Session(osDebug bool) error {
	// Don't get AWS session unless we need it for AccessKey, SecretKey.
	if c.AccessKey != "" && c.SecretKey != "" {
		// Setup AWS/S3 client/config information for Swift S3 buckets
		log.Println("[INFO] Building Swift S3 auth structure")
		creds, err := c.GetCredentials()
		if err != nil {
			return err
		}
		// Call Get to check for credential provider. If nothing found, we'll get an
		// error, and we can present it nicely to the user
		cp, err := creds.Get()
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoCredentialProviders" {
				return fmt.Errorf(`no valid credential sources found for Swift S3 Provider.
  Please see https://terraform.io/docs/providers/aws/index.html for more information on
  providing credentials for the S3 Provider`)
			}

			return fmt.Errorf("error loading credentials for Swift S3 Provider: %s", err)
		}

		log.Printf("[INFO] Swift S3 Auth provider used: %q", cp.ProviderName)

		awsConfig := &aws.Config{
			Credentials: creds,
			Region:      aws.String(c.GetRegion(nil)),
			// MaxRetries:       aws.Int(c.MaxRetries),
			HTTPClient: cleanhttp.DefaultClient(),
			// S3ForcePathStyle: aws.Bool(c.S3ForcePathStyle),
		}

		if osDebug {
			awsConfig.LogLevel = aws.LogLevel(aws.LogDebugWithHTTPBody | aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors)
			awsConfig.Logger = awsLogger{}
		}

		if c.Insecure {
			transport := awsConfig.HTTPClient.Transport.(*http.Transport)
			transport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		// Set up base session for AWS/Swift S3
		c.s3sess, err = session.NewSession(awsConfig)
		if err != nil {
			return fmt.Errorf("error creating Swift S3 session: %s", err)
		}
	}
	return nil
}

var validEndpoints = []string{
	"internal", "internalURL",
	"admin", "adminURL",
	"public", "publicURL",
	"",
}

func (c *Config) validateEndpoint() error {
	for _, endpoint := range validEndpoints {
		if c.EndpointType == endpoint {
			return nil
		}
	}
	return fmt.Errorf("invalid endpoint type provided: %s", c.EndpointType)
}

// validateProject checks that `Project`(`Tenant`) value is set
func (c *Config) validateProject() error {
	if c.TenantName == "" && c.TenantID == "" && c.DelegatedProject == "" {
		return errors.New("no project name/id or delegated project is provided")
	}
	return nil
}

func buildClientByToken(c *Config) error {
	var pao, dao golangsdk.AuthOptions

	if c.AgencyDomainName != "" && c.AgencyName != "" {
		pao = golangsdk.AuthOptions{
			AgencyName:       c.AgencyName,
			AgencyDomainName: c.AgencyDomainName,
			DelegatedProject: c.DelegatedProject,
		}

		dao = golangsdk.AuthOptions{
			AgencyName:       c.AgencyName,
			AgencyDomainName: c.AgencyDomainName,
		}
	} else {
		pao = golangsdk.AuthOptions{
			DomainID:   c.DomainID,
			DomainName: c.DomainName,
			TenantID:   c.TenantID,
			TenantName: c.TenantName,
		}

		dao = golangsdk.AuthOptions{
			DomainID:   c.DomainID,
			DomainName: c.DomainName,
		}
	}

	for _, ao := range []*golangsdk.AuthOptions{&pao, &dao} {
		ao.IdentityEndpoint = c.IdentityEndpoint
		ao.TokenID = c.Token
	}
	return c.genClients(pao, dao)
}

func buildClientByAKSK(c *Config) error {
	var pao, dao golangsdk.AKSKAuthOptions

	if c.AgencyDomainName != "" && c.AgencyName != "" {
		pao = golangsdk.AKSKAuthOptions{
			DomainID:         c.DomainID,
			Domain:           c.DomainName,
			AgencyName:       c.AgencyName,
			AgencyDomainName: c.AgencyDomainName,
			DelegatedProject: c.DelegatedProject,
		}

		dao = golangsdk.AKSKAuthOptions{
			DomainID:         c.DomainID,
			Domain:           c.DomainName,
			AgencyName:       c.AgencyName,
			AgencyDomainName: c.AgencyDomainName,
		}
	} else {
		pao = golangsdk.AKSKAuthOptions{
			ProjectName: c.TenantName,
			ProjectId:   c.TenantID,
		}

		dao = golangsdk.AKSKAuthOptions{
			DomainID: c.DomainID,
			Domain:   c.DomainName,
		}
	}
	if c.SecurityToken != "" {
		dao.ProjectId = c.TenantID
		dao.ProjectName = c.TenantName
	}
	for _, ao := range []*golangsdk.AKSKAuthOptions{&pao, &dao} {
		ao.IdentityEndpoint = c.IdentityEndpoint
		ao.AccessKey = c.AccessKey
		ao.SecretKey = c.SecretKey
		if c.SecurityToken != "" {
			ao.SecurityToken = c.SecurityToken
		}
	}
	return c.genClients(pao, dao)
}

func buildClientByPassword(c *Config) error {
	var pao, dao golangsdk.AuthOptions

	if c.AgencyDomainName != "" && c.AgencyName != "" {
		pao = golangsdk.AuthOptions{
			DomainID:         c.DomainID,
			DomainName:       c.DomainName,
			AgencyName:       c.AgencyName,
			AgencyDomainName: c.AgencyDomainName,
			DelegatedProject: c.DelegatedProject,
		}

		dao = golangsdk.AuthOptions{
			DomainID:         c.DomainID,
			DomainName:       c.DomainName,
			AgencyName:       c.AgencyName,
			AgencyDomainName: c.AgencyDomainName,
		}
	} else {
		pao = golangsdk.AuthOptions{
			DomainID:   c.DomainID,
			DomainName: c.DomainName,
			TenantID:   c.TenantID,
			TenantName: c.TenantName,
		}

		dao = golangsdk.AuthOptions{
			DomainID:   c.DomainID,
			DomainName: c.DomainName,
		}
	}

	for _, ao := range []*golangsdk.AuthOptions{&pao, &dao} {
		ao.IdentityEndpoint = c.IdentityEndpoint
		ao.Password = c.Password
		ao.Username = c.Username
		ao.UserID = c.UserID
		ao.Passcode = c.Passcode
		ao.AllowReauth = c.AllowReauth
	}
	return c.genClients(pao, dao)
}

func (c *Config) genClients(pao, dao golangsdk.AuthOptionsProvider) error {
	client, err := c.genClient(pao)
	if err != nil {
		return fmt.Errorf("error generating project client: %w", err)
	}
	c.HwClient = client

	client, err = c.genClient(dao)
	if err != nil {
		return fmt.Errorf("error generating domain client: %w", err)
	}
	c.DomainClient = client
	return nil
}

func (c *Config) genClient(ao golangsdk.AuthOptionsProvider) (*golangsdk.ProviderClient, error) {
	client, err := openstack.NewClient(ao.GetIdentityEndpoint())
	if err != nil {
		return nil, err
	}

	// Set UserAgent
	client.UserAgent.Prepend(c.UserAgent)

	config, err := c.generateTLSConfig()
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment, TLSClientConfig: config}

	// if OS_DEBUG is set, log the requests and responses
	var osDebug bool
	if os.Getenv("OS_DEBUG") != "" {
		osDebug = true
	}

	client.MaxBackoffRetries = pointerto.Int(c.MaxBackoffRetries)
	defaultBackoffTimeout := time.Duration(c.BackoffRetryTimeout) * time.Second
	client.BackoffRetryTimeout = &defaultBackoffTimeout

	client.HTTPClient = http.Client{
		Transport: &RoundTripper{
			Rt:         transport,
			OsDebug:    osDebug,
			MaxRetries: c.MaxRetries,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if client.AKSKAuthOptions.AccessKey != "" {
				golangsdk.ReSign(req, golangsdk.SignOptions{
					AccessKey: client.AKSKAuthOptions.AccessKey,
					SecretKey: client.AKSKAuthOptions.SecretKey,
				})
			}
			return nil
		},
	}

	// If using Swift Authentication, there's no need to validate authentication normally.
	if !c.Swauth {
		err = openstack.Authenticate(client, ao)
		if err != nil {
			return nil, err
		}
	}

	c.Region = client.RegionID

	return client, nil
}

type awsLogger struct{}

func (l awsLogger) Log(args ...interface{}) {
	tokens := make([]string, 0, len(args))
	for _, arg := range args {
		if token, ok := arg.(string); ok {
			tokens = append(tokens, token)
		}
	}
	log.Printf("[DEBUG] [aws-sdk-go] %s", strings.Join(tokens, " "))
}

func (c *Config) determineRegion(region string) string {
	// If a resource-level region was not specified, and a provider-level region was set,
	// use the provider-level region.
	if region == "" && c.Region != "" {
		region = c.Region
	}

	log.Printf("[DEBUG] OpenTelekomCloud Region is: %s", region)
	return region
}

func (c *Config) S3Client(region string) (*s3.S3, error) {
	if c.s3sess == nil {
		return nil, fmt.Errorf("missing credentials for Swift S3 Provider, need access_key and secret_key values for provider")
	}

	client, err := openstack.NewOBSService(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
	if err != nil {
		return nil, err
	}

	awsS3Sess := c.s3sess.Copy(&aws.Config{Endpoint: aws.String(client.Endpoint)})
	s3conn := s3.New(awsS3Sess)

	return s3conn, err
}

func setUpOBSLogging() {
	// init log
	if os.Getenv("OS_DEBUG") != "" {
		var logfile = "./.obs-sdk.log"
		// maxLogSize:10M, backups:10
		if err := obs.InitLog(logfile, 1024*1024*10, 10, obs.LEVEL_DEBUG, false); err != nil {
			log.Printf("[WARN] initial obs sdk log failed: %s", err)
		}
	}
}

// issueTemporaryCredentials creates temporary AK/SK, which can be used to auth in OBS when AK/SK is not provided
func (c *Config) issueTemporaryCredentials() (*credentials.TemporaryCredential, error) {
	if c.AccessKey != "" && c.SecretKey != "" {
		return &credentials.TemporaryCredential{
			AccessKey:     c.AccessKey,
			SecretKey:     c.SecretKey,
			SecurityToken: c.SecurityToken,
		}, nil
	}
	client, err := c.IdentityV3Client()
	if err != nil {
		return nil, fmt.Errorf("error creating identity v3 domain client: %s", err)
	}
	credential, err := credentials.CreateTemporary(client, credentials.CreateTemporaryOpts{
		Methods: []string{"token"},
		Token:   client.Token(),
	}).Extract()
	if err != nil {
		return nil, fmt.Errorf("error creating temporary AK/SK: %s", err)
	}
	return credential, nil
}

func (c *Config) NewObjectStorageClient(region string) (*obs.ObsClient, error) {
	cred, err := c.issueTemporaryCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to construct OBS client without AK/SK: %s", err)
	}

	client, err := openstack.NewOBSService(c.HwClient, golangsdk.EndpointOpts{
		Region:       c.determineRegion(region),
		Availability: c.getEndpointType(),
	})
	if err != nil {
		return nil, err
	}

	setUpOBSLogging()
	proxyConfigure, err := obsProxyConf()
	if err != nil {
		return nil, err
	}

	return obs.New(
		cred.AccessKey, cred.SecretKey, client.Endpoint,
		obs.WithSecurityToken(cred.SecurityToken), obs.WithSignature(obs.SignatureObs), proxyConfigure,
	)
}

func obsProxyConf() (obs.Configurer, error) {
	proxyConfigure := obs.WithProxyUrl("")
	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")

	if httpsProxy != "" {
		_, err := url.ParseRequestURI(httpsProxy)
		if err != nil {
			return proxyConfigure, err
		}
		return obs.WithProxyUrl(httpsProxy), nil
	} else if httpProxy != "" {
		_, err := url.ParseRequestURI(httpProxy)
		if err != nil {
			return proxyConfigure, err
		}
		return obs.WithProxyUrl(httpProxy), nil
	}
	return proxyConfigure, nil
}

func (c *Config) BlockStorageV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewBlockStorageV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) BlockStorageV3Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewBlockStorageV3(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) CbrV3Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewCBRService(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DisV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDISServiceV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DrsV3Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDRSServiceV3(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) ComputeV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewComputeV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       c.determineRegion(region),
		Availability: c.getEndpointType(),
	})
}

func (c *Config) ComputeV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewComputeV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DnsV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDNSV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) GaussDBV3Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewGaussDBV3(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) IdentityV3Client(_ ...string) (*golangsdk.ServiceClient, error) {
	return openstack.NewIdentityV3(c.DomainClient, golangsdk.EndpointOpts{
		Availability: c.getEndpointType(),
	})
}

// IdentityV30Client - provides client is used for use with endpoints with invalid "v3.0" URLs
func (c *Config) IdentityV30Client() (*golangsdk.ServiceClient, error) {
	service, err := openstack.NewIdentityV3(c.DomainClient, golangsdk.EndpointOpts{
		Availability: c.getEndpointType(),
	})
	if err != nil {
		return nil, err
	}
	service.Endpoint = strings.Replace(service.IdentityEndpoint, "v3/", "v3.0/", 1)
	return service, nil
}

func (c *Config) RegionIdentityV3Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewIdentityV3(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) ImageV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewIMSV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) ImageV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewIMSV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) NetworkingV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewNetworkV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) NetworkingV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewNetworkV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) NetworkingV3Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewVpcV3(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) SmnV2Client(projectName ProjectName) (*golangsdk.ServiceClient, error) {
	newConfig, err := reconfigProjectName(*c, projectName)
	if err != nil {
		return nil, err
	}
	return openstack.NewSMNV2(newConfig.HwClient, golangsdk.EndpointOpts{
		Region:       c.GetRegion(nil),
		Availability: c.getEndpointType(),
	})
}

func (c *Config) SmnV2TagClient(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewSMNV2Tags(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) CesV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewCESClient(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) getEndpointType() golangsdk.Availability {
	if c.EndpointType == "internal" || c.EndpointType == "internalURL" {
		return golangsdk.AvailabilityInternal
	}
	if c.EndpointType == "admin" || c.EndpointType == "adminURL" {
		return golangsdk.AvailabilityAdmin
	}
	return golangsdk.AvailabilityPublic
}

func (c *Config) KmsKeyV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewKMSV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) NatV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewNatV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) OrchestrationV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewOrchestrationV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) SfsV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewSharedFileSystemV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) SfsTurboV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewSharedFileSystemTurboV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) VbsV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewVBS(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) AutoscalingV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewAutoScalingV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) AutoscalingV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewAutoScalingV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) CsbsV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewCSBSService(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DCaaSV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDCaaSV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DCaaSV3Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDCaaSV3(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DdmV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDDMV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DdmV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDDMV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DdmV3Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDDMV3(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DehV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDeHServiceV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DmsV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDMSServiceV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DmsV11Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDMSServiceV11(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DmsV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDMSServiceV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) MrsV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewMapReduceV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) ElbV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewELBV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) ElbV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewELBV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) ElbV3Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewELBV3(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) RdsV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewRDSV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) AntiddosV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewAntiDDoSV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) CtsV1Client(projectName ProjectName) (*golangsdk.ServiceClient, error) {
	newConfig, err := reconfigProjectName(*c, projectName)
	if err != nil {
		return nil, err
	}
	return openstack.NewCTSV1(newConfig.HwClient, golangsdk.EndpointOpts{
		Region:       c.GetRegion(nil),
		Availability: c.getEndpointType(),
	})
}

func (c *Config) CtsV2Client(projectName ProjectName) (*golangsdk.ServiceClient, error) {
	newConfig, err := reconfigProjectName(*c, projectName)
	if err != nil {
		return nil, err
	}
	return openstack.NewCTSV2(newConfig.HwClient, golangsdk.EndpointOpts{
		Region:       c.GetRegion(nil),
		Availability: c.getEndpointType(),
	})
}

func (c *Config) CtsV3Client(projectName ProjectName) (*golangsdk.ServiceClient, error) {
	newConfig, err := reconfigProjectName(*c, projectName)
	if err != nil {
		return nil, err
	}
	return openstack.NewCTSV3(newConfig.HwClient, golangsdk.EndpointOpts{
		Region:       c.GetRegion(nil),
		Availability: c.getEndpointType(),
	})
}

func (c *Config) CssV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewCSSService(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) CceV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewCCEv1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) CceV3Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewCCE(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) CceV3AddonClient(region string) (*golangsdk.ServiceClient, error) {
	client, err := c.CceV3Client(region)
	if err != nil {
		return nil, err
	}
	client.ResourceBase = fmt.Sprintf("%sapi/v3/", client.Endpoint)
	return client, nil
}

func (c *Config) DcsV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDCSServiceV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DcsV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDCSServiceV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) RdsTagV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewRdsTagV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) WafV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewWAFV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) WafDedicatedV1Client(region string) (*golangsdk.ServiceClient, error) {
	if region != "eu-ch2" {
		return openstack.NewWAFDV1(c.HwClient, golangsdk.EndpointOpts{
			Region:       region,
			Availability: c.getEndpointType(),
		})
	} else {
		return openstack.NewWAFDSwissV1(c.HwClient, golangsdk.EndpointOpts{
			Region:       region,
			Availability: c.getEndpointType(),
		})
	}
}

func (c *Config) RdsV3Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewRDSV3(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) RmsV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewRmsServiceV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) SdrsV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewSDRSV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) LtsV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewLTSV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DdsV3Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDDSServiceV3(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) SwrV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewSWRV2(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) VpcEpV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewVpcEpV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DwsV1Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewDWSV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) APIGWV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewAPIGW(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) FuncGraphV2Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewFuncGraph(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) ErV3Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewERServiceV3(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) DwsV2Client(region string) (*golangsdk.ServiceClient, error) {
	service, err := openstack.NewDWSV1(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
	if err != nil {
		return nil, err
	}
	service.ResourceBase = strings.Replace(service.ResourceBase, "v1.0/", "v2/", 1)
	return service, nil
}

func (c *Config) TmsV1Client() (*golangsdk.ServiceClient, error) {
	service, err := c.IdentityV3Client()
	if err != nil {
		return nil, err
	}
	service.Endpoint = strings.Replace(service.Endpoint, "v3/", "v1.0/", 1)
	service.Endpoint = strings.Replace(service.Endpoint, "iam", "tms", 1)
	return service, nil
}

func (c *Config) EvpnV5Client(region string) (*golangsdk.ServiceClient, error) {
	return openstack.NewEVPNServiceV3(c.HwClient, golangsdk.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func reconfigProjectName(src Config, projectName ProjectName) (*Config, error) {
	config := &Config{}
	if err := copier.Copy(config, &src); err != nil {
		return nil, err
	}
	config.TenantName = string(projectName)
	if err := config.LoadAndValidate(); err != nil {
		return nil, err
	}
	return config, nil
}

type SchemaOrDiff interface {
	GetOk(key string) (interface{}, bool)
	Get(key string) interface{}
}

// GetRegion returns the region that was specified in the resource. If a
// region was not set, the provider-level region is checked. The provider-level
// region can either be set by the region argument or by OS_REGION_NAME.
func (c *Config) GetRegion(d SchemaOrDiff) string {
	if d != nil {
		if v, ok := d.GetOk("region"); ok {
			return v.(string)
		}
	}
	if v := c.Region; v != "" {
		return v
	}
	tenantName := string(c.GetProjectName(d))
	if region := strings.Split(tenantName, "_")[0]; region != "" {
		return region
	}

	return strings.Split(c.IdentityEndpoint, ".")[1]
}

type ProjectName string

// GetProjectName returns the project name that was specified in the resource.
func (c *Config) GetProjectName(d SchemaOrDiff) ProjectName {
	if d != nil {
		if v, ok := d.GetOk("project_name"); ok {
			return ProjectName(v.(string))
		}
	}
	tenantName := c.TenantName
	if tenantName == "" {
		tenantName = c.DelegatedProject
	}
	return ProjectName(tenantName)
}

func SetOptionalEndpoint(cfg *aws.Config) string {
	endpoint := os.Getenv("AWS_METADATA_URL")
	if endpoint != "" {
		log.Printf("[INFO] Setting custom metadata endpoint: %q", endpoint)
		cfg.Endpoint = aws.String(endpoint)
		return endpoint
	}
	return ""
}
