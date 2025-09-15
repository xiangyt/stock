#!/bin/bash

# 配置验证脚本

set -e

echo "🔍 配置验证工具"
echo "================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查配置文件是否存在
check_config_files() {
    echo "📁 检查配置文件..."
    
    if [ -f "configs/app.yaml" ]; then
        echo -e "${GREEN}✅ configs/app.yaml 存在${NC}"
    else
        echo -e "${RED}❌ configs/app.yaml 不存在${NC}"
        echo "请运行: cp configs/app.yaml.example configs/app.yaml"
        exit 1
    fi
}

# 验证 YAML 语法
validate_yaml_syntax() {
    echo "📝 验证 YAML 语法..."
    
    # 使用 Go 程序验证 YAML
    cat > validate_yaml.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "stock/internal/config"
)

func main() {
    _, err := config.Load()
    if err != nil {
        log.Fatalf("YAML 语法错误: %v", err)
    }
    fmt.Println("✅ YAML 语法正确")
}
EOF
    
    if go run validate_yaml.go 2>/dev/null; then
        echo -e "${GREEN}✅ YAML 语法验证通过${NC}"
    else
        echo -e "${RED}❌ YAML 语法验证失败${NC}"
        go run validate_yaml.go
        rm -f validate_yaml.go
        exit 1
    fi
    
    rm -f validate_yaml.go
}

# 测试数据库连接
test_database_connection() {
    echo "🗄️  测试数据库连接..."
    
    # 创建数据库连接测试程序
    cat > test_db.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "stock/internal/config"
    "stock/internal/database"
    "stock/internal/utils"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("配置加载失败: %v", err)
    }
    
    logger := utils.NewLogger(cfg.Log)
    
    db, err := database.NewDatabase(&cfg.Database, logger)
    if err != nil {
        log.Fatalf("数据库连接失败: %v", err)
    }
    defer db.Close()
    
    if err := db.HealthCheck(); err != nil {
        log.Fatalf("数据库健康检查失败: %v", err)
    }
    
    fmt.Println("✅ 数据库连接成功")
    
    // 显示连接统计
    stats := db.GetStats()
    fmt.Printf("连接统计: %+v\n", stats)
}
EOF
    
    if go run test_db.go 2>/dev/null; then
        echo -e "${GREEN}✅ 数据库连接测试通过${NC}"
    else
        echo -e "${YELLOW}⚠️  数据库连接测试失败${NC}"
        echo "请检查数据库配置和服务状态"
        go run test_db.go
    fi
    
    rm -f test_db.go
}

# 检查端口占用
check_ports() {
    echo "🔌 检查端口占用..."
    
    # 创建端口检查程序
    cat > check_ports.go << 'EOF'
package main

import (
    "fmt"
    "net"
    "stock/internal/config"
    "strconv"
    "time"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        fmt.Printf("配置加载失败: %v\n", err)
        return
    }
    
    // 检查应用端口
    appPort := strconv.Itoa(cfg.App.Port)
    if isPortInUse("localhost", appPort) {
        fmt.Printf("⚠️  端口 %s 已被占用\n", appPort)
    } else {
        fmt.Printf("✅ 端口 %s 可用\n", appPort)
    }
    
    // 检查监控端口
    if cfg.Metrics.Enabled {
        metricsPort := strconv.Itoa(cfg.Metrics.Port)
        if isPortInUse("localhost", metricsPort) {
            fmt.Printf("⚠️  监控端口 %s 已被占用\n", metricsPort)
        } else {
            fmt.Printf("✅ 监控端口 %s 可用\n", metricsPort)
        }
    }
}

func isPortInUse(host, port string) bool {
    conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), time.Second)
    if err != nil {
        return false
    }
    conn.Close()
    return true
}
EOF
    
    go run check_ports.go
    rm -f check_ports.go
}

# 验证环境变量
check_environment_variables() {
    echo "🌍 检查环境变量..."
    
    # 检查关键环境变量
    critical_vars=("STOCK_DATABASE_PASSWORD" "STOCK_JWT_SECRET")
    
    for var in "${critical_vars[@]}"; do
        if [ -n "${!var}" ]; then
            echo -e "${GREEN}✅ $var 已设置${NC}"
        else
            echo -e "${YELLOW}⚠️  $var 未设置，将使用配置文件默认值${NC}"
        fi
    done
}

# 检查日志目录权限
check_log_permissions() {
    echo "📋 检查日志目录权限..."
    
    if [ -d "logs" ]; then
        if [ -w "logs" ]; then
            echo -e "${GREEN}✅ logs 目录可写${NC}"
        else
            echo -e "${RED}❌ logs 目录不可写${NC}"
            echo "请运行: chmod 755 logs"
        fi
    else
        echo -e "${YELLOW}⚠️  logs 目录不存在，将自动创建${NC}"
        mkdir -p logs
    fi
}

# 显示配置摘要
show_config_summary() {
    echo ""
    echo "📊 配置摘要"
    echo "================================"
    
    cat > show_config.go << 'EOF'
package main

import (
    "fmt"
    "stock/internal/config"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        fmt.Printf("配置加载失败: %v\n", err)
        return
    }
    
    fmt.Printf("应用名称: %s\n", cfg.App.Name)
    fmt.Printf("应用版本: %s\n", cfg.App.Version)
    fmt.Printf("运行环境: %s\n", cfg.App.Env)
    fmt.Printf("监听端口: %d\n", cfg.App.Port)
    fmt.Printf("调试模式: %t\n", cfg.App.Debug)
    fmt.Printf("数据库: %s://%s@%s:%d/%s\n", 
        cfg.Database.Driver, cfg.Database.User, cfg.Database.Host, 
        cfg.Database.Port, cfg.Database.Name)
    fmt.Printf("Redis: %s:%d (DB:%d)\n", 
        cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.DB)
    fmt.Printf("日志级别: %s\n", cfg.Log.Level)
    fmt.Printf("日志格式: %s\n", cfg.Log.Format)
    fmt.Printf("监控启用: %t\n", cfg.Metrics.Enabled)
    if cfg.Metrics.Enabled {
        fmt.Printf("监控端口: %d\n", cfg.Metrics.Port)
    }
}
EOF
    
    go run show_config.go
    rm -f show_config.go
}

# 主函数
main() {
    check_config_files
    validate_yaml_syntax
    check_environment_variables
    check_log_permissions
    check_ports
    test_database_connection
    show_config_summary
    
    echo ""
    echo -e "${GREEN}🎉 配置验证完成！${NC}"
}

# 运行主函数
main