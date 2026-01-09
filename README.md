# 医学问答助手

基于检索增强生成（RAG）技术的医学问答助手项目。

## 项目结构

```
medical-QA-assistant/
├── backend/          # Go + Gin + Gorm 后端
├── frontend/         # Vue3 + Vite 前端
├── PROJECT.md       # 项目设计文档
└── TODO.md          # 项目待办事项
```

## 快速开始

### 后端设置

1. 进入后端目录：
```bash
cd backend
```

2. 安装依赖：
```bash
go mod download
```

3. 创建 MySQL 数据库：
```sql
CREATE DATABASE medical_qa;
```

4. 配置环境变量（复制 `.env.example` 到 `.env` 并修改）：
```bash
cp .env.example .env
```

5. 运行后端服务：
```bash
go run cmd/server/main.go
```

后端服务将在 `http://localhost:8081` 启动。

### 前端设置

1. 进入前端目录：
```bash
cd frontend
```

2. 安装依赖：
```bash
npm install
```

3. 启动开发服务器：
```bash
npm run dev
```

前端应用将在 `http://localhost:3000` 启动。


## 功能特性

- ✅ 用户注册和登录
- ✅ JWT Token 认证
- ✅ 密码加密存储
- ✅ 前端路由保护
- ⏳ 医学文档管理（待实现）
- ✅ RAG 问答功能
- ⏳ 对话历史（待实现）

## 技术栈

### 后端
- Go 1.21+
- Gin Web 框架
- Gorm ORM
- MySQL 数据库
- JWT 认证

### 前端
- Vue 3
- Vite
- Vue Router
- Pinia
- Axios

## 开发说明

详细的项目设计文档请参考 [PROJECT.md](./PROJECT.md)。
