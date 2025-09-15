package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"stock/internal/config"
	"stock/internal/service"
	"stock/internal/utils"
)

func main() {
	var (
		command  = flag.String("cmd", "", "Command to execute: init-db, migrate, update-data, select-stocks")
		strategy = flag.String("strategy", "technical", "Selection strategy: technical, fundamental, combined")
		limit    = flag.Int("limit", 20, "Number of stocks to select")
		source   = flag.String("source", "tushare", "Data source: tushare, akshare, yahoo")
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
	logger := utils.NewLogger(cfg.Log)

	// 初始化服务
	services, err := service.NewServices(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}

	// 执行命令
	switch *command {
	case "init-db":
		err = initDatabase(services)
	case "migrate":
		err = migrateDatabase(services)
	case "update-data":
		err = updateData(services, *source)
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

func updateData(services *service.Services, source string) error {
	fmt.Printf("Updating data from %s...\n", source)
	return services.DataCollector.UpdateAllData(source)
}

func selectStocks(services *service.Services, strategy string, limit int) error {
	fmt.Printf("Selecting stocks using %s strategy (limit: %d)...\n", strategy, limit)
	results, err := services.StrategyEngine.ExecuteStrategy(strategy, limit)
	if err != nil {
		return err
	}

	fmt.Printf("Selected %d stocks:\n", len(results))
	for i, result := range results {
		fmt.Printf("%d. %s (%s) - Score: %.2f\n",
			i+1, result.Stock.Name, result.Stock.Code, result.Score)
	}

	return nil
}
