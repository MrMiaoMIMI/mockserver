---
name: backend-conventions
description: Use this skill when update frontend code.
---

# Backend Skill

## 目标

在这个仓库里新增或修改后端能力时，优先遵守已有代码形态，而不是引入一套通用但不贴仓库实际的 Go 服务模板。

适用范围：

- `cmd/server/**`
- `internal/dao/**`
- `internal/service/**`
- `internal/model/**`
- `internal/config/**`

## 项目后端现状

当前后端是 Go HTTP 服务，真实调用链是：

`router -> controller -> view -> service -> dao`

启动和装配方式：

- `cmd/server/main.go` 负责加载 config、进入启动流程，`cmd/server` 包内文件负责对象装配和生命周期管理。
- 每个后端 package 可以用独立的 `module.go` 暴露 `Module = fx.Module(...)`。
- Fx 只应该出现在 package 的 `module.go` 或 `cmd/server` 生命周期装配文件中；service、dao、controller、view 等业务构造函数保持普通 Go 函数，方便单测直接构造。

核心技术与约定：

- 核心依赖公共库：github.com/MrMiaoMIMI/goshared >= v0.0.7
- Web 框架：Gin
- 依赖装配：`go.uber.org/fx`，各 package 暴露自己的 `Module`，`cmd/server` 只负责组合 package modules 和 HTTP lifecycle
- 数据访问：`github.com/MrMiaoMIMI/goshared/db/dbspi.Manager` + `dbhelper.NewSoftDeleteTableStore`
- db变更：不要 `AutoMigrate`，通过在 docs 目录维护 sql 来迭代db变更
- 日志：`github.com/MrMiaoMIMI/goshared/logger`
- 认证：JWT，`github.com/golang-jwt/jwt/v5` >= v5.3.0
- 链路追踪：需要结合日志，设置 trace_id
- 由于需要打log和控制接口链路，因此除了一些通用的函数，各个层级之间需要透传ctx。

## 分层职责

### 1. router 层

职责：

- 组织路由分组。
- 区分公开路由、鉴权路由、debug 路由。
- 绑定 controller 方法。

约定：

- 所有 API 挂在 `/mockserver/api/v1` 下。
- debug 路由使用可选登录中间件，不应混入正式业务接口。

不要做：

- 不在 router 写业务判断。
- 不在 router 拼装请求结构体。

### 2. controller 层

职责：

- 解析 path/query/body。
- 调用 view 层。
- 把错误映射为 HTTP 响应。

约定：

- 统一复用 `github.com/MrMiaoMIMI/goshared/util/serverresp`。
- `query` 用 `ShouldBindQuery`。
- `json body` 用 `ShouldBindJSON`。
- controller 尽量保持“薄”，只做入参解析和响应输出。

错误处理约定：

- 参考 `github.com/MrMiaoMIMI/goshared/util/serverresp` 和 `servererr`
- 参数错误：`400`
- 未找到：`404`
- 未授权：由中间件处理
- 其余业务失败：当前项目大多返回 `500`

注意：

- 现有 controller 的错误分类比较粗，新增代码应先保持一致性；如果要细化错误分层，应成批统一调整，不要只改单个接口风格。

### 3. view 层

职责：

- 连接 request/response 与 service/bo。
- 组装查询条件、分页、排序。
- 做跨 service 的编排和数据 enrichment。
- 负责 bo -> response 转换。

这是本项目里最容易被误解的一层。它不只是“展示层”，而是接口编排层。

约定：

- query 条件用 `fmo + dbhelper.Q(...)` 构造。
- 分页统一走 `request.PageRequest.ToPaginationConfig()`。
- 分页响应统一走 `ConvertToPaginatedResponse(...)`。
- 列表结果统一在 view 层转成 response。
- 跨模块补充字段，例如 `ProjectName` 这类 enrichment，优先在 view 层做。
- 创建/更新前的存在性检查、唯一性检查，优先在 view 层协调 service 完成。

不要做：

- 不把 HTTP 细节带进 service。
- 不直接操作 DAO。

### 4. service 层

职责：

- 承载核心业务逻辑。
- 维护 BO/DO 转换。
- 组织 DAO 调用。
- 在需要时读取上下文中的用户信息。

约定：

- 方法签名统一 `ctx context.Context` 放第一位。
- service 对外暴露 interface，构造函数返回 interface。
- 更新逻辑优先使用 `dbhelper.NewUpdater()` + `fmo` 字段。
- 查询逻辑优先使用 DAO 的 `FindWithoutDeleted / CountWithoutDeleted / ExistsById / GetById`。
- 审计字段由 service 写入，常见方式：
  - 创建：`bo.NewCommonBoForCreate(userEmail, userEmail)`
  - 更新：`updater.Add(fmo.Updater, userEmail)`
- 软删优先使用 `SoftDeleteById`。
- 错误信息保留调用语义，使用 `fmt.Errorf("...: %w", err)` 包装。

特别约定：

- 当前项目通过 `middleware.GetUserEmail(ctx)` 从上下文取操作者。
- 资源配置类字段通常需要在 service 内完成序列化/反序列化，不要把 JSON 细节泄漏到更高层。

### 5. dao 层

职责：

- 提供表级访问入口。
- 复用 `dbhelper.NewManager(...)` 创建 `dbspi.Manager`，再通过 `dbhelper.NewSoftDeleteTableStore(&do.Xxx{}, dbhelper.WithManager(manager))` 生成标准 CRUD 能力。
- `creator`、`updater`、`ctime`、`mtime` 通过 common-field autofill 自动维护；Gin middleware 把 operator 写入 `dbspi.WithOperator(ctx, email)`。

约定：

- DAO 保持薄封装，不承载业务规则。
- 每张表都有单独的 DAO interface。
- `dao.DB` 聚合所有表 DAO，供 service 注入使用。
- 新增模型时，通常需要同步补齐：
  - `model/do`
  - `model/fmo`
  - `dao/interfaces.go`
  - `dao/db.go`

## 模型约定

项目把模型拆成多层，新增字段或模块时要保持同步：

- `model/request`：接口请求结构
- `model/response`：接口响应结构
- `model/bo`：service/view 间业务对象
- `model/do`：数据库对象
- `model/fmo`：数据库字段管理对象
- `model/eo`：枚举、常量

具体规则：

- request 用于 HTTP 入参，不要复用 BO/DO 代替。
- response 用于 HTTP 出参，不要直接返回 BO/DO。
- DO 必须带清晰的 `gorm` tag 和 `TableName()`。
- FMO 必须覆盖该表需要参与查询/更新的字段。
- 可选更新字段优先用指针，便于区分“未传”和“零值”。

## API 与分页约定

统一响应格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

分页约定：

- request 复用 `PageRequest`
- response 复用 `PaginatedResponse[T]`
- 排序优先显式追加，例如按 `id desc`

注意：

- README 里的分页描述与当前代码不完全一致。当前代码实际默认值是：
  - `page_index=1`
  - `page_size=20`

## 日志、鉴权与上下文

日志约定：

- 不用 `fmt.Println`。
- 用 `github.com/MrMiaoMIMI/goshared/logger` 输出结构化日志。
- 尽量传递 `ctx`，让 trace_id 自动进入日志。

鉴权约定：

- 正式业务接口默认要求 JWT。
- 用户信息通过 middleware 注入上下文，不要手工解析 token。

trace 约定：

- `TraceMiddleware` 会生成或透传 `X-Trace-ID`。
- 需要打日志的 service/view 代码，应沿用请求上下文，不要新建空 context。

## 新增一个后端模块的建议步骤

1. 先定义 `model/do` 和 `TableName()`
2. 定义 `model/bo`、`model/request`、`model/response`
3. 定义 `model/fmo`
4. 扩展 `dao/interfaces.go`、`dao/impl.go`、`dao/db.go`
5. 扩展 `service/interfaces.go` 并实现具体 service
6. 实现 view，负责 request/response 编排
7. 实现 controller
8. 在 `router.go` 注册路由
9. 如果有 schema 变更，补 `doc/*.sql`

## 禁止事项

- 不直接在 controller 调 DAO。
- 不直接把 gin 的 `Context` 透传为领域模型的一部分。
- 不让 request/response 结构污染 service/dao 层签名。
- 不在多个层重复做同一份转换逻辑。
- 不新增一套平行的日志、错误响应或分页格式。

## 提交前自检

- 路由是否挂在正确的 auth group 下
- request/response/bo/do/fmo 是否成套补齐
- controller 是否只做解析和响应
- view 是否负责 query 组装与 response 转换
- service 是否负责业务规则与审计字段
- DAO 是否保持薄封装
- 分页是否复用 `PageRequest` 和 `PaginatedResponse`
- 日志是否走 `github.com/MrMiaoMIMI/goshared/logger`
- schema 变更是否补了 `doc/*.sql`
