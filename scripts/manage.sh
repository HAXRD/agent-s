#!/bin/bash

# 统一服务管理脚本
# 用于启动、停止、重启和检查后端、前端服务状态

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKEND_PID_FILE="$PROJECT_ROOT/server.pid"
FRONTEND_PID_FILE="$PROJECT_ROOT/frontend.pid"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查环境变量文件
check_env_file() {
    if [ ! -f "$PROJECT_ROOT/.env" ]; then
        echo -e "${YELLOW}警告: .env 文件不存在，使用默认配置${NC}"
        if [ -f "$PROJECT_ROOT/.env.example" ]; then
            echo -e "${YELLOW}提示: 可以复制 .env.example 到 .env 并修改配置${NC}"
        fi
    fi
}

# 检查数据库状态
check_database() {
    if [ -f "$SCRIPT_DIR/db.sh" ]; then
        "$SCRIPT_DIR/db.sh" status > /dev/null 2>&1
        return $?
    fi
    return 1
}

# 启动后端服务
start_backend() {
    if [ -f "$BACKEND_PID_FILE" ]; then
        PID=$(cat "$BACKEND_PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            echo -e "${YELLOW}后端服务已在运行 (PID: $PID)${NC}"
            return 0
        else
            rm -f "$BACKEND_PID_FILE"
        fi
    fi

    echo "启动后端服务..."
    check_env_file

    # 检查数据库连接（可选）
    if ! check_database; then
        echo -e "${YELLOW}警告: 数据库可能未启动，后端服务可能无法连接数据库${NC}"
        echo -e "${YELLOW}提示: 运行 ./scripts/db.sh start 启动数据库${NC}"
    fi

    cd "$PROJECT_ROOT" || exit 1
    
    # 后台启动后端服务
    nohup go run cmd/server/main.go > server.log 2>&1 &
    BACKEND_PID=$!
    echo $BACKEND_PID > "$BACKEND_PID_FILE"
    
    sleep 2
    
    if ps -p "$BACKEND_PID" > /dev/null 2>&1; then
        echo -e "${GREEN}后端服务已启动 (PID: $BACKEND_PID)${NC}"
        echo "日志文件: $PROJECT_ROOT/server.log"
    else
        echo -e "${RED}后端服务启动失败${NC}"
        rm -f "$BACKEND_PID_FILE"
        return 1
    fi
}

# 停止后端服务
stop_backend() {
    if [ ! -f "$BACKEND_PID_FILE" ]; then
        echo -e "${YELLOW}后端服务未运行${NC}"
        return 0
    fi

    PID=$(cat "$BACKEND_PID_FILE")
    
    if ! ps -p "$PID" > /dev/null 2>&1; then
        echo -e "${YELLOW}后端服务未运行 (PID 文件存在但进程不存在)${NC}"
        rm -f "$BACKEND_PID_FILE"
        return 0
    fi

    echo "停止后端服务 (PID: $PID)..."
    
    # 优雅停止（发送 SIGTERM）
    kill -TERM "$PID" 2>/dev/null
    
    # 等待进程结束
    for i in {1..10}; do
        if ! ps -p "$PID" > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done
    
    # 如果还在运行，强制停止
    if ps -p "$PID" > /dev/null 2>&1; then
        echo -e "${YELLOW}强制停止后端服务...${NC}"
        kill -KILL "$PID" 2>/dev/null
    fi
    
    rm -f "$BACKEND_PID_FILE"
    echo -e "${GREEN}后端服务已停止${NC}"
}

# 启动前端服务
start_frontend() {
    if [ -f "$FRONTEND_PID_FILE" ]; then
        PID=$(cat "$FRONTEND_PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            echo -e "${YELLOW}前端服务已在运行 (PID: $PID)${NC}"
            return 0
        else
            rm -f "$FRONTEND_PID_FILE"
        fi
    fi

    if [ ! -d "$PROJECT_ROOT/frontend" ]; then
        echo -e "${RED}错误: frontend 目录不存在${NC}"
        return 1
    fi

    echo "启动前端服务..."
    check_env_file

    cd "$PROJECT_ROOT/frontend" || exit 1
    
    # 检查 node_modules
    if [ ! -d "node_modules" ]; then
        echo "安装前端依赖..."
        npm install
    fi
    
    # 后台启动前端服务
    nohup npm run dev > ../frontend.log 2>&1 &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > "$FRONTEND_PID_FILE"
    
    sleep 2
    
    if ps -p "$FRONTEND_PID" > /dev/null 2>&1; then
        echo -e "${GREEN}前端服务已启动 (PID: $FRONTEND_PID)${NC}"
        echo "日志文件: $PROJECT_ROOT/frontend.log"
        echo "前端地址: http://localhost:5173"
    else
        echo -e "${RED}前端服务启动失败${NC}"
        rm -f "$FRONTEND_PID_FILE"
        return 1
    fi
}

# 停止前端服务
stop_frontend() {
    if [ ! -f "$FRONTEND_PID_FILE" ]; then
        echo -e "${YELLOW}前端服务未运行${NC}"
        return 0
    fi

    PID=$(cat "$FRONTEND_PID_FILE")
    
    if ! ps -p "$PID" > /dev/null 2>&1; then
        echo -e "${YELLOW}前端服务未运行 (PID 文件存在但进程不存在)${NC}"
        rm -f "$FRONTEND_PID_FILE"
        return 0
    fi

    echo "停止前端服务 (PID: $PID)..."
    
    # 停止进程及其子进程
    kill -TERM "$PID" 2>/dev/null
    
    # 等待进程结束
    for i in {1..10}; do
        if ! ps -p "$PID" > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done
    
    # 如果还在运行，强制停止
    if ps -p "$PID" > /dev/null 2>&1; then
        echo -e "${YELLOW}强制停止前端服务...${NC}"
        kill -KILL "$PID" 2>/dev/null
        # 也尝试停止相关的 node 进程
        pkill -f "vite" 2>/dev/null
    fi
    
    rm -f "$FRONTEND_PID_FILE"
    echo -e "${GREEN}前端服务已停止${NC}"
}

# 检查服务状态
check_status() {
    echo "=== 服务状态 ==="
    echo ""
    
    # 检查后端服务
    echo -n "后端服务: "
    if [ -f "$BACKEND_PID_FILE" ]; then
        PID=$(cat "$BACKEND_PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            echo -e "${GREEN}运行中 (PID: $PID)${NC}"
        else
            echo -e "${RED}已停止 (PID 文件存在但进程不存在)${NC}"
        fi
    else
        echo -e "${YELLOW}未运行${NC}"
    fi
    
    # 检查前端服务
    echo -n "前端服务: "
    if [ -f "$FRONTEND_PID_FILE" ]; then
        PID=$(cat "$FRONTEND_PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            echo -e "${GREEN}运行中 (PID: $PID)${NC}"
        else
            echo -e "${RED}已停止 (PID 文件存在但进程不存在)${NC}"
        fi
    else
        echo -e "${YELLOW}未运行${NC}"
    fi
    
    # 检查数据库状态
    echo -n "数据库: "
    if check_database; then
        echo -e "${GREEN}运行中${NC}"
    else
        echo -e "${YELLOW}未运行${NC}"
    fi
    
    echo ""
}

# 主逻辑
case "$1" in
    start)
        start_backend
        start_frontend
        ;;
    stop)
        stop_backend
        stop_frontend
        ;;
    restart)
        echo "重启服务..."
        stop_backend
        stop_frontend
        sleep 2
        start_backend
        start_frontend
        ;;
    status)
        check_status
        ;;
    start-backend)
        start_backend
        ;;
    stop-backend)
        stop_backend
        ;;
    start-frontend)
        start_frontend
        ;;
    stop-frontend)
        stop_frontend
        ;;
    *)
        echo "用法: $0 {start|stop|restart|status|start-backend|stop-backend|start-frontend|stop-frontend}"
        echo ""
        echo "命令说明："
        echo "  start            - 启动所有服务（后端和前端）"
        echo "  stop             - 停止所有服务"
        echo "  restart          - 重启所有服务"
        echo "  status           - 查看所有服务状态"
        echo "  start-backend    - 仅启动后端服务"
        echo "  stop-backend     - 仅停止后端服务"
        echo "  start-frontend   - 仅启动前端服务"
        echo "  stop-frontend    - 仅停止前端服务"
        exit 1
        ;;
esac

