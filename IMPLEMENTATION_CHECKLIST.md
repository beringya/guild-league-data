# 开发任务计划与验收对照

本清单按最新要求同步：项目技术路线参考 `sub2api`，目标实现为 Go/Gin/Ent + Vue/Vite/TailwindCSS + PostgreSQL/Redis + Docker Compose 独立部署。当前本地没有部署环境，因此本轮验证以源码、配置、文档和脚本静态检查为准。

## 任务状态

| 模块 | 状态 | 验收口径 |
|---|---|---|
| 技术栈调整 | 已完成 | 旧 FastAPI/React 实现归档到 `legacy/fastapi-react/`，新 `backend/` 与 `frontend/` 指向 Go/Vue。 |
| 后端工程 | 已完成 | Go 服务包含 Gin 路由、配置加载、PostgreSQL 连接、Redis 连接、启动引导和 API 分层。 |
| Ent 模型 | 已完成 | `ent/schema` 覆盖用户、会话、帮会、比赛、统计、评分规则、范围版本、头像与导入日志。 |
| 数据迁移 | 已完成 | SQL migration 可初始化 PostgreSQL 表、索引、约束和默认设置。 |
| 认证安全 | 已完成 | 管理员随机初始密码、强制改密、Redis 会话、CSRF、登出、改密撤销会话、登录限流。 |
| CSV 导入 | 已完成 | UTF-8 BOM/UTF-8/GB18030、重复表头、字段别名、两个帮会、十三职业、预览缓存、确认入库。 |
| 分析算法 | 已完成 | KDA、参团率、伤害占比、职业区间标准化、范围建议、六维综合分、并列排名、多场聚合。 |
| API 接口 | 已完成 | 覆盖设计文档 OpenAPI 轮廓中的登录、导入、比赛、排名、对比、设置、规则、头像、备份接口。 |
| 前端工程 | 已完成 | Vue 3 + Vite + TypeScript + TailwindCSS，复用设计资源，包含路由、状态、API 客户端和页面。 |
| 图表与页面 | 已完成 | 登录、概览、个人排名、玩家详情、团内 TOP3、对手对比、团队对比、导入、历史、设置。 |
| Docker 部署 | 已完成 | `docker-compose.yml` 默认启动 app/postgres/redis，独立网络和卷，端口默认 18080。 |
| 自动化安装 | 已完成 | `scripts/install.sh` 自动生成密钥、配置 `.env`、创建 systemd 服务、启动 Compose 并输出运维提示。 |
| 打包卸载 | 已完成 | `scripts/package.sh` 生成离线包，`scripts/uninstall.sh` 默认保留数据并支持显式清理。 |
| Gitee 规范 | 已完成 | README、LICENSE、CONTRIBUTING、Issue/PR 模板、忽略规则与交付说明完整。 |

## 后端验收项

- `GET /api/health` 返回应用、PostgreSQL、Redis 健康状态。
- `POST /api/auth/login` 成功后设置 HttpOnly Cookie；失败触发限流计数。
- `GET /api/auth/me` 返回当前管理员信息与是否需要改密。
- `POST /api/auth/change-password` 修改密码后撤销旧会话。
- `POST /api/battles/import/preview` 只解析和校验 CSV，不入库。
- `POST /api/battles/import/confirm` 按预览 token 入库并生成分析结果。
- `GET /api/battles`、`GET /api/battles/{id}`、`DELETE /api/battles/{id}` 管理历史比赛。
- `GET /api/battles/{id}/overview` 提供首页总览和双方关键指标。
- `GET /api/battles/{id}/rankings` 支持 side/career/team/search/page。
- `GET /api/battles/{id}/players/{stat_id}` 返回六维分析、同职业比较和评分解释。
- `GET /api/battles/{id}/team-top3` 按 `所在团长` 展示每团前三名个人。
- `GET /api/battles/{id}/guild-comparison` 返回双方总量、人均、职业人均和结论。
- `GET /api/battles/{id}/squad-comparison` 返回双方分团汇总与人均对比。
- `GET /api/rankings/history` 支持多场历史榜、日期、帮会、职业、最低场次、排序和分页。
- `GET/PUT /api/settings` 管理默认本帮会、阈值、备份、会话等设置。
- `GET/POST /api/scoring-rules` 和 `POST /api/scoring-rules/{version}/publish` 管理职业规则版本。
- `POST /api/scoring-rules/range-suggestions` 生成职业范围建议。
- `GET/POST /api/scoring-ranges` 查看和发布冻结范围版本。
- `PUT/DELETE /api/players/{id}/avatar` 与 `PUT/DELETE /api/careers/{career}/avatar` 管理头像。
- `POST /api/backups` 创建数据库备份或导出包。

## 前端验收项

- 登录页能处理首次管理员改密提示、错误提示和加载状态。
- 主界面固定侧栏包含概览、排名、团内 TOP3、对手对比、团队对比、个人分析、导入、历史、设置。
- 概览页展示最新比赛、双方 KPI、职业结构、优势不足和导入入口。
- 个人排名页支持单场/多场切换、帮会、职业、分团、搜索、分页和排序。
- 玩家详情页展示六维雷达、单项贡献、同职业本帮/对手平均、百分位和评分解释。
- 团内 TOP3 页按每个 `所在团长` 分组展示前三名，不误导为团队名次。
- 对手帮会对比页展示总量、人均、职业人均和结论。
- 团队数据对比页展示双方分团汇总和人均指标。
- 数据导入页包含文件选择、解析预览、帮会选择、校验结果和确认入库。
- 历史页可查看、删除、重新分析和进入历史总榜。
- 设置页可管理职业六维范围、规则版本、默认本帮会、头像、备份和管理员安全。

## 部署验收项

- `.env.example` 包含 `APP_PORT`、`GIN_MODE`、`DATABASE_DSN`、`REDIS_ADDR`、`POSTGRES_*`、`SESSION_SECRET` 等必需配置。
- `Dockerfile` 使用前端构建阶段、Go 构建阶段和最小运行镜像阶段。
- `docker-compose.yml` 使用固定项目内服务名 `app`、`postgres`、`redis`，并创建项目专属卷。
- 安装脚本可在离线包目录执行，不要求宿主机已有 PostgreSQL/Redis。
- systemd 服务只管理本项目 Compose，不影响其他项目容器。

## 未执行的本地验证

- 未执行 `go test ./...`。
- 未执行 `npm install` / `npm run build`。
- 未启动 `docker compose up -d --build`。
- 未进行浏览器端交互验收。

后续具备部署环境后建议执行：

```bash
cp .env.example .env
docker compose up -d --build
docker compose logs -f app
```

并使用 `设计文档/data/sample_battle.csv` 或 `联赛初始数据/` 中的 CSV 完成导入回归。
