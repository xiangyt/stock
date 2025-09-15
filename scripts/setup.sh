#!/bin/bash

# 智能选股系统配置初始化脚本

set -e

echo "🚀 智能选股系统配置初始化"
echo "================================"

# 检查必要的工具
check_requirements() {
    echo "📋 检查系统要求..."
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        echo "❌ Go 未安装，请先安装 Go 1.19+"
        exit 1
    fi
    
    echo "✅ Go 版本: $(go version)"
    
    # 检查 MySQL 客户端（可选）
    if command -v mysql &> /dev/null; then
        echo "✅ MySQL 客户端已安装"
    else
        echo "⚠️  MySQL 客户端未安装，无法测试数据库连接"
    fi
    
    # 检查 Redis 客户端（可选）
    if command -v redis-cli &> /dev/null; then
        echo "✅ Redis 客户端已安装"
    else
        echo "⚠️  Redis 客户端未安装，无法测试 Redis 连接"
    fi
}

# 创建必要的目录
create_directories() {
    echo "📁 创建必要的目录..."
    
    mkdir -p logs
    mkdir -p bin
    mkdir -p data
    
    echo "✅ 目录创建完成"
}

# 复制配置文件模板
setup_config() {
    echo "⚙️  设置配置文件..."
    
    # 复制主配置文件
    if [ ! -f "configs/app.yaml" ]; then
        cp configs/app.yaml.example configs/app.yaml
        echo "✅ 已创建 configs/app.yaml"
    else
        echo "⚠️  configs/app.yaml 已存在，跳过"
    fi
    
    # 复制环境变量文件
    if [ ! -f ".env" ]; then
        cp .env.example .env
        echo "✅ 已创建 .env"
    else
        echo "⚠️  .env 已存在，跳过"
    fi
    
    # 复制 Docker Compose 文件
    if [ ! -f "docker-compose.yml" ]; then
        cp docker-compose.yml.example docker-compose.yml
        echo "✅ 已创建 docker-compose.yml"
    else
        echo "⚠️  docker-compose.yml 已存在，跳过"
    fi
}

# 构建应用
build_app() {
    echo "🔨 构建应用..."
    
    echo "构建 API 服务..."
    go build -o bin/api cmd/api/main.go
    
    echo "构建 Web 服务..."
    go build -o bin/web cmd/web/main.go
    
    echo "构建 CLI 工具..."
    go build -o bin/cli cmd/cli/main.go
    
    echo "✅ 应用构建完成"
}

# 测试配置
test_config() {
    echo "🧪 测试配置..."
    
    # 创建临时测试文件
    cat > test_config.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "stock/internal/config"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("配置加载失败: %v", err)
    }
    
    fmt.Println("✅ 配置加载成功")
    fmt.Printf("应用名称: %s\n", cfg.App.Name)
    fmt.Printf("数据库: %s@%s:%d/%s\n", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
}
EOF
    
    if go run test_config.go; then
        echo "✅ 配置测试通过"
    else
        echo "❌ 配置测试失败"
        rm -f test_config.go
        exit 1
    fi
    
    rm -f test_config.go
}

# 显示下一步操作
show_next_steps() {
    echo ""
    echo "🎉 初始化完成！"
    echo "================================"
    echo ""
    echo "📝 下一步操作："
    echo ""
    echo "1. 修改配置文件："
    echo "   - 编辑 configs/app.yaml 设置数据库连接等配置"
    echo "   - 编辑 .env 设置环境变量（可选）"
    echo ""
    echo "2. 初始化数据库："
    echo "   ./bin/cli -cmd init-db"
    echo ""
    echo "3. 运行数据库迁移："
    echo "   ./bin/cli -cmd migrate"
    echo ""
    echo "4. 启动服务："
    echo "   # API 服务"
    echo "   ./bin/api"
    echo ""
    echo "   # Web 服务"
    echo "   ./bin/web"
    echo ""
    echo "5. 使用 Docker（可选）："
    echo "   docker-compose up -d"
    echo ""
    echo "📚 更多信息请查看："
    echo "   - configs/README.md - 配置说明"
    echo "   - README.md - 项目文档"
    echo ""
}

# 主函数
main() {
    check_requirements
    create_directories
    setup_config
    build_app
    test_config
    show_next_steps
}

# 运行主函数
main