# Bug：GORM JSON 字段名大小写问题

## 问题
前端页面无法显示订阅源和文章，但数据库有数据、API 返回 200。

## 原因
Go 的 GORM 序列化 JSON 时默认使用结构体字段名（首字母大写），导致 API 返回 `{"Title": "xxx"}` 而前端期望 `{"title": "xxx"}`。

## 修复
在 models.go 的结构体字段添加 `json` tag：

```go
type Feed struct {
    Title string `json:"title"`  // 指定小写 JSON 字段名
}
```

## 教训
Go 后端返回 JSON 给前端时，**必须**添加 `json` tag，否则字段名为大写，前端无法解析。

## 排查命令
```bash
# 检查 API 返回的字段名
curl -s http://localhost/api/feeds -H "Authorization: Bearer $TOKEN" | jq 'keys'
```

---

## 复发：同一个错误又犯了一次

**日期**: 2026-03-06 18:04

**问题**: 订阅源筛选不工作，控制台打印 `selectFeed called with: undefined undefined`

**原因**: `gorm.Model` 自动添加的 `ID` 字段没有 json tag，GORM 默认序列化为大写 `ID`，前端用 `feed.id` 取值时得到 `undefined`。

**之前的修复不完整**: 只给自定义字段加了 `json` tag，但 `gorm.Model` 内嵌的 `ID` 字段被忽略了。

**正确做法**:

```go
// ❌ 错误 - gorm.Model 的 ID 会序列化为 "ID"
type Feed struct {
    gorm.Model  // 包含 ID, CreatedAt, UpdatedAt, DeletedAt
    Title string `json:"title"`
}

// ✅ 正确 - 显式定义 ID 字段覆盖 gorm.Model
type Feed struct {
    ID        uint   `gorm:"primarykey" json:"id"`
    Title     string `json:"title"`
    // 不用 gorm.Model，手动添加需要的字段
}
```

**教训**: 
1. **不要相信"修复完成"** - 要验证 API 实际返回的数据格式
2. **gorm.Model 是个坑** - 它的字段没有 json tag，会输出大写字段名
3. **测试命令**: `curl -s http://localhost/api/feeds | jq '.[0] | keys'` 确认字段名
