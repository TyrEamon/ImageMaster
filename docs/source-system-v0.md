# ImageMaster Source v0

## 目标

这份文档定义的是 `ImageMaster` 在线漫画源的 `v0` 约定。

当前阶段先以 `Go` 内置源的形式实现，等搜索、详情、阅读、下载链路稳定后，再把同一套约定外置成独立源文件。

这份规范的重点不是“兼容 Miru”，而是借鉴它的思路，做一套更适合 `ImageMaster` 自己的源系统。

## 当前路线

现在的推进顺序是：

1. 先做内置源
2. 跑通 `搜索 -> 详情 -> 章节 -> 在线阅读`
3. 把在线章节接到本地下载器
4. 再抽成外置源
5. 最后才考虑“安装源 / 更新源”

也就是说：

- 现在的 `Baozi`、`MangaDex` 是内置源样板
- 将来的外置源必须尽量对齐这份 `v0` 结构

## 设计原则

### 1. 源只负责解析

源负责：

- 搜索作品
- 获取作品详情
- 获取章节列表
- 获取章节图片 URL
- 在必要时返回图片请求头

### 2. 软件本体负责通用能力

软件本体负责：

- 下载
- 并发与重试
- 保存目录
- 历史记录
- 本地漫画库刷新
- UI 展示

这条边界很重要。

不要把“如何保存文件”写进源里。
源只负责把“可下载的数据”交给软件本体。

### 3. 先兼容内置，再兼容外置

外置源不是先决条件。

先把内置源接口稳定下来，再去做加载器、源目录、安装与更新。

## 核心能力

`v0` 只定义 3 个核心能力：

- `search`
- `detail`
- `read`

对应到当前 Go 代码里，大致是：

- `Search(query, page)`
- `Detail(itemID)`
- `Images(chapterID)`

下载暂时不算源能力，而是本体基于 `Images` 结果继续完成。

## 元数据结构

每个源至少要提供一份摘要信息：

```go
type Summary struct {
    ID           string
    Name         string
    Type         string
    Language     string
    Website      string
    Version      string
    BuiltIn      bool
    Capabilities []string
    Description  string
}
```

字段说明：

- `ID`: 全局唯一源 ID，例如 `baozi`
- `Name`: 显示名称
- `Type`: 当前先固定为 `manga`
- `Language`: 主要语言，例如 `zh`、`en`、`all`
- `Website`: 源站主页
- `Version`: 源版本号，不等于软件版本号
- `BuiltIn`: 当前是否为内置源
- `Capabilities`: 支持的能力列表
- `Description`: 简短描述

## 搜索结构

```go
type SearchItem struct {
    ID             string
    Title          string
    Cover          string
    Summary        string
    PrimaryLabel   string
    SecondaryLabel string
    DetailURL      string
}
```

要求：

- `ID` 必须稳定，后续可直接拿去请求详情
- `DetailURL` 是源站外链，方便调试和兜底
- `Cover` 可以为空

## 详情结构

```go
type ChapterItem struct {
    ID           string
    Name         string
    URL          string
    Index        int
    UpdatedLabel string
}

type DetailItem struct {
    ID        string
    Title     string
    Cover     string
    Summary   string
    Author    string
    Status    string
    Tags      []string
    DetailURL string
    Chapters  []ChapterItem
}
```

要求：

- `ChapterItem.ID` 必须能直接进入阅读接口
- `ChapterItem.URL` 保留原始章节地址，便于调试与外链打开
- `Index` 用于后续排序和“下一话”逻辑

## 阅读结构

```go
type ImageEntry struct {
    URL     string
    Referer string
    Headers map[string]string
}

type ImageResult struct {
    ComicTitle   string
    ChapterTitle string
    ChapterURL   string
    Images       []string
    Entries      []ImageEntry
    HasNext      bool
    NextURL      string
}
```

说明：

- `Images` 是兼容字段，给当前前端直接显示用
- `Entries` 是未来下载器真正应该优先消费的结构
- `Referer` 和 `Headers` 是为后面防盗链、图片下载适配预留的
- 如果源不需要特殊请求头，`Entries` 也建议带上 `URL`

推荐规则：

- 前端阅读器可继续直接用 `Images`
- 下载器未来优先读 `Entries`
- 如果 `Entries` 为空，再退回 `Images`

## 能力声明

当前约定的能力值：

- `search`
- `detail`
- `read`

未来如果要扩展，再新增：

- `download`
- `library`
- `auth`

但这些暂时不要提前做复杂。

## 错误处理

`v0` 阶段先保持简单：

- 找不到源：直接返回错误
- 源不支持某能力：返回明确错误
- 页面结构变化：返回解析失败错误

建议错误文案带上源名和动作，例如：

- `baozi detail failed: ...`
- `source mangadex does not support detail yet`

## 外置源的目标形态

虽然现在还没正式做外置源，但未来建议长成这样：

```text
sources/
  baozi/
    manifest.json
    index.js
```

其中：

- `manifest.json` 放元数据
- `index.js` 或其它入口文件实现 `search / detail / images`

未来外置源的运行时，不一定要完全兼容 Miru。

更推荐的是保留自己的轻量接口，例如：

```js
export default {
  meta: {
    id: 'baozi',
    name: 'Baozi',
    version: '0.1.0',
    type: 'manga',
    language: 'zh'
  },
  async search(query, page, ctx) {},
  async detail(itemId, ctx) {},
  async images(chapterId, ctx) {}
}
```

## 现在不做什么

为了避免过度设计，下面这些暂时不做：

- 不兼容 Miru 运行时
- 不做插件市场
- 不做远程源仓库安装
- 不做复杂权限系统
- 不把下载器塞进源

## 下一步建议

最合理的下一个节点是：

1. 让在线章节支持“下载到本地漫画库”
2. 下载器优先适配 `ImageEntry`
3. 再补 1 到 2 个真实源
4. 等源接口稳定后，再外置

这样后面再做源目录、源安装、源更新，就不会反复返工。
