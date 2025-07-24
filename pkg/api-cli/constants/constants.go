// Package constants provides shared constants for the Virtuoso API CLI
package constants

import "time"

// HTTP and Authentication constants
const (
	// AuthorizationHeaderPrefix is the prefix for authorization header
	AuthorizationHeaderPrefix = "Bearer "

	// ContentTypeJSON is the content type for JSON requests
	ContentTypeJSON = "application/json"

	// HeaderContentType is the content type header name
	HeaderContentType = "Content-Type"

	// HeaderAuthorization is the authorization header name
	HeaderAuthorization = "Authorization"

	// HeaderXAppVersion is the custom app version header
	HeaderXAppVersion = "X-App-Version"
)

// Timeout constants
const (
	// DefaultTimeoutMs is the default timeout in milliseconds for wait operations
	DefaultTimeoutMs = 30000 // 30 seconds

	// DefaultHTTPTimeout is the default timeout for HTTP requests
	DefaultHTTPTimeout = 30 * time.Second

	// DefaultRetryCount is the default number of retries
	DefaultRetryCount = 3

	// DefaultRetryWait is the default wait time between retries
	DefaultRetryWait = 1 * time.Second
)

// Conversion constants
const (
	// MillisecondsPerSecond is used for time conversions
	MillisecondsPerSecond = 1000
)

// Default values
const (
	// DefaultOrganizationID is the default organization ID
	DefaultOrganizationID = "2242"

	// DefaultBaseURL is the default API base URL
	DefaultBaseURL = "https://api-app2.virtuoso.qa/api"

	// DefaultClientID is the default client ID header value
	DefaultClientID = "api-cli-generator"

	// DefaultClientName is the default client name header value
	DefaultClientName = "api-cli-generator"
)

// Business rule constants
const (
	// DefaultInitialCheckpointName is the default name for initial checkpoints
	DefaultInitialCheckpointName = "INITIAL_CHECKPOINT"

	// DefaultMaxStepsPerCheckpoint is the default maximum steps per checkpoint
	DefaultMaxStepsPerCheckpoint = 20
)

// AI and test configuration defaults
const (
	// DefaultBatchDir is the default directory for test batches
	DefaultBatchDir = "./test-batches"

	// DefaultTemplateDir is the default directory for templates
	DefaultTemplateDir = "./examples"

	// DefaultContextDepth is the default context depth for AI responses
	DefaultContextDepth = 3
)
