# ImageMaster

一个强大的漫画/图片查看和下载工具。

## 功能特点

- 本地漫画库管理
- 多种网站图片爬取支持
  - 通用网页爬虫
  - E-Hentai 支持
  - Telegraph 支持
- 并发下载
- 自动排序图片
- 图片预览

## 项目结构

```
ImageMaster/
├── core/                 # 核心功能
│   ├── crawler/          # 爬虫模块
│   │   ├── parsers/      # 不同网站的解析器
│   │   │   ├── ehentai.go  # E-Hentai 解析器
│   │   │   └── telegraph.go # Telegraph 解析器
│   │   └── crawler.go    # 爬虫工厂和接口
│   ├── downloader/       # 下载器
│   │   └── downloader.go # 下载器实现
│   ├── getter/           # 图片获取器
│   │   ├── local.go      # 本地图片获取器
│   │   ├── models.go     # 数据模型
│   │   └── remote.go     # 远程图片获取器
│   ├── logger/           # 日志模块
│   │   └── logger.go     # 日志实现
│   └── viewer/           # 查看器
│       └── viewer.go     # 查看器实现
└── main.go               # 入口文件
```

## 使用方法

1. 启动应用
2. 选择或添加漫画库
3. 浏览漫画
4. 或者输入网址爬取新漫画

## 技术栈

- Golang
- Wails框架
- 并发编程

## About

This is the official Wails Svelte-TS template.

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.
