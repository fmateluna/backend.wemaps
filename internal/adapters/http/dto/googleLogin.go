package dto

type LoginGoogle struct {
	User          User          `json:"user,omitempty"`
	ProviderID    string        `json:"providerId,omitempty"`
	TokenResponse TokenResponse `json:"_tokenResponse,omitempty"`
	OperationType string        `json:"operationType,omitempty"`
}
type ProviderData struct {
	ProviderID  string `json:"providerId,omitempty"`
	UID         string `json:"uid,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email,omitempty"`
	PhoneNumber any    `json:"phoneNumber,omitempty"`
	PhotoURL    string `json:"photoURL,omitempty"`
}
type StsTokenManager struct {
	RefreshToken   string `json:"refreshToken,omitempty"`
	AccessToken    string `json:"accessToken,omitempty"`
	ExpirationTime int64  `json:"expirationTime,omitempty"`
}
type User struct {
	UID             string          `json:"uid,omitempty"`
	Email           string          `json:"email,omitempty"`
	EmailVerified   bool            `json:"emailVerified,omitempty"`
	DisplayName     string          `json:"displayName,omitempty"`
	IsAnonymous     bool            `json:"isAnonymous,omitempty"`
	PhotoURL        string          `json:"photoURL,omitempty"`
	ProviderData    []ProviderData  `json:"providerData,omitempty"`
	StsTokenManager StsTokenManager `json:"stsTokenManager,omitempty"`
	CreatedAt       string          `json:"createdAt,omitempty"`
	LastLoginAt     string          `json:"lastLoginAt,omitempty"`
	APIKey          string          `json:"apiKey,omitempty"`
	AppName         string          `json:"appName,omitempty"`
}
type TokenResponse struct {
	FederatedID      string `json:"federatedId,omitempty"`
	ProviderID       string `json:"providerId,omitempty"`
	Email            string `json:"email,omitempty"`
	EmailVerified    bool   `json:"emailVerified,omitempty"`
	FirstName        string `json:"firstName,omitempty"`
	FullName         string `json:"fullName,omitempty"`
	LastName         string `json:"lastName,omitempty"`
	PhotoURL         string `json:"photoUrl,omitempty"`
	LocalID          string `json:"localId,omitempty"`
	DisplayName      string `json:"displayName,omitempty"`
	IDToken          string `json:"idToken,omitempty"`
	Context          string `json:"context,omitempty"`
	OauthAccessToken string `json:"oauthAccessToken,omitempty"`
	OauthExpireIn    int    `json:"oauthExpireIn,omitempty"`
	RefreshToken     string `json:"refreshToken,omitempty"`
	ExpiresIn        string `json:"expiresIn,omitempty"`
	OauthIDToken     string `json:"oauthIdToken,omitempty"`
	RawUserInfo      string `json:"rawUserInfo,omitempty"`
	Kind             string `json:"kind,omitempty"`
}
