# XEHN 构建溯源与二进制等价性验证

本文记录「出厂二进制能否由公开源码复现」的实测结论。所有结论都附可复现命令与观测量；观测、推断、未知三类明确区分。

## 1. 被测对象

| 项 | 值 |
|---|---|
| 文件 | `H:\Program Files\XEHN\xehn.exe` |
| 大小 | 20,978,176 字节 |
| SHA-256 | `A016C0FC32307A0BBA6272FE9B0C3379069940A8806A19E9B8C469A519361D87` |
| PE 架构 | Machine `0x8664`（x86-64），Subsystem `2`（windowsgui） |
| 构建工具链 | `go1.27.0`（观测：`go version -m`） |

## 2. 构建参数（观测）

`pack.ps1:83` 定义了唯一的发布构建命令：

```powershell
go build -trimpath -ldflags "-s -w -H windowsgui -X main.version=$Version -X main.windowsGUIBuild=1 -X main.buildStamp=$stamp" -o $bin .
```

关键点：

- `-trimpath` 去除本机路径，因此两个独立构建可以逐段对比；
- `-s -w` 剥离符号表与 DWARF，这就是出厂二进制 **20.9 MB**、而朴素 `go build` 产出 **30.0 MB** 的原因（观测：朴素构建 30,034,432 字节）；
- `-X main.buildStamp=...` 注入打包时间，出厂值为 `20261030-2143`（观测：二进制内字符串偏移 9242216）。

## 3. 工厂构建信息（观测）

`go version -m "H:\Program Files\XEHN\xehn.exe"` 输出：

```
path  github.com/xhwarp/warp-client
mod   github.com/xhwarp/warp-client  v0.0.0-20261030130144-463792ab6313+dirty
build -buildmode=exe
build -compiler=gc
build -trimpath=true
build CGO_ENABLED=0
build GOARCH=amd64
build GOOS=windows
build GOAMD64=v1
build vcs=git
build vcs.revision=463792ab6313e322055879fca2bfc095bfd31aad
build vcs.time=2026-10-30T13:01:44Z
build vcs.modified=true
```

**观测结论**：出厂二进制由提交 `463792a` 构建，且打包时工作区存在**未提交改动**（`+dirty` / `vcs.modified=true`）。

## 4. 复现实验

用 `G:\work\toolchain\go\bin\go.exe`（go1.27.3）在干净工作树中按 `pack.ps1:83` 的参数构建：

| 构建源 | 提交 | 产物大小 | 与工厂差值 |
|---|---|---|---|
| 工厂二进制 | `463792a` + dirty | 20,978,176 | — |
| 源码构建 A | `463792a`（干净） | 20,996,608 | +18,432 |
| 源码构建 B | `c319e7c`（`463792a` 的子提交） | 20,997,120 | +18,944 |

工具链差异（go1.27.0 vs go1.27.3）足以解释约 18–19 KB 的体积差（观测：同源不同工具链的产物体积不同）。

## 5. 等价性判定（核心证据）

剥离符号后无法用 `go tool nm` / `objdump`（观测：均报 `no symbols` / `no runtime.pclntab symbol found`），因此改用**应用特征字符串集合**比对：

- 提取两个二进制中全部 ≥6 字符的可打印 ASCII 串；
- 用应用专属标记筛选（模块路径 `xhwarp`/`warp-client`、`xehn`、`mihomo`、`psiphon`、`masque`、`amnezia`、`hezhanleiok`，以及所有 CJK 文案）；
- 对工厂二进制与每个源码构建做集合差。

结果：

| 对照文案 | 应用字符串数 | 工厂独有 | 源码独有 |
|---|---|---|---|
| `463792a` vs 工厂 | 895 vs 895 | **1** | **1** |
| `c319e7c` vs 工厂 | 895 vs 895 | **0** | **0** |

`463792a` 的那 1 条差异是一段相邻字符串的合并/拆分边界不同（同一条 UTF-8 文案被 Go 串接器切成不同候选），不是新增或删除的逻辑。`c319e7c` 对照为**完全零差异**。

### 结论

> **出厂二进制 = 提交 `c319e7c` 的内容，用 `pack.ps1:83` 的参数在本机构建。**

`+dirty` 的来源：打包时工作区停在 `463792a`，但 `gui_i18n_windows.go` 与 `gui_page_settings_windows.go` 里放着随后作为 `c319e7c` 提交的两条 DNS 泄露提示文案（观测：`git show c319e7c` 只改这两个文件、+9/−1 行；且 `c319e7c` 的提交时间 2026-10-30 21:47:35 晚于二进制打包时间 21:43）。打完包后再提交，于是二进制带上 `+dirty` 而源码树最终停在 `c319e7c`。

这也解释了 `vcs.time=2026-10-30T13:01:44Z`（= 21:01:44 +08:00，`463792a` 的提交时间）与 `buildStamp=20261030-2143`（21:43 打包）的时间顺序：源码 HEAD 是 `463792a`，工作区改了文件，21:43 打包，21:47 提交成 `c319e7c`。

## 6. 未知与边界

- **未验证**：go1.27.0 与 go1.27.3 各自的 stdlib 改动是否影响运行时行为。仅观测到体积差异，未做逐函数反汇编对比（两个二进制均被 `-s -w` 完全剥离）。
- **未验证**：打包机是否为同一台。体积差与工具链版本一致，但无法从二进制反推构建主机。
- **不适用**：源码构建产物的 SHA-256 必然与出厂值不同（go build 默认不保证可重复构建，含 buildID 与时间戳）。因此**不能**用哈希相等性作为复现判据，本文用字符串集合与体积作为判据是合适的替代。

## 7. 复现命令

```powershell
$env:GOCACHE='G:\work\cache\gocache'
$env:GOMODCACHE='G:\work\cache\gomod'
$env:GOPATH='G:\work\cache\gopath'
$env:GOPROXY='https://goproxy.cn,direct'
$env:GOOS='windows'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'
$go='G:\work\toolchain\go\bin\go.exe'

git worktree add --detach G:\work\wt-c319e7c c319e7c
cd G:\work\wt-c319e7c
& $go build -trimpath -ldflags "-s -w -H windowsgui -X main.version=0.3.0-202610302143 -X main.windowsGUIBuild=1 -X main.buildStamp=20261030-2143" -o G:\work\build\xehn-c319e7c-flags.exe .
& $go version -m G:\work\build\xehn-c319e7c-flags.exe
```

## 8. 基线健康度（观测）

在干净 `c319e7c` 工作树上：

- `go build ./...` → **exit 0**
- `go test ./...` → **exit 0**（`warp-client` 0.9s、`internal/cache`、`internal/core` 通过；`tools/mklogo` 无测试）
- `go vet ./...` → **exit 1**，4 条既有告警：
  - `gui_chrome_windows.go:292` `possible misuse of unsafe.Pointer`
  - `gui_chrome_windows.go:303` `possible misuse of unsafe.Pointer`
  - `connect.go:1196` `cancelWarm is not used on all paths (possible context leak)`
  - `connect.go:1205` `may be reached without using the cancelWarm var`

这 4 条在改动前就存在，与本次工作无关。
