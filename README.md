# Gmail 邮件转发系统

基于 Go + Gin + GORM + MySQL 开发的智能邮件转发系统，能够根据邮件主题自动匹配关键字并转发至指定邮箱地址。

## 功能特性

- 🔄 自动拉取 Gmail 未读邮件
- 🎯 基于主题关键字的智能匹配
- 📧 自动转发邮件至指定邮箱地址
- 🌐 RESTful API 接口
- ⏰ 可配置的定时检查任务（默认5分钟）
- 🐳 Docker 容器化部署
- ⚡ 批量规则加载优化性能
- 🔄 SMTP发送重试机制

## 邮件主题格式

系统支持以下主题格式的邮件转发：

```
关键字 - 邮箱地址
```

示例：
- `订单通知 - user@example.com`
- `客户投诉 - support@company.com`
- `系统报警 - admin@domain.com`

## 系统架构

### 核心模块

1. **IMAP 邮件拉取模块** - 连接 Gmail 获取未读邮件
2. **邮件处理器** - 解析主题、匹配规则、执行转发  
3. **SMTP 转发模块** - 发送转发邮件
4. **REST API** - 管理转发对象和规则
5. **定时调度器** - 自动定期检查新邮件
6. **数据库层** - GORM + MySQL 数据持久化

### 数据模型

- **recipients** - 转发对象管理（自动创建，存储邮箱地址）
- **forwarding_rules** - 转发规则配置（关键字匹配）

## 快速开始

### 环境要求

- Go 1.23+
- MySQL 8.0+
- Gmail 账户（需启用应用专用密码）

### 配置 Gmail

1. 开启 Gmail 两步验证
2. 生成应用专用密码
3. 配置环境变量

### 安装部署

#### 1. 克隆项目

```bash
git clone <repository-url>
cd gmail-forwarding
```

#### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，填入实际配置
```

#### 3. Docker 部署（推荐）

```bash
# 1. 复制并配置环境变量
cp .env.example .env
# 编辑 .env 文件，设置真实的配置信息

# 2. 启动服务
docker-compose up -d
```

#### 4. 本地开发

```bash
# 安装依赖
go mod tidy

# 启动服务
go run cmd/server/main.go
```

## API 接口

### 核心功能

- `GET /health` - 系统健康检查
- `POST /api/process` - 手动触发邮件处理

### 转发对象管理

- `GET /api/recipients` - 获取所有转发对象
- `GET /api/recipients/:id` - 获取指定转发对象
- `POST /api/recipients` - 创建转发对象
- `PUT /api/recipients/:id` - 更新转发对象
- `DELETE /api/recipients/:id` - 删除转发对象

### 转发规则管理

- `GET /api/rules` - 获取所有转发规则
- `GET /api/rules/:id` - 获取指定转发规则
- `POST /api/rules` - 创建转发规则
- `PUT /api/rules/:id` - 更新转发规则
- `DELETE /api/rules/:id` - 删除转发规则

### 示例用法

```bash
# 创建转发规则
curl -X POST http://localhost:8080/api/rules \
  -H "Content-Type: application/json" \
  -d '{"keyword": "订单通知", "active": true}'

# 手动触发邮件处理
curl -X POST http://localhost:8080/api/process
```

## 配置说明

| 环境变量 | 描述 | 默认值 |
|---------|------|--------|
| GMAIL_USER | Gmail 账户 | - |
| GMAIL_APP_PASSWORD | 应用专用密码 | - |
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 3306 |
| DB_USER | 数据库用户 | gmail_user |
| DB_PASSWORD | 数据库密码 | - |
| DB_NAME | 数据库名 | gmail_forwarding |
| APP_PORT | 应用端口 | 8080 |
| CHECK_INTERVAL | 检查间隔 | 5m |

## 技术栈

- **后端框架**: Gin Web Framework
- **ORM**: GORM v2
- **数据库**: MySQL 8.0
- **定时任务**: robfig/cron
- **邮件处理**: go-imap, go-message
- **容器化**: Docker + Docker Compose

## 系统特性

### 性能优化

- **批量规则加载** - 启动时一次性加载所有规则，避免每封邮件查询数据库
- **内存匹配** - 规则匹配在内存中进行，提高处理速度
- **SMTP重试机制** - 发送失败时自动重试3次，提高成功率

### 自动化特性

- **自动创建收件人** - 首次出现的邮箱地址自动创建收件人记录
- **定时处理** - 每5分钟自动检查未读邮件
- **邮件标记** - 处理后自动标记邮件为已读

### 部署特性

- **Docker一键部署** - 包含MySQL数据库的完整部署方案
- **环境变量配置** - 敏感信息通过环境变量配置
- **健康检查** - 提供健康检查端点监控系统状态

## 工作流程

1. **定时检查** - 系统每5分钟检查Gmail未读邮件
2. **主题解析** - 使用正则表达式解析"关键字 - 邮箱地址"格式
3. **规则匹配** - 在内存中快速匹配关键字规则
4. **自动转发** - 匹配成功后自动转发邮件到指定邮箱
5. **记录管理** - 自动创建和维护收件人记录

## 开发状态

- [x] 项目基础架构搭建
- [x] 数据库模型设计  
- [x] IMAP 邮件拉取
- [x] 邮件主题解析
- [x] SMTP 邮件转发
- [x] REST API 接口
- [x] 定时任务调度
- [x] Docker 容器化
- [x] 性能优化（批量规则加载）
- [x] SMTP EOF错误修复
- [x] 代码清理和重构