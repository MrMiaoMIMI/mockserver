# MockServer 用户使用说明书

这份文档面向 MockServer 的日常使用者，目标是让你在本地快速启动服务、创建规则、发布到 runtime，并用前端、Postman 或 curl 完成调试。

## 1. MockServer 是什么

MockServer 用规则来模拟协议调用结果。HTTP runtime 适合本地调试 HTTP 规则，SDK decision endpoint 则面向 mockinject 和更多协议：

- 用 `ruleset` 管理一组 mock 规则。
- 用 `selector` 先筛选哪些请求属于这个 ruleset。
- 用 `rule.when` 精确判断某个请求是否命中。
- 用 `rule.action` 返回静态响应、模板响应、CEL 动态响应、顺序响应或 webhook 响应。
- 用 `publish` 把 draft 规则发布到 runtime。
- 用 `simulate` 在不真正调用 runtime 的情况下解释为什么命中或不命中。
- 用 `rollback` 回滚到历史发布快照。
- 用代码注册的 `ProtocolSpec` 约束协议字段，当前包含 HTTP、cache 和 SPEX。

## 2. 你需要准备什么

本地最小使用需要 Go 和一个 MySQL 数据库。当前本地测试常用示例：

```bash
go run ./cmd/server
```

服务默认读取 `etc/server.yml`，MySQL 连接配置和 JWT 登录配置都建议维护在这个文件里。HTTP 监听端口优先读取环境变量 `PORT`，未设置时默认使用 `8080`。

如果要使用前端页面，需要 Node.js：

```bash
cd web
npm install
npm run dev
```

运行时只支持 MySQL 存储；本地快速体验也需要配置 MySQL 连接信息。

## 3. 快速启动

### 3.1 只启动后端

```bash
go run ./cmd/server
```

默认监听：

```text
http://127.0.0.1:8080
```

### 3.2 启动前端

另开一个终端：

```bash
cd web
npm install
npm run dev
```

访问：

```text
http://127.0.0.1:6173/
```

前端开发服务默认会把 `/mockserver` 代理到：

```text
http://localhost:8080
```

如果后端不在 `8080`，可以这样启动前端：

```bash
VITE_MOCKSERVER_PROXY_TARGET=http://127.0.0.1:18080 npm run dev
```

## 4. 推荐的使用方式

### 4.1 前端页面

适合日常配置和排障。

主要入口：

- `规则集工作台`：创建、编辑、保存 draft ruleset。
- `Ruleset 表单`：编辑 `name/enabled/protocol/namespace/selector`，Ruleset ID 由后端创建时自动生成。
- `Rule 管理`：新增、更新、删除、启用、禁用、调整优先级。
- `Simulate`：输入 event，查看命中结果和 explain。
- `Rollback`：查看 snapshot，预演回滚 diff，再执行回滚。
- `Traffic Inspector`：查看 SDK decision traffic，包括协议、namespace、命中结果、ruleset/rule、trace、原始 event、decision 和可查询索引字段。

前端配置 selector 和 rule condition 时会根据 `ProtocolSpec` 自动切换字段、operator 和 value 输入方式：

- 选择不同协议后，字段下拉只展示该协议注册的字段和 selector。
- `exists`、`not_exists`、`is_null`、`is_not_null` 不需要填写 value。
- `in`、`not_in` 使用列表输入，每个 chip 会保存为数组项。
- 数字比较使用数字输入，布尔字段使用 true/false 选择。
- `regex` 会在前端检查正则表达式是否能编译。
- JSON 根字段使用 JSON 输入；动态 JSON 子字段如 `request.body.status` 默认按普通值填写。
- `request.query`、`request.headers` 这类动态字段会提供 key/path 输入，header key 会归一化为小写，query/header 可选择 `[0]` 或 `[*]`。

如果后端启用了 JWT，进入前端后先用登录页填写邮箱。发布和回滚时，后端会默认使用 JWT 中的邮箱作为 operator；也可以在请求里显式传入：

- `Operator`：操作者，例如 `admin@example.com`。
- `Trace ID`：本次操作的追踪 ID，例如 `manual-release-001`。

### 4.2 Postman

仓库已经提供可导入文件：

```text
examples/mockserver.postman_collection.json
```

导入 Postman 后，确认 collection 变量：

- `base_url`：默认 `http://127.0.0.1:8080`。
- `admin_token`：如果后端没有开启鉴权，可以留空。
- `operator`：发布/回滚时记录到 snapshot audit。
- `trace_id`：用于追踪本次调试请求。

Postman collection 已包含：

- draft ruleset 创建、查询、校验。
- simulate full explain / summary / selector miss。
- publish、runtime 调用、published simulate。
- published snapshots、rollback preview、rollback。
- 单条 rule 的增删改、启用禁用、优先级调整。
- `respond` + `sequence` renderer 示例。
- `respond` + `webhook` renderer 示例。
- runtime metrics 查询。

建议第一次按文件夹顺序执行：

```text
01 Draft Ruleset -> 02 Simulate -> 03 Publish And Runtime
```

### 4.3 curl

适合脚本化和快速验证。

导入 draft：

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets \
  -H 'Content-Type: application/json' \
  --data-binary @examples/ruleset.json
```

校验 draft：

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets/http-default/validate
```

发布 draft：

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets/http-default/publish \
  -H 'Content-Type: application/json' \
  -H 'X-Mockserver-Operator: admin@example.com' \
  -H 'X-Trace-ID: release-001' \
  -d '{"reason":"first release"}'
```

调用 runtime：

```bash
curl 'http://127.0.0.1:8080/mockserver/runtime/default/http/api/v1/debug?q1=qv1' \
  -H 'Host: demo.com' \
  -H 'X-Trace-ID: runtime-001'
```

查看 HTTP runtime metrics：

```bash
curl http://127.0.0.1:8080/mockserver/api/v1/admin/metrics/runtime
```

查看 SDK decision traffic：

```bash
curl 'http://127.0.0.1:8080/mockserver/api/v1/admin/traffic/events?limit=50&offset=0'
curl 'http://127.0.0.1:8080/mockserver/api/v1/admin/traffic/events/35'
```

SDK decision traffic 会落库；ruleset/rule 管理页面里的 simulate 结果直接展示在页面，不写入 traffic 表。
Traffic index 默认只保存低基数、常用于定位的字段，例如 HTTP `method/host/path`、cache `operation/key`、SPEX `cmd`、`decision.response.payload.status` 和 `decision.response.payload.code`；HTTP `query/header/body`、cache `value`、SPEX `param/req` 等高基数或大字段保留在原始 event JSON 中，不默认展开到索引表。

## 5. 核心概念

### 5.1 Draft 和 Published

MockServer 有两层规则状态：

- `draft`：编辑态，修改后不会立刻影响 runtime。
- `published`：运行态，runtime 只使用 published snapshot。

常见流程：

```text
创建/修改 draft -> validate -> simulate -> publish -> runtime 生效
```

如果发现发布有问题：

```text
查看 snapshots -> rollback preview -> rollback
```

### 5.2 Ruleset

`ruleset` 是一组规则的集合。示例：

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
  "rules": []
}
```

关键字段：

- `id`：规则集唯一 ID，由后端创建 draft 时生成，后续用于发布、回滚、trace 和 metrics。
- `enabled`：关闭后整个 ruleset 不参与匹配。
- `protocol`：当前支持 `http`、`cache` 和 `spex`。
- `namespace`：运行时命名空间，runtime URL 中会用到。
- `selector`：ruleset 级粗筛。
- `rules`：具体规则列表。

### 5.3 Runtime URL

runtime 路径格式：

```text
/mockserver/runtime/{namespace}/http/{actual_path}
```

示例：

```text
/mockserver/runtime/default/http/api/v1/debug
```

真正参与规则匹配的 path 是：

```text
/api/v1/debug
```

不是完整的 `/mockserver/runtime/default/http/api/v1/debug`。

### 5.4 Selector

`selector` 用来快速判断一个请求是否属于某个 ruleset。

支持字段：

```json
{
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
}
```

如果一个请求同时命中多个 ruleset，MockServer 会先选 selector 更严格的 ruleset，再只匹配该 ruleset 内部的 rules。selector 字段来自当前协议的 `ProtocolSpec.selectors`，例如 HTTP 支持 `request.host` 和 `request.path`，cache 支持 `request.operation` 和 `request.key`，SPEX 支持 `request.cmd`。当前严格度按 selector 条件累加，`eq` 高于 `prefix/suffix`，再高于 `contains/exists`，最后用 `ruleset.id` 倒序做稳定兜底。

如果 selector 没命中，ruleset 内部的 rule 不会继续匹配。排查这类问题时，用 `simulate` 看 `rule_set_explanations.selector_checks`。

### 5.5 Rule

`rule` 是实际命中逻辑：

```json
{
  "id": "debug-api",
  "enabled": true,
  "priority": 100,
  "when": {
    "all": [
      {"field": "request.method", "op": "eq", "value": "GET"},
      {"field": "request.path", "op": "eq", "value": "/api/v1/debug"}
    ]
  },
  "action": {
    "type": "respond",
    "renderer": "static",
    "response": {
      "payload": {
        "status": 200,
        "body": {
          "message": "ok"
        }
      }
    }
  }
}
```

关键字段：

- `id`：规则唯一 ID。
- `enabled`：关闭后该 rule 不参与匹配。
- `priority`：优先级，数值越大越优先。
- `when`：条件树。
- `action`：命中后的响应动作。

## 6. 条件怎么写

### 6.1 all / any / not

所有条件都满足：

```json
{
  "all": [
    {"field": "request.method", "op": "eq", "value": "GET"},
    {"field": "request.path", "op": "prefix", "value": "/api/"}
  ]
}
```

任一条件满足：

```json
{
  "any": [
    {"field": "request.query.env[*]", "op": "contains", "value": "test"},
    {"field": "request.headers.x-env[*]", "op": "contains", "value": "test"}
  ]
}
```

条件取反：

```json
{
  "not": {
    "field": "request.path",
    "op": "prefix",
    "value": "/internal/"
  }
}
```

### 6.2 常用字段路径

字段路径来自 mockserver 代码注册的 `ProtocolSpec`。用户配置 ruleset/rule/namespace，不动态配置协议字段。

通用字段：

```text
protocol
namespace
meta.trace_id
meta.source
meta.extra.<key>
```

HTTP 请求基本信息：

```text
request.method
request.scheme
request.host
request.original_host
request.path
request.client_ip
request.raw_body
```

Query 和 Header：

```text
request.query.q1
request.query.q1[*]
request.query.q1[0]
request.headers.x-env
request.headers.x-env[*]
request.headers.x-env[0]
```

Body：

```text
request.body.user.id
request.body.items[0].id
request.body.items[*].id
```

HTTP body 解析规则：

- body 可解析为 JSON 时，同时设置 `request.body` 和 `request.raw_body`
- body 不能解析为 JSON 时，只设置 `request.raw_body`
- body 为空时，不设置 `request.body` 和 `request.raw_body`

Cache：

```text
request.operation
request.key
request.ttl_ms
request.value
request.value.user.id
```

SPEX：

```text
request.cmd
request.req
request.req.order_id
request.param
```

### 6.3 常用操作符

精确匹配：

```json
{"field": "request.method", "op": "eq", "value": "GET"}
```

包含：

```json
{"field": "request.query.tags[*]", "op": "contains", "value": "vip"}
```

前缀：

```json
{"field": "request.path", "op": "prefix", "value": "/api/"}
```

正则：

```json
{"field": "request.path", "op": "regex", "value": "^/api/v[0-9]+/users$"}
```

数字比较：

```json
{"field": "request.body.score", "op": "gte", "value": 90}
```

当前支持的操作符：

```text
eq, ne, in, not_in, contains, not_contains, exists, not_exists,
is_null, is_not_null, prefix, suffix, regex, gt, gte, lt, lte
```

`exists` 表示路径存在，即使值是 null。`not_exists` 表示路径不存在。`is_null` 表示路径不存在，或路径存在但值是 JSON null / Go nil。`is_not_null` 表示路径存在且值不是 null。

value 填写规则：

- `eq`、`ne`：按字段类型填写单个值，例如 `"GET"`、`90`、`true` 或合法 JSON。
- `in`、`not_in`：填写数组，例如 `["GET", "POST"]`；前端会用 chip 输入辅助生成数组。
- `contains`、`not_contains`：通常填写要包含的单个字符串、数字或 JSON 子值。
- `prefix`、`suffix`、`regex`：填写字符串；`regex` 不需要写 `/.../` 包裹。
- `gt`、`gte`、`lt`、`lte`：填写数字。
- `exists`、`not_exists`、`is_null`、`is_not_null`：不填写 value。

### 6.4 CEL 条件

复杂条件可以用 `expr`：

```json
{
  "expr": "request.headers[\"x-env\"][0] == \"test\" && request.body.score >= 90"
}
```

适合表达跨字段组合判断。简单条件优先用 `field + op + value`，更容易解释和排查。

## 7. Action 怎么写

### 7.1 Static Respond

返回固定响应：

```json
{
  "type": "respond",
  "renderer": "static",
  "response": {
    "payload": {
      "status": 200,
      "headers": {
        "content-type": ["application/json"]
      },
      "body": {
        "message": "ok"
      }
    }
  }
}
```

### 7.2 Template Respond

根据请求动态渲染响应：

```json
{
  "type": "respond",
  "renderer": "template",
  "response_template": "{\"status\":200,\"headers\":{\"content-type\":[\"application/json\"]},\"body\":{\"message\":\"hello {{ query . \"q1\" }}\",\"path\":\"{{ field . \"request.path\" }}\"}}"
}
```

常用 helper：

```text
field . "request.path"
query . "q1"
queryAll . "q1"
header . "x-env"
headerAll . "x-env"
body . "user.id"
first ...
toJSON ...
```

### 7.3 CEL Respond

用 CEL 表达式生成完整协议 response payload：

```json
{
  "type": "respond",
  "renderer": "cel",
  "response_expression": "{\"status\": 200, \"headers\": {\"content-type\": [\"application/json\"]}, \"body\": {\"path\": request.path, \"score\": request.body.score}}"
}
```

### 7.4 Sequence Respond

按调用次数返回不同结果，适合模拟轮询状态：

```json
{
  "type": "respond",
  "renderer": "sequence",
  "sequence_strategy": "last",
  "sequence": [
    {
      "response": {
        "payload": {
          "status": 200,
          "body": {
            "state": "pending"
          }
        }
      }
    },
    {
      "response": {
        "payload": {
          "status": 200,
          "body": {
            "state": "running"
          }
        }
      }
    },
    {
      "response": {
        "payload": {
          "status": 200,
          "body": {
            "state": "done"
          }
        }
      }
    }
  ]
}
```

`sequence_strategy` 支持：

- `last`：到最后一步后一直返回最后一步。
- `loop`：到最后一步后从第一步重新开始。

注意：当前 sequence 计数是进程内状态，服务重启后会重置。

### 7.5 Webhook Respond

把响应生成交给外部服务：

```json
{
  "type": "respond",
  "renderer": "webhook",
  "webhook": {
    "url": "https://postman-echo.com/post",
    "method": "POST",
    "timeout_ms": 3000,
    "headers": {
      "content-type": ["application/json"],
      "x-from": ["mockserver"]
    }
  }
}
```

MockServer 会把当前 event 发送给 webhook，再把 webhook 返回的协议 response payload 作为 runtime 响应返回。

注意：生产环境使用 webhook 前，建议先补域名 allowlist、内网地址拦截和响应大小限制。

## 8. Simulate 和 Explain

`simulate` 用来回答这些问题：

- 为什么请求没有命中 ruleset？
- 为什么 selector 把 ruleset 筛掉了？
- 哪些 rule 是候选？
- 每个条件为什么命中或不命中？
- 最终会返回什么响应？

模拟 draft：

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets/http-default/simulate \
  -H 'Content-Type: application/json' \
  -d '{
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
    },
    "explain_only": true,
    "explain_summary": true,
    "explain_max_depth": 2
  }'
```

常用调试选项：

- `explain_only=true`：只看命中解释，不真正执行 action。
- `explain_summary=true`：只返回命中摘要和失败原因概览。
- `explain_compact=true`：减少 expected、actual、rendered_result 等大字段。
- `explain_max_depth=2`：限制条件树 explain 深度。

推荐排障顺序：

```text
先看 rule_set_explanations.selector_checks
再看 candidate_rules
再看 rule_explanations.condition
最后看 action_info
```

## 9. 发布、快照和回滚

### 9.1 发布

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/rulesets/http-default/publish \
  -H 'Content-Type: application/json' \
  -H 'X-Mockserver-Operator: admin@example.com' \
  -H 'X-Trace-ID: release-001' \
  -d '{"reason":"release debug api mock"}'
```

发布后会生成一个 snapshot。snapshot 中会记录：

- ruleset 内容。
- 发布时间。
- 操作者。
- 发布原因。
- trace id。

### 9.2 查看快照

```bash
curl http://127.0.0.1:8080/mockserver/api/v1/admin/published/rulesets/http-default/snapshots
```

### 9.3 回滚预演

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/published/rulesets/http-default/rollback/preview \
  -H 'Content-Type: application/json' \
  -d '{
    "snapshot_id": "目标 snapshot id",
    "explain_summary": true,
    "explain_max_depth": 2
  }'
```

预演结果会包含：

- 目标 snapshot 是否能通过校验。
- 与当前 published 的 diff。
- ruleset 字段级 diff。
- rule 字段级 diff。
- 可选 event 的模拟命中结果。

### 9.4 执行回滚

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/published/rulesets/http-default/rollback \
  -H 'Content-Type: application/json' \
  -H 'X-Mockserver-Operator: admin@example.com' \
  -H 'X-Trace-ID: rollback-001' \
  -d '{
    "snapshot_id": "目标 snapshot id",
    "reason": "restore previous stable mock"
  }'
```

## 10. 鉴权和权限

默认 `etc/server.yml` 启用 JWT debug 登录。前端登录页会调用 `/mockserver/api/v1/auth/debug/login`，用手动填写的邮箱换取 JWT。

JWT 配置：

```yaml
auth:
  jwt_secret: mockserver-debug-secret
  debug_login_enabled: true
```

JWT 请求：

```text
Authorization: Bearer <jwt>
```

admin API 只支持 JWT 鉴权；runtime mock 接口不受 JWT 影响。

## 11. 存储方式

### 11.1 MySQL 存储

当前唯一支持的运行时存储方式。推荐在 `etc/server.yml` 中维护：

```yaml
db:
  database_groups:
    default:
      host: "127.0.0.1"
      port: 3306
      user: "root"
      password: "123456"
      database_name: "mockserver_db"
      max_open_conns: 20
      max_idle_conns: 5
      conn_max_lifetime_seconds: 3600
      debug: false
```

特点：

- draft、published、snapshot、namespace 配置都会落库。
- 支持服务重启后恢复数据。
- 后端对象装配使用 `go.uber.org/fx v1.24.0`，各后端 package 暴露自己的 `Module`，HTTP server 由 Fx lifecycle 负责启动和优雅关闭。
- DB 访问层使用 `github.com/MrMiaoMIMI/goshared/db/dbspi.Manager` + `dbhelper.NewSoftDeleteTableStore`。
- `ctime`、`mtime`、`publish_time` 使用 UnixMilli；`creator`、`updater` 通过请求上下文自动填充。
- 服务启动时不会执行 DDL 或迁移，数据库和表结构需要用户自行创建。

schema 参考文件：

```text
docs/db_schema.sql
```

## 12. 常见问题

### 12.1 runtime 返回 404

优先检查：

- 是否已经 `publish`，只保存 draft 不会影响 runtime。
- runtime URL 的 namespace 是否正确。
- 请求是否带了正确的 `Host` header。
- selector 的 `request.host` / `request.path` 等协议字段是否命中。
- rule 的 `when` 条件是否命中。

建议直接用 simulate：

```bash
curl -X POST http://127.0.0.1:8080/mockserver/api/v1/admin/published/simulate \
  -H 'Content-Type: application/json' \
  -d '{
    "explain_only": true,
    "explain_summary": false,
    "explain_max_depth": 4,
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

### 12.2 前端接口 404

检查：

- 后端是否启动在 `8080`。
- 前端是否通过 `npm run dev` 启动在 `6173`。
- Vite proxy 是否指向正确后端。
- 浏览器请求路径是否以 `/mockserver` 开头。

如果后端端口不是 `8080`：

```bash
cd web
VITE_MOCKSERVER_PROXY_TARGET=http://127.0.0.1:18080 npm run dev
```

### 12.3 admin 接口返回 401/403

检查：

- 是否在 `etc/server.yml` 设置了 `auth.jwt_secret`。
- 请求是否携带 `Authorization: Bearer <jwt>`。
- JWT 是否有效或签名不匹配。

常见情况：

- 没有先通过 debug login 获取 JWT。
- 使用了旧的 `X-Mockserver-Admin-Token` header。
- 本地服务重启后更换了 `auth.jwt_secret`，旧 JWT 会失效。

### 12.4 template renderer 没有按预期渲染

建议：

- 先用 `simulate`。
- 设置 `explain_only=false` 看实际渲染结果。
- 检查 helper 路径是否正确。
- JSON body 模板要保证渲染后仍是合法 JSON。

### 12.5 webhook renderer 调不通

检查：

- webhook URL 是否是完整的 `http://` 或 `https://` URL。
- 外部服务是否能从 MockServer 所在机器访问。
- `timeout_ms` 是否太短。
- webhook 服务返回的 body 是否是预期格式。

本地安全提醒：当前 webhook 适合受控环境调试，生产前建议补强 SSRF 防护。

## 13. API 速查

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/mockserver/api/v1/admin/protocols` | 查询协议字段和 selector 规格 |
| `POST` | `/mockserver/api/v1/admin/rulesets` | 创建或更新 draft ruleset |
| `GET` | `/mockserver/api/v1/admin/rulesets` | 查询 draft 列表 |
| `GET` | `/mockserver/api/v1/admin/rulesets/{id}` | 查询 draft 详情 |
| `POST` | `/mockserver/api/v1/admin/rulesets/{id}/validate` | 校验 draft |
| `POST` | `/mockserver/api/v1/admin/rulesets/{id}/simulate` | 模拟 draft 命中 |
| `POST` | `/mockserver/api/v1/admin/rulesets/{id}/publish` | 发布 draft |
| `POST` | `/mockserver/api/v1/admin/rulesets/{id}/rules` | 新增 rule |
| `PUT` | `/mockserver/api/v1/admin/rulesets/{id}/rules/{rule_id}` | 更新 rule |
| `DELETE` | `/mockserver/api/v1/admin/rulesets/{id}/rules/{rule_id}` | 删除 rule |
| `POST` | `/mockserver/api/v1/admin/rulesets/{id}/rules/{rule_id}/enable` | 启用 rule |
| `POST` | `/mockserver/api/v1/admin/rulesets/{id}/rules/{rule_id}/disable` | 禁用 rule |
| `POST` | `/mockserver/api/v1/admin/rulesets/{id}/rules/{rule_id}/priority` | 调整 rule 优先级 |
| `GET` | `/mockserver/api/v1/admin/published/rulesets` | 查询 published 列表 |
| `GET` | `/mockserver/api/v1/admin/published/rulesets/{id}` | 查询当前 published |
| `GET` | `/mockserver/api/v1/admin/published/rulesets/{id}/snapshots` | 查询历史 snapshots |
| `POST` | `/mockserver/api/v1/admin/published/simulate` | 模拟当前 published 命中 |
| `POST` | `/mockserver/api/v1/admin/published/rulesets/{id}/rollback/preview` | 回滚预演 |
| `POST` | `/mockserver/api/v1/admin/published/rulesets/{id}/rollback` | 执行回滚 |
| `GET` | `/mockserver/api/v1/admin/metrics/runtime` | 查询 runtime metrics |
| `GET` | `/mockserver/api/v1/admin/traffic/events` | 查询 SDK decision traffic 事件 |
| `ANY` | `/mockserver/runtime/{namespace}/http/{actual_path}` | runtime mock 调用入口 |

## 14. 推荐上手路径

第一次使用建议按这个顺序：

1. 后端启动：先手工创建数据库和表结构，再在 `etc/server.yml` 配置 `db.database_groups.default`，然后执行 `go run ./cmd/server`。
2. 启动前端：`cd web && npm install && npm run dev`。
3. 在前端登录并创建一条 rule，保存 draft。
4. 发布 rule 后调 runtime：`GET /mockserver/runtime/default/http/api/v1/debug?q1=qv1`，带 `Host: demo.com`。
5. 导入 Postman collection：`examples/mockserver.postman_collection.json`。
6. 执行 `02 Simulate`，理解 explain 输出。
7. validate、simulate、publish。
8. 再调 runtime，确认发布生效。
9. 查看 snapshots，执行 rollback preview。
10. 需要持久化时使用默认 MySQL 存储。
