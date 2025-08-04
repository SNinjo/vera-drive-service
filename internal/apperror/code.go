package apperror

// the error code format is <status>_<service_id>_<serial_number>
// Service ID: 02 - drive service

const (
	// middleware
	CodeIdentityServiceUnavailable = "401_02_001"
	CodeInvalidClaimsInUserToken   = "401_02_002"

	// url package
	CodeURLNotFound          = "404_02_003"
	CodeURLAccessDenied      = "403_02_004"
	CodeURLNameAlreadyExists = "400_02_005"
)
