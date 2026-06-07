package resolveprincipal

type Principal struct {
	UserID             string
	TenantID           string
	Email              string
	DisplayName        string
	Role               string
	MustChangePassword bool
	ViaAPIToken        bool
	APITokenID         string
	Scopes             []string
}
