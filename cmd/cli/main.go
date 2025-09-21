package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"stock/internal/logger"

	"stock/internal/config"
	"stock/internal/service"
)

func main() {
	var (
		command  = flag.String("cmd", "", "Command to execute: init-db, migrate, update-data, select-stocks")
		strategy = flag.String("strategy", "technical", "Selection strategy: technical, fundamental, combined")
		limit    = flag.Int("limit", 20, "Number of stocks to select")
	)
	flag.Parse()

	if *command == "" {
		printUsage()
		os.Exit(1)
	}

	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	log := logger.NewLogger(cfg.Log)

	// 初始化服务
	services, err := service.NewServices(cfg, log)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}

	// 执行命令
	switch *command {
	case "init-db":
		err = initDatabase(services)
	case "migrate":
		err = migrateDatabase(services)
	case "select-stocks":
		err = selectStocks(services, *strategy, *limit)
	default:
		fmt.Printf("Unknown command: %s\n", *command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		logger.Errorf("Command failed: %v", err)
		os.Exit(1)
	}

	logger.Info("Command completed successfully")
}

func printUsage() {
	fmt.Println("Usage: cli -cmd <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  init-db      Initialize database")
	fmt.Println("  migrate      Run database migration")
	fmt.Println("  update-data  Update stock data")
	fmt.Println("  select-stocks Execute stock selection")
	fmt.Println("\nOptions:")
	fmt.Println("  -strategy    Selection strategy (technical, fundamental, combined)")
	fmt.Println("  -limit       Number of stocks to select")
	fmt.Println("  -source      Data source (tushare, akshare, yahoo)")
}

func initDatabase(services *service.Services) error {
	fmt.Println("Initializing database...")
	return services.Database.InitDB()
}

func migrateDatabase(services *service.Services) error {
	fmt.Println("Running database migration...")
	return services.Database.Migrate()
}

func selectStocks(services *service.Services, strategy string, limit int) error {
	// todo

	return nil
}
