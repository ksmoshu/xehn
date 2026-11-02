# REA 检测报告：XEHN 出厂二进制

对 `H:\Program Files\XEHN\xehn.exe` 的只读取证结果。每条结论标注**观测**（可直接用命令复现）、**推断**（基于观测的合理结论）、**未知**（无法从现有证据判定）。

## 1. 二进制身份（观测）

| 项 | 值 |
|---|---|
| 路径 | `H:\Program Files\XEHN\xehn.exe` |
| 大小 | 20,978,176 字节 |
| SHA-256 | `A016C0FC32307A0BBA6272FE9B0C3379069940A8806A19E9B8C469A519361D87` |
| 格式 | PE x86-64，Machine `0x8664`，Subsystem `2`（windowsgui） |
| 工具链 | `go1.27.0` |
| 模块 | `github.com/xhwarp/warp-client` |
| 模块版本 | `v0.0.0-20261030130144-463792ab6313+dirty` |
| 构建（VCS） | `vcs.revision=463792ab6313e322055879fca2bfc095bfd31aad`，`vcs.time=2026-10-30T13:01:44Z`，`vcs.modified=true` |
| 打包戳 | `20261030-2143`（二进制内偏移 9,242,216） |
| 符号 | 已剥离（`go tool nm` 报 `no symbols`，`objdump` 报 `no runtime.pclntab symbol found`） |

复现命令：

```powershell
go version -m "H:\Program Files\XEHN\xehn.exe"
```

## 2. 构建溯源结论

**观测**：用公开源码 `c319e7c` + `pack.ps1:83` 的参数重建，与出厂二进制的应用特征字符串集合比对为 **895 vs 895，零差异**。

**推断**：出厂二进制 = 公开源码 `c319e7c` 的内容。`+dirty` 来自打包时工作区停在 `463792a`，但 `gui_i18n_windows.go` 与 `gui_page_settings_windows.go` 中已含有随后作为 `c319e7c` 提交的两条 DNS 泄露文案。

详见 [`build-parity.md`](build-parity.md)。

**未发现**：任何未在公开源码中出现的额外依赖、隐藏网络端点或额外逻辑。

## 3. 依赖（观测）

`go version -m` 报告的 43 个依赖与 `go.mod` **逐项一致**，无出厂独有依赖。关键项：

| 依赖 | 版本 | 作用 |
|---|---|---|
| `github.com/amnezia-vpn/amneziawg-go` | `v0.2.20-0.20260709022233` | AmneziaWG（抗封锁 WireGuard） |
| `github.com/Diniboy1123/usque` | `v1.5.1-0.20260720063354` | MASQUE 协议 |
| `github.com/Diniboy1123/connect-ip-go` | `v0.0.0-20260613064811` | MASQUE 隧道承载 |
| `github.com/quic-go/quic-go` | `v0.60.0` | MASQUE H2/HTTP3 |
| `github.com/lxn/walk` + `lxn/win` | 2021 快照 | 原生 Win32 GUI |
| `golang.zx2c4.com/wintun` | `v0.0.0-20230126` | TUN 虚拟网卡 |
| `golang.zx2c4.com/wireguard` | `v0.0.0-20260522` | WG 协议 |
| `gvisor.dev/gvisor` | `v0.0.0-20260616` | 用户态网络栈 |
| `github.com/charmbracelet/bubbletea` | `v1.3.10` | TUI |
| `github.com/spf13/cobra` | `v1.10.2` | CLI |

## 4. 网络端点（观测，源码交叉引用）

以下为源码中硬编码的全部出站端点，按用途分组。源码路径相对仓库根。

### 4.1 WARP 账号注册（Cloudflare 官方）

| 端点 | 源码位置 | 说明 |
|---|---|---|
| `https://api.cloudflareclient.com/v0a4005/reg` | `register.go:27` | WireGuard/AmneziaWG 账号注册 |
| `https://api.cloudflareclient.com/` | `register.go:28` | API 可达性探测 |
| `https://api.cloudflareclient.com/v0a4471/reg` | `masque.go:163` | MASQUE 账号注册 |

### 4.2 端点扫描与验证

| 端点 | 源码位置 | 说明 |
|---|---|---|
| `http://www.gstatic.com/generate_204` | `client.go`, `connect.go`, `gui_native_windows.go` | 隧道连通性探针 |
| `https://api.ipify.org` | `client.go:826` | 出口 IP 回显 |
| `https://speed.cloudflare.com/meta` | `client.go:820` | 出口地区/城市 |
| `https://speed.cloudflare.com/__down` | `client.go` | 扫描测速 |
| `https://www.cloudflare.com/cdn-cgi/trace` | `client.go:827` | Colo 信息 |
| `https://dns.google/resolve` | `client.go:963` | DNS 泄露检测 |

### 4.3 VPN Gate 中继（第三方）

| 端点 | 源码位置 | 说明 |
|---|---|---|
| `https://www.vpngate.net/api/iphone/` | `vpngate.go:30` | 拉取节点列表（CSV） |
| `http://ip-api.com/batch` | `vpngate.go:59` | 批量地区/ISP 标注 |

### 4.4 Python 与自建中继

| 端点 | 源码位置 | 说明 |
|---|---|---|
| `https://edge-client-api.vercel.app` | `register.go:33` | 默认 relay（WARP 注册备用通道） |
| `http://ip-api.com/json/` | `gui_native_windows.go:2094` | GUI 地区显示 |

### 4.5 更新与资源下载

| 端点 | 源码位置 | 说明 |
|---|---|---|
| `https://api.github.com/repos/hezhanleiok/xehn/releases/latest` | `version.go:36`, `update.go:22` | 客户端版本检查 |
| `https://api.github.com/repos/MetaCubeX/mihomo/releases` | `download.go` | mihomo 核心下载 |
| `https://github.com/` | `geodata.go` | geosite/geoip 下载 |

### 4.6 本机回环

`http://127.0.0.1` 在 `client.go`、`connect.go`、`core.go`、`vpngate.go`、`mihomo_integration_test.go` 中用于：
- mixed port（默认 7890）— 代理入站
- external-controller（默认 9090）— mihomo REST API

### 4.7 已移除

**观测**：`https://raw.githubusercontent.com/hezhanleiok/xehn/main/control.json` 曾是每 15 分钟的远程停用名单拉取地址（源码 `control.go:30`）。本次改动已**彻底删除**该通道及其全部代码；重建后的二进制中 `control.json`、`control-state.json`、`xehn-control-v1`、`stop_all`、`stopped_codes` 均不存在。

## 5. 嵌入资源（观测）

| 资源 | 大小 | 说明 |
|---|---|---|
| `rsrc_windows_amd64.syso` | 79,714 字节 | Win32 manifest（comctl32 v6 + PerMonitorV2 DPI）+ 程序图标 |
| `logo.ico` | 78,282 字节 | 多尺寸图标，由 `rsrc` 烘焙进 exe |
| `assets/logo.png` | 261,464 字节 | 界面内品牌图 |
| `core/geosite.dat` | 4,353,539 字节 | 分流规则（外部文件，非嵌入） |
| `core/geoip.metadb` | — | GeoIP 数据库（外部文件） |

**观测**：`app.manifest` 要求 `Microsoft.Windows.Common-Controls` v6 —— walk 的原生控件依赖它，缺失会导致 `NewMainWindow` 失败。

## 6. 协议常量（观测）

二进制与源码中包含以下协议参数名（AmneziaWG 混淆参数、MASQUE 配置）：

- AmneziaWG：`Jc`、`Jmin`、`Jmax`、`S1`、`S2`、`H1`-`H4`、`I1`
- MASQUE：`congestion-controller: bbr`、`ip-stack: gvisor`、`network: h2`
- WARP-in-WARP：`dialer-proxy` 链式配置

## 7. 控制码通道（已移除，历史记录）

移除前的行为边界（观测自源码，非运行时观测）：

| 项 | 行为 |
|---|---|
| 控制码生成 | 由 `HKLM\SOFTWARE\Microsoft\Cryptography\MachineGuid` 派生（`control_code_windows.go`） |
| 持久化 | `settings.json` 的 `control_code` 字段 |
| 名单拉取 | 启动时 + 每 15 分钟，`control.json` |
| 失败模式 | fail-open：拉取失败沿用上次结论 |
| 命中效果 | 拒绝连接、拆掉在线会话、弹窗 |

**推断**：这是一个可远程致客户端不可用的通道。名单本身当前为空（`stop_all: false`，`stopped_codes: []`），但机制本身构成单点控制风险。

## 8. 未知与边界

- **未做运行时动态观测**：本报告全部为静态证据（源码 + 二进制字符串 + buildinfo）。未抓包、未 hook、未观察真实 TLS 流量。
- **未验证** `edge-client-api.vercel.app` 是否为作者自建、其后端行为如何。
- **未验证** `ip-api.com` 为明文 HTTP（`http://`），意味着节点地区查询不经加密；这是既有设计，不在本次改动范围。
- **不能**用 SHA-256 相等性判断复现成功：Go 构建含 buildID 与时间戳，不同环境必然哈希不同。

## 9. 结论

出厂二进制是公开源码的忠实构建，**无隐藏逻辑、无额外依赖、无远程控制残余**（本次改动后）。唯一曾被识别的远程控制面（控制码停用名单）已完全移除。
