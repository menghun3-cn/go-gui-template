# Magpie 技术栈与实现原理研究

## 1. 结论

[yetone/magpie](https://github.com/yetone/magpie) 是一个使用 **Go + Wails v3** 的桌面应用：

- GUI 外壳：**Wails v3**，本质是“Web 前端 + 系统 WebView + Go 后端”。
- 前端：内嵌的**纯静态 HTML / CSS / JavaScript**（`internal/gui/assets`），不是 Electron。
- 后端：**Go**，同时提供本地 HTTP gateway、TUI、CLI、GUI。
- 跨平台 WebView：
  - Windows：**WebView2（Microsoft Edge Runtime）**
  - macOS：**WKWebView**
  - Linux：**WebKitGTK 4.1（GTK3）**
- 自动更新：**自研更新模块**拉取独立发布仓库 `yetone/magpie-releases` 的资源，校验 SHA-256 后替换自身并重启。
- 单文件形态：Windows/Linux 发布 **单个可执行文件**，前端资源通过 `go:embed` 嵌入到二进制里；macOS 发布 `.app` 或压缩包。
- 自启动：通过桌面平台机制实现 **“开机/登录时启动”**（Wails / app 层），不是靠固定的系统服务。
- 语言构成：Go 约 63%，JavaScript/HTML/CSS 约 35%，另有少量 C/Objective-C/Shell 用于平台原生胶水。

## 2. 仓库基本信息

| 项目 | 值 |
| --- | --- |
| 仓库 | `https://github.com/yetone/magpie` |
| 默认分支 | `main` |
| 许可证 | MIT |
| 主语言 | Go |
| 最新 activity | 2026-10 仍活跃 |
| 发布源码仓库 | `yetone/magpie-releases`（构建、签名、上传安装包） |
| 更新 feed | `usemagpie.ai/api/latest` |

## 3. 技术栈明细

### 3.1 后端 / 语言

- **Go**：所有核心业务都用 Go 写，包括本地 gateway、配置、更新、GUI shell、daemon、CLI、TUI。
- 主要 Go 依赖（来自 `go.mod`）：
  - `github.com/wailsapp/wails/v3`：桌面 GUI 框架
  - `github.com/charmbracelet/bubbletea` + `bubbles` + `lipgloss`：TUI
  - `modernc.org/sqlite`：SQLite（避免 CGO）
  - `github.com/tidwall/gjson`：JSON 解析
  - `github.com/go-ole/go-ole`：Windows COM/OLE
  - `golang.org/x/sys`：平台系统调用
  - `github.com/klauspost/compress`：压缩
  - `gopkg.in/yaml.v3` / `github.com/pelletier/go-toml/v2`：配置格式

### 3.2 前端

- 不是 React/Vue/Electron。
- 页面文件放在 `internal/gui/assets/`：
  - `index.html`：单页入口
  - `app.js` / `app.css`
  - `kernel.js`：GUI 内核（窗口通信、错误边界、插件 runtime）
  - 大量 SVG/PNG 图标
- 前端通过 `//go:embed` 直接嵌入 Go 二进制，不依赖外部静态服务器。
- `internal/gui/web.go` 提供 HTTP handler：
  - 嵌入的前端资源
  - `/api/*` JSON API
  - 页面内容 hash version 缓存
- 兼容性要求：同一个页面需要跑在 Chromium（Windows WebView2）、WebKit（macOS、Linux WebKitGTK）。
- GUI 测试用 Playwright 的 Chromium 和 WebKit。

### 3.3 GUI 外壳

使用 **Wails v3**（beta），而不是 Tauri / Electron / Fyne / Walk：

- Wails 负责创建原生窗口、托盘图标、系统菜单、WebView 生命周期。
- Go 端通过 Wails 提供“native host”能力：
  - 托盘图标（tray icon）
  - 面板/窗口显示隐藏
  - 打开 URL 或文件夹
  - 剪贴板
  - 窗口自适应、标题栏
  - 单实例 handover
  - 登录时启动
- Windows 上 Wails 使用 **WebView2 Runtime**；如果未安装，magpie 会在启动前用系统 `MessageBox` 提示安装。
- Linux 上默认使用 GTK4/WebKitGTK6，但 magpie 通过 `gtk3` tag 降到 **GTK3 + WebKitGTK 4.1** 以获得更广兼容。
- macOS 上构建成菜单栏应用（`LSUIElement`，不占 Dock），并有大量 AppKit 适配代码。

## 4. 单 exe 原理

Windows 和 Linux 桌面版是**单文件可执行程序**：

1. 全部前端资源用 `//go:embed internal/gui/assets` 编进二进制。
2. Go 编译时 `-trimpath -ldflags="-s -w -X main.version=... -H windowsgui"`：
   - 去掉符号表，缩小体积
   - 写入编译时版本号
   - Windows 下 `-H windowsgui` 隐藏控制台
3. Windows 图标/清单/版本资源由 `github.com/tc-hib/go-winres` 生成 `.syso` 后链接进 exe。
4. Windows/Linux 交叉/原生编译成 `dist/magpie-windows-amd64.exe`、`dist/magpie-linux-amd64` 这样的单文件。
5. Windows 不嵌 WebView2 runtime，而是要求系统安装 WebView2；这样可以显著减小体积并避免绑定私有 runtime license。
6. 用户数据存在用户目录（`.config/magpie` 等），不放在 exe 旁边，因此 exe 可以“便携”但更新/配置走用户目录。

## 5. 自动更新原理

magpie 的桌面 app 更新链路（`internal/gui/update.go`、`internal/update/*`）：

1. 检查更新 feed：
   - 默认 `usemagpie.ai/api/latest`
   - 也可配置 `--proxy` / `--mirror`
2. 拿到的 release 数据结构包含：
   - `version`
   - 每个资产路径
   - 每个资产的 **SHA-256**
3. 桌面版下载对应平台资产：
   - macOS：`magpie-darwin-<arch>.zip`（整个 `.app` 压缩包）
   - Windows/Linux：对应单文件二进制
4. 下载后放到“待安装位置”，校验 SHA-256。
5. 通过“同签名”校验：macOS 上接受的更新必须由同一开发者签名。
6. `Stage` 在目标目录旁准备新版本。
7. `Install` 替换当前 bundle/二进制：
   - 普通权限能替换就直接替换
   - 权限不足时 `NeedsAdmin` / `InstallAsAdmin` 走提权路径
8. `Relaunch` 重新启动应用：
   - 保存当前进程 pid
   - 等自身退出
   - 用当前参数/窗口状态重新打开新版本

### 关键技术点

- 更新源是**独立 release 仓库** `yetone/magpie-releases`，主仓库只触发构建并镜像发布。
- 主仓库 push tag 后，通过 GitHub Actions + `repository_dispatch` 通知 `magpie-releases` 构建发布。
- self-update 只从 `magpie-releases` 读取，与主仓库 Releases 解耦。
- 更新不一定要求安装器；Windows/Linux 直接替换 exe 即可。
- 干净替换正在运行的 exe 会受 Windows 文件锁限制，因此代码里有“被占用/等待/提权”处理，并且支持通过重启打开新版本。

## 6. 自启动 / “在线更新后自启动”原理

“更新后自启动”包含两层含义，magpie 都有对应实现：

### 6.1 更新完成后自动重启

- `internal/update/update.go` 的 `Relaunch`：
  - 计算下一个版本启动命令
  - 当前进程退出后启动新 binary / bundle
- 因此“在线更新完”不会停在 `success` 页面，而是直接拉起新版本。

### 6.2 开机/登录时自启动

- GUI shell 的 Wails native host 提供 `open at login` 能力。
- “Open at login” 是设置项，不只是登录会话的启动项。
- 平台实现：
  - macOS：LaunchServices / Login Items
  - Windows：注册表/启动项（Wails 或宿主层实现）
  - Linux：通常是 `.desktop` autostart（如果启用）
- magpie 同时有 `internal/daemon`：
  - macOS：`launchd`
  - Linux：`systemd`
  - 但这部分更偏后台 daemon/常驻能力，GUI 的登录启动不一定依赖 daemon。

## 7. CLI / TUI / Web 模式

单一 Go 二进制还提供：

- `magpie` / `magpie app`：启动 GUI
- `magpie tray`：只启动托盘
- `magpie panel`：面板
- `magpie web`：在内嵌本地 HTTP server 上打开网页版
- `magpie tui`：Charm 系 TUI
- `magpie update`：命令行更新
- 其他 provider/group/quota/settings 命令

这可以用 `-tags nogui` 或运行时分支实现。GUI 代码在 `internal/gui`，CLI 在 `internal/cli`，业务在 `internal/gateway` 等。

## 8. 对我们模板的落地启示

我们后续目标是“Go GUI 模板工程，能自动更新、能打包单 exe”，建议按 magpie 的架构做精简版：

```text
go-gui-template/
  cmd/your-app/
    main.go            # -ldflags 注入 version
  internal/gui/
    app.go             # Wails/WebView shell + 托盘 + 窗口
    assets/            # 前端静态资源，go:embed
    update/            # 更新检查、下载、校验、替换、重启
    server/
  internal/update/
    feed.go
    download.go
    verify.go
    replace_win.go     # Windows 替换/提权/等待文件锁
    relaunch_win.go
    replace_unix.go
    relaunch_unix.go
  internal/platform/
    autostart_win.go
    autostart_mac.go
    autostart_linux.go
  build/
    windows/winres.json
    icons/
  scripts/
    release.ps1
    update-server-schema.json
  Makefile
  go.mod
```

### 模板默认技术选型建议

- **GUI 框架**：Wails v3（与 magpie 同款，避免学习和 Wails 之外的坑）
- **前端**：小工具直接用原生 HTML/JS/CSS + `go:embed`；页面复杂再考虑 React/Vite 构建产物嵌入
- **自动更新**：复用 magpie 思路：
  - 更新 feed：`GET https://your-host/app/releases/latest.json`
  - feed 返回 `{version, url, sha256, notes}`
  - Windows/Linux 下载单文件 exe/二进制，校验 SHA-256，替换自身，重启动
  - macOS 下载 zip bundle，替换 `.app`
- **单 exe**：
  - Windows 用 `go-winres` 生成图标/清单/版本资源
  - `go build -tags production -ldflags "-s -w -H windowsgui"` 得到单 exe
  - 依赖系统 WebView2（体积小）
- **自启动**：
  - Windows：注册表 `HKCU\...\Run`
  - Linux：`~/.config/autostart/*.desktop`
  - macOS：Login Items / launchd
- **测试**：
  - Go 单测
  - Playwright Chromium/WebKit 跑 GUI 页面

## 9. 已验证的仓库证据

| 证据 | 来源 |
| --- | --- |
| Go 1.26 + Wails v3 | `go.mod` |
| 前端 `go:embed` | `internal/gui/app.go` / `assets` |
| 生产 tag `production` | Makefile |
| Windows 单 exe | Makefile `release-windows`、`winres.json` |
| WebView2 检测 | `internal/gui/webview2.go` |
| Linux WebKitGTK 4.1/GTK3 | Makefile、`internal/gui/webkit_linux.go` |
| 更新 feed 与 SHA-256 | `internal/update/update.go` |
| 更新后 relaunch | `internal/update/update.go`、`Relaunch` |
| 独立 release 仓库 | `.github/workflows/release.yml` |
| 登录启动 | `docs/subsystems/gui-shell.md` 的 Wails native host 表 |

## 10. 备注

- 在线更新的“签名认证”：macOS 有明显 `sameSigner` 校验；Windows 建议在模板里加代码签名或至少固定校验发布清单哈希。
- Wails v3 仍是 beta，需要在模板中锁定版本，避免升级破坏。
- 单 exe 不等于把 WebView2 runtime 打包进去；如果业务要求完全不依赖系统 runtime，需要额外考虑 WebView2 FixedVersion / Evergreen bootstrapper，会显著增大体积。
