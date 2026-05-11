# 背景

我需要新实现一个mockserver，核心功能是提供通用的能力，提供通用、可定义、易拓展的mock能力，可以集成各种各样的mock（eg:http）。

这是一个新的仓库，还未写任何代码，所以需要你帮忙完整实现所有需求。

## 核心需求

1. go语言实现。
2. 参考 backend_conventions skills 来设计后端。
3. 实现一套完整 基于json结构体的强大的规则引擎。
4. 实现相关的服务接口（http接口）

### 关于 基于json结构体的强大的规则引擎

核心思路：
1. 不限定mockserver本身可以进行哪些特定接口（http、rpc、db、cache等）的mock，这个主要依赖用户如何使用。
2. 规则引擎包含：命中组规则、组规则、响应规则。
3. 一个完整的流程是：用户请求，传入一个json -> 服务端解析json -> 根据规则确定命中的组 -> 根据规则确认命中的规则 -> 根据配置返回该规则的响应。

### 举例

#### 输入

```json
{
  "type": "http",
  "request": {
    "host": "https://demo.com",
    "path": "/api/v1/debug",
    "query_params": [
      {
        "key": "q1",
        "value": "qv1"
      }
    ],
    "headers": [
      {
        "key": "h1",
        "value": "hv1"
      }
    ]
  },
  "meta_data": {}
}
```

#### 确认命中的组

group rule:
```json
{
  "rules": [
    {
      "field": "type",
      "operator": "eq",
      "value": "http"
    },
    {
      "field": "request.header[*].key",
      "operator": "eq",
      "value": "h1"
    }
  ]
}
```

#### 确认命中的响应

```json
{
  "rules": [
    {
      "field": "path",
      "operator": "eq",
      "value": "/api/v1/debug"
    }
  ]
}
```

#### 根据配置构造响应并返回

核心支持2种返回：
1. 返回固定结果
2. 支持自定义的go脚本，用户需要实现某个指定的的函数，根据go脚本执行结果返回数据。

