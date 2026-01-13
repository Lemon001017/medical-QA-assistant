# 快速启动指南

## 前置要求

- Go 1.21 或更高版本
- Node.js 16+ 和 npm
- MySQL 5.7+ 或 MySQL 8.0+

## 步骤 1: 设置数据库

1. 登录 MySQL：
```bash
mysql -u root -p
```

2. 创建数据库：
```sql
CREATE DATABASE medical_qa CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
EXIT;
```

## 步骤 1.5: 启动 Chroma 向量数据库

Chroma 用于存储文档的向量嵌入。你可以通过以下方式启动：

**方式 1: 使用 Docker（推荐）**
```bash
docker pull chromadb/chroma
docker run -p 8000:8000 chromadb/chroma
```

**方式 2: 使用 Python（需要先安装 Python 和 pip）**
```bash
pip install chromadb
chroma run --path ./chroma_data --port 8000
```

Chroma 将在 `http://localhost:8000` 启动。

## 步骤 2: 启动后端

1. 进入后端目录：
```bash
cd backend
```

2. 安装 Go 依赖：
```bash
go mod download
```

3. 配置环境变量（创建 `.env` 文件，程序会自动加载）：
```bash
cat > .env << 'EOF'
# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=medical_qa
JWT_SECRET=your-secret-key-change-in-production
PORT=8081

# 对话模型提供商（默认openai）
LLM_PROVIDER=openai

OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_API_KEY=
OPENAI_MODEL=gpt-4o-mini
OPENAI_EMBEDDING_MODEL=text-embedding-3-small

DEEPSEEK_BASE_URL=https://api.deepseek.com/v1
DEEPSEEK_API_KEY=
DEEPSEEK_MODEL=deepseek-chat

# 向量数据库配置
CHROMA_BASE_URL=http://localhost:8000
CHROMA_COLLECTION=medical_documents

# 嵌入模型配置
ALIYUN_EMBEDDING_MODEL=text-embedding-v4
ALIYUN_EMBEDDING_KEY=
ALIYUN_EMBEDDING_BASEURL=https://dashscope.aliyuncs.com/compatible-mode/v1
EOF
```

请将 `your_mysql_password` 替换为你的 MySQL 密码，并按需填写 `OPENAI_API_KEY`。

**重要提示：**
- `OPENAI_API_KEY` 是必需的，用于生成文档嵌入向量和问答
- `CHROMA_BASE_URL` 默认是 `http://localhost:8000`，如果 Chroma 运行在其他地址，请修改

4. 启动后端服务（会自动尝试加载当前目录下的 `.env`）：
```bash
go run cmd/server/main.go
```

后端将在 `http://localhost:8081` 启动。

## 步骤 3: 启动前端

1. 打开新的终端窗口，进入前端目录：
```bash
cd frontend
```

2. 安装 npm 依赖：
```bash
npm install
```

3. 启动开发服务器：
```bash
npm run dev
```

前端将在 `http://localhost:3000` 启动。

## 步骤 4: 测试注册和登录

1. 打开浏览器访问 `http://localhost:3000`
2. 点击"立即注册"创建新账号
3. 填写注册信息：
   - 用户名：至少 3 个字符
   - 邮箱：有效的邮箱地址
   - 密码：至少 6 个字符
4. 注册成功后会自动跳转到首页
5. 可以点击"退出"按钮，然后使用刚才注册的账号登录

## 常见问题

### 后端无法连接数据库
- 检查 MySQL 服务是否运行
- 确认 `.env` 文件中的数据库配置正确
- 确认数据库 `medical_qa` 已创建

### Chroma 连接失败
- 确认 Chroma 服务正在运行（默认 `http://localhost:8000`）
- 检查 `CHROMA_BASE_URL` 环境变量是否正确
- 如果使用 Docker，确认端口映射正确（`-p 8000:8000`）

### 前端无法连接后端
- 确认后端服务正在运行（`http://localhost:8081`）
- 检查浏览器控制台是否有 CORS 错误
- 确认 `vite.config.js` 中的代理配置正确

### 端口被占用
- 后端默认端口：8081，可在 `.env` 中修改 `PORT`
- 前端默认端口：3000，可在 `vite.config.js` 中修改
