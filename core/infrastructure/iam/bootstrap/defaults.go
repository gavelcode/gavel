// Package bootstrap holds the well-known identity of the default tenant and its
// first administrator. First-boot seeding (gavel-server) and the test kit both
// seed this exact identity, and login-flow tests authenticate against it, so the
// values must be defined once — a divergence between the two would let tests
// pass against an admin production never creates.
package bootstrap

const (
	DefaultTenantSlug        = "default"
	DefaultTenantDisplayName = "Default"
	DefaultAdminEmail        = "admin@gavel.local"
	DefaultAdminDisplayName  = "Administrator"
)
