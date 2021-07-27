package horizon

// generated from https://code-stg.vmware.com/apis/1169/view-rest-api

// AuthLoginRequest is used to perform a full authentication against the Horizon
// API server
type AuthLoginRequest struct {
	// Domain
	Domain string `json:"domain"`
	// User Name
	Username string `json:"username"`
	// User password
	Password string `json:"password"`
}

// RefreshTokenRequest is used to get a new Access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AccessToken contains the new access token returned from a successful token
// refresh
type AccessToken struct {
	// Access Token to be used in API calls.
	AccessToken string `json:"access_token"`
}

// AuthTokens contains authentication details with access and refresh token
type AuthTokens struct {
	// Access Token to be used in API calls.
	AccessToken string `json:"access_token"`
	// Refresh Token to be used to get a new Access token.
	RefreshToken string `json:"refresh_token"`
}

// AuditEventAttributeInfo contains extended event attribute information
type AuditEventAttributeInfo struct {
	// Key value pairs representing Extended attributes related to the event.
	EventData map[string]interface{} `json:"event_data,omitempty"`
	// Unique id representing an event.
	ID int64 `json:"id,omitempty"`
}

// AuditEventSummary contains information about audit events
type AuditEventSummary struct {
	// Application Pool associated with this event. Will be unset if there is no
	// application association for this event. Supported Filters : 'Equals'.
	ApplicationPoolName string `json:"application_pool_name,omitempty"`
	// Desktop Pool associated with this event. Will be unset if there is no desktop
	// association for this event. Supported Filters : 'Equals'.
	DesktopPoolName string `json:"desktop_pool_name,omitempty"`
	// Unique id representing an event. Supported Filters : 'Equals'.
	ID int64 `json:"id,omitempty"`
	// FQDN of the machine in the Pod that has logged this event. Supported Filters
	// : 'Equals'.
	MachineDNSName string `json:"machine_dns_name,omitempty"`
	// Machine associated with this event. Will be unset if there is no machine
	// association for this event. Supported Filters : 'Equals'.
	MachineID string `json:"machine_id,omitempty"`
	// Audit event message.
	Message string `json:"message,omitempty"`
	// Horizon component that has logged this event. Supported Filters : 'Equals'.
	Module string `json:"module,omitempty"`
	// Severity type of the event. Supported Filters : 'Equals'. * INFO: Audit event
	// is of INFO severity. * WARNING: Audit event is of WARNING severity * ERROR:
	// Audit event is of ERROR severity * AUDIT_SUCCESS: Audit event is of
	// AUDIT_SUCCESS severity * AUDIT_FAIL: Audit event is of AUDIT_FAIL severity *
	// UNKNOWN: Not able to identify severity
	Severity string `json:"severity,omitempty"`
	// Time at which the event occurred. Supported Filters : 'Equals'.
	Time int64 `json:"time,omitempty"`
	// Event name that corresponds to an item in the message catalog. Supported
	// Filters : 'Equals'.
	Type string `json:"type,omitempty"`
	// Sid of the user associated with this event. Supported Filters : 'Equals'.
	UserID string `json:"user_id,omitempty"`
}

// BetweenFilter is a range filter. It can be used to filter on int64
// timestamps.
type BetweenFilter struct {
	Type      string      `json:"type,omitempty"`
	FromValue interface{} `json:"fromValue,omitempty"`
	Name      string      `json:"name,omitempty"`
	ToValue   interface{} `json:"toValue,omitempty"`
}

// Timestamp is time since unix epoch (UTC) in milliseconds (as defined by
// Horizon spec)
type Timestamp int64
