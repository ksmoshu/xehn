# 本次迭代改动说明

## 1. 移除控制码远程停用机制（已完成）

**问题**：每台设备携带一个由 Windows MachineGuid 派生的控制码，客户端每 15 分钟从
`raw.githubusercontent.com/hezhanleiok/xehn/main/control.json` 拉取停用名单；命中即拒绝连接、
拆掉在线会话、弹窗说明。这是一条可远程致客户端不可用的通道。

**证据**：`git show` 前 `control.go`（269 行）、`control_code_windows.go`、
`control_code_other.go`、`control_test.go`、`control.json` 均存在；出厂二进制内可见
`control.json` / `control-state.json` / `xehn-control-v1` / `stop_all` 等字符串
（偏移见 `docs/rea-detection.md`）。

**改法**：删除上述 5 个文件；移除 `settings.go` 的 `ControlCode` 字段与规范化；
移除 `gui_native_windows.go` 的模型字段、15 分钟控制循环、启动检查与 `doConnect` 闸门；
移除「复制控制码 / 立即检查」动作与「客户端已停用」对话框；`About` 页的远程控制卡片并入
整行的「客户端更新」卡；清理 17 条 i18n 文案；同步更新 `README.md`、`README-portable.txt`、
`docs/blog/xehn-usage.html`。

**验证**：`go build ./...` exit 0；`go test ./...` exit 0；重新构建的二进制中
`control.json`、`control-state.json`、`控制码`、`远程控制`、`stop_all`、`stopped_codes`、
`xehn-control-v1` 字符串**全部不存在**。净删 715 行，未引入新依赖。

**风险**：`settings.json` 里已有的 `control_code` 字段会被忽略（JSON 解析不校验未知字段），
不做数据迁移；用户无感知。

**回滚**：`git revert` 该提交即可。

---

## 2. 健康守护：修复重连风暴（已完成）

**问题**：出厂日志（`H:\Program Files\XEHN\xehn.log`）显示一个稳定复发的循环——
每约 15 分钟「数据面探测失败 3/3 → OpenVPN 出口失效 → 连续切换出口 → 大面积握手失败 →
没有可用的 VPN Gate 节点 → 结束会话 → 断开」。实测统计：

| 指标 | 数量 |
|---|---|
| 修复事件（探测失败 3/3） | 28 |
| 出口切换尝试 | 55 |
| 修复失败 | 29 |
| 不同失败节点 | 24 |

24 个不同节点全部失败，说明故障不在单个节点，而是**整体出口链路不稳**；但代码对每次
探测失败都执行完整的 4 轮节点轮换，失败后**直接 `giveUp` 断开**，把一次几十秒的网络抖动
放大成"用户必须手动重新连接"。

**证据**（改动前代码位置）：
- `health.go:775` 主循环，`fails` 计数；`judgeHealth` 连续 3 次失败即判定端点死亡。
- `health.go:889` OpenVPN 分支：`repairOpenVPNOnce` 失败后直接
  `logf("没有可用的 VPN Gate 节点，结束会话"); giveUp(logf); return nil`。
- `health.go:1037` 默认（WARP）分支同样以 `giveUp(logf); return nil` 收尾。
- `health.go:626` `const maxAttempts = 4` 硬编码，与检查周期无关。

**改法**：

1. **指数退避**（新增 `repairBackoff` 纯函数）：连续修复失败后的等待为
   2m → 4m → 8m → 16m → 30m（封顶）。成功一次即清零。
2. **保留会话而非断开**：OpenVPN 与默认分支在候选池耗尽后，不再 `giveUp`，
   而是保留当前会话、记录退避截止时间，下一轮尝试重新扫描。VPN Gate 中继本就
   自行轮换，几分钟后刷新列表常能连上。
3. **健康满窗重置**（`repairBackoffReset = 10m`）：数据面连续正常 10 分钟才清除
   `repairFails`，避免"通一下又挂"被误判为恢复。
4. **计数器与截止时间同步**：清零 `repairFails` 时一并清空 `nextRepairAt`，
   否则残留的未来截止时间会继续压制本应允许的修复。

**涉及文件**：`health.go`（新增 `repairBackoffBase`/`repairBackoffMax`/`repairBackoffReset`
常量、`repairBackoff()`、`insuranceWait()`，改写主循环与两个放弃分支）；
新增 `repair_backoff_test.go`。

**测试**：4 个用例覆盖退避表（0→0、1→2m…5+→30m 封顶）、单调性、计数器推进与截止时间前移、
首次调用必为正等待。`go test -run 'RepairBackoff|InsuranceWait' -v` 全 PASS。

**验证**：`go build ./...` exit 0；`go test ./...` exit 0；`go vet .` 仅保留 4 条改动前既有告警
（`connect.go:1196`、`connect.go:1205`、`gui_chrome_windows.go:292`、`:303`）。

**风险**：`giveUp` 不再调用，意味着长期彻底断网时客户端会保持一个"已连接但无网"的会话
最长 30 分钟才重试一次。这是刻意的取舍；用户仍可手动断开重连，且日志会明确写出
「本轮修复失败，保留当前会话，Xm 后重试」。

**回滚**：`git revert` 该提交。

---

## 3. 配置校验加固（已完成）

**问题**：GUI 写入的 `settings.json` 有三处未校验就进入渲染器：

1. `APIPort == MixedPort` 时用 `APIPort = MixedPort + 1` 消解冲突，而 `MixedPort = 65535`
   会把它推成 **65536**，一个非法端口。
2. `dns_custom` / `domain_direct` / `domain_proxy` **完全无校验**：手改 settings.json 写入
   非 IP 的 DNS 项会让 mihomo 拒绝整份配置；域名项若带 `||`、`/`、`:`、引号等字符，
   会生成永不匹配的 `DOMAIN-SUFFIX` 规则或破坏 YAML。

**证据**：改动前 `settings.go:318-320`（端口 +1 溢出）、`settings.go:437`（`o.dns` 直取原文）、
`settings.go:444-445`（域名列表直达渲染器），`normalize()`（`settings.go:237-353`）中无相关校验。

**改法**：

- `normalize()` 增加三行守卫，接入两个新函数：
  - `validDNSServers()` —— `net.ParseIP` 过滤，非法项丢弃，输出统一为 `逗号+空格` 形式；
  - `validDomainList()` / `validDomainToken()` —— 只接受 `字母数字 - _ .` 与 `*.` 通配前缀，
    拒绝空标签、首尾点、连字符开头/结尾的标签、以及 YAML 指示符。
- 端口冲突改为就近避让：能 +1 就 +1，`65535` 时改为 −1。

**测试**：新增 `settings_validate_test.go`，16 个非法域名用例、10 个 DNS 用例、6 个域名列表用例、
5 组端口用例。**测试立即抓到一个真实缺陷**：`-example.com` 曾被放行（DNS 标签不得以连字符
开头），已修复。全 PASS。

**验证**：`go build ./...` exit 0；`go test ./...` exit 0；`go vet` 仅 4 条既有告警。

**风险**：非法值现在会被**静默丢弃**而非报错。这是刻意的：一个手改错的文件不应让程序无法启动。
用户若发现某项设置"没生效"，日志与设置页仍显示规范化后的值。

---

## 4. 系统代理还原：审计结论（未改动）

**审计结果**：既有实现已足够健壮，本次未改动。理由：

- `SetSystemProxy` 会捕获完整快照（含 PAC URL 与 auto-detect），并识别"快照其实是我们自己
  上次留下的"并替换为干净基线（`proxy_windows.go:126-130`）。
- 还原走三级降级（直写注册表 → WinINet API → `reg.exe`），每次写入都**读回验证**
  （`applyProxyValues`），避免安全软件静默回滚被当成成功。
- 快照丢失时有 `SystemProxyEnabledAt()` 兜底清理（`connect.go:1436-1445`）。
- 仍失败时打印明确告警，提示手动关闭或提权运行。
- 启动时 `healStaleProxy()` 处理"上次被强杀"的残留（`gui_native_windows.go:531-550`），
  且先用 TCP 探测确认端口无人监听，避免误清正在工作的会话。

---

## 5. 剩余未做项

- 4 条既有 `go vet` 告警（`connect.go:1196/1205` 的 context 泄漏、`gui_chrome_windows.go:292/303`
  的 `unsafe.Pointer`）未处理。属既有问题，且 `unsafe.Pointer` 涉及 Win32 回调，改动风险高于收益。
- `GOOS=linux` 编译失败（`shortVersion`、`tr` 定义在 `_windows.go` 中）。已在原始 `c319e7c`
  树上复现同样失败，确认为既有问题；本项目为 Windows 专用。
