# 一键安装脚本

脚本风格参考 `sub2api` 的安装体验。本项目默认使用 Docker Compose 独立部署，启动 `app`、`postgres`、`redis` 三个项目专属服务，不复用宿主机已有 PostgreSQL、Redis、网络、卷或容器。

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

安装脚本会自动生成 `.env`、`POSTGRES_PASSWORD`、`SESSION_SECRET` 和 `DATABASE_DSN`，启动 Docker Compose，并创建 `nsh-guild-analytics.service`。

## 卸载

```bash
sudo bash scripts/uninstall.sh
```

默认保留安装目录和 Docker 数据卷。确认删除数据：

```bash
sudo REMOVE_DATA=true bash scripts/uninstall.sh
```
