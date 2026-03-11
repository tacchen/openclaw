# Git 提交规范

本项目遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范。

## 提交格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

## Type 类型说明

| 类型 | 说明 | 示例 |
|------|------|------|
| **feat** | 新功能 | feat: 添加文章收藏功能 |
| **fix** | 修复 Bug | fix: 修复夜间模式文字颜色问题 |
| **docs** | 文档变更 | docs: 更新 README 安装说明 |
| **style** | 代码格式（不影响功能） | style: 格式化代码缩进 |
| **refactor** | 重构代码（不是新功能也不是修复） | refactor: 重构 API 响应处理 |
| **perf** | 性能优化 | perf: 优化文章列表加载速度 |
| **test** | 添加测试 | test: 添加用户认证单元测试 |
| **build** | 构建系统或依赖变更 | build: 升级 node 版本 |
| **ci** | CI 配置变更 | ci: 添加 GitHub Actions 配置 |
| **chore** | 其他杂项 | chore: 更新 .gitignore |
| **revert** | 回滚提交 | revert: 回滚登录功能修改 |

## Scope（可选）

表示影响范围，如：`feat(backend):`、`fix(frontend):`、`refactor(api):`

## 示例

```bash
# 新功能
git commit -m "feat: 添加 RSS 订阅源分类功能"

# 修复 Bug
git commit -m "fix: 修复夜间模式侧边栏文字颜色"

# 重构
git commit -m "refactor: 重构文章列表组件"

# 带 scope
git commit -m "feat(api): 添加文章搜索接口"
```

## 配置 Husky（可选）

如需强制规范，可安装 husky + commitlint：

```bash
npm install -D husky @commitlint/cli @commitlint/config-conventional
npx husky init
echo "npx --no -- commitlint --edit \$1" > .husky/commit-msg
```
