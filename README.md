# go-gui-template

Go + Wails v3 的单 exe 桌面应用模板，参考 magpie 的架构实现：

- Windows 单 exe / GUI
- 前端资源通过 `go:embed` 内嵌
- 自更新：feed + SHA-256 校验 + 替换 exe
- Windows 注册表自启动
- Linux `.desktop` autostart（macOS 占位待接）

## 环境

- Go 1.26（建议装到非系统目录，避免动全机 PATH）
- Windows 需要 WebView2 Runtime

## 模块

```text
cmd/gogui         入口，支持 gogui / gogui update
internal/app      组装
internal/gui      Wails 壳 + 前端资源 go:embed
internal/server   内嵌资源 HTTP + JSON API
internal/update   更新 feed / SHA-256 / 替换自身
internal/platform 自启动
```

## 构建

```powershell
.\scripts\build-win.ps1
```

或：

```bash
make build
```

## 更新 feed 格式

```json
{
  "version": "1.0.0",
  "assets": [
    {
      "os": "windows",
      "arch": "amd64",
      "name": "gogui-windows-amd64.exe",
      "url": "https://example.com/releases/1.0.0/gogui-windows-amd64.exe",
      "sha256": "<64 hex>"
    }
  ]
}
```

Signature/签名校验是生产上线前必须补充的部分；目前只有 SHA-256，不能防被篡改的 feed。
