# ImageMaster

## 安卓端软件：https://github.com/TyrEamon/comicr

一个基于 `Wails + Go + Vue 3` 的本地漫画/图片下载与管理工具，支持多站点链接抓取、本地漫画库、批量解压与简繁统一搜索。

## 功能特点

- 本地漫画库管理：添加多个漫画库目录，切换活动漫画库，按文件夹浏览本地内容
- 多站点链接下载：输入作品链接后创建下载任务，并写入历史记录
- 下载任务管理：支持任务列表、完成/取消状态提示与历史查看
- Extract 解压管理：扫描漫画库中的 `zip / cbz / 7z / rar`，区分待解压与判定已解压，支持单个解压和批量解压
- 搜索增强：首页搜索支持简繁统一和常见日文字形归一，查找标题更顺手
- 版本信息：设置页可查看当前版本号、提交哈希和构建时间

## 支持站点

当前代码里明确注册的站点包括：

- `e-hentai.org`
- `exhentai.org`
- `telegra.ph`
- `telegraph.com`
- `wnacg.com`
- `nhentai.xxx`
- `hitomi.la`
- `18comic.vip`
- `18comic.org`

常见链接样式示例：

```text
https://e-hentai.org/g/{gallery-id}/{token}/
https://exhentai.org/g/{gallery-id}/{token}/
https://telegra.ph/{slug}
https://www.wnacg.com/photos-index-aid-{id}.html
https://nhentai.xxx/g/{id}/
https://hitomi.la/{category}/{slug}-{id}.html
https://18comic.vip/photo/{id}
```

说明：

- 建议优先使用具体作品页，不要直接使用首页、分类页或搜索页
- `18comic` 目前属于 best-effort 适配，站点结构变化、访问限制或 `403` 都可能导致失败
- README 早期提到的“通用网页爬虫”在当前代码里并没有完整落地，请以实际注册的解析器为准

## 项目结构

```text
ImageMaster/
├─ core/
│  ├─ archive/        # 解压扫描与批量解压
│  ├─ config/         # 配置读写
│  ├─ crawler/        # 爬虫入口与各站点解析器
│  ├─ download/       # 下载器
│  ├─ history/        # 下载历史
│  ├─ library/        # 本地漫画库扫描与管理
│  ├─ logger/         # 日志
│  ├─ meta/           # 版本信息
│  ├─ request/        # 请求封装
│  ├─ task/           # 下载任务管理
│  ├─ types/          # 接口与公共类型
│  └─ utils/          # 通用工具
├─ front/             # Vue 3 前端
├─ .github/workflows/ # GitHub Actions 构建
├─ main.go            # Wails 应用入口
└─ version.go         # 默认版本号与构建信息
```

## 使用方法

1. 首次启动后，先在 `Setting` 中配置下载目录、漫画库目录和代理
2. 在 `Home` 页浏览当前活动漫画库，并使用搜索框查找本地本子
3. 在 `Download` 页输入支持站点的作品链接，创建下载任务
4. 在 `Extract` 页扫描漫画库中的压缩包，并按需单个或批量解压
5. 需要排查问题时，可在 `Setting` 页查看日志目录、当前日志文件和版本信息

## 技术栈

- Go
- Wails v2
- Vue 3
- Pinia
- Vue Router
- Tailwind CSS 4
- opencc-js
- pnpm / Vite

## About

ImageMaster 是一个桌面应用，不是传统网站。前端界面由 `Vue 3` 实现，后端能力由 `Go` 提供，再通过 `Wails` 打包成 Windows 桌面程序。

## Live Development

本地开发前请先安装：

- Go
- Node.js
- pnpm
- Wails CLI

开发模式启动方式：

```bash
cd front
pnpm install
cd ..
wails dev
```

如果希望只单独调试前端，也可以在 `front` 目录里运行：

```bash
pnpm install
pnpm dev
```

## Building

本地构建可执行文件：

```bash
wails build
```

项目内也提供了 GitHub Actions 工作流：

- 手动触发 `Build Windows EXE`
- 或推送 `v*` 标签后自动构建
- 构建产物会带版本号，例如 `ImageMaster-0.2.0.exe`

## 配置文件

默认配置文件位置：

```text
%AppData%\imagemaster
```

这是一个无扩展名的 JSON 文件，常见字段包括：

- `output_dir`：下载目录
- `libraries`：漫画库目录列表
- `active_library`：当前活动漫画库
- `proxy_url`：代理地址
- `bandizip_path`：Bandizip CLI 路径

## 解压说明

`Extract` 页依赖外部 `Bandizip` 命令行工具。

建议优先填写：

```text
D:\bandizip\bz.exe
```

如果填写的是 `Bandizip.exe`，程序会优先尝试同目录下的 `bz.exe`。解压状态判定规则为：

- 目标目录中已有子文件夹，视为已解压
- 目标目录中已有非压缩包文件，例如图片，视为已解压
- 其余情况视为待解压

## 注意事项

- 部分站点可能需要代理、登录状态或更稳定的网络环境
- `18comic` 等站点可能因页面结构变化或访问限制导致失败
- `Bandizip` 解压到某些同步盘目录时可能不稳定，建议优先使用 `bz.exe`
- 当前 Windows 构建依赖系统 `WebView2`
