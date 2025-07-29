# Gmail 邮件转发系统

基于 Go + Gin + GORM + MySQL 开发的智能邮件转发系统，能够根据邮件主题自动匹配关键字并转发至指定对象。

## 功能特性

- 🔄 自动拉取 Gmail 未读邮件
- 🎯 基于主题关键字的智能匹配
- 📧 自动转发邮件至指定对象
- 📊 完整的转发日志记录
- 🌐 RESTful API 接口
- ⏰ 可配置的定时检查任务
- 🐳 Docker 容器化部署

## 邮件主题格式

系统支持以下主题格式的邮件转发：

```
关键字 - 转发对象名称
```

示例：
- `订单通知 - 张三`
- `客户投诉 - 李四`
- `系统报警 - 运维团队`

## 系统架构

### 核心模块

1. **IMAP 邮件拉取模块** - 连接 Gmail 获取未读邮件
2. **邮件处理器** - 解析主题、匹配规则、执行转发  
3. **SMTP 转发模块** - 发送转发邮件
4. **REST API** - 管理转发对象和规则
5. **定时调度器** - 自动定期检查新邮件
6. **数据库层** - GORM + MySQL 数据持久化

### 数据模型

- **recipients** - 转发对象管理
- **forwarding_rules** - 转发规则配置  
- **email_logs** - 邮件处理日志

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

### 转发对象管理

- `GET /api/recipients` - 获取所有转发对象
- `POST /api/recipients` - 创建转发对象
- `PUT /api/recipients/:id` - 更新转发对象
- `DELETE /api/recipients/:id` - 删除转发对象

### 转发规则管理

- `GET /api/rules` - 获取所有转发规则
- `POST /api/rules` - 创建转发规则
- `PUT /api/rules/:id` - 更新转发规则
- `DELETE /api/rules/:id` - 删除转发规则

### 日志查询

- `GET /api/logs` - 获取转发日志（支持分页和状态过滤）
- `GET /api/logs/stats` - 获取统计信息

### 手动触发

- `POST /api/process` - 手动触发邮件处理

## 配置说明

| 环境变量 | 描述 | 默认值 |
|---------|------|--------|
| GMAIL_USER | Gmail 账户 | - |
| GMAIL_APP_PASSWORD | 应用专用密码 | - |
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 3306 |
| DB_USER | 数据库用户 | root |
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

## 面试考察重点

### 核心技术能力

1. **Gmail API 集成**
   - IMAP 协议使用
   - 应用专用密码认证
   - 邮件内容解析

2. **数据库设计**
   - GORM 模型定义
   - 外键关联设计
   - AutoMigrate 使用

3. **邮件处理逻辑**
   - 正则表达式解析
   - 规则匹配算法
   - 错误处理机制

4. **API 设计**
   - RESTful 规范
   - 统一响应格式
   - 参数验证

5. **系统架构**
   - 模块化设计
   - 配置管理
   - 日志记录

### 扩展思考

- 如何处理大量邮件的性能优化？
- 如何实现邮件转发的重试机制？
- 如何防止重复转发？
- 如何支持更复杂的转发规则？

## 开发计划

- [x] 项目基础架构搭建
- [x] 数据库模型设计
- [x] IMAP 邮件拉取
- [x] 邮件主题解析
- [x] SMTP 邮件转发  
- [x] REST API 接口
- [x] 定时任务调度
- [x] Docker 容器化
- [ ] 单元测试
- [ ] 性能优化
- [ ] 监控告警