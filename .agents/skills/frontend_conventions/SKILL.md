---
name: frontend-conventions
description: Use this skill when update frontend code.
---

# Frontend Skill

## 目标

在这个仓库里新增或修改前端能力时，优先遵守当前 Vue 应用已经形成的代码组织、请求链路、状态管理和样式体系，而不是引入另一套新的页面模板或 UI 语言。

适用范围：
- `web/src/**`
- `web/package.json`
- `web/vite.config.ts`
- `web/eslint.config.ts`
- `web/.prettierrc.json`

## 项目现状

前端真实技术栈：
- Vue 3
- TypeScript
- Vite
- Element Plus
- Pinia
- Vue Router
- Axios
- SCSS
- `@form-create/element-ui`
- `@form-create/designer`

关键入口：
- `web/src/main.ts`
- `web/src/router/index.ts`
- `web/src/utils/request.ts`
- `web/src/assets/design-system.scss`
- `web/src/assets/element-overrides.scss`

项目不是一个“纯 CRUD 页面集合”，而是包含普通管理页和资源工具工作台两类前端形态：
- 经典列表/表单/详情页
- 类 IDE 的工作台页：`resource-tool-workspace`

## 代码组织约定

按当前仓库真实结构工作：

- `src/api`：接口调用封装，按业务域拆分
- `src/types`：前端领域类型，与后端 response/request 保持字段对齐
- `src/store`：Pinia 业务 store，主状态目录
- `src/components/common`：跨业务通用组件
- `src/components/<domain>`：业务域组件
- `src/views/<domain>`：页面级视图
- `src/utils`：请求层、解析器、通用工具函数
- `src/assets` / `src/styles`：设计系统、覆盖样式、SCSS mixin

当前仓库同时存在 `src/store` 和 `src/stores`：
- `src/store` 是实际业务目录
- `src/stores` 目前是 Vue 模板遗留

后续新增状态逻辑时：
- 默认放到 `src/store`
- 不继续往 `src/stores` 扩写新业务代码

## 代码风格约定

从现有 ESLint / Prettier 配置看，当前前端代码风格应保持：
- `script setup` + TypeScript
- 单引号
- 不写分号
- `printWidth = 100`
- import 路径优先使用 `@/`

参考：
- `web/eslint.config.ts`
- `web/.prettierrc.json`
- `web/vite.config.ts`

新增代码时应遵守：
- 优先组合式 API，不写 Options API
- 优先 `const xxx = ref(...) / computed(...) / reactive(...)`
- 组件内逻辑按“imports -> store/router/route -> state -> computed -> methods -> lifecycle”排序
- 类型尽量从 `@/types` 导入，不在页面里临时内联一堆匿名结构

## 路由约定

路由统一在 `web/src/router/index.ts` 注册。

当前真实约定：
- 登录页用 `meta.public`
- 主应用页面挂在 `AppLayout` 下
- 需要登录的页面依赖父级或自身 `meta.requiresAuth`
- 菜单隐藏页用 `meta.hidden`
- 页面标题来自 `meta.title`

路由守卫行为：
- 公开页直接放行
- 已登录访问 `/login` 或 `/debug-login` 时自动跳 `/dashboard`
- 需要认证但无 token 时跳转到 `/login?redirect=...`

新增路由时：
- 必须补 `meta.title`
- 如果是主应用页，优先放到 `AppLayout` children 下
- 若不希望出现在菜单里，用现有的 `meta.hidden`
- 不另发明 `requiresLogin`、`authOnly` 之类平行字段

参考：
- `web/src/router/index.ts`

## 请求层约定

统一请求入口是：
- `web/src/utils/request.ts`

这里已经做了几件关键事：
- 基于 axios 创建唯一实例
- 自动序列化数组 query 参数
- 自动从 `localStorage.auth_token` 注入 `Authorization: Bearer <token>`
- 自动解包后端 `CommonResponse`
- 只有 `code === 0` 视为成功
- `401` 时清理本地登录态并跳登录页

因此新增 API 时要遵守：
- 一律基于 `http.get/post/put/delete`
- API 方法的返回值写“解包后的 data 类型”，不要写 `CommonResponse<T>`
- 让请求层负责通用错误处理，不在每个 API 文件里重复写拦截逻辑

推荐形态：

```ts
export const xxxApi = {
  list(params: ListXxxParams = {}): Promise<ListXxxResponse> {
    return http.get('/api/v1/xxx', { params })
  },
  get(id: number): Promise<XxxResponse> {
    return http.get(`/api/v1/xxx/${id}`)
  },
}
```

注意：
- 当前少量 API 方法会对返回值再做一次 `.then(...)` 提取字段，这在“后端返回结构与页面实际需要不完全一致”时是允许的。
- 不要再创建第二个 axios 实例。

参考：
- `web/src/utils/request.ts`
- `web/src/api/resource.ts`

## 类型约定

类型统一放在 `src/types`，按业务域拆分，例如：
- `resource.ts`
- `resource-tool.ts`
- `snapshot.ts`
- `team.ts`
- `common.ts`

约定：
- response 类型命名与后端保持接近，如 `ResourceResponse`
- 列表类型优先复用 `PaginatedResponse<T>`
- 查询参数继承 `BaseQueryParams`
- 表单态类型允许和 response 类型分开，例如 `ResourceFormData`

原因：
- 页面表单经常需要 string 化的 JSON、临时字段、额外 UI 状态
- 不应强迫表单直接复用后端响应结构

新增类型时：
- 先考虑复用 `PaginationParams / BaseQueryParams / PaginatedResponse`
- 与后端字段保持 snake_case 对齐，减少转换层
- 不要在页面里硬编码大量 `any`

当前允许但应谨慎的情况：
- 复杂动态配置、工作台执行结果这类弱结构数据，可以保留 `Record<string, any>` 或 `any`
- 但普通 CRUD 领域对象应优先补完整类型

参考：
- `web/src/types/common.ts`
- `web/src/types/resource.ts`

## API、Store、View 的职责边界

当前项目的前端链路更接近：

`api -> store -> view/component`

### 1. api 层

职责：
- 直接对接后端接口
- 定义单个业务域的 HTTP 方法

不要做：
- 不维护页面 UI 状态
- 不直接弹 `ElMessage`
- 不直接读写页面组件状态

### 2. store 层

职责：
- 持有跨页面或页面级共享状态
- 封装业务操作
- 管理 loading、缓存、局部本地持久化
- 在需要时弹统一的成功/失败提示

当前 store 的典型形态：
- 使用 Pinia setup store
- `ref` 管状态
- `computed` 派生只读视图
- action 内调用 API
- 适度维护本地缓存，如数组、map、当前实体

已有稳定模式：
- 列表 store：`resources + total + loading`
- 详情读取后会回填本地缓存
- 某些 store 会把 UI 偏好或上下文写入 `localStorage`

例如：
- `auth` store 保存 token / user
- `team` store 保存 `selected_team_id`
- `resourceToolWorkspace` store 保存面板折叠状态、右侧模式、选中的变量集

约定：
- 涉及跨页面复用、上下文切换、工作台状态时，用 store
- 单页非常局部的临时状态，留在组件内即可
- store 名称用 `useXxxStore`
- store id 用稳定字符串，例如 `'resource'`、`'auth'`

参考：
- `web/src/store/resource.ts`
- `web/src/store/auth.ts`
- `web/src/store/team.ts`
- `web/src/store/resource-tool-workspace.ts`

### 3. view / component 层

职责：
- 组织页面布局
- 组合 store、route、router、业务组件
- 管理搜索表单、当前页码、弹窗开关等页面态

页面里常见职责：
- 收集 query form
- 调用 store 的 fetch/create/update/delete
- 处理分页、筛选、跳转
- 处理局部 UI 交互

不要做：
- 不在 `views` 里直接重写一遍 axios 请求逻辑
- 不把大块公共 UI 复制到多个页面

## 页面分型约定

当前前端页面大体分 4 类，新增页面时尽量向这几种模式靠拢。

### 1. List 页

典型页面：
- `ResourceList.vue`
- `ScriptList.vue`
- `SnapshotList.vue`

共同模式：
- `PageContainer`
- 顶部 `header-actions`
- `filter-bar`
- `table-wrapper`
- `pagination-bar`
- 本地 `currentPage / pageSize`
- `loadData()` 负责把搜索条件和分页合并后调用 store

约定：
- 搜索条件清洗空值再发请求
- 切换筛选时重置到第一页
- 删除成功后重新刷新列表
- 表格操作列按钮文案尽量统一：查看 / 编辑 / 删除

### 2. Form 页

典型页面：
- `ResourceForm.vue`
- `ProjectForm.vue`
- `ScriptFormPage.vue`

共同模式：
- `PageContainer`
- `el-form` + `FormRules`
- 编辑态通过 `route.params.id` 判断
- 提交按钮统一带 `loading`
- 新增和编辑尽量复用同一个页面

约定：
- 表单展示态和提交态可以不同
- JSON 或复杂配置尽量复用 `ConfigEditor` / `ResourceConfigEditor`
- 先前端校验，再调用 API 或 store
- 成功后统一跳详情页或列表页，不留在不确定状态

### 3. Detail 页

典型页面：
- `ResourceDetail.vue`
- `ProjectDetail.vue`
- `ScriptDetail.vue`

共同模式：
- `PageContainer`
- header 区域放返回 / 编辑等动作
- 主体展示摘要、元信息、结构化详情
- 加载失败时显示骨架或空态

约定：
- 详情页优先做“读视图”，不要强塞编辑逻辑
- 复制、下载、查看关联项这类只读增强能力可放详情页

### 4. Workspace 页

典型页面：
- `ResourceToolWorkspace.vue`

这是仓库里的特殊页面模型，不应按普通 CRUD 页处理。

共同模式：
- 页面本身只负责三栏布局和路由参数
- 大量交互状态下沉到专用 store
- 左右面板宽度、折叠状态、活动 tab 等是工作台级状态

约定：
- 工作台状态优先放到 `resource-tool-workspace` store
- 工作台不要引入普通列表页那套布局习惯
- 对多面板、多 tab、多结果缓存的逻辑，优先抽成 store 和业务组件，不堆在单个 `.vue` 文件里

## 通用组件约定

优先复用已有通用组件，而不是每个页面重新造一套：

- 布局容器：`PageContainer`
- 面包屑：`PageBreadcrumb`
- 描述展示：`DescriptionCell`、`DescriptionDisplay`
- JSON/配置编辑：`ConfigEditor`、`JsonEditor`
- 模板变量展示：`TemplateVarsCard`
- 关联项展示：`RelatedToolsCard`

选择原则：
- 多个业务域都要用的组件，放 `components/common`
- 单一业务域使用的组件，放 `components/<domain>`
- 通用组件应接受清晰 props 和 events，不耦合某个 store

注意：
- `ConfigEditor` / `JsonEditor` 已经是现成能力，不要退回到大块 `textarea`
- Teleport 场景或覆盖第三方组件时，少量非 `scoped` 样式是允许的，但应有明确理由

参考：
- `web/src/components/common/PageContainer.vue`
- `web/src/components/common/README.md`

## 布局与导航约定

主应用布局由 `AppLayout` 提供：
- 左侧边栏
- 顶部导航
- 主内容区 `router-view`
- 路由切换动画

已有稳定行为：
- 侧边栏折叠状态存在 `localStorage`
- 小屏默认偏向折叠
- team 切换会触发工作台页的特定重置逻辑

因此新增全局行为时：
- 先看是否应放在 `AppLayout` / `AppHeader` / `AppSidebar`
- 不要在业务页面里各自实现“全局”导航状态

参考：
- `web/src/components/layout/AppLayout.vue`

## 样式系统约定

### 1. 设计系统

全局设计 token 定义在：
- `web/src/assets/design-system.scss`

里面已经定义：
- `--ms-*` 色彩、间距、圆角、阴影、字体、动画、尺寸、z-index

### 2. Element Plus 覆盖

统一覆盖在：
- `web/src/assets/element-overrides.scss`

这层负责把 `--el-*` 映射到 `--ms-*`。

### 3. SCSS 基础设施

当前 Vite 已启用全局 SCSS mixin 注入：
- `@use "@/styles/mixins" as *;`

这意味着：
- 业务组件默认可以用项目级 mixin
- 共享布局模式应优先沉淀到 `src/styles/*`

### 4. 样式编写规则

默认规则：
- 组件样式优先 `<style lang="scss" scoped>`
- 使用设计系统变量，不直接硬编码颜色、圆角、阴影
- 优先复用已存在的列表页/详情页/卡片样式模式

尽量避免：
- 到处写 hex 颜色
- 到处写内联 `style="..."`
- 重复定义 `list-container / filter-bar / pagination-bar` 这一类公共模式

如果需要新增通用样式模式：
- 优先加到 `src/styles/*` 的 mixin 或共享样式里
- 再在具体组件里引用
- 不要复制 10 份近似 CSS

这个方向和仓库内已有前端样式总结是一致的。

## 交互提示与错误处理约定

当前项目里，用户提示大多通过 `ElMessage` / `ElMessageBox` 实现。

约定：
- 删除确认统一用 `ElMessageBox.confirm`
- 成功提示通常在 store action 内或页面提交成功后给出
- 失败提示不要重复轰炸；请求层已经处理了一部分通用错误

建议：
- 领域动作的成功/失败提示优先集中在 store
- 页面只处理自己独有的交互提示
- 对可恢复错误，保留原始错误消息优先于硬编码“失败”

## 鉴权与上下文约定

### 1. Auth

登录态来源：
- `localStorage.auth_token`
- `localStorage.auth_user`

认证相关逻辑：
- 请求头注入在 `request.ts`
- 路由拦截在 `router/index.ts`
- 用户状态管理在 `store/auth.ts`

约定：
- 不在每个页面里手工判断 token
- 不在每个 API 里手工拼 Authorization

### 2. Team Context

这个项目的前端不是简单的“全局 project 选择器”模式，而是有 team 上下文。

`team` store 负责：
- 我的团队列表
- 当前选中的 team
- 当前 team 的 project ids
- 当前用户是否 admin
- team 切换版本号

约定：
- 涉及 project 作用域的页面，在加载前先考虑 `ensureTeamContext()`
- 当页面支持 team 过滤时，优先复用 `selectedTeamId`
- 不要在各页重复实现 team 初始化逻辑

参考：
- `web/src/store/team.ts`

## 与后端协同的约定

这个仓库是前后端分离，但结构耦合度其实比较高，前端应围绕后端现实来做，而不是假设存在 OpenAPI 自动生成层。

约定：
- 字段命名尽量保持和后端 response 一致，默认使用 snake_case
- 列表接口统一兼容 `results + pagination`
- 认证失败依赖后端 `401`
- 特殊 FE 定义接口要复用，如资源配置字段定义接口

特别是资源相关模块：
- 不只是 CRUD，还依赖后端返回的配置 schema / define
- 前端应优先消费这些定义来构建动态配置界面

## 新增一个前端业务模块的建议步骤

1. 先确认后端是否已经有稳定接口和字段结构
2. 在 `src/types/<domain>.ts` 增加 response / request / params 类型
3. 在 `src/api/<domain>.ts` 增加 API 封装
4. 评估是否需要 `src/store/<domain>.ts`
5. 实现 `views/<domain>/List.vue / Form.vue / Detail.vue` 或实际需要的页面
6. 复用现有 `PageContainer`、表格、编辑器、说明类组件
7. 在 `router/index.ts` 注册路由和 `meta.title`
8. 如果有导航入口，再接入侧边栏或业务入口
9. 补齐必要的筛选、分页、空态、加载态

## 当前仓库里的不一致点

写新代码时，优先向“更稳定的一侧”收敛，不继续扩散这些遗留问题：

- `src/store` 和 `src/stores` 并存：新代码统一用 `src/store`
- barrel export 不完整：例如部分 store / api 仍被直接路径导入，新增代码按现有模块实际情况处理，不强行一次性大重构
- 样式模式有重复：新代码优先复用和抽取，不复制已有重复块
- 某些接口的返回值和页面预期存在历史兼容写法：新增模块先以“类型清晰、接口单一职责”为目标

## 禁止事项

- 不新建第二套请求封装
- 不绕开 `request.ts` 直接在页面里写裸 axios
- 不在 `views` 里堆一整套可复用组件
- 不继续往 `src/stores` 放新业务逻辑
- 不大量硬编码颜色、间距、圆角
- 不在多个页面复制同一份列表页 CSS
- 不手工在每个页面里管理 token 注入和 401 跳转
- 不随意改动后端字段名再在前端做一层私有映射，除非有明确业务价值

## 提交前自检

- 路由是否补了 `meta.title`
- API 是否复用 `http` 实例
- 返回类型是否写成解包后的 `Promise<T>`
- 类型是否放在 `src/types`
- 状态是否应进 `store` 而不是页面本地乱存
- 页面是否复用了 `PageContainer` 和现有通用组件
- 样式是否使用 `--ms-*` 设计 token
- 是否避免了重复的列表页 / 详情页样式块
- 鉴权相关逻辑是否沿用 `request.ts + router + auth store`
- 涉及 team / project 上下文时，是否复用了 `team` store
