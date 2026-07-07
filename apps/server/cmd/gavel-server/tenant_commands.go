package main

import (
	"context"
	"crypto/rand"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/usegavel/gavel/apps/server/internal/platform/config"
	"github.com/usegavel/gavel/apps/server/internal/platform/firstadmin"
	tenantactivate "github.com/usegavel/gavel/core/application/iam/tenant/activate"
	tenantprovision "github.com/usegavel/gavel/core/application/iam/tenant/provision"
	tenantsuspend "github.com/usegavel/gavel/core/application/iam/tenant/suspend"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/infrastructure/iam/argon2"
	"github.com/usegavel/gavel/core/infrastructure/iam/bootstrap"
	pgiam "github.com/usegavel/gavel/core/infrastructure/iam/postgres"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

// tenantCmd groups the operator-only tenant lifecycle commands. Provisioning and
// suspending/activating a tenant crosses the tenant boundary, so it belongs to
// whoever operates the host — the same privilege as serve/migrate — not to any
// in-tenant admin or an HTTP endpoint.
func tenantCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "tenant",
		Short: "Manage tenants (operator-only)",
	}
	for _, sub := range []*cobra.Command{provisionTenantCmd(), suspendTenantCmd(), activateTenantCmd()} {
		// A runtime failure (e.g. a taken slug) is not a usage error, so don't
		// dump the flags help on top of it — the error message stands on its own.
		sub.SilenceUsage = true
		command.AddCommand(sub)
	}
	return command
}

func provisionTenantCmd() *cobra.Command {
	var tenantSlug, displayName, adminEmail, adminName, adminPassword string
	command := &cobra.Command{
		Use:   "provision",
		Short: "Provision a new tenant together with its first administrator",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dbConn, logger, err := openOperatorDB(cmd.Context())
			if err != nil {
				return err
			}
			defer func() { _ = dbConn.Close() }()

			password, generated, err := firstadmin.ResolvePassword(adminPassword, rand.Reader)
			if err != nil {
				return err
			}

			handler := tenantprovision.NewHandler(
				pgiam.NewTenantProvisioner(dbConn), argon2.New(rand.Reader))
			provisionCmd, err := tenantprovision.NewCommand(
				tenantSlug, displayName, adminEmail, adminName, password, time.Now().UTC())
			if err != nil {
				return err
			}
			result, err := handler.Execute(cmd.Context(), provisionCmd)
			if err != nil {
				return err
			}

			if generated {
				logger.Warn("generated a one-time admin password; change it after first login",
					"admin_password", password)
			}
			logger.Info("tenant provisioned",
				"slug", tenantSlug, "tenant_id", result.TenantID, "admin_user_id", result.AdminUserID)
			return nil
		},
	}
	command.Flags().StringVar(&tenantSlug, "slug", "", "unique slug for the tenant (required)")
	command.Flags().StringVar(&displayName, "display-name", "", "human-readable tenant name (required)")
	command.Flags().StringVar(&adminEmail, "admin-email", "", "email of the first administrator (required)")
	command.Flags().StringVar(&adminName, "admin-name", bootstrap.DefaultAdminDisplayName, "display name of the first administrator")
	command.Flags().StringVar(&adminPassword, "admin-password", "", "admin password; generated and logged once if unset")
	mustMarkRequired(command, "slug", "display-name", "admin-email")
	return command
}

func suspendTenantCmd() *cobra.Command {
	var tenantSlug string
	command := &cobra.Command{
		Use:   "suspend",
		Short: "Suspend a tenant, blocking authentication for its users",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dbConn, logger, err := openOperatorDB(cmd.Context())
			if err != nil {
				return err
			}
			defer func() { _ = dbConn.Close() }()

			tenants := pgiam.NewTenantRepo(dbConn)
			found, err := tenantBySlug(cmd.Context(), tenants, tenantSlug)
			if err != nil {
				return err
			}
			suspendCmd, err := tenantsuspend.NewCommand(found.ID().String(), time.Now().UTC())
			if err != nil {
				return err
			}
			result, err := tenantsuspend.NewHandler(tenants).Execute(cmd.Context(), suspendCmd)
			if err != nil {
				return err
			}
			logger.Info("tenant suspended", "slug", tenantSlug, "tenant_id", result.TenantID)
			return nil
		},
	}
	command.Flags().StringVar(&tenantSlug, "slug", "", "slug of the tenant (required)")
	mustMarkRequired(command, "slug")
	return command
}

func activateTenantCmd() *cobra.Command {
	var tenantSlug string
	command := &cobra.Command{
		Use:   "activate",
		Short: "Reactivate a suspended tenant",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dbConn, logger, err := openOperatorDB(cmd.Context())
			if err != nil {
				return err
			}
			defer func() { _ = dbConn.Close() }()

			tenants := pgiam.NewTenantRepo(dbConn)
			found, err := tenantBySlug(cmd.Context(), tenants, tenantSlug)
			if err != nil {
				return err
			}
			activateCmd, err := tenantactivate.NewCommand(found.ID().String(), time.Now().UTC())
			if err != nil {
				return err
			}
			result, err := tenantactivate.NewHandler(tenants).Execute(cmd.Context(), activateCmd)
			if err != nil {
				return err
			}
			logger.Info("tenant activated", "slug", tenantSlug, "tenant_id", result.TenantID)
			return nil
		},
	}
	command.Flags().StringVar(&tenantSlug, "slug", "", "slug of the tenant (required)")
	mustMarkRequired(command, "slug")
	return command
}

// openOperatorDB opens and migrates the database for an operator command, the
// same as serve, so `tenant provision` against a never-migrated database gets
// the schema applied instead of a raw "relation does not exist" error.
func openOperatorDB(ctx context.Context) (*database.DB, *slog.Logger, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	dbConn, err := openAndMigrateDB(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}
	return dbConn, slog.New(slog.NewTextHandler(os.Stderr, nil)), nil
}

func tenantBySlug(ctx context.Context, tenants *pgiam.TenantRepo, slugRaw string) (tenant.Tenant, error) {
	slugVO, err := tenant.NewSlug(slugRaw)
	if err != nil {
		return tenant.Tenant{}, err
	}
	return tenants.BySlug(ctx, slugVO)
}

func mustMarkRequired(command *cobra.Command, names ...string) {
	for _, flagName := range names {
		_ = command.MarkFlagRequired(flagName)
	}
}
