package api

import "time"

// ResultVO is the generic API response wrapper.
type ResultVO[T any] struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Data      T      `json:"data"`
	Timestamp string `json:"timestamp"`
}

// PageDTO is a paginated list response.
type PageDTO[T any] struct {
	Total int64 `json:"total"`
	List  []T   `json:"list"`
}

// --- Auth ---

// LoginRequest is the request body for login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// --- User ---

// UserProfile represents a user's profile.
type UserProfile struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Role        int    `json:"role"`
	Avatar      string `json:"avatar,omitempty"`
}

// UpdateProfileRequest for PATCH /api/v1/user/profile
type UpdateProfileRequest struct {
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email,omitempty"`
	OldPassword string `json:"oldPassword,omitempty"`
	NewPassword string `json:"newPassword,omitempty"`
}

// Session represents a user session.
type Session struct {
	UUID        string `json:"uuid"`
	UserAgent   string `json:"userAgent"`
	IP          string `json:"ip"`
	LoginAt     string `json:"loginAt"`
	ExpireAt    string `json:"expireAt"`
	CurrentFlag bool   `json:"currentFlag"`
}

// --- Certificates ---

// CACert represents a CA certificate.
type CACert struct {
	UUID        string    `json:"uuid"`
	Algorithm   string    `json:"algorithm"`
	KeySize     int       `json:"keySize"`
	CN          string    `json:"cn"`
	Country     string    `json:"country"`
	Province    string    `json:"province"`
	City        string    `json:"city"`
	Org         string    `json:"org"`
	OrgUnit     string    `json:"orgUnit"`
	Email       string    `json:"email"`
	NotBefore   time.Time `json:"notBefore"`
	NotAfter    time.Time `json:"notAfter"`
	IsCA        bool      `json:"isCA"`
	ParentCA    string    `json:"parentCa,omitempty"`
	Available   bool      `json:"available"`
	Comment     string    `json:"comment,omitempty"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// SSLCert represents an SSL certificate.
type SSLCert struct {
	UUID      string    `json:"uuid"`
	Algorithm string    `json:"algorithm"`
	KeySize   int       `json:"keySize"`
	CN        string    `json:"cn"`
	Country   string    `json:"country"`
	Province  string    `json:"province"`
	City      string    `json:"city"`
	Org       string    `json:"org"`
	OrgUnit   string    `json:"orgUnit"`
	Email     string    `json:"email"`
	SANs      []string  `json:"sans,omitempty"`
	NotBefore time.Time `json:"notBefore"`
	NotAfter  time.Time `json:"notAfter"`
	CA        string    `json:"ca"`
	Comment   string    `json:"comment,omitempty"`
	Owner     string    `json:"owner"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CertContent holds PEM-encoded certificate content.
type CertContent struct {
	Certificate string `json:"certificate"`
	Chain       string `json:"chain,omitempty"`
}

// PrivKeyResponse holds an encrypted private key.
type PrivKeyResponse struct {
	PrivateKey string `json:"privateKey"`
}

// RequestSSLCertRequest is the request to issue a new SSL cert.
type RequestSSLCertRequest struct {
	Algorithm string   `json:"algorithm"`
	KeySize   int      `json:"keySize"`
	CN        string   `json:"cn"`
	Country   string   `json:"country,omitempty"`
	Province  string   `json:"province,omitempty"`
	City      string   `json:"city,omitempty"`
	Org       string   `json:"org,omitempty"`
	OrgUnit   string   `json:"orgUnit,omitempty"`
	Email     string   `json:"email,omitempty"`
	SANs      []string `json:"sans,omitempty"`
	CAUuid    string   `json:"caUuid"`
	ExpireDays int     `json:"expireDays"`
	Comment   string   `json:"comment,omitempty"`
	Password  string   `json:"password"`
}

// RenewSSLCertRequest is the request to renew an SSL cert.
type RenewSSLCertRequest struct {
	ExpireDays int    `json:"expireDays"`
	Password   string `json:"password"`
}

// UpdateCommentRequest updates a comment.
type UpdateCommentRequest struct {
	Comment string `json:"comment"`
}

// GetPrivKeyRequest requests a private key.
type GetPrivKeyRequest struct {
	Password string `json:"password"`
}

// AnalyzeCertRequest is the request to analyze a certificate.
type AnalyzeCertRequest struct {
	Certificate string `json:"certificate"`
}

// AnalyzePrivKeyRequest is the request to analyze a private key.
type AnalyzePrivKeyRequest struct {
	PrivateKey string `json:"privateKey"`
	Password   string `json:"password,omitempty"`
}

// CertAnalysis holds the result of a certificate analysis.
type CertAnalysis struct {
	Subject     map[string]string `json:"subject"`
	Issuer      map[string]string `json:"issuer"`
	SANs        []string          `json:"sans,omitempty"`
	NotBefore   string            `json:"notBefore"`
	NotAfter    string            `json:"notAfter"`
	Algorithm   string            `json:"algorithm"`
	KeySize     int               `json:"keySize"`
	IsCA        bool              `json:"isCA"`
	Fingerprint string            `json:"fingerprint"`
}

// PrivKeyAnalysis holds the result of a private key analysis.
type PrivKeyAnalysis struct {
	Algorithm string `json:"algorithm"`
	KeySize   int    `json:"keySize"`
}

// ConvertPEMtoPFXRequest converts PEM to PFX.
type ConvertPEMtoPFXRequest struct {
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"privateKey"`
	Password    string `json:"password"`
}

// ConvertPEMtoDERRequest converts PEM to DER.
type ConvertPEMtoDERRequest struct {
	Certificate string `json:"certificate"`
}

// ConvertDERtoPEMRequest converts DER to PEM (base64-encoded DER).
type ConvertDERtoPEMRequest struct {
	Certificate string `json:"certificate"`
}

// ConvertResult holds a converted certificate.
type ConvertResult struct {
	Data string `json:"data"`
}

// --- Admin ---

// AdminUser represents a user in the admin view.
type AdminUser struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Role        int    `json:"role"`
	Enabled     bool   `json:"enabled"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// RequestCACertRequest is the request to create a CA certificate.
type RequestCACertRequest struct {
	Algorithm  string `json:"algorithm"`
	KeySize    int    `json:"keySize"`
	CN         string `json:"cn"`
	Country    string `json:"country,omitempty"`
	Province   string `json:"province,omitempty"`
	City       string `json:"city,omitempty"`
	Org        string `json:"org,omitempty"`
	OrgUnit    string `json:"orgUnit,omitempty"`
	Email      string `json:"email,omitempty"`
	ExpireDays int    `json:"expireDays"`
	ParentCA   string `json:"parentCa,omitempty"`
	Comment    string `json:"comment,omitempty"`
	Password   string `json:"password"`
}

// RenewCACertRequest is the request to renew a CA certificate.
type RenewCACertRequest struct {
	ExpireDays int    `json:"expireDays"`
	Password   string `json:"password"`
}

// ImportCACertRequest is the request to import a CA certificate.
type ImportCACertRequest struct {
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"privateKey"`
	Password    string `json:"password"`
	Comment     string `json:"comment,omitempty"`
}

// ToggleAvailableRequest toggles CA availability.
type ToggleAvailableRequest struct {
	Available bool `json:"available"`
}

// BindUsersRequest binds users to a CA.
type BindUsersRequest struct {
	Usernames []string `json:"usernames"`
}

// --- Superadmin ---

// AllSession represents a session in the superadmin view.
type AllSession struct {
	UUID      string `json:"uuid"`
	Username  string `json:"username"`
	UserAgent string `json:"userAgent"`
	IP        string `json:"ip"`
	LoginAt   string `json:"loginAt"`
	ExpireAt  string `json:"expireAt"`
}

// CreateUserRequest creates a new user.
type CreateUserRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        int    `json:"role"`
}

// BatchCreateUsersRequest creates multiple users.
type BatchCreateUsersRequest struct {
	Users []CreateUserRequest `json:"users"`
}

// BatchDeleteUsersRequest deletes multiple users.
type BatchDeleteUsersRequest struct {
	Usernames []string `json:"usernames"`
}

// UpdateUserRequest updates user info.
type UpdateUserRequest struct {
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email,omitempty"`
	Password    string `json:"password,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

// UpdateUserRoleRequest updates a user's role.
type UpdateUserRoleRequest struct {
	Username string `json:"username"`
	Role     int    `json:"role"`
}

// RoleName returns a human-readable role name.
func RoleName(role int) string {
	switch role {
	case 0:
		return "User"
	case 1:
		return "Admin"
	case 2:
		return "Superadmin"
	default:
		return "Unknown"
	}
}
