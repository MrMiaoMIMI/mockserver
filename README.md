# MockServer

一个基于 Go 的通用 mockserver 骨架，当前实现的是 V1 最小闭环：

- 统一 `Event` 输入模型
- 代码注册 `ProtocolSpec`，当前包含 HTTP、cache 和 SPEX
- `RuleSet -> Rule -> Action` 规则结构
- `draft / published` 双态
- `selector` 粗筛 + 条件树精匹配
- `static_response` / `template_response` / `cel_response`
- HTTP 管理接口
- admin token 鉴权和 read / write / publish 权限分层
- publish / rollback snapshot 审计 metadata
- DB schema / DAO repository 第一阶段骨架
- HTTP runtime adapter
- SDK decision endpoint 和 cache tracer bullet
- 唯一 `X-Trace-ID`
- runtime 命中日志和内存 metrics
- SDK decision traffic 落库和前端 Traffic Inspector

面向使用者的快速上手说明见：

- [用户使用说明书](docs/user_guide.md)
- [Mock SDK Quickstart](docs/mocksdk.md)

当前版本先把架构闭环做通，已经接入部分 V2 能力。仍未包含：

- 任意 Go 脚本执行
- 用户体系、登录、JWT、完整 RBAC 和租户隔离
- 独立审计表、不可变审计日志和审批流

## 目录结构

```text
cmd/server                     启动入口、依赖装配和 HTTP 生命周期
internal/config                基础配置
internal/adapter/httpadapter   HTTP 请求规范化
internal/controller            管理面 / 运行面接口
internal/engine                规则校验、编译、匹配、动作执行
internal/router                路由注册
internal/service               规则服务
internal/dao                   MySQL 数据访问与规则仓储
internal/view                  请求/响应编排
internal/model/bo              核心领域对象
internal/model/request         管理接口请求模型
internal/model/response        管理接口响应模型
mockprotocol                   协议字段与 selector 规格
docs/user_guide.md             用户使用说明书
docs/mocksdk.md                外部 Mock SDK 快速接入说明
examples/ruleset.json          示例规则
examples/mockserver.postman_collection.json  Postman 调试集合
```

## 启动

服务默认读取 `etc/server.yml`。其中包含监听地址、启动规则源、MySQL DSN、连接池和 admin token 配置。

```bash
go run ./cmd/server
```

也可以显式指定配置文件：

```bash
MOCKSERVER_CONFIG_FILE=etc/server.yml go run ./cmd/server
```

配置文件示例：

```yaml
server:
  address: ":8080"

bootstrap:
  ruleset_file: ""

db:
  driver: mysql
  dsn: "root:123456@tcp(127.0.0.1:3306)/mockserver_db?charset=utf8mb4&parseTime=True&loc=Local"
  init_schema: true
  max_open_conns: 20
  max_idle_conns: 5
  conn_max_lifetime_seconds: 3600
  debug: false
```

环境变量仍然可用于部署时覆盖配置，例如：

```bash
MOCKSERVER_ADDR=:18080 \
MOCKSERVER_DB_DEBUG=true \
go run ./cmd/server
```

后端对象装配使用 `go.uber.org/fx v1.24.0`，各后端 package 暴露自己的 `Module`，由 `cmd/server` 入口统一组合；DB 访问层使用 `github.com/MrMiaoMIMI/goshared v0.0.10` 的 `dbspi.Manager` + `dbhelper.NewSoftDeleteTableStore`，并通过 common-field autofill 维护 `creator`、`updater`、`ctime`、`mtime`。当前 Go 基线为 `1.25.9`。

运行时只支持 MySQL 存储，已移除其他存储分支。

默认会执行内置 schema 初始化；如需手动管理迁移，可以在 `etc/server.yml` 里设置：

```yaml
db:
  init_schema: false
```

如果希望启动时自动加载并发布规则，可以在 `bootstrap.ruleset_file` 里配置单文件、目录，或逗号分隔的多个路径：

```yaml
bootstrap:
  ruleset_file: ./examples/ruleset.json
```

```yaml
bootstrap:
  ruleset_file: ./examples
```

```yaml
bootstrap:
  ruleset_file: ./examples/ruleset.json,./examples/ruleset-cel.json
```

默认 admin 接口不鉴权，便于本地开发。一旦配置任意 admin token，所有 `/mockserver/api/v1/admin/*` 接口都会启用 token 鉴权；runtime mock 接口不受 admin token 影响。

```yaml
admin:
  token: admin-secret
```

也可以拆分权限 token：

```yaml
admin:
  read_token: read-secret
  write_token: write-secret
  publish_token: publish-secret
```

- `admin.token`：超级 token，具备 read / write / publish 全部权限
- `admin.read_token`：只读权限，可访问查询、validate、simulate、rollback preview、metrics
- `admin.write_token`：读写 draft 权限，可创建/更新 ruleset 和 rule，但不能 publish / rollback
- `admin.publish_token`：发布权限，可 publish / rollback，也可读

请求时使用 `Authorization: Bearer <token>` 或 `X-Mockserver-Admin-Token: <token>`。

发布和回滚会在生成的新 snapshot 中记录 `audit` metadata。操作者来自 `X-Mockserver-Operator` 或 `X-Operator`，trace 来自 `X-Trace-ID`，原因来自请求 body 的 `reason` 字段；如果未传操作者，会记录为 `anonymous`。

## 当前规则模型

核心对象：

- `RuleSet`
规则集合，带协议、命名空间、选择器和规则列表。

- `Rule`
单条规则，包含优先级、条件和动作。

- `Condition`
条件树，当前支持 `all / any / not / predicate / expr`。

- `Action`
当前支持 `static_response`、`template_response`、`cel_response`、`sequence_response` 和 `webhook_response`。

协议字段由代码注册的 `ProtocolSpec` 提供。用户配置 ruleset/rule/namespace，不在数据库中动态配置协议字段。第一版公开 `fields` 和 `selectors`，不公开 matcher 的内部索引提示。

前端会基于 `ProtocolSpec` 辅助配置 selector 和 rule condition：协议字段和 operator 通过下拉选择，动态字段会提供 key/path 输入，`exists`/`is_null` 等 operator 不需要 value，`in`/`not_in` 使用列表输入，数字/布尔/JSON/regex 会切换到对应的输入组件。

## 当前支持的字段路径

通用字段：

- `protocol`
- `namespace`
- `meta.trace_id`
- `meta.source`
- `meta.extra.<key>`

HTTP 字段：

- `request.method`
- `request.scheme`
- `request.host`
- `request.original_host`
- `request.path`
- `request.client_ip`
- `request.raw_body`
- `request.headers.<key>`
- `request.headers.<key>[*]`
- `request.headers.<key>[0]`
- `request.query.<key>`
- `request.query.<key>[*]`
- `request.query.<key>[0]`
- `request.body.xxx`
- `request.body.items[0].id`
- `request.body.items[*].id`

HTTP body 解析规则：

- body 可解析为 JSON 时，同时设置 `request.body` 和 `request.raw_body`
- body 不能解析为 JSON 时，只设置 `request.raw_body`
- body 为空时，不设置 `request.body` 和 `request.raw_body`

Cache 字段：

- `request.operation`
- `request.key`
- `request.ttl_ms`
- `request.value`
- `request.value.<key>`

SPEX 字段：

- `request.cmd`
- `request.req`
- `request.req.<key>`
- `request.param`

## 当前支持的操作符

- `eq`
- `ne`
- `in`
- `not_in`
- `contains`
- `not_contains`
- `exists`
- `not_exists`
- `is_null`
- `is_not_null`
- `prefix`
- `suffix`
- `regex`
- `gt`
- `gte`
- `lt`
- `lte`

`exists` 表示路径存在，即使值为 null。`not_exists` 表示路径不存在。`is_null` 表示路径不存在，或路径存在但值是 JSON null / Go nil。`is_not_null` 表示路径存在且值不是 null。

## 当前校验能力

`validate`、`publish`、启动加载和持久化恢复都会走同一套规则校验。当前会拒绝：

- 空 ruleset 或重复 rule id
- 非 `100-599` 的 HTTP status
- 非法 response header 名称，或包含 CR/LF 的 header value
- 重复 selector 条目、非 `/` 开头的 path selector
- 非法 regex，或超过 512 字符的 regex pattern
- 非法 CEL 表达式、非法字段路径、未知操作符、协议字段不支持的操作符

## 当前 CEL 能力

条件节点可以直接写：

```json
{
  "expr": "request.headers[\"x-env\"][0] == \"test\" && request.body.score >= 90"
}
```

响应动作可以直接写：

```json
{
  "type": "cel_response",
  "status": 200,
  "body_expression": "{\"message\": request.path, \"score\": request.body.score}"
}
```

当前可用变量：

- `event`
- `request`
- `meta`
- `protocol`
- `namespace`

## 当前 Template Helper

`template_response` 现在支持一组更稳定的 helper，不需要再手写很长的 `index (index ...)`：

- `field . "request.path"`
- `query . "q1"`
- `queryAll . "q1"`
- `header . "x-env"`
- `headerAll . "x-env"`
- `body . "user.id"`
- `first ...`
- `toJSON ...`

示例：

```json
{
  "type": "template_response",
  "status": 200,
  "body_template": "{\"message\":\"hello {{ query . \\\"q1\\\" }}\",\"path\":\"{{ field . \\\"request.path\\\" }}\",\"user\":{{ toJSON (body . \\\"user\\\") }}}"
}
```

## 管理接口

统一前缀：

```text
/mockserver/api/v1/admin
```

如果启动时配置了 admin token，下面所有接口都需要携带 `Authorization: Bearer <token>` 或 `X-Mockserver-Admin-Token: <token>`。

接口列表：

- `GET /mockserver/api/v1/admin/protocols`
查询已注册协议的字段和 selector 规格

- `POST /mockserver/api/v1/admin/rulesets`
创建或更新 draft 规则集

- `GET /mockserver/api/v1/admin/rulesets`
查询 draft 规则集列表

- `GET /mockserver/api/v1/admin/rulesets/{id}`
查询单个 draft 规则集

- `POST /mockserver/api/v1/admin/rulesets/{id}/validate`
校验 draft 规则集

- `POST /mockserver/api/v1/admin/rulesets/{id}/publish`
发布 draft 规则集到 runtime

- `POST /mockserver/api/v1/admin/rulesets/{id}/simulate`
拿一份输入 event 做命中模拟

- `POST /mockserver/api/v1/admin/rulesets/{id}/rules`
向某个 draft ruleset 新增一条 rule

- `PUT /mockserver/api/v1/admin/rulesets/{id}/rules/{rule_id}`
更新某个 draft rule

- `DELETE /mockserver/api/v1/admin/rulesets/{id}/rules/{rule_id}`
删除某个 draft rule

- `POST /mockserver/api/v1/admin/rulesets/{id}/rules/{rule_id}/enable`
启用某个 draft rule

- `POST /mockserver/api/v1/admin/rulesets/{id}/rules/{rule_id}/disable`
禁用某个 draft rule

- `POST /mockserver/api/v1/admin/rulesets/{id}/rules/{rule_id}/priority`
调整某个 draft rule 的优先级

- `POST /mockserver/api/v1/admin/published/simulate`
用当前 published 快照做一次全局命中模拟，返回结果与 runtime 匹配链路一致

- `GET /mockserver/api/v1/admin/published/rulesets`
查询当前正在 runtime 生效的 published 规则集

- `GET /mockserver/api/v1/admin/published/rulesets/{id}`
查询某个规则集当前生效的 published 快照

- `GET /mockserver/api/v1/admin/published/rulesets/{id}/snapshots`
查询某个规则集的历史发布快照

- `POST /mockserver/api/v1/admin/published/rulesets/{id}/rollback`
回滚当前 runtime 到某个历史 snapshot

- `POST /mockserver/api/v1/admin/published/rulesets/{id}/rollback/preview`
预演某个历史 snapshot 是否能正常编译，并可选带 event 做一次命中模拟

- `GET /mockserver/api/v1/admin/metrics/runtime`
查询 runtime 请求数、命中数、错误数、耗时和 ruleset/rule 维度命中统计

- `GET /mockserver/api/v1/admin/traffic/events`
查询 SDK decision traffic 事件，支持按时间、协议、namespace、outcome、ruleset/rule、trace 和协议索引字段过滤

## Runtime 接口

统一前缀：

```text
/mockserver/runtime/{namespace}/http/{actual_path}
```

例如：

```text
GET /mockserver/runtime/default/http/api/v1/debug?q1=qv1
```

注意：

- `runtime` 路径里的 `default` 是 namespace。
- `http` 表示当前 adapter 类型。
- 真正参与规则匹配的 path 是 `/api/v1/debug`，不是整条 runtime path。

如果一个请求同时命中多个 published ruleset，MockServer 会先按 selector 严格程度选出一个 ruleset，再只匹配该 ruleset 内部的 rules。当前排序规则是：

- `selector.all` 只允许使用当前协议在 `ProtocolSpec.selectors` 中注册的字段
- selector 条件全部命中后才会进入该 ruleset 的 rule 匹配
- 严格度按 selector 条件累加，`eq` 高于 `prefix/suffix`，再高于 `contains/exists`
- 严格度相同时按 `ruleset.id` 倒序兜底，保证结果稳定

## 快速体验

### 1. 启动服务

```bash
go run ./cmd/server
```

或者在 `etc/server.yml` 中配置规则源后启动：

```yaml
bootstrap:
  ruleset_file: ./examples/ruleset.json
```

### 2. 导入示例规则

如果你上一条已经用 `bootstrap.ruleset_file` 自动加载了规则，这一步可以跳过。

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets \
  -H 'Content-Type: application/json' \
  --data-binary @examples/ruleset.json
```

### 3. 校验规则

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets/http-default/validate
```

### 4. 发布规则

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets/http-default/publish \
  -H 'Content-Type: application/json' \
  -H 'X-Mockserver-Operator: admin@example.com' \
  -d '{
    "reason": "release debug api mock"
  }'
```

### 5. 查看 draft 列表

```bash
curl http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets
```

### 6. 管理单条 draft rule

新增一条 rule：

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets/http-default/rules \
  -H 'Content-Type: application/json' \
  -d '{
    "rule": {
      "id": "new-debug-api",
      "enabled": true,
      "priority": 200,
      "when": {
        "all": [
          {"field": "request.method", "op": "eq", "value": "GET"},
          {"field": "request.path", "op": "eq", "value": "/api/v1/new-debug"}
        ]
      },
      "action": {
        "type": "static_response",
        "status": 200,
        "body": {"message": "new rule"}
      }
    }
  }'
```

调整优先级、禁用、启用或删除：

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets/http-default/rules/new-debug-api/priority \
  -H 'Content-Type: application/json' \
  -d '{"priority":300}'

curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets/http-default/rules/new-debug-api/disable
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets/http-default/rules/new-debug-api/enable
curl -X DELETE http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets/http-default/rules/new-debug-api
```

这些接口只修改 draft，不会影响 runtime；仍然需要 `publish` 后才会生效。

### 7. 查看当前 published 列表

```bash
curl http://127.0.0.1:8080/mockserver/api/v1/admin/published/rulesets
```

### 8. 全局模拟当前 published 命中

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/published/simulate \
  -H 'Content-Type: application/json' \
  -d '{
    "explain_summary": true,
    "event": {
      "protocol": "http",
      "namespace": "default",
      "request": {
        "method": "GET",
        "host": "demo.com",
        "path": "/api/v1/debug"
      }
    }
  }'
```

### 9. 查看某个规则集的发布快照

```bash
curl http://127.0.0.1:8080/mockserver/api/v1/admin/published/rulesets/http-default
curl http://127.0.0.1:8080/mockserver/api/v1/admin/published/rulesets/http-default/snapshots
```

### 10. 回滚到某个 snapshot

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/published/rulesets/http-default/rollback \
  -H 'Content-Type: application/json' \
  -H 'X-Mockserver-Operator: admin@example.com' \
  -d '{
    "snapshot_id": "snap_9f2a4c7b1d3e5f60",
    "reason": "restore previous stable mock"
  }'
```

### 11. 预演某个 snapshot 的回滚效果

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/published/rulesets/http-default/rollback/preview \
  -H 'Content-Type: application/json' \
  -d '{
    "snapshot_id": "snap_9f2a4c7b1d3e5f60",
    "explain_compact": true,
    "explain_max_depth": 1,
    "event": {
      "protocol": "http",
      "namespace": "default",
      "request": {
        "method": "GET",
        "host": "demo.com",
        "path": "/api/v1/debug"
      }
    }
  }'
```

`rollback/preview` 现在除了 `valid/validation/simulation`，还会返回一份 `diff` 摘要：

- 当前 snapshot 和目标 snapshot 的 ID
- ruleset 级别是否有变化
- `ruleset_field_diffs`：ruleset 级字段路径，例如 `ruleset.selector.all[0].value`、`ruleset.namespace`、`ruleset.protocol`
- 哪些 rule 被新增、删除、修改
- 某条 rule 的 `condition` 或 `action` 是否变化
- `field_diffs`：具体变化字段路径，例如 `action.body.version`、`action.status`、`when.all[0].field`

这可以直接帮助判断“回滚后真正变的是哪条规则、是不是 action 变了”。

### 12. 模拟单个 draft ruleset 命中

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets/http-default/simulate \
  -H 'Content-Type: application/json' \
  -d '{
    "explain_only": true,
    "explain_compact": true,
    "explain_summary": true,
    "explain_max_depth": 1,
    "event": {
      "protocol": "http",
      "namespace": "default",
      "request": {
        "method": "GET",
        "host": "demo.com",
        "path": "/api/v1/debug",
        "query": {
          "q1": ["qv1"]
        }
      }
    }
  }'
```

`simulate` 的返回结果现在除了 `matched/trace/response`，还会带一份 `explain`：

- `candidate_rules`
- `rule_explanations`
- `rule_set_explanations`
- ruleset selector 的逐项检查结果 `selector_checks`
- 每个条件节点的 `matched/message`
- `expr` 条件的 CEL 求值结果
- 命中规则的 `action_info`

如果传 `explain_only=true`：

- 命中链仍然会完整计算
- 但不会真正执行 `template_response` 或 `cel_response`
- `action_info.message` 会明确标记为跳过执行

如果再配合：

- `explain_compact=true`
  会去掉 `expected`、`actual`、`rendered_result` 这类更大的 explain 字段

- `explain_summary=true`
  会只保留最终命中的 ruleset/rule 和失败原因概览，不再返回 `selector_checks`、`candidate_rules`、深层 `children`

- `explain_max_depth=1`
  会裁剪过深的嵌套条件 explain，只保留前几层

这对排查“为什么没命中”“为什么命中了错误规则”“是 ruleset selector 先被筛掉，还是 rule 条件没过”“模板/CEL 最终渲染了什么”很有帮助，同时能避免重模板或复杂 CEL 在排障时真正执行。

### 13. 调用 runtime

`examples/ruleset.json` 里有一条 selector：

- `request.host eq "demo.com"`
- `request.path prefix "/api/"`

所以 runtime 调用时要带 `Host: demo.com`：

```bash
curl 'http://127.0.0.1:8080/mockserver/runtime/default/http/api/v1/debug?q1=qv1' \
  -H 'Host: demo.com'
```

预期返回：

```json
{"message":"hello qv1","path":"/api/v1/debug"}
```

并且响应 header 里会带：

- `X-Mockserver-Ruleset`
- `X-Mockserver-Rule`
- `X-Trace-ID`

如果请求没有带 `X-Trace-ID`，服务会自动生成一个唯一 trace id。HTTP runtime 请求还会输出一条 JSON 格式命中日志，并写入内存 metrics。

查看 runtime metrics：

```bash
curl http://127.0.0.1:8080/mockserver/api/v1/admin/metrics/runtime
```

查看 SDK decision traffic：

```bash
curl 'http://127.0.0.1:8080/mockserver/api/v1/admin/traffic/events?limit=50&offset=0'
curl 'http://127.0.0.1:8080/mockserver/api/v1/admin/traffic/events/35'
```

## 一个最小规则示例

```json
{
  "name": "http default",
  "enabled": true,
  "protocol": "http",
  "namespace": "default",
  "selector": {
    "all": [
      {
        "field": "request.host",
        "op": "eq",
        "value": "demo.com"
      },
      {
        "field": "request.path",
        "op": "prefix",
        "value": "/api/"
      }
    ]
  },
  "rules": [
    {
      "id": "debug-api",
      "enabled": true,
      "priority": 100,
      "when": {
        "all": [
          {
            "field": "request.method",
            "op": "eq",
            "value": "GET"
          },
          {
            "field": "request.path",
            "op": "eq",
            "value": "/api/v1/debug"
          }
        ]
      },
      "action": {
        "type": "static_response",
        "status": 200,
        "body": {
          "message": "ok"
        }
      }
    }
  ]
}
```

## 已知限制

- 运行时只支持 MySQL 存储。
- 当前启动时支持单文件、目录扫描和逗号分隔多路径，但仍然只支持 `.json` 规则文件。
- 当前版本快照支持查询、预演、回滚和 MySQL 存储；DB schema 位于 `docs/db_schema.sql`。
- 后端启动装配已切到 `go.uber.org/fx v1.24.0`；HTTP server 由 Fx lifecycle 负责启动和优雅关闭。
- DB 访问层已切到 `goshared/db/dbspi.Manager` + `dbhelper.NewSoftDeleteTableStore`；提供 `MOCKSERVER_MYSQL_TEST_DSN` 时会执行真实 MySQL 集成测试。
- admin token 已支持 read / write / publish 权限分层，但还没有用户体系、登录、JWT、RBAC 和租户隔离。
- snapshot 已记录 publish / rollback 审计 metadata，但还没有独立审计表、不可变审计日志和审批流。
- `template_response` 已支持常用 helper，但还没有做模板沙箱、模板限流和更强的调试信息。
- `request.body` 的路径访问目前只覆盖基础 JSON 对象场景。
- runtime 目前只实现了 HTTP adapter。
- runtime metrics 当前是内存型，服务重启后会清零，仅作为 HTTP runtime 调试指标。
- SDK decision traffic 会落库到 `mockserver_traffic_event_tab` 和 `mockserver_traffic_event_index_tab`；admin simulate 流量不落库。
- traffic index 默认只展开低基数定位字段；HTTP `query/header/body`、cache `value`、SPEX `param/req` 等高基数或大字段保留在原始 event JSON 中。

## 已有验证

当前已通过：

```bash
GOCACHE=$(pwd)/.gocache go test ./...
```

如果要跑 MySQL DAO 集成测试，需要提供测试 DSN：
该测试会覆盖 DAO repository 和 admin/runtime HTTP 端到端链路。

```bash
MOCKSERVER_MYSQL_TEST_DSN='root:123456@tcp(127.0.0.1:3306)/mockserver_db?charset=utf8mb4&parseTime=True&loc=Local' \
GOCACHE=$(pwd)/.gocache go test ./...
```
