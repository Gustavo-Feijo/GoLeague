package main

import (
	"goleague/pkg/config"
	"goleague/scheduler/jobs"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if os.Getenv("ENVIRONMENT") != "docker" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
	config.LoadEnv()

	log.Println("Starting scheduler.")

	// Create a new scheduler with options
	s, err := gocron.NewScheduler(
		gocron.WithLocation(time.UTC),
	)
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}

	// Register champion cache revalidation job - once per day at 3:00 AM
	_, err = s.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(4, 0, 0),
			),
		),
		gocron.NewTask(
			jobs.RevalidateCache,
		),
		gocron.WithName("cache-revalidation"),
		gocron.WithTags("cache"),
	)
	if err != nil {
		log.Fatalf("Failed to create cache job: %v", err)
	}

	// Register champion cache revalidation job - once per day at 3:00 AM
	_, err = s.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(3, 0, 0),
			),
		),
		gocron.NewTask(
			jobs.RecalculateMatchRating,
		),
		gocron.WithName("match-rating-revalidation"),
		gocron.WithTags("rating"),
	)
	if err != nil {
		log.Fatalf("Failed to create match rating revalidation job: %v", err)
	}

	// Start the scheduler
	s.Start()

	defer func() {
		// Shutdown the scheduler when main() exits
		err := s.Shutdown()
		if err != nil {
			log.Printf("Error shutting down scheduler: %v", err)
		}
	}()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigChan
	log.Println("Shutting down scheduler...")
}
