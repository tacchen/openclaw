# Docker 构建缓存问题

## 问题

Docker 构建时经常使用缓存，导致前端代码更新后没有生效。

## 现象

1. 修改了前端代码（如 Home.vue）
2. 运行 `docker compose build`
3. 构建日志显示 `CACHED`
4. 部署后前端代码没有变化

## 原因

Docker 的多阶段构建会缓存每一层：
- 前端依赖 `npm install` 被缓存
- 前端构建 `npm run build` 被缓存
- 只有文件变化才会重新构建

## 解决方案

### 方案1：强制不使用缓存（推荐）

```bash
docker compose build --no-cache
docker compose down && docker compose up -d
```

### 方案2：只重新构建前端

```bash
# 删除镜像后重新构建
docker rmi rss-reader-app
docker compose build
docker compose up -d
```

### 方案3：使用 build 参数

```bash
docker compose build --pull
```

## 检查前端是否更新

部署后执行：

```bash
# 检查 JS 文件是否包含新代码
docker exec rss-reader-app-1 cat ./frontend/assets/Home-*.js | grep "新关键词"

# 检查 CSS 是否包含新样式
docker exec rss-reader-app-1 cat ./frontend/assets/Home-*.css | grep "新样式类名"
```

## 最佳实践

**每次修改前端代码后**：
1. 运行 `docker compose build --no-cache`
2. 检查容器中的文件是否更新
3. 提醒用户强制刷新浏览器（Ctrl+Shift+R）

**快速命令**：
```bash
cd /home/prj/rss-reader && docker compose build --no-cache && docker compose down && docker compose up -d
```

## 时间线

| 日期 | 问题 | 教训 |
|------|------|------|
| 2026-03-06 | GORM JSON 字段名大小写问题 | 要验证 API 返回的实际数据 |
| 2026-03-07 | 前端代码更新后没生效 | 每次都要检查 `--no-cache` |
