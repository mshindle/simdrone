package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mshindle/simdrone/internal/mock"
	"github.com/spf13/cobra"
)

var (
	mockURL     string
	mockTotal   int
	mockWorkers int
	mockDrones  int
)

// mockCmd represents the mock command
var mockCmd = &cobra.Command{
	Use:   "mock",
	Short: "Simulate drone activity",
	Long:  `Simulates multiple drones sending telemetry, position and alert events to a command handler.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		sim := &mock.Simulation{
			URL:           mockURL,
			TotalMessages: mockTotal,
			WorkerCount:   mockWorkers,
			DroneCount:    mockDrones,
		}

		fmt.Printf("Starting simulation with %d drones, sending %d messages using %d workers to %s\n",
			mockDrones, mockTotal, mockWorkers, mockURL)

		if err := sim.Run(ctx); err != nil {
			fmt.Printf("Simulation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Simulation completed.")
	},
}

func init() {
	rootCmd.AddCommand(mockCmd)

	mockCmd.Flags().StringVarP(&mockURL, "url", "u", "http://localhost:8080", "URL of the command handler")
	mockCmd.Flags().IntVarP(&mockTotal, "total", "t", 1000, "Total number of messages to send")
	mockCmd.Flags().IntVarP(&mockWorkers, "workers", "w", 2, "Number of worker routines")
	mockCmd.Flags().IntVarP(&mockDrones, "drones", "d", 5, "Number of drones reporting simultaneously")
}
