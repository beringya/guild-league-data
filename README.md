# 逆水寒帮会联赛数据分析平台

基于设计文档实现的 Docker Web 应用，包含 FastAPI + SQLite 后端、React + Vite 前端、CSV 导入分析、职业六维评分、单场/多场排名、帮会与分团对比、设置和备份能力。

当前交付按用户要求暂不在本机编译运行；代码、部署文件和测试样例已按设计文档组织完成。

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

## 样例数据

可使用 `联赛初始数据/` 或 `设计文档/data/sample_battle.csv` 中的 CSV 进行导入预览。导入时选择本帮会后，系统会生成职业范围版本并完成分析入库。
