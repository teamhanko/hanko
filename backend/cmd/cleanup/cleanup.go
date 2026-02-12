package cleanup

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"log"
	"sort"
	"strings"
	"time"
)

// options holds user-provided CLI options
type options struct {
	tables     []string // List of tables to clean up
	configFile string   // Path to configuration file
	pageSize   int      // The number of entities to query at once
	run        bool     // Whether to execute cleanup or simulate
}

// handlerParam holds the necessary parameters for cleanup operations
type handlerParam struct {
	table   string
	config  *config.Config
	storage persistence.Storage
	options *options
}

// handlerFunc defines the function signature for cleanup handlers
type handlerFunc func(handlerParam) error

// Table names used for cleanup operations
const (
	tableAuditLogs           = "audit_logs"
	tableFlows               = "flows"
	tableWebauthnSessionData = "webauthn_session_data"
)

// Map of table names to their respective cleanup handlers
var handler = map[string]handlerFunc{
	tableFlows: func(param handlerParam) error {
		return cleanup[models.Flow](param, param.storage.GetFlowPersister(), time.Now().UTC())
	},
	tableAuditLogs: func(param handlerParam) error {
		duration, err := time.ParseDuration(param.config.AuditLog.Retention)
		if err != nil {
			return fmt.Errorf("failed to parse the retention duration: %w", err)
		}

		return cleanup[models.AuditLog](param, param.storage.GetAuditLogPersister(), time.Now().Add(-duration).UTC())
	},
	tableWebauthnSessionData: func(param handlerParam) error {
		return cleanup[models.WebauthnSessionData](param, param.storage.GetWebauthnSessionDataPersister(), time.Now().UTC())
	},
}

// allowedTables is a list of table names that can be cleaned up
var allowedTables = func() []string {
	keys := make([]string, 0, len(handler))
	for key := range handler {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}()

// isTableAllowed checks if a given table name exists in the allowed list
func isTableAllowed(table string) bool {
	for _, allowed := range allowedTables {
		if table == allowed {
			return true
		}
	}
	return false
}

// validateTables checks if the specified table names exist in the allowed list
func validateTables(tables []string) error {
	var invalidTables []string

	for _, table := range tables {
		if !isTableAllowed(table) {
			invalidTables = append(invalidTables, table)
		}
	}

	if len(invalidTables) > 0 {
		return fmt.Errorf("invalid table name(s): %s - allowed values: %s",
			strings.Join(invalidTables, ", "), strings.Join(allowedTables, ", "))
	}

	return nil
}

// newCleanupCommand creates the Cobra command for database cleanup
func newCleanupCommand() *cobra.Command {
	opts := &options{}

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Cleanup the database.",
		Long:  `Cleans up the database by deleting expired entities.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(opts.tables) == 0 {
				opts.tables = allowedTables
				return nil
			}

			return validateTables(opts.tables)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(&opts.configFile)
			if err != nil {
				log.Fatal(err)
			}

			storage, err := persistence.New(cfg.Database)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Cleaning up table(s): %s...\n", strings.Join(opts.tables, ", "))

			for _, table := range opts.tables {
				param := handlerParam{
					table:   table,
					config:  cfg,
					storage: storage,
					options: opts,
				}
				err = handler[table](param)
				if err != nil {
					log.Fatal(err)
				}
			}

			log.Println("Cleanup completed.")

			if !opts.run {
				log.Println("This was a dry-run; add --run to the command to really delete the data.")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.configFile, "config", "c", config.DefaultConfigFilePath, "path to config file")
	cmd.Flags().StringSliceVarP(&opts.tables, "tables", "t", []string{}, fmt.Sprintf("specify individual tables to clean up (comma-separated) - allowed values: %s", strings.Join(allowedTables, ", ")))
	cmd.Flags().IntVarP(&opts.pageSize, "page-size", "s", 512, "the number of entities to query at once")
	cmd.Flags().BoolVar(&opts.run, "run", false, "execute the cleanup process instead of simulating")

	return cmd
}

// cleanup performs the cleanup operation for a given table and persister
func cleanup[T any](param handlerParam, persister persistence.Cleanup[T], cutoffTime time.Time) error {
	var (
		page    = 1
		deleted = 0
	)

	for {
		items, err := persister.FindExpired(cutoffTime, page, param.options.pageSize)
		if err != nil {
			return err
		}

		if len(items) > 0 {
			for _, item := range items {
				if param.options.run {
					err = persister.Delete(item)
					if err != nil {
						return err
					}
				}

				deleted++
			}

			log.Printf("Deleted %d %s in total.", deleted, param.table)

			if !param.options.run {
				page++
			}
		}

		if len(items) < param.options.pageSize {
			break
		}
	}

	return nil
}

// RegisterCommands registers the cleanup command with the parent command
func RegisterCommands(parent *cobra.Command) {
	cmd := newCleanupCommand()
	parent.AddCommand(cmd)
}
