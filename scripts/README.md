# 一键安装脚本

脚本风格参考 `sub2api` 的安装体验。本项目默认使用 Docker Compose 独立部署，启动 `app`、`postgres`、`redis` 三个项目专属服务，不复用宿主机已有 PostgreSQL、Redis、网络、卷或容器。应用本体从 GHCR 拉取发布镜像，不在部署服务器上构建源码。

## 安装

```bash
tar -xzf nsh-guild-analytics-*.tar.gz
cd nsh-guild-analytics-*
sudo bash scripts/install.sh
```

默认访问端口为 `18080`，可通过环境变量修改：

```bash
sudo APP_PORT=19090 bash scripts/install.sh
```

安装脚本会自动生成 `.env`、`POSTGRES_PASSWORD`、`SESSION_SECRET` 和 `DATABASE_DSN`，拉取 `ghcr.io/beringya/guild-league-data:<version>` 镜像，启动 Docker Compose，并创建 `nsh-guild-analytics.service`。

## 更新

网页登录后可在左上角版本面板自动检查 GitHub Release。通过官方安装脚本部署时，点击“自动更新”会让服务端拉取新镜像并重启 `app` 服务。

服务器手动更新也可以执行：

```bash
cd /opt/nsh-guild-analytics
docker compose pull app
docker compose up -d --no-build app
```

## 卸载

```bash
sudo bash scripts/uninstall.sh
```

默认保留安装目录和 Docker 数据卷。确认删除数据：

```bash
sudo REMOVE_DATA=true bash scripts/uninstall.sh
```
