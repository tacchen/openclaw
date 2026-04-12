#!/bin/bash
# 定时同步仓库脚本

REPO_DIR="/home/prj/rss-reader"
LOG_FILE="/home/prj/rss-reader/sync.log"

cd $REPO_DIR || exit 1

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting sync..." >> $LOG_FILE

# 拉取最新代码
git fetch origin >> $LOG_FILE 2>&1
git pull origin main >> $LOG_FILE 2>&1

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Sync completed" >> $LOG_FILE
