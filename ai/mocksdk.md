# 概述

一个通用的sdk，提供通用的能力方便用户可以调用 mockserver 的接口，将原始请求转换成决策输入，从而判断是否命中 ruleset 和 rules，并决策出 Action（返回 Response 还是告知 mockinject 转发原始请求）。mocksdk 不负责执行真实转发。

## 核心设计

- 提供给 mockinject 使用
- 制定通用的输入协议，让 mockinject 可以将原始请求转换成通用的结构数据，然后调用 mockserver 的接口来决策。
- 制定通用的输出协议，将从 mockserver 获取到的决策返回给 mockinject，从而方便 mockinject 控制业务进程的行为。
- 决策包含2类：返回Response 和 告知 mockinject 转发原始请求。
- 转发原始请求的具体实现由 mockinject 结合业务技术栈负责，例如 HTTP client、SPEX/gRPC client、cache client 或 MQ client。
- Event 使用通用 `request map[string]any`，字段由 mockserver 代码注册的 `ProtocolSpec` 约束；mocksdk 通过代码 normalizer 将具体协议调用转换成 Event，不从数据库动态读取 ProtocolSpec 来转换。
- 第一版提供 HTTP normalizer 和 cache event builder。HTTP 提供 `request.method/path/headers/query/body/raw_body` 等字段；cache 提供 `request.operation/key/ttl_ms/value`。
- 支持通过环境变量设置特殊信息，并且优先级最高，例如：namespace_id、mockserver_host。
