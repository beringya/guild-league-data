# 一键安装脚本

脚本风格参考 `sub2api` 的安装体验，但本项目默认使用 Docker Compose 独立部署，不直接依赖宿主已有 PostgreSQL/Redis，避免影响服务器其他项目。

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

## PostgreSQL 可选模式

默认 SQLite 足够 MVP 使用，数据库文件保存在安装目录 `data/app.db`。如果希望使用独立 PostgreSQL 容器：

```bash
sudo USE_POSTGRES=true bash scripts/install.sh
```

该模式会启动当前项目独立的 PostgreSQL 容器和数据目录，不复用服务器其他项目的数据库。

## 卸载

```bash
sudo bash scripts/uninstall.sh
```

默认保留数据。确认删除数据：

```bash
sudo REMOVE_DATA=true bash scripts/uninstall.sh
```
