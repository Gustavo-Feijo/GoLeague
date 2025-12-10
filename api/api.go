package main

import (
	"context"
	"goleague/api/cache"
	"goleague/api/modules"
	"goleague/api/routes"
	"goleague/pkg/config"
	"goleague/pkg/database"
	"goleague/pkg/redis"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Simple testing for the Redis.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Couldn't initialize the configuration: %v", err)
	}

	moduleDeps, cleanup, err := initializeModuleDependencies(cfg)
	if err != nil {
		log.Fatalf("Couldn't initialize dependencies: %v", err)
	}
	defer cleanup()

	// Runs the migrations.
	rawDb, err := moduleDeps.DB.DB()
	if err != nil {
		log.Fatalf("Couldn't get raw db connection: %v", err)
	}

	if err = database.RunMigrations(cfg, rawDb); err != nil {
		log.Fatalf("Couldn't run migrations: %v", err)
	}

	// Create a module with all necessary handlers.
	module, err := modules.NewModule(moduleDeps)
	if err != nil {
		log.Fatalf("Error to start the modules: %v", err)
	}

	// Create a new router with the routes setup.
	router := routes.NewRouter(module.Router)
	router.SetupRoutes(
		module.ChampionHandler,
		module.MatchHandler,
		module.TierlistHandler,
		module.PlayerHandler,
	)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        router.Engine, // router is *gin.Engine which implements http.Handler
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed to start:", err)
	}
}

// initializeModuleDependencies starts all necessary dependencies.
func initializeModuleDependencies(config *config.Config) (*modules.ModuleDependencies, func(), error) {
	var cleanupFuncs []func()

	cleanup := func() {
		for i := len(cleanupFuncs) - 1; i >= 0; i-- {
			cleanupFuncs[i]()
		}
	}
	// Connect to the fetcher grpc.
	fetcherHost := config.Grpc.Host + ":" + config.Grpc.Port
	grpcClient, err := grpc.NewClient(fetcherHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	// Creates the database connection.
	db, err := database.NewConnection(config.Database.DSN)
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	// Creates the redis connection.
	// Can run without, could implement it to not start withou.
	redis, err := redis.NewClient(config.Redis)
	if err != nil {
		log.Printf("Warning: Error connecting to Redis: %v", err)
	} else {
		cleanupFuncs = append(cleanupFuncs, func() { redis.Client.Close() })
	}

	memCache := cache.NewMemCache()

	if sqlDB, err := db.DB(); err == nil {
		cleanupFuncs = append(cleanupFuncs, func() { sqlDB.Close() })
	}

	// Create a instance  just for initializing.
	championCache := cache.NewChampionCache(db, redis, memCache)
	champCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	championCache.Initialize(champCtx)

	cleanupFuncs = append(cleanupFuncs, func() { grpcClient.Close() })
	cleanupFuncs = append(cleanupFuncs, func() { memCache.Close() })

	// Pass down the dependencies.
	moduleDeps := &modules.ModuleDependencies{
		DB:            db,
		ChampionCache: championCache,
		GrpcClient:    grpcClient,
		MemCache:      memCache,
		Redis:         redis,
	}

	return moduleDeps, cleanup, nil
}
