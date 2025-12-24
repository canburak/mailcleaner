package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/mailcleaner/mailcleaner/internal/config"
	"github.com/mailcleaner/mailcleaner/internal/imap"
	"github.com/mailcleaner/mailcleaner/internal/scheduler"
)

var (
	configFile string
	dryRun     bool
	version    = "dev"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mailcleaner",
		Short: "Automate email organization with user-defined rules",
		Long: `MailCleaner connects to your email mailbox and executes user-defined rules
on a schedule. This allows you to automate email organization, cleanup,
and management tasks without manual intervention.`,
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "path to configuration file")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "show what would be done without making changes")

	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(daemonCmd())
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(listFoldersCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Execute all rules once",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile)
			if err != nil {
				return err
			}

			if dryRun {
				log.Println("DRY RUN MODE - no changes will be made")
			}

			sched := scheduler.New(cfg, dryRun)
			return sched.RunOnce()
		},
	}
}

func daemonCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "daemon",
		Short: "Run as a daemon with scheduled execution",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile)
			if err != nil {
				return err
			}

			if dryRun {
				log.Println("DRY RUN MODE - no changes will be made")
			}

			sched := scheduler.New(cfg, dryRun)

			// Handle graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigChan
				log.Println("Shutting down...")
				sched.Stop()
				os.Exit(0)
			}()

			log.Println("Starting MailCleaner daemon...")
			return sched.Start()
		},
	}
}

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile)
			if err != nil {
				return err
			}

			fmt.Println("Configuration is valid!")
			fmt.Printf("  Accounts: %d\n", len(cfg.Accounts))
			fmt.Printf("  Rules: %d\n", len(cfg.Rules))

			if cfg.Schedule.Cron != "" {
				fmt.Printf("  Schedule: cron(%s)\n", cfg.Schedule.Cron)
			} else {
				fmt.Printf("  Schedule: every %v\n", cfg.Schedule.GetInterval())
			}

			return nil
		},
	}
}

func listFoldersCmd() *cobra.Command {
	var accountName string

	cmd := &cobra.Command{
		Use:   "list-folders",
		Short: "List all folders in a mailbox",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile)
			if err != nil {
				return err
			}

			var account *config.Account
			if accountName != "" {
				account = cfg.GetAccount(accountName)
				if account == nil {
					return fmt.Errorf("account not found: %s", accountName)
				}
			} else if len(cfg.Accounts) > 0 {
				account = &cfg.Accounts[0]
			} else {
				return fmt.Errorf("no accounts configured")
			}

			client, err := imap.NewClient(account)
			if err != nil {
				return err
			}
			defer client.Close()

			folders, err := client.ListFolders()
			if err != nil {
				return err
			}

			fmt.Printf("Folders for %s:\n", account.Name)
			for _, folder := range folders {
				fmt.Printf("  - %s\n", folder)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&accountName, "account", "a", "", "account name (uses first account if not specified)")
	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("MailCleaner %s\n", version)
		},
	}
}
