# 逆水寒帮会联赛数据分析平台

一个可 Docker 独立部署的 Web 后台，用于导入和分析逆水寒帮会联赛 CSV。正式技术栈为 Go + Gin + Ent 模型、Vue 3 + TypeScript + Vite + TailwindCSS、PostgreSQL 15+、Redis 7+、ECharts 和 Docker Compose。

当前本地没有部署环境，按项目要求未执行编译、测试、安装依赖或启动 Docker；本仓库已补齐源码、部署配置、脚本和静态检查口径。

## 目录

- `backend/`：Go/Gin API、PostgreSQL migration、Redis 会话与预览缓存、CSV 导入、评分分析、Ent schema。
- `frontend/`：Vue 3 后台应用、ECharts 图表、导入流程、排名、对比、历史和设置页面。
- `deployment/`：容器入口脚本。
- `scripts/`：一键安装、卸载、打包脚本。
- `data/`：运行时上传目录占位，不提交用户数据。
- `backups/`：备份目录占位，不提交备份文件。
- `legacy/fastapi-react/`：旧 FastAPI/React 实现归档，仅作参考。
- `设计文档/`：权威需求、设计图、样例数据和资源。

## 核心能力

- 首次启动自动创建 `admin`，随机密码只在首次日志显示一次，并强制首次改密。
- HttpOnly SameSite Cookie 会话、Redis 会话索引、CSRF 防护、登录失败限流。
- CSV 预览支持 UTF-8 BOM、UTF-8、GB18030、重复表头、字段别名、空行、数字逗号、两个帮会和十三职业校验。
- 确认导入后写入 PostgreSQL，并生成 KDA、参团率、伤害占比、职业区间转化率、六维评分和排名。
- 覆盖概览、个人排名、多场总榜、玩家详情、团内 TOP3、对手对比、团队对比、导入、历史、设置页面。
- Docker Compose 默认启动独立 `app`、`postgres`、`redis`、独立网络和独立卷，宿主机默认端口 `18080`。

## Docker 部署

```bash
cp .env.example .env
# 至少替换 POSTGRES_PASSWORD 和 SESSION_SECRET；一键安装脚本会自动生成
docker compose up -d --build
docker compose logs app
```

访问地址：

```text
http://服务器地址:18080
```

首次管理员账号为 `admin`。初始随机密码通过首次启动日志查看：

```bash
docker compose logs app
```

## 一键安装包

打包：

```bash
bash scripts/package.sh
```

安装：

```bash
tar -xzf release/nsh-guild-analytics-*.tar.gz
cd nsh-guild-analytics-*
sudo bash scripts/install.sh
```

卸载默认保留数据：

```bash
sudo bash scripts/uninstall.sh
```

确认删除安装目录和本项目 Docker 卷：

```bash
sudo REMOVE_DATA=true bash scripts/uninstall.sh
```

## 样例数据

可使用 `联赛初始数据/` 或 `设计文档/data/sample_battle.csv` 中的 CSV 进行导入预览。导入时选择本帮会后，系统会生成本场职业范围建议并完成分析入库。

## 静态验证说明

本轮按目标文件要求未执行：

- `go test`
- `npm install`
- `npm run build`
- `docker compose up`
- 本地服务启动

已进行源码和文件级静态检查：目录结构、OpenAPI 覆盖、Docker/Compose/env 脚本一致性、前后端入口、正式技术栈描述和敏感文件忽略规则。

## Gitee 交付规范

仓库包含 `.gitignore`、`.dockerignore`、`LICENSE`、`CONTRIBUTING.md`、`.gitee/ISSUE_TEMPLATE.zh-CN.md` 和 `.gitee/PULL_REQUEST_TEMPLATE.zh-CN.md`。提交前确认未提交 `.env`、数据库目录、上传文件、备份文件、依赖目录或构建产物。
