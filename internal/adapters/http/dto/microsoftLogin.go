package dto

// AuthResponse represents the top-level JSON structure
type LoginMicrosoft struct {
	User          UserMS          `json:"user"`
	ProviderID    string          `json:"providerId"`
	TokenResponse TokenResponseMS `json:"_tokenResponse"`
	OperationType string          `json:"operationType"`
}

// User represents the user object in the JSON
type UserMS struct {
	UID             string          `json:"uid"`
	Email           string          `json:"email"`
	EmailVerified   bool            `json:"emailVerified"`
	DisplayName     string          `json:"displayName"`
	IsAnonymous     bool            `json:"isAnonymous"`
	ProviderDataMS  []ProviderData  `json:"providerData"`
	STSTokenManager STSTokenManager `json:"stsTokenManager"`
	CreatedAt       string          `json:"createdAt"`
	LastLoginAt     string          `json:"lastLoginAt"`
	APIKey          string          `json:"apiKey"`
	AppName         string          `json:"appName"`
}

// ProviderData represents the providerData array elements
type ProviderDataMS struct {
	ProviderID  string  `json:"providerId"`
	UID         string  `json:"uid"`
	DisplayName string  `json:"displayName"`
	Email       string  `json:"email"`
	PhoneNumber *string `json:"phoneNumber"`
	PhotoURL    *string `json:"photoURL"`
}

// STSTokenManager represents the stsTokenManager object
type STSTokenManager struct {
	RefreshToken   string `json:"refreshToken"`
	AccessToken    string `json:"accessToken"`
	ExpirationTime int64  `json:"expirationTime"`
}

// TokenResponse represents the _tokenResponse object
type TokenResponseMS struct {
	FederatedID      string `json:"federatedId"`
	ProviderID       string `json:"providerId"`
	Email            string `json:"email"`
	EmailVerified    bool   `json:"emailVerified"`
	FirstName        string `json:"firstName"`
	FullName         string `json:"fullName"`
	LastName         string `json:"lastName"`
	LocalID          string `json:"localId"`
	DisplayName      string `json:"displayName"`
	IDToken          string `json:"idToken"`
	Context          string `json:"context"`
	OAuthAccessToken string `json:"oauthAccessToken"`
	OAuthExpireIn    int    `json:"oauthExpireIn"`
	RefreshToken     string `json:"refreshToken"`
	ExpiresIn        string `json:"expiresIn"`
	OAuthIDToken     string `json:"oauthIdToken"`
	RawUserInfo      string `json:"rawUserInfo"`
	Kind             string `json:"kind"`
	PendingToken     string `json:"pendingToken"`
}
