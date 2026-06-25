# 逆水寒帮会联赛数据分析平台

一个可 Docker 独立部署的 Web 后台，用于导入和分析逆水寒帮会联赛 CSV。正式技术栈为 Go + Gin + Ent 模型、Vue 3 + TypeScript + Vite + TailwindCSS、PostgreSQL 15+、Redis 7+、ECharts 和 Docker Compose。

适合在一台 Linux 服务器上独立部署，默认通过 Docker Compose 启动应用、PostgreSQL 和 Redis。

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

## 快速安装

服务器需要已安装 Docker 和 Docker Compose v2。下载 GitHub Release 中的 `nsh-guild-analytics-*.tar.gz` 后执行。安装包只包含部署脚本和 Compose 文件，应用镜像会从 GHCR 拉取，不在服务器本地构建源码。

```bash
tar -xzf nsh-guild-analytics-*.tar.gz
cd nsh-guild-analytics-*
sudo bash scripts/install.sh
```

访问地址：

```text
http://服务器地址:18080
```

首次管理员账号为 `admin`。初始随机密码通过首次启动日志查看：

```bash
cd /opt/nsh-guild-analytics
docker compose logs app
```

## 本地开发构建

普通用户部署不要走源码构建。维护方或开发者需要本地调试镜像时，可以使用构建覆盖文件：

```bash
cp .env.example .env
docker compose -f docker-compose.yml -f docker-compose.build.yml up -d --build
docker compose logs app
```

## 发布流程

发布流程和 `sub2api` 类似：

- 维护方在源码仓库开发。
- 推送 `v*` tag 触发 GitHub Actions。
- Actions 先构建前端，再由 GoReleaser 构建 Linux amd64/arm64 后端二进制和多架构 Docker 镜像。
- 镜像推送到 `ghcr.io/beringya/guild-league-data`，同时生成 GitHub Release。
- Release 上传轻量安装包，部署服务器只拉镜像运行。

发布 GitHub 版本：

```bash
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions 会自动生成 Release，推送 `ghcr.io/beringya/guild-league-data:<version>` 和 `latest` 镜像，并上传 `.tar.gz` 安装包和 `.sha256` 校验文件。

## 卸载

默认保留数据：

```bash
sudo bash scripts/uninstall.sh
```

确认删除安装目录和本项目 Docker 卷：

```bash
sudo REMOVE_DATA=true bash scripts/uninstall.sh
```

## 版本更新提示

登录后，左上角版本徽标会调用 `GET /api/system/version` 检查更新。检测到远程版本高于当前 `APP_VERSION` 时，侧边栏会提示新版本。通过官方安装脚本部署时，页面会显示“自动更新”按钮：服务端会拉取新发布的镜像，更新 `.env` 中的 `APP_IMAGE` 和 `APP_VERSION`，然后用 Docker Compose 重建 `app` 服务。

推荐直接对接 GitHub Latest Release：

```env
APP_VERSION=1.0.0
APP_IMAGE=ghcr.io/beringya/guild-league-data:1.0.0
APP_IMAGE_REPOSITORY=ghcr.io/beringya/guild-league-data
UPDATE_GITHUB_REPO=beringya/guild-league-data
UPDATE_INSTALL_COMMAND=cd /opt/nsh-guild-analytics && docker compose pull app && docker compose up -d --no-build app
UPDATE_APPLY_ENABLED=true
UPDATE_APPLY_COMMAND=/app/bin/update-image.sh
UPDATE_CHECK_TIMEOUT=3s
```

也可以使用自定义 manifest，例如把 `UPDATE_CHECK_URL` 指向 GitHub raw 或自己的静态文件：

```json
{
  "version": "1.0.1",
  "channel": "stable",
  "release_url": "https://github.com/beringya/guild-league-data/releases/tag/v1.0.1",
  "download_url": "https://github.com/beringya/guild-league-data/releases/tag/v1.0.1",
  "checksum": "sha256:...",
  "notes": "更新说明"
}
```

如果同时配置了 `UPDATE_GITHUB_REPO` 和 `UPDATE_CHECK_URL`，系统优先读取 GitHub 最新 Release。

## 样例数据

可使用 `联赛初始数据/` 或 `设计文档/data/sample_battle.csv` 中的 CSV 进行导入预览。导入时选择本帮会后，系统会生成本场职业范围建议并完成分析入库。

## 交付规范

仓库包含 `.gitignore`、`.dockerignore`、`LICENSE`、`CONTRIBUTING.md` 和 GitHub Issue/PR 模板。提交前确认未提交 `.env`、数据库目录、上传文件、备份文件、依赖目录或构建产物。
