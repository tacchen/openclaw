#!/bin/bash
# 测试脚本

set -e

echo "🧪 Running tests..."

# 设置颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 1. 编译检查
echo -e "${YELLOW}→ Checking compilation...${NC}"
if ! go build ./...; then
    echo -e "${RED}❌ Compilation failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Compilation successful${NC}"

# 2. 运行 Repository 层测试
echo -e "${YELLOW}→ Running Repository tests...${NC}"
if ! go test ./internal/repository/ -v -cover; then
    echo -e "${RED}❌ Repository tests failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Repository tests passed${NC}"

# 3. 运行所有测试
echo -e "${YELLOW}→ Running all tests...${NC}"
if ! go test ./... -v -cover; then
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ All tests passed${NC}"

echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
