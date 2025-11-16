#!/bin/bash

# 数据库 Docker 管理脚本

case "$1" in
    start)
        echo "启动 PostgreSQL 数据库..."
        docker-compose up -d postgres
        echo "等待数据库就绪..."
        sleep 5
        echo "数据库已启动"
        ;;
    stop)
        echo "停止 PostgreSQL 数据库..."
        docker-compose stop postgres
        echo "数据库已停止"
        ;;
    restart)
        echo "重启 PostgreSQL 数据库..."
        docker-compose restart postgres
        echo "数据库已重启"
        ;;
    status)
        echo "检查数据库状态..."
        docker-compose ps postgres
        ;;
    logs)
        echo "查看数据库日志..."
        docker-compose logs -f postgres
        ;;
    shell)
        echo "连接到 PostgreSQL 数据库..."
        docker-compose exec postgres psql -U postgres -d crypto_monitor
        ;;
    reset)
        echo "重置数据库（删除所有数据）..."
        read -p "确认要删除所有数据吗？(y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker-compose down -v
            docker-compose up -d postgres
            echo "数据库已重置"
        else
            echo "操作已取消"
        fi
        ;;
    backup)
        BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).sql"
        echo "备份数据库到 $BACKUP_FILE..."
        docker-compose exec -T postgres pg_dump -U postgres crypto_monitor > "$BACKUP_FILE"
        echo "备份完成: $BACKUP_FILE"
        ;;
    restore)
        if [ -z "$2" ]; then
            echo "用法: $0 restore <backup_file.sql>"
            exit 1
        fi
        echo "从 $2 恢复数据库..."
        docker-compose exec -T postgres psql -U postgres -d crypto_monitor < "$2"
        echo "恢复完成"
        ;;
    *)
        echo "用法: $0 {start|stop|restart|status|logs|shell|reset|backup|restore}"
        echo ""
        echo "命令说明："
        echo "  start    - 启动数据库"
        echo "  stop     - 停止数据库"
        echo "  restart  - 重启数据库"
        echo "  status   - 查看数据库状态"
        echo "  logs     - 查看数据库日志"
        echo "  shell    - 连接到数据库命令行"
        echo "  reset    - 重置数据库（删除所有数据）"
        echo "  backup   - 备份数据库"
        echo "  restore  - 从备份文件恢复数据库"
        exit 1
        ;;
esac

