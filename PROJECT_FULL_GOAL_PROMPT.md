# 逆水寒帮会联赛数据分析平台：完整项目目标提示词

请在当前工作区 `F:\工作\帮会联赛数据` 中完整开发“逆水寒帮会联赛数据分析平台”。这是一个从设计文档到源码、部署、打包、Gitee 提交的一体化项目目标，不是只完成下一步任务。

## 1. 必须先阅读的项目资料

开始开发前，必须完整查看并理解以下资料：

- `F:\工作\帮会联赛数据\设计文档\`
- `F:\工作\帮会联赛数据\设计文档\逆水寒帮会联赛数据分析平台_开发说明_v4.md`
- `F:\工作\帮会联赛数据\设计文档\README.md`
- `F:\工作\帮会联赛数据\设计文档\docs\`
- `F:\工作\帮会联赛数据\设计文档\api\openapi-outline.yaml`
- `F:\工作\帮会联赛数据\设计文档\database\schema.sql`
- `F:\工作\帮会联赛数据\设计文档\config\`
- `F:\工作\帮会联赛数据\设计文档\design\`
- `F:\工作\帮会联赛数据\设计文档\assets\`
- `F:\工作\帮会联赛数据\联赛初始数据\`
- `F:\工作\帮会联赛数据\DEVELOPMENT_PLAN.md`
- `F:\工作\帮会联赛数据\IMPLEMENTATION_CHECKLIST.md`

设计文档是权威业务来源。如果源码、计划、README 与设计文档冲突，以设计文档和用户最新要求为准，并同步修正文档与代码。

## 2. 完整项目目标

将当前项目完整开发为一个可 Docker 独立部署、可一键安装、可自动创建管理员账号密码、可导入和分析逆水寒帮会联赛 CSV 的 Web 平台。

最终交付必须包含：

- 完整后端源码
- 完整前端源码
- 数据库模型与迁移
- CSV 导入、清洗、预览、确认入库
- 职业评分、六维分析、排名、对比、多场历史总榜
- 登录、安全、管理员初始化与强制改密
- 设置、评分规则、范围版本、头像、备份功能
- Dockerfile
- Docker Compose 独立部署配置
- `.env.example`
- 一键安装脚本
- 卸载脚本
- 打包脚本
- README 与开发说明
- Gitee 常用规范文件
- 最终自动提交并推送到已配置的 Gitee remote

当前本地没有部署环境，暂时不要编译运行，不要启动 Docker，不要安装依赖。开发完成后做静态文件级检查，并明确说明未执行编译运行。

## 3. 技术栈要求

参考 `https://github.com/Wei-Shaw/sub2api` 的技术栈和安装体验实现本项目。

正式技术栈：

- 后端：Go + Gin + Ent
- 前端：Vue 3 + TypeScript + Vite + TailwindCSS
- 数据库：PostgreSQL 15+
- 缓存/会话/限流/导入预览缓存：Redis 7+
- 图表：ECharts
- 部署：Docker Compose
- 安装：类似 sub2api 的自动化安装脚本

不要把 FastAPI + React + SQLite 作为正式实现目标。旧实现如存在，只能放在 `legacy/` 中作为参考。

## 4. Docker 独立部署要求

一键安装必须默认使用 Docker Compose 独立部署，不能影响服务器其他项目。

必须满足：

- 独立 compose project
- 独立 Docker network
- 独立 Docker volume
- 独立容器名或服务名
- 默认宿主机端口不要使用常见 `8080`，建议 `18080`
- PostgreSQL 和 Redis 默认由本项目 Compose 独立启动
- 不复用宿主机已有 PostgreSQL、Redis、网络、卷或容器
- `.env` 中可修改端口、数据库密码、会话密钥、安装路径等配置
- 卸载脚本默认保留数据，只有显式 `REMOVE_DATA=true` 才清理数据

## 5. 自动化安装与打包要求

提供类似 sub2api 的一键安装体验。

必须实现：

- `scripts/install.sh`
  - 检查 Docker 和 Docker Compose
  - 支持离线解包后执行
  - 自动创建安装目录
  - 自动生成 `.env`
  - 自动生成 `POSTGRES_PASSWORD`
  - 自动生成 `SESSION_SECRET`
  - 启动 Docker Compose
  - 创建 systemd 服务
  - 输出访问地址、管理员账号、查看初始密码日志的命令

- `scripts/uninstall.sh`
  - 停止本项目 systemd 服务
  - 停止本项目 Docker Compose
  - 默认保留数据
  - 显式确认后才删除安装目录和数据卷

- `scripts/package.sh`
  - 生成可上传服务器并一键安装的发布包
  - 不打包 `.env`、数据库文件、上传文件、备份文件、node_modules、编译产物、临时日志

## 6. 账号与安全要求

- 首次启动自动创建管理员账号 `admin`
- 初始密码随机生成，只在首次初始化日志中显示一次
- 首次登录后强制修改密码
- 密码必须使用强哈希保存
- 使用 HttpOnly SameSite Cookie 会话
- 会话存储或会话索引使用 Redis
- 写操作需要 CSRF 防护
- 登录失败需要限流
- 修改密码后撤销其他有效会话
- 除健康检查和登录接口外，业务 API 默认需要认证

## 7. CSV 导入要求

CSV 字段包括：

- 帮会名
- 玩家
- 等级
- 职业
- 所在团长
- 击败
- 助攻
- 战备资源
- 对玩家伤害
- 对建筑伤害
- 治疗值
- 承受伤害
- 重伤
- 青灯焚骨
- 化羽
- 控制

导入必须支持：

- UTF-8 BOM
- UTF-8
- GB18030
- 重复表头清理
- 空行清理
- 数字逗号清理
- 字段别名
- 非法数字错误提示
- 必填字段校验
- 两个帮会校验
- 十三个已确认职业校验
- 未配置职业阻止导入并提示
- 文件 SHA-256 去重
- 从文件名推断比赛时间
- 预览阶段不入库
- 确认阶段由管理员选择本帮会，另一方自动为对手

## 8. 核心分析规则

- KDA = `(击败 + 助攻) / max(重伤, 1)`
- 参团率 = `(击败 + 助攻) / 所在帮会总击败`
- 玩家伤害占比 = `个人对玩家伤害 / 帮会总对玩家伤害`
- 建筑伤害占比 = `个人对建筑伤害 / 帮会总对建筑伤害`
- 伤害占比和职业区间转化率必须分开展示
- 每个职业固定 6 个分析维度
- 综合分只使用该职业启用且权重大于 0 的维度
- 维度分 = 按职业范围标准化到 0-100
- 综合分 = 各启用维度分乘权重后求和
- 支持并列排序和稳定排序
- 支持个人评分解释 JSON
- 支持同职业本帮平均、同职业对手平均、同职业百分位

职业范围建议：

- 首次无正式范围时，按当前导入文件中双方同职业样本生成建议
- `n >= 20` 使用 P5/P95
- `3 <= n < 20` 使用 min/max 加安全余量
- `n < 3` 给出高风险提示
- 建议必须由管理员确认发布
- 发布后冻结，不随新比赛自动漂移
- 历史比赛只有显式重新分析才更新派生得分

## 9. 十三职业评分权重

- 素问：治疗值 55%、承受伤害 25%、化羽 20%
- 铁衣：控制 60%、承受伤害 40%
- 神相：对玩家伤害 50%、对建筑伤害 50%
- 血河：对玩家伤害 50%、对建筑伤害 50%
- 沧澜：对玩家伤害 50%、对建筑伤害 50%
- 玄机：对玩家伤害 50%、对建筑伤害 50%
- 云瑶：对玩家伤害 50%、对建筑伤害 50%
- 碎梦：击败 50%、对玩家伤害 30%、对建筑伤害 20%
- 九灵：青灯焚骨 60%、对玩家伤害 20%、对建筑伤害 20%
- 鸿音：控制 55%、治疗值 45%
- 潮光：对玩家伤害 50%、对建筑伤害 50%
- 荒羽：对玩家伤害 50%、对建筑伤害 50%
- 龙吟：对玩家伤害 50%、对建筑伤害 50%

## 10. 页面与功能要求

前端必须是实际可用后台应用，不要做营销落地页。

必须实现页面：

- 登录页
- 首页概览
- 个人排名
- 多场历史总榜
- 玩家详情 / 个人六维分析
- 团内 TOP3
- 对手帮会对比
- 团队数据对比
- 数据导入
- 历史记录
- 设置

页面要求：

- 复用设计文档中的品牌、背景、图标、头像等资源
- 使用粉色、可爱、柔和、清晰的后台界面风格
- 使用固定侧栏和顶部操作栏
- 使用真实 API 数据流
- 空数据时提供导入入口
- 使用 ECharts 展示柱状图、雷达图、趋势图和结构图
- “团队”指 CSV 字段 `所在团长`
- “团内 TOP3”是每个分团内部前三名个人，不是给团队排名

## 11. API 范围

至少覆盖设计文档 OpenAPI 轮廓中的接口：

- `GET /api/health`
- `POST /api/auth/login`
- `POST /api/auth/logout`
- `GET /api/auth/me`
- `POST /api/auth/change-password`
- `POST /api/battles/import/preview`
- `POST /api/battles/import/confirm`
- `GET /api/battles`
- `GET /api/battles/{id}`
- `DELETE /api/battles/{id}`
- `POST /api/battles/{id}/reanalyze`
- `GET /api/battles/{id}/overview`
- `GET /api/battles/{id}/rankings`
- `GET /api/battles/{id}/players/{stat_id}`
- `GET /api/battles/{id}/team-top3`
- `GET /api/battles/{id}/guild-comparison`
- `GET /api/battles/{id}/squad-comparison`
- `GET /api/rankings/history`
- `GET /api/settings`
- `PUT /api/settings`
- `GET /api/scoring-rules`
- `POST /api/scoring-rules`
- `POST /api/scoring-rules/validate`
- `POST /api/scoring-rules/range-suggestions`
- `POST /api/scoring-rules/{version}/publish`
- `GET /api/scoring-ranges`
- `POST /api/scoring-ranges`
- `PUT /api/players/{player_id}/avatar`
- `DELETE /api/players/{player_id}/avatar`
- `PUT /api/careers/{career}/avatar`
- `DELETE /api/careers/{career}/avatar`
- `POST /api/backups`

## 12. Gitee 规范与提交要求

需要参考 Gitee 常见仓库规范，补齐并维护：

- `.gitignore`
- `.dockerignore`
- `LICENSE`
- `CONTRIBUTING.md`
- `.gitee/ISSUE_TEMPLATE.zh-CN.md`
- `.gitee/PULL_REQUEST_TEMPLATE.zh-CN.md`
- README
- 开发计划
- 任务清单
- 部署说明

开发任务完成后，自动执行：

1. `git status`
2. 合理暂存本次任务文件
3. 提交代码
4. 推送到已配置的 Gitee remote

建议提交信息：

```bash
feat: implement guild analytics platform
```

如果本机没有 git、remote 未配置、凭证失败或推送失败，需要说明原因，并给出用户后续可执行的命令。

## 13. 验证要求

由于当前本地没有部署环境：

- 不要执行 `go test`
- 不要执行 `npm install`
- 不要执行 `npm run build`
- 不要执行 `docker compose up`
- 不要启动本地服务

但必须执行静态检查：

- 检查文件结构是否完整
- 检查 README、开发计划、设计说明、任务清单是否一致
- 检查不再把 FastAPI/React/SQLite 描述为正式目标
- 检查 Dockerfile、docker-compose、`.env.example`、安装脚本是否一致
- 检查后端路由、前端页面、部署脚本和打包脚本是否覆盖目标
- 检查没有误提交 `.env`、数据库、上传文件、备份文件、依赖目录或编译产物

最终回复必须说明：

- 完成了哪些模块
- 哪些文件是关键入口
- 因无本地部署环境未执行哪些命令
- 如何一键安装
- 是否已提交并推送到 Gitee

## 14. 工作方式要求

- 直接实施，不要只给方案
- 不要删除用户数据、设计文档或样例 CSV
- 不要回退用户已有改动
- 保持旧实现可追溯，可归档到 `legacy/`
- 文档、代码、脚本要同步更新
- 最终目标是完整可交付项目，而不是只完成脚手架
