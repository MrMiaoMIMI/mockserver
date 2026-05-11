# MockServer 方案评审与完整设计（V2）

## 1. 结论先行

当前 `ai/init.md` 里的方向是对的，但还不够稳。

它的优点是：

- 抓住了核心问题：要做的是“通用 mock 能力”，不是只做 HTTP mock。
- 规则与响应解耦的方向是合理的。
- 用 JSON 作为规则配置载体，便于存储、传输、管理接口化，这一点是成立的。

它的主要问题是：

- “基于 JSON 的强大规则引擎”这个说法太泛，容易把 JSON 当成一种编程语言来用，后续会变得非常难维护。
- “先命中组，再命中规则，再生成响应”这个抽象还不够稳定，组与规则的职责边界不清晰。
- “支持自定义 Go 脚本”风险很高，安全性、跨平台、运行时隔离、调试成本都会很差。
- 当前输入模型和匹配模型偏弱，尤其是数组、嵌套对象、可组合条件、优先级冲突、性能优化这些关键点还没有定义清楚。

最终建议：

- 保留 `JSON 作为规则 DSL 载体`，但不要把它设计成“任意 JSON + 任意字段 + 任意脚本”的松散模式。
- 核心引擎采用 `规范化输入 + 声明式规则 AST + 编译后匹配` 的方案。
- 不建议把“用户上传 Go 脚本”作为主方案。
- 建议把动态计算能力拆成 4 个层次：`static`、`template`、`CEL expression`、`external webhook`。
- 规则组织方式建议从“命中组规则 / 组规则 / 响应规则”重构为：`RuleSet -> Rule -> Action`。

这会比当前草案更容易扩展、更容易调试，也更适合后续支持 HTTP 之外的协议。

## 2. 对当前方案的逐项评审

### 2.1 基于 JSON 的规则引擎是否合适

答案是：`合适，但只适合作为 DSL 和配置格式，不适合作为计算模型本身。`

适合的原因：

- JSON 易于落库存储，也适合走 HTTP API 管理。
- 规则可以版本化、审计、导入导出。
- 对前端、平台端、控制台都友好。
- Go 原生支持 JSON 序列化，工程成本低。

不适合直接无限放大的原因：

- JSON 本身没有类型系统表达能力，复杂逻辑会非常啰嗦。
- 一旦支持 `all/any/not + 数组匹配 + 正则 + 函数 + 脚本`，DSL 很快会失控。
- 如果字段路径、操作符和值类型不做强约束，线上会出现大量“规则可保存但不可正确执行”的情况。
- 如果直接对 `map[string]any` 做运行时解释，性能、可测试性、错误定位都会很差。

所以正确姿势是：

- 对外：JSON。
- 对内：强类型 Go 结构体 + 编译后的规则对象。
- 规则保存时做校验。
- 规则加载时做编译。
- 请求命中时只执行编译后的规则。

### 2.2 “先命中组，再命中规则”是否合适

答案是：`思路可以保留，但抽象层级建议调整。`

当前草案的问题不是“两阶段匹配”本身，而是“组”这个概念太模糊。

它可能代表：

- 协议类型分组，例如 HTTP / RPC / DB。
- 场景分组，例如登录场景、风控场景、回调场景。
- 租户分组，例如不同业务线或环境。
- 响应集合分组，例如同一个请求命中后进入不同返回分支。

这些其实不是一个维度，不应该混成一个 `group`。

更稳的设计是：

- `RuleSet`：规则集合，负责领域隔离和生命周期管理。
- `Selector`：粗粒度筛选，用于快速缩小候选集。
- `Rule`：精确匹配条件。
- `Action`：命中后的行为。

即：

1. 先进入某个 `RuleSet`
2. 通过 `Selector` 做粗筛
3. 在候选 `Rule` 里做精确匹配
4. 执行 `Action`

这样保留了“两阶段匹配”的效率优势，但概念更清晰。

### 2.3 自定义 Go 脚本是否合适

答案是：`不建议作为主方案，最多作为受控扩展能力。`

主要问题：

- Go 并不是脚本语言。
- `plugin` 方案跨平台和运维成本都很差。
- `yaegi` 这类 Go 解释器可行，但稳定性、性能、兼容性、安全边界都不理想。
- 用户脚本一旦可以直接访问运行时、文件系统、网络，风险会很高。
- 排错极难，尤其是在动态加载、热更新、上下文隔离方面。

更合理的替代方案：

1. `static`
固定状态码、固定 header、固定 body。

2. `template`
用受限模板从请求上下文中取值构造响应。

3. `expression`
使用 `CEL` 做条件判断和轻量计算。

4. `webhook`
把复杂逻辑转发给外部计算服务，mockserver 只负责输入输出契约。

如果一定要保留 Go 扩展能力，建议只保留：

- `builtin handler registry`

即服务启动时注册若干受控的 Go 函数，例如：

- `buildUserProfile`
- `buildOrderState`
- `signResponse`

规则里只能引用白名单函数名，不能上传任意 Go 代码。

这比“用户提交 Go 脚本”安全得多，也更适合生产环境。

## 3. 推荐的总体架构

### 3.1 设计目标

目标不是“先做一个 HTTP mockserver”，而是：

- 核心引擎协议无关。
- HTTP 只是第一个 adapter。
- 规则可管理、可验证、可调试、可回放。
- 匹配过程稳定且可解释。
- 响应生成可从简单到复杂渐进增强。

### 3.2 核心分层

参考 `backend_conventions` 的分层思想，但不机械照搬数据库和鉴权假设，建议使用下面的结构：

- `cmd/server`
启动入口、配置加载、依赖装配。

- `internal/router`
HTTP 路由注册。

- `internal/controller`
解析管理接口和 runtime 接口请求，输出统一响应。

- `internal/view`
负责 request/response 组装，以及调试接口输出转换。

- `internal/service`
规则管理、规则发布、模拟执行、命中调试。

- `internal/engine`
规则编译、匹配、动作执行。

- `internal/adapter`
协议适配层，例如 `httpadapter`。

- `internal/dao`
基于 `dbhelper` / `dbspi` 的 MySQL 数据访问与规则仓储。

- `model/request`
管理接口请求模型。

- `model/response`
管理接口响应模型。

- `model/bo`
服务内部业务对象。

- `model/eo`
枚举、常量、操作符定义。

核心原则：

- `engine` 不依赖 Gin，不依赖 HTTP 细节。
- `adapter` 负责把协议请求转换成统一输入。
- `service` 不直接感知某条规则的底层执行细节。
- `dao` 负责持久化、快照和 BO/DO 转换，不负责规则匹配。

## 4. 核心对象模型

### 4.1 运行时统一输入模型

不建议直接把“用户传来的任意 JSON”塞给规则引擎。

建议引入一个规范化后的统一输入 `Event`：

```json
{
  "protocol": "http",
  "operation": "request",
  "namespace": "default",
  "request": {
    "method": "GET",
    "scheme": "https",
    "host": "demo.com",
    "path": "/api/v1/debug",
    "query": {
      "q1": ["qv1"]
    },
    "headers": {
      "h1": ["hv1"]
    },
    "body": {},
    "raw_body": "",
    "client_ip": "127.0.0.1"
  },
  "meta": {
    "trace_id": "xxx"
  }
}
```

关键建议：

- `headers` 和 `query` 统一成 `map[string][]string`，不要保留数组对数组的搜索式结构。
- `host` 不要带 scheme，`scheme` 单独存。
- `path` 必须是标准化后的 path。
- `body` 尽量解析为 JSON 对象；解析失败时保留 `raw_body`。
- 所有协议最终都映射到统一 `Event`，而不是每种协议都自己发明一套规则解释逻辑。

### 4.2 规则集合模型

推荐核心模型：

```json
{
  "name": "http default ruleset",
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

这里的 `selector` 不是完整规则，只做粗筛。

### 4.3 精确匹配规则模型

```json
{
  "id": "rule-debug-api",
  "name": "debug api",
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
      },
      {
        "field": "request.headers.h1[*]",
        "op": "contains",
        "value": "hv1"
      }
    ]
  },
  "action": {
    "type": "static_response",
    "status": 200,
    "headers": {
      "content-type": ["application/json"]
    },
    "body": {
      "code": 0,
      "message": "ok"
    }
  }
}
```

### 4.4 条件表达式 AST

建议规则条件采用统一 AST，而不是“平铺 rules 数组”。

推荐结构：

- `all`
- `any`
- `not`
- `predicate`

例如：

```json
{
  "all": [
    {
      "field": "protocol",
      "op": "eq",
      "value": "http"
    },
    {
      "any": [
        {
          "field": "request.path",
          "op": "prefix",
          "value": "/api/v1/"
        },
        {
          "field": "request.path",
          "op": "regex",
          "value": "^/internal/"
        }
      ]
    }
  ]
}
```

这是后续可扩展的关键。否则只靠 `rules: []` 很快会表达不出复杂条件。

## 5. 匹配设计建议

### 5.1 字段路径设计

当前草案里的 `request.header[*].key` 这种写法不推荐。

原因：

- 它迫使引擎在数组结构里做低效扫描。
- 对 headers/query 这种天然 key-value 的场景表达不自然。
- 易产生歧义，例如 key 大小写、重复值、多值顺序。

更好的写法是：

- `request.headers.h1[*]`
- `request.query.q1[*]`
- `request.body.user.id`

也就是：

- map 用点路径。
- 数组用 `[*]`。
- body 对象走 JSON path 风格的字段访问。

### 5.2 操作符设计

建议首批支持这些操作符：

- `eq`
- `ne`
- `in`
- `not_in`
- `contains`
- `not_contains`
- `exists`
- `not_exists`
- `prefix`
- `suffix`
- `regex`
- `gt`
- `gte`
- `lt`
- `lte`

不要一开始就支持太多自定义函数。先把一阶操作符做好。

### 5.3 匹配顺序设计

推荐匹配流程：

1. 规范化输入 `Event`
2. 通过 `protocol + namespace + method + host + path` 做候选粗筛
3. 根据优先级和特异性排序候选规则
4. 执行 `when` AST
5. 命中后执行 `action`

排序建议：

- 第一关键字：`priority` 降序
- 第二关键字：`specificity_score` 降序
- 第三关键字：`created_at` 升序

这里的 `specificity_score` 可以由规则编译时计算，例如：

- `path eq` 比 `path prefix` 更具体
- `header eq` 比 `header exists` 更具体
- `body exact` 比 `body regex` 更具体

这样比纯 `priority` 更稳，不容易出现“后加一条规则把老规则全打穿”的问题。

### 5.4 性能优化建议

如果规则量上来，不能每次请求都全表扫描。

建议编译期构建索引：

- `protocol -> ruleset`
- `namespace -> ruleset`
- `host -> candidate rules`
- `path prefix trie`

高成本条件，例如：

- regex
- body 深层匹配
- expression

必须放到最后阶段执行。

简单说就是：

- 先用便宜条件筛选。
- 再用昂贵条件确认。

这部分会直接决定后期是否能承受大量 mock 规则。

## 6. 响应动作设计

### 6.1 推荐的 Action 类型

建议不要只支持“固定返回”和“Go 脚本”两种。

推荐至少支持以下 5 类：

1. `static_response`
最基础，直接返回状态码、header、body。

2. `template_response`
基于请求上下文做变量替换。

3. `expression_response`
用 `CEL` 计算字段或动态 body。

4. `sequence_response`
同一规则按次数或状态顺序返回不同结果。

5. `webhook_response`
调用外部服务获取动态结果。

其中：

- `static_response` 是最低成本默认选项。
- `template_response` 解决 70% 的“响应里带请求字段”需求。
- `expression_response` 解决轻量动态逻辑。
- `sequence_response` 很适合 mock 登录态、轮询态、任务状态变化。
- `webhook_response` 解决真正复杂的动态生成问题。

### 6.2 模板能力建议

如果用模板，建议不要直接暴露完整 `text/template` 能力给用户。

建议提供受限变量上下文，例如：

- `request.method`
- `request.path`
- `request.query.q1[0]`
- `request.headers.x_trace_id[0]`
- `meta.trace_id`
- `now.unix`

以及少量白名单函数，例如：

- `default`
- `join`
- `lower`
- `upper`
- `json`

### 6.3 动态逻辑引擎建议

推荐优先级：

1. `CEL`
适合条件判断、轻量计算、字段拼装。

2. `Starlark`
如果后续需要稍复杂的脚本能力，可以考虑。

3. `Webhook`
适合复杂逻辑和外部依赖。

不建议优先做：

- 用户上传 Go 脚本
- 运行 Go plugin
- 直接执行 shell

## 7. 为什么推荐 CEL，而不是 Go 脚本

如果目标是“规则引擎里的动态表达式”，`CEL` 比 Go 脚本更适合。

原因：

- 它天然就是表达式语言，不是通用编程语言。
- 可控，容易限制能力边界。
- 适合做条件与字段计算。
- 可以提前编译和校验。
- 运行时行为更容易预测。

但也要注意边界：

- `CEL` 很适合表达式，不适合大段业务逻辑。
- 一旦需要复杂 I/O、外部调用、复杂状态机，应该转向 `webhook` 或受控扩展。

所以最佳组合通常是：

- 匹配条件：`AST + CEL`
- 响应拼装：`template + CEL`
- 复杂动态逻辑：`webhook`

## 8. 对“group 规则”的重构建议

如果你非常想保留 group 概念，可以这样重构：

- `Namespace`
用于租户、环境、业务域隔离。

- `RuleSet`
用于协议和场景级组织。

- `Rule`
用于单条匹配规则。

- `Action`
用于命中后的行为。

不要再把“组规则”和“命中组规则”拆成两种结构。

推荐直接统一为：

- `RuleSet.selector`
- `Rule.when`

也就是：

- 组级做粗筛。
- 规则级做精匹配。

这个比原设计更简单，也更容易解释给使用方。

## 9. 管理接口设计建议

既然你明确需要 HTTP 接口，建议管理面和运行面分开。

### 9.1 管理接口

建议前缀：

- `/mockserver/api/v1/admin/...`

建议接口：

- `POST /rulesets`
创建规则集

- `PUT /rulesets/:id`
更新规则集

- `GET /rulesets/:id`
查询规则集

- `GET /rulesets`
分页查询规则集

- `POST /rulesets/:id/validate`
校验规则集

- `POST /rulesets/:id/publish`
发布规则集

- `POST /rulesets/:id/simulate`
用输入样例模拟命中结果

- `POST /rulesets/:id/enable`
启用规则集

- `POST /rulesets/:id/disable`
禁用规则集

### 9.2 运行接口

建议前缀：

- `/mockserver/runtime/...`

例如 HTTP adapter：

- `/mockserver/runtime/http/*path`

也可以更进一步支持按 namespace 或 app 区分：

- `/mockserver/runtime/:namespace/http/*path`

### 9.3 非常建议增加的调试接口

- `POST /debug/match`
输入一个 event，返回命中的 ruleset、候选规则、最终命中规则、未命中原因。

这个接口极其重要。

没有它，规则一多就会进入“为什么没命中 / 为什么命中了错误规则”的排障地狱。

## 10. 存储与发布设计

不要让“编辑中的规则”和“运行中的规则”共用同一份内存结构。

建议采用双态模型：

- `draft`
- `published`

流程建议：

1. 用户保存 draft
2. 服务端校验 draft
3. 编译 draft
4. 发布成功后生成 compiled snapshot
5. runtime 原子切换到新 snapshot

这样可以保证：

- 发布前规则不影响线上流量
- 运行时不需要边查库边解释规则
- 支持版本回滚

运行时推荐：

- 维护一个只读 `compiled snapshot`
- 通过原子替换完成热更新

## 11. 可观测性与调试设计

这个项目很容易低估 debug 能力的重要性。

建议至少提供：

- 请求级 `trace_id`
- 命中规则日志
- 候选规则数量
- 最终 action 类型
- 执行耗时
- expression 或 template 执行错误明细

建议 runtime 返回时可选附带调试 header，例如：

- `X-Mockserver-Ruleset`
- `X-Mockserver-Rule`
- `X-Mockserver-Action`

开发环境开启，生产环境可配置关闭。

## 12. 安全性建议

如果系统后续要给多人使用，这一部分必须提前设计。

重点风险：

- 用户写的 regex 造成性能问题
- 用户写的 expression 造成高 CPU 消耗
- webhook 访问内网地址带来 SSRF 风险
- 动态模板泄露敏感字段
- 任意脚本执行带来宿主机风险

建议措施：

- 限制 regex 长度和复杂度
- expression 编译时和执行时都要限时
- webhook 做域名白名单和超时控制
- 动态能力按租户或环境做能力开关
- 明确禁止任意 Go 代码执行

## 13. 推荐的第一版范围

不要一开始就把“全协议 + 强脚本 + 状态机 + 数据库存储”全部做完。

推荐 V1 范围：

- 仅支持 `HTTP adapter`
- 规则 DSL 为 `RuleSet -> Rule -> Action`
- 支持 `all/any/not`
- 支持常见操作符
- 支持 `static_response`
- 支持 `template_response`
- 支持 `simulate`
- 支持 `publish`
- 支持 MySQL 存储和启动规则文件加载

推荐 V2：

- 支持 `CEL`
- 支持 `sequence_response`
- 支持 DB 存储
- 支持版本回滚

推荐 V3：

- 支持 `webhook_response`
- 支持更多 protocol adapter
- 支持多租户和权限控制

## 14. 推荐的数据结构方向

下面是一套更适合长期演进的核心 JSON 方向。

### 14.1 RuleSet

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

### 14.2 Rule

```json
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
      },
      {
        "field": "request.query.q1[*]",
        "op": "contains",
        "value": "qv1"
      },
      {
        "field": "request.headers.h1[*]",
        "op": "contains",
        "value": "hv1"
      }
    ]
  },
  "action": {
    "type": "template_response",
    "status": 200,
    "headers": {
      "content-type": ["application/json"]
    },
    "body_template": "{\"message\":\"hello {{ request.query.q1[0] }}\"}"
  }
}
```

## 15. 最终建议

综合来看，推荐你采用下面这套决策：

### 15.1 应该保留的

- 用 Go 实现
- 用 JSON 作为规则配置载体
- 提供 HTTP 管理接口
- 保留“先粗筛，再精匹配”的思想

### 15.2 应该调整的

- 不要把核心模型设计成“组规则 / 响应规则”两层散装结构
- 不要把 headers/query 设计成低效数组匹配结构
- 不要把任意 Go 脚本作为默认动态能力
- 不要在运行时直接解释未经编译的 JSON 规则

### 15.3 应该新增的

- 统一 `Event` 输入模型
- `RuleSet -> Rule -> Action` 模型
- 条件 AST
- 编译期校验与索引
- simulate/debug 能力
- draft/published 双态发布模型

## 16. 一句话判断

如果只问“当前方案是否 ok”，答案是：

`可以作为起点，但不建议直接照着实现。`

更优解不是抛弃 JSON，而是：

- `JSON DSL + 强类型编译 + 两阶段匹配 + 受控动态能力`

如果后续要正式开做代码，实现顺序建议是：

1. 先定 `Event`、`RuleSet`、`Rule`、`Action` 四个核心模型
2. 再定条件 AST 和操作符
3. 再实现编译器和匹配器
4. 再补 HTTP adapter 和管理接口
5. 最后再加 CEL / sequence / webhook

这会比“先写一套 JSON 规则引擎，再往里面塞脚本”稳得多。
