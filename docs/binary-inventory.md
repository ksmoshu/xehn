# 二进制与依赖清单

XEHN 出厂二进制 `H:\Program Files\XEHN\xehn.exe` 的依赖、端点与外部组件清单。全部为静态观测，可用第 8 节命令复现。

## 1. 构建元数据

| 字段 | 值 |
|---|---|
| 模块路径 | `github.com/xhwarp/warp-client` |
| 模块版本 | `v0.0.0-20261030130144-463792ab6313+dirty` |
| Go 工具链 | `go1.27.0` |
| GOOS / GOARCH | `windows` / `amd64` |
| GOAMD64 | `v1` |
| CGO_ENABLED | `0` |
| trimpath | `true` |
| buildmode | `exe` |
| VCS | `git`，`vcs.revision=463792ab6313…`，`vcs.modified=true` |
| 大小 | 20,978,176 字节 |
| SHA-256 | `A016C0FC32307A0BBA6272FE9B0C3379069940A8806A19E9B8C469A519361D87` |

## 2. 直接依赖（7）

| 模块 | 版本 | 用途 |
|---|---|---|
| `github.com/amnezia-vpn/amneziawg-go` | `v0.2.20-0.20260709022233-c1e9bb3758e7` | AmneziaWG 协议（抗 DPI 的 WireGuard） |
| `github.com/charmbracelet/bubbles` | `v0.21.0` | TUI 组件 |
| `github.com/charmbracelet/bubbletea` | `v1.3.10` | TUI 框架 |
| `github.com/charmbracelet/lipgloss` | `v1.1.0` | TUI 样式 |
| `github.com/muesli/termenv` | `v0.16.0` | 终端能力探测 |
| `golang.org/x/crypto` | `v0.53.0` | 加密原语 |
| `golang.org/x/net` | `v0.56.0` | HTTP/2、IDNA |
| `golang.org/x/sys` | `v0.46.0` | Windows 系统调用 |

## 3. 间接依赖（部分，共 43 项）

| 模块 | 版本 | 用途 |
|---|---|---|
| `github.com/Diniboy1123/connect-ip-go` | `v0.0.0-20260613064811-66cba32d7d33` | MASQUE 的 IP 隧道承载 |
| `github.com/Diniboy1123/usque` | `v1.5.1-0.20260720063354-6aa03fc97d12` | MASQUE 客户端实现 |
| `github.com/quic-go/quic-go` | `v0.60.0` | QUIC / HTTP3 |
| `github.com/quic-go/qpack` | `v0.6.0` | QPACK 头压缩 |
| `gvisor.dev/gvisor` | `v0.0.0-20260616165937-8e4bc62602eb` | 用户态网络栈 |
| `golang.zx2c4.com/wintun` | `v0.0.0-20230126152724-0fa3db229ce2` | TUN 虚拟网卡驱动 |
| `golang.zx2c4.com/wireguard` | `v0.0.0-20260522210424-ecfc5a8d5446` | WireGuard 实现 |
| `github.com/lxn/walk` | `v0.0.0-20210112085537-c389da54e794` | 原生 Win32 GUI 控件 |
| `github.com/lxn/win` | `v0.0.0-20210218163916-a377121e959e` | Win32 API 绑定 |
| `github.com/songgao/water` | `v0.0.0-20200317203138` | TUN/TAP 抽象 |
| `github.com/spf13/cobra` | `v1.10.2` | CLI 框架 |
| `github.com/spf13/pflag` | `v1.0.10` | 命令行解析 |
| `github.com/txthinking/socks5` | `v0.0.0-20260601051520` | SOCKS5 服务 |
| `github.com/things-go/go-socks5` | `v0.1.1` | SOCKS5 服务 |
| `github.com/vishvananda/netlink` | `v1.3.1` | Linux 网络配置 |
| `github.com/patrickmn/go-cache` | `v2.1.0+incompatible` | 内存缓存 |
| `github.com/google/btree` | `v1.1.3` | B 树索引 |
| `github.com/dunglas/httpsfv` | `v1.1.0` | HTTP Structured Field Values |
| `github.com/yosida95/uritemplate/v3` | `v3.0.2` | URI 模板 |
| `gopkg.in/Knetic/govaluate.v3` | `v3.0.0` | 表达式求值 |
| `github.com/inconshreveable/mousetrap` | `v1.1.0` | Windows 双击检测 |

> 完整 43 项见 `go.mod` 与 `go version -m` 输出。**无一为出厂独有**，即二进制依赖与公开源码完全一致。

## 4. 网络端点清单

### 4.1 Cloudflare WARP API

| URL | 方法 | 源码 | 频率 |
|---|---|---|---|
| `https://api.cloudflareclient.com/v0a4005/reg` | POST | `register.go:27` | 首次注册 / 换账号 |
| `https://api.cloudflareclient.com/` | GET | `register.go:28` | 可达性预检 |
| `https://api.cloudflareclient.com/v0a4471/reg` | POST | `masque.go:163` | MASQUE 账号注册 |

### 4.2 自建/第三方 relay

| URL | 源码 | 说明 |
|---|---|---|
| `https://edge-client-api.vercel.app` | `register.go:33` | 默认 relay；WARP 注册备用通道 |

### 4.3 连通性与出口信息

| URL | 源码 | 说明 |
|---|---|---|
| `http://www.gstatic.com/generate_204` | `client.go`, `connect.go`, `gui_native_windows.go` | 204 探针（HTTP 明文） |
| `https://api.ipify.org` | `client.go:826` | 出口 IP |
| `https://speed.cloudflare.com/meta` | `client.go:820` | 地区/城市 |
| `https://speed.cloudflare.com/__down` | `client.go` | 扫描测速 |
| `https://www.cloudflare.com/cdn-cgi/trace` | `client.go:827` | Colo 节点 |
| `https://dns.google/resolve` | `client.go:963` | DNS 泄露检测 |

### 4.4 VPN Gate

| URL | 源码 | 说明 |
|---|---|---|
| `https://www.vpngate.net/api/iphone/` | `vpngate.go:30` | 节点列表（CSV） |
| `http://ip-api.com/batch` | `vpngate.go:59` | 批量地区标注（**明文 HTTP**） |
| `http://ip-api.com/json/` | `gui_native_windows.go:2094` | GUI 地区显示（**明文 HTTP**） |

### 4.5 更新与资源

| URL | 源码 | 说明 |
|---|---|---|
| `https://api.github.com/repos/hezhanleiok/xehn/releases/latest` | `version.go:36`, `update.go:22` | 客户端版本检查 |
| `https://api.github.com/repos/MetaCubeX/mihomo/releases` | `download.go` | 核心下载 |
| `https://github.com/`（geosite/geoip 路径拼接） | `geodata.go` | Geo 数据下载 |

### 4.6 本机回环

| 端点 | 默认值 | 用途 |
|---|---|---|
| `127.0.0.1:7890` | `mixed-port` | HTTP + SOCKS 混合代理入站 |
| `127.0.0.1:9090` | `external-controller` | mihomo REST API |

### 4.7 已移除

| URL | 移除前源码 | 状态 |
|---|---|---|
| `https://raw.githubusercontent.com/hezhanleiok/xehn/main/control.json` | `control.go:30` | **已删除** |

## 5. 外部可执行文件（运行时，非嵌入）

| 文件 | 大小 | 许可 | 集成方式 |
|---|---|---|---|
| `core\mihomo.exe` | 61,247,488 字节 | GPL-3.0 | 独立进程 + `config.yaml` + REST 控制 |
| `core\psiphon.exe` | 22,463,488 字节 | GPL-3.0 | 独立进程 + 本地 SOCKS5 |

两者均**不**链接进主程序（法律上按「独立进程 ⇒ 聚合」处理，见 `THIRD_PARTY_NOTICES.md`）。

## 6. 运行时生成的本地文件

| 文件 | 内容 | 敏感性 |
|---|---|---|
| `warpscout-account.json` | WARP 账号私钥 + token | **高**（设备身份） |
| `settings.json` | 用户偏好 | 低 |
| `state.json` | 当前会话（含 `proxy_snapshot` base64） | 低 |
| `core\config.yaml` | 生成的 mihomo 配置（含 WG 私钥） | **中** |
| `cache.db` / `cache.json` | 扫描结果 | 低 |
| `vpngate-*.json` | VPN Gate 节点与测速 | 低 |
| `health-owner` | 健康守护者 pid | 低 |
| `xehn.log` | 运行日志 | 低 |

> `pack.ps1` 打包时会**拒绝**把上述任何文件打进 ZIP，并额外扫描二进制中是否嵌入了本机 WARP 私钥（`pack.ps1:98-114`）。

## 7. 嵌入资源

| 资源 | 大小 | 说明 |
|---|---|---|
| `rsrc_windows_amd64.syso` | 79,714 字节 | manifest + 图标（`rsrc` 烘焙） |
| `logo.ico` | 78,282 字节 | 多尺寸程序图标 |
| `assets/logo.png` | 261,464 字节 | GUI 内品牌图 |
| `app.manifest` | 1,252 字节 | comctl32 v6 + PerMonitorV2 DPI |

## 8. 复现命令

```powershell
# 构建信息（模块 + 依赖 + VCS）
go version -m "H:\Program Files\XEHN\xehn.exe"

# 哈希与大小
Get-FileHash "H:\Program Files\XEHN\xehn.exe" -Algorithm SHA256
(Get-Item "H:\Program Files\XEHN\xehn.exe").Length

# PE 头（架构 / 子系统）
$fs=[IO.File]::OpenRead("H:\Program Files\XEHN\xehn.exe"); $br=New-Object IO.BinaryReader($fs)
$fs.Position=0x3C; $pe=$br.ReadInt32(); $fs.Position=$pe+4
"Machine=0x{0:X4}" -f $br.ReadUInt16()
$fs.Position=$pe+0x18+68; "Subsystem=$($br.ReadUInt16())"
$br.Close(); $fs.Close()

# 源码端点（在仓库根执行）
Get-ChildItem -Filter *.go -Recurse | Select-String 'https?://[a-zA-Z0-9./_\-]+' -AllMatches
```
