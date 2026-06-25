# 实现验收对照

本文件按设计文档 v1.5-final 对当前代码实现做交付对照。当前环境没有可用本地部署/编译运行环境，因此本轮验证以源码、配置和静态检查为准，未启动 Web 服务、未执行前端构建、未运行测试。

## 已实现

- 项目形态：`backend/` FastAPI + SQLite，`frontend/` React + Vite，`Dockerfile` 单容器多阶段构建，`docker-compose.yml` 暴露 8080。
- 首次管理员：`backend/app/core/bootstrap.py` 启动时创建 `admin`，生成随机密码并仅在首次初始化日志输出，`force_password_change=1`。
- 认证安全：`backend/app/api/routes/auth.py` 实现登录、退出、当前用户、改密、失败限流；`backend/app/api/deps.py` 实现 Cookie 会话和 CSRF 校验。
- 数据库模型：`backend/app/core/models.py` 覆盖用户、会话、设置、帮会、比赛、玩家、统计、评分规则、范围版本、头像和导入日志。
- CSV 导入：`backend/app/services/csv_import.py` 支持 UTF-8 BOM、UTF-8、GB18030、字段别名、重复表头、空数字警告、非法数字错误、两个帮会校验、未知职业阻止、文件名时间解析。
- 评分算法：`backend/app/services/analysis.py` 实现 KDA、参团率、伤害占比、职业范围建议、六维标准化、综合分、并列排序、团内 TOP3、双方对比、分团对比、规则化结论和多场聚合辅助函数。
- 默认职业规则：`backend/app/core/defaults.py` 实现十三个职业的六维槽位与已确认权重。
- 业务 API：`/api/battles/import/preview`、`/confirm`、历史、概览、排名、玩家详情、团内 TOP3、帮会对比、分团对比、多场总榜、设置、评分规则、范围版本、备份、头像兜底。
- 前端页面：登录、首页概览、个人排名/多场总榜、玩家详情、团内 TOP3、帮会对比、分团对比、导入、历史、设置。
- 视觉资源：正式前端复用 `设计文档/assets` 中的品牌、背景和图标资源，样式主题位于 `frontend/src/styles/app.css`。
- 测试样例：`backend/tests/` 覆盖样例 CSV 人数、重复表头、职业数量、文件名时间推断、KDA/参团率/综合分有限值和素问权重。

## 预留但已留接口

- 玩家/职业头像上传已支持 PNG/JPG/WebP 保存到 `data/avatars/`，默认稳定随机 SVG 头像仍作为兜底。
- 指定比赛使用当前启用评分规则与职业范围重新分析已接入历史页操作。
- 手动生成范围建议与发布冻结范围版本已接入设置页操作。

## 未执行的本地验证

- 未执行 `npm install` / `npm run build`。
- 未执行 `pytest`。
- 未启动 `uvicorn` 或 Docker Compose。
- 尝试 `python -m compileall backend` 时，当前沙箱无法访问 `python.exe`，因此未完成语法编译检查。

后续在具备部署环境后建议执行：

```bash
cp .env.example .env
docker compose build
docker compose up -d
docker compose logs app
```

并用 `设计文档/data/sample_battle.csv` 或 `联赛初始数据/` 中的 CSV 完成导入回归。
