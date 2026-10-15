# 模板使用说明

## 已实现骨架

- `internal/update`：更新检查、下载、SHA-256 校验、替换当前 exe。
- `internal/platform`：Windows 注册表自启动、Linux autostart。
- `internal/gui`：Wails v3 窗口，assets 用 `go:embed` 打包。
- `internal/server`：内嵌前端和 `/api`。
- `scripts/build-win.ps1`：Windows 单 exe 构建。

## 下一步

- 接入真正的更新签名（macOS 签名校验、Windows Authenticode）。
- 配置实际 feed URL。
- macOS 自启动接入 launchd/SMAppService。
- macOS `.app` 打包、公证、notarize。

## GUI 启动方式

```powershell
go run ./cmd/gogui
```

构建时默认启用 GUI。`-tags nogui` 可只编译 CLI。

## 自动更新流程

1. 启动时或点击“检查更新”。
2. `internal/update` 拉取 feed。
3. 对比版本号。
4. 下载平台资产。
5. 校验 SHA-256。
6. 替换当前 exe（Windows 先将旧 exe 改名再放入新 exe）。
7. 由调用方负责 relaunch。模板中 CLI 尚未自动 relaunch，可在 `replaceBinary` 成功后启动新进程。
