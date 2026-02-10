package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "genomdb",
	Short: "Genomic datastore",
	Long: `GenomDB is a distributed storage system for SAM and BAM files.

Get started by running:
genomdb start configs/config-node1.yml`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("Error executing root command: %d", err)
	}
}

func init() {}
