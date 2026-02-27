package api

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
}

// UpdateProfileRequest for PATCH /api/v1/user/profile
type UpdateProfileRequest struct {
DisplayName string `json:"displayName,omitempty"`
Email       string `json:"email,omitempty"`
OldPassword string `json:"oldPassword,omitempty"`
NewPassword string `json:"newPassword,omitempty"`
}

// LoginRecord represents a user login/session record (LoginRecordDTO).
type LoginRecord struct {
UUID      string `json:"uuid"`
Username  string `json:"username"`
IPAddress string `json:"ipAddress"`
Region    string `json:"region"`
City      string `json:"city"`
Browser   string `json:"browser"`
OS        string `json:"os"`
LoginTime string `json:"loginTime"`
IsOnline  bool   `json:"isOnline"`
}

// --- Certificates ---

// CACert represents a CA certificate info DTO (CaInfoDTO from API).
type CACert struct {
UUID       string `json:"uuid"`
Owner      string `json:"owner"`
AllowSubCa bool   `json:"allowSubCa"`
ParentCa   string `json:"parentCa"`
Comment    string `json:"comment"`
Available  bool   `json:"available"`
NotBefore  string `json:"notBefore"`
NotAfter   string `json:"notAfter"`
}

// CAType returns the certificate type: "Root CA", "Int CA", or "Leaf CA".
func (c *CACert) CAType() string {
if c.ParentCa == "" {
return "Root CA"
} else if c.AllowSubCa {
return "Int CA"
}
return "Leaf CA"
}

// SSLCert represents an SSL certificate info DTO (CertInfoDTO from API).
type SSLCert struct {
UUID       string `json:"uuid"`
CaUUID     string `json:"caUuid"`
Owner      string `json:"owner"`
Comment    string `json:"comment"`
NotBefore  string `json:"notBefore"`
NotAfter   string `json:"notAfter"`
CreatedAt  string `json:"createdAt"`
ModifiedAt string `json:"modifiedAt"`
}

// PrivKeyResponse holds an encrypted private key.
type PrivKeyResponse struct {
PrivateKey string `json:"privateKey"`
}

// RequestSSLCertRequest is the request to issue a new SSL cert (matches API DTO).
type RequestSSLCertRequest struct {
CaUUID             string           `json:"caUuid"`
Algorithm          string           `json:"algorithm,omitempty"`
KeySize            int              `json:"keySize,omitempty"`
Country            string           `json:"country"`
Province           string           `json:"province"`
City               string           `json:"city"`
Organization       string           `json:"organization"`
OrganizationalUnit string           `json:"organizationalUnit"`
CommonName         string           `json:"commonName"`
Expiry             int              `json:"expiry"`
SubjectAltNames    []SubjectAltName `json:"subjectAltNames,omitempty"`
Comment            string           `json:"comment,omitempty"`
}

// SubjectAltName represents a SAN entry.
type SubjectAltName struct {
Type  string `json:"type"`
Value string `json:"value"`
}

// RenewSSLCertRequest is the request to renew an SSL cert.
type RenewSSLCertRequest struct {
Expiry int `json:"expiry"`
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
Cert string `json:"cert"`
}

// AnalyzePrivKeyRequest is the request to analyze a private key.
type AnalyzePrivKeyRequest struct {
PrivKey  string `json:"privkey"`
Password string `json:"password,omitempty"`
}

// CertAnalysis holds the result of a certificate analysis.
// The fields are a superset of all possible API response fields.
type CertAnalysis struct {
Subject      string                 `json:"subject,omitempty"`
Issuer       string                 `json:"issuer,omitempty"`
NotBefore    string                 `json:"notBefore"`
NotAfter     string                 `json:"notAfter"`
SerialNumber string                 `json:"serialNumber,omitempty"`
Algorithm    string                 `json:"algorithm"`
IsCA         bool                   `json:"isCA"`
Fingerprint  string                 `json:"fingerprint,omitempty"`
PublicKey    map[string]interface{} `json:"publicKey,omitempty"`
Extensions   map[string]string      `json:"extensions,omitempty"`
SANs         []string               `json:"subjectAltNames,omitempty"`
}

// PrivKeyAnalysis holds the result of a private key analysis.
type PrivKeyAnalysis struct {
Algorithm string `json:"algorithm"`
KeySize   int    `json:"keySize"`
}

// ConvertPEMtoPFXRequest converts PEM to PFX.
type ConvertPEMtoPFXRequest struct {
Cert     string `json:"cert"`
PrivKey  string `json:"privkey"`
Password string `json:"password"`
}

// ConvertRequest for PEM↔DER conversions.
type ConvertRequest struct {
Cert    string `json:"cert,omitempty"`
PrivKey string `json:"privkey,omitempty"`
}

// ConvertResult holds a converted certificate.
type ConvertResult struct {
Data string `json:"data"`
}

// --- Admin ---

// AdminUser represents a user in the admin view (UserProfileDTO from API).
type AdminUser struct {
Username    string `json:"username"`
DisplayName string `json:"displayName"`
Email       string `json:"email"`
Role        int    `json:"role"`
}

// RequestCACertRequest is the request to create a CA certificate.
type RequestCACertRequest struct {
CaUUID             string `json:"caUuid,omitempty"`
AllowSubCa         bool   `json:"allowSubCa"`
Algorithm          string `json:"algorithm,omitempty"`
KeySize            int    `json:"keySize,omitempty"`
Country            string `json:"country"`
Province           string `json:"province"`
City               string `json:"city"`
Organization       string `json:"organization"`
OrganizationalUnit string `json:"organizationalUnit"`
CommonName         string `json:"commonName"`
Expiry             int    `json:"expiry"`
Comment            string `json:"comment,omitempty"`
}

// RenewCACertRequest is the request to renew a CA certificate.
type RenewCACertRequest struct {
Expiry int `json:"expiry"`
}

// ImportCACertRequest is the request to import a CA certificate.
type ImportCACertRequest struct {
Certificate string `json:"certificate"`
PrivKey     string `json:"privkey"`
Comment     string `json:"comment,omitempty"`
}

// CABindingDTO represents a CA-User binding.
type CABindingDTO struct {
CaUUID   string `json:"caUuid"`
Username string `json:"username"`
}

// --- Superadmin ---

// CreateUserRequest creates a new user.
type CreateUserRequest struct {
Username    string `json:"username"`
DisplayName string `json:"displayName"`
Email       string `json:"email"`
Password    string `json:"password"`
Role        int    `json:"role"`
}

// UpdateUserRoleRequest updates a user's role.
type UpdateUserRoleRequest struct {
Username string `json:"username"`
Role     int    `json:"role"`
}

// UpdateSuperadminUserRequest updates user info (superadmin).
type UpdateSuperadminUserRequest struct {
DisplayName string `json:"displayName,omitempty"`
Email       string `json:"email,omitempty"`
Password    string `json:"password,omitempty"`
}

// RoleName returns a human-readable role name.
// API role values: 1=User, 2=Admin, 3=Superadmin.
func RoleName(role int) string {
switch role {
case 1:
return "User"
case 2:
return "Admin"
case 3:
return "Superadmin"
default:
return "Unknown"
}
}

// BindUsersRequest binds users to a CA.
type BindUsersRequest struct {
Usernames []string `json:"usernames"`
}

// ToggleAvailableRequest toggles CA availability.
type ToggleAvailableRequest struct {
Available bool `json:"available"`
}

// BatchDeleteUsersRequest deletes multiple users.
type BatchDeleteUsersRequest struct {
Usernames []string `json:"usernames"`
}
