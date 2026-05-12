# Mock 工具链

## 工具链

设想：mock工具链需要包含3个部分：

- mockserver
- mocksdk
- mockinject

### mockserver

即当前 cmd/server 服务的方式部署起来，用户可以进行ruleset/rules/namespace等管理。

### mocksdk

一个通用的sdk，提供通用的能力方便用户可以调用 mockserver 的接口，将原始请求转换成决策输入，从而判断是否命中 ruleset 和 rules，并决策出 Action（返回 Response 还是告知 mockinject 转发原始请求）。mocksdk 不负责执行真实转发。

mocksdk 已拆分为独立仓库和 Go module：`github.com/MrMiaoMIMI/mocksdk`。它需要支持较低版本的业务仓库 Go toolchain，当前基线为 Go 1.17，并且只使用 Go 标准库依赖。

### mockinject

特指一些用户客制化的工具，该工具会结合业务代码所用的技术栈，在代码编译前进行代码注入（例如替换默认的 http client，从而能够获取到原始的 http 请求，并能够控制 http 请求），注入的代码会调用 mocksdk 将原始请求发送给 mockserver 从而获得 Action。
由于是编译前进行代码注入，因此可以认为是对业务代码无侵入（注入的代码不会被提交）并达到mock的效果。

特别注意：mockinject 不属于本仓库要实现的范畴，因此只需要知道其概念，不需要实现它。

## 原理步骤

1. 使用 mockinject 编译业务代码，包括代码注入和代码编译。
2. 启动业务代码。
3. 业务代码调用 HTTP、cache、RPC 或 MQ 等协议能力时，被注入的代码接管，调用 mocksdk 的协议 normalizer 或 event builder 将原始调用转换成通用 Event，并发送给 mockserver。
4. mockserver 解析请求内容，并作出决策，然后通过http接口返回决策结果。
5. 注入的代码拿到决策结果进行 mock response 或者结合具体技术栈转发原始请求。

## 协议字段约束

mockserver 使用代码注册的 `ProtocolSpec` 描述每种协议的可用字段、动态路径、字段类型和 selectors。`ProtocolSpec` 不是用户在数据库中动态配置的协议定义；用户配置的是 ruleset/rule/namespace。

mocksdk 负责把具体技术栈的原始调用转换成 Event：

- HTTP：从 `http.Request` 投影出 `request.method`、`request.host`、`request.path`、`request.query`、`request.headers`、`request.body`、`request.raw_body` 等字段。
- Cache：通过 `mocksdk/cacheadapter.Event` 投影出 `request.operation`、`request.key`、`request.ttl_ms`、`request.value`。

mockinject 仍然负责真正执行原协议调用或转发，mockserver 和独立的 mocksdk module 只返回决策。
