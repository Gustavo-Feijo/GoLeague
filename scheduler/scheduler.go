package main

import (
	"goleague/pkg/config"
	"goleague/pkg/database"
	"goleague/scheduler/jobs"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Couldn't initialize the configuration: %v", err)
	}

	db, err := database.NewConnection(cfg.Database.DSN)
	if err != nil {
		log.Fatal(err)
	}

	// Runs the migrations.
	rawDb, err := db.DB()
	if err != nil {
		log.Fatalf("Couldn't get raw db connection: %v", err)
	}

	if err := database.RunMigrations(cfg, rawDb); err != nil {
		log.Fatal(err)
	}

	log.Println("Starting scheduler.")

	// Create a new scheduler with options.
	s, err := gocron.NewScheduler(
		gocron.WithLocation(time.UTC),
	)
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}

	// Register champion cache revalidation job - once per day at 3:00 AM.
	_, err = s.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(4, 0, 0),
			),
		),
		gocron.NewTask(
			jobs.RevalidateCache,
			cfg,
		),
		gocron.WithName("cache-revalidation"),
		gocron.WithTags("cache"),
		gocron.JobOption(gocron.WithStartImmediately()),
	)
	if err != nil {
		log.Fatalf("Failed to create cache job: %v", err)
	}

	// Register champion cache revalidation job - once per day at 3:00 AM.
	_, err = s.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(3, 0, 0),
			),
		),
		gocron.NewTask(
			jobs.RecalculateMatchRating,
			cfg,
		),
		gocron.WithName("match-rating-revalidation"),
		gocron.WithTags("rating"),
		gocron.JobOption(gocron.WithStartImmediately()),
	)
	if err != nil {
		log.Fatalf("Failed to create match rating revalidation job: %v", err)
	}

	_, err = s.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(4, 0, 0),
			),
		),
		gocron.NewTask(
			jobs.RecalculateFetchPriority,
			cfg,
		),
		gocron.WithName("fetch-priority-revalidation"),
		gocron.WithTags("priority"),
		gocron.JobOption(gocron.WithStartImmediately()),
	)
	if err != nil {
		log.Fatalf("Failed to create fetch priority revalidation job: %v", err)
	}

	// Start the scheduler.
	s.Start()

	defer func() {
		// Shutdown the scheduler when main() exits.
		err := s.Shutdown()
		if err != nil {
			log.Printf("Error shutting down scheduler: %v", err)
		}
	}()

	// Setup signal handling for graceful shutdown.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal.
	<-sigChan
	log.Println("Shutting down scheduler...")
}
