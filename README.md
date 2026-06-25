# 逆水寒帮会联赛数据分析平台

基于设计文档实现的 Docker Web 应用，包含 FastAPI + SQLite 后端、React + Vite 前端、CSV 导入分析、职业六维评分、单场/多场排名、帮会与分团对比、设置和备份能力。

当前交付按用户要求暂不在本机编译运行；代码、部署文件、测试样例和一键安装包脚本已按设计文档组织完成。

## 目录

- `backend/`：FastAPI、SQLite、认证、CSV 导入、评分分析和 API。
- `frontend/`：React 页面、图表、导入流程和粉色主题界面。
- `deployment/`：容器入口脚本。
- `data/`：运行时 SQLite 和上传文件目录。
- `backups/`：SQLite 在线备份目录。
- `设计文档/`：原始需求、设计图、样例数据和资源。

## Docker 部署

```bash
cp .env.example .env
docker compose up -d --build
docker compose logs app
```

首次启动会自动创建 `admin`，随机密码只会在首次初始化日志中显示一次，并要求首次登录后修改密码。

默认宿主机端口为 `18080`，不会占用常见的 `8080`。可在 `.env` 中修改 `APP_PORT`。

## 一键安装包

安装体验参考 `sub2api` 的脚本安装方式，但本项目使用独立 Docker Compose 部署，默认不会复用服务器已有数据库、网络或容器名称。

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

安装目录默认为 `/opt/nsh-guild-analytics`，会创建 systemd 服务 `nsh-guild-analytics.service`，并使用独立 compose project `nsh-guild-analytics`。

卸载：

```bash
sudo bash scripts/uninstall.sh
```

默认保留数据。若确认删除：

```bash
sudo REMOVE_DATA=true bash scripts/uninstall.sh
```

## 数据库选择

- 默认：SQLite，文件位于独立安装目录的 `data/app.db`，最适合当前单管理员、低并发 MVP，也最不影响服务器其他项目。
- 可选：独立 PostgreSQL 容器。使用 `USE_POSTGRES=true bash scripts/install.sh` 或 `docker compose -f docker-compose.yml -f docker-compose.postgres.yml up -d --build`。该模式不会复用服务器其他 PostgreSQL 实例。

## 样例数据

可使用 `联赛初始数据/` 或 `设计文档/data/sample_battle.csv` 中的 CSV 进行导入预览。导入时选择本帮会后，系统会生成职业范围版本并完成分析入库。

## Gitee 交付规范

仓库包含 `.gitignore`、`.dockerignore`、`LICENSE`、`CONTRIBUTING.md`、`.gitee/ISSUE_TEMPLATE.zh-CN.md` 和 `.gitee/PULL_REQUEST_TEMPLATE.zh-CN.md`。上传 Gitee 前请确认未提交 `.env`、数据库、上传文件或备份文件。
