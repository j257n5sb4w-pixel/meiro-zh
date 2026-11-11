# Meiro 中文 + 代理补丁（Windows）

基于 [elianiva/meiro](https://github.com/elianiva/meiro) **0.6.2**，遵循上游 **GPL-3.0**。
这个仓库是完整源码（已打补丁），不是只有 diff——克隆下来直接就能编。

---

## 一、直接用（覆盖 exe 即可）

1. 先装好官方 Meiro 0.6.2（本机是 `D:\Meiro`）。
2. 到本仓库 **Releases** 下载 `Meiro-0.6.2-zh.1-win64.zip` 并解压出 `Meiro.exe`。
3. 备份后覆盖：把下载的文件改名/覆盖到 `D:\Meiro\Meiro.exe`。
4. 打开 Meiro → **设置 → 系统 → 网络**，填代理地址（例如 `http://127.0.0.1:7897`）→ **保存** → **测试**。
5. 登录：**设置 → 账户 → 登录**，弹出的对话框里同样有代理框。

> 覆盖前建议先 `copy Meiro.exe Meiro.exe.bak`。原始 0.6.2 的 exe 见上游 Release。

---

## 二、改了什么

### 1. 代理（这是原来完全没有的东西）

| 位置 | 说明 |
|---|---|
| 登录对话框 | 在「高级」上方，粘 Cookie 之前就能先设代理 |
| 设置 → 系统 → 网络 | 正式入口，带**保存**和**测试**按钮 |
| 顶栏胶囊 | 设了代理才出现，点一下直接跳到上面那张卡片 |

生效范围（这点很重要，很多播放器的"代理"只对 API 生效，播放时照样被墙）：

- YouTube Music API（`youtube/` 的 HTTP client）
- 封面、歌词（LRCLIB / BetterLyrics）、Last.fm
- 音频缓存下载（`internal/audio`，走 `http.DefaultTransport`）
- **ffmpeg**（播放时的取流，通过 `HTTP_PROXY`/`HTTPS_PROXY` 环境变量）
- **yt-dlp**（`--proxy` 之外还加了环境变量，两条路都通）
- **登录用的浏览器窗口**：Chromium 系加 `--proxy-server=`，Firefox 写 `user_pref("network.proxy.*")`

支持的写法：`http://host:port`、`https://`、`socks5://`、`socks5h://`、`socks4://`；不写 scheme 默认 http。
`localhost` / `127.0.0.1` / `::1` 永远直连，不会被代理套娃。

「测试」按钮是真的向 `https://music.youtube.com/` 发一次 HEAD 请求，不是假装成功。

### 2. 中文界面

- 约 **434 条**整句词条，外加一套运行时拼接规则（`已添加“X”到媒体库`、`无法喜欢“X”：…`、`正在播放 X`、`已从队列中移除 N 首歌曲` 这类模板）。
- 翻译发生在**绘制那一刻**，所以同一个会话切语言不用重启。
- 设置 → 外观与声音 → **语言**：`自动 / English / 中文`。
- YouTube 内容也跟着切：请求会带 `HL=zh-CN`，所以分区标题是「为你推荐的音乐视频」而不是「Music videos for you」。
- 没翻译到的字符串回退成英文，不会变成空白。

---

## 三、自己构建

```powershell
powershell -ExecutionPolicy Bypass -File .\build.ps1
```

产物：`dist\windows-amd64\Meiro.exe`

### ⚠️ 必须用 `mygo build`，不能用 `go build`

`mygo build` 会注入 `-X github.com/egoist/mygo.packageIdentifier=com.elianiva.meiro` 等 ldflags。
Windows 上保存的 cookie 是 **用这个 identifier 当服务名**存进「凭据管理器」的，所以：

- 用 `go build` 编出来的 exe 能跑，但读不出已保存的登录态，报
  `the key of ...\cookie.txt is not in the system credential store`；
- 用 `mygo build` 编出来的就能正常解密、自动恢复会话。

`build.ps1` 已经替你调好了 mygo。

---

## 四、升级到上游新版本

```bash
git clone https://github.com/elianiva/meiro.git meiro-new
cd meiro-new
git apply --3way /path/to/patches/meiro-0.6.2-zh-proxy.patch
```

冲突一般出在这几个文件，手动合一下就行：

- `signin.go`（代理框插在「高级」上方）
- `settingspage.go`（网络卡片、语言卡片的位置）
- `shell.go`（顶栏胶囊）
- `m3/theme.go`、`m3/button.go`、`m3/misc.go`、`m3/nav.go`、`m3/field.go`（翻译注入点）

然后跑 `build.ps1`。

---

## 五、改动清单

**新增文件（6 个）**

```
internal/proxy/proxy.go     代理地址、Transport、子进程环境变量
internal/lang/lang.go       语言选择、T()、HL()
internal/lang/zh.go         词条表（434 条）
internal/lang/parts.go      运行时拼接规则 + 正则模板
internal/lang/lang_test.go  翻译测试
network.go                  代理设置卡片、代理字段、顶栏胶囊
```

**修改文件**：29 个（完整 diff 见 `patches/`）。核心几处：

- `youtube/client.go`：`transport.Proxy = proxy.Proxy` + 内容语言跟随
- `internal/proc/proc.go`、`player/player.go`：子进程注入代理环境变量
- `internal/cookies/window.go`：登录窗口带代理
- `internal/settings/settings.go`：新增 `proxy`、`language` 字段
- `m3/*`：绘制层统一走 `lang.T()`

---

## 六、已知限制

- **无障碍标签**（屏幕阅读器读的那层）仍是英文。上游的 UI 测试就是靠英文文本定位控件的（`Tester.Click("Sign in with a browser window")`），改了会挂一片测试，所以刻意保留。
- 歌曲名、专辑名、歌手名属于内容，不翻译。
- 播放栏里的 `158K views` 这类文案来自 YouTube 返回的数据，会随内容语言变化。
- 词条表是按 0.6.2 的界面抓的；上游加新界面文案时，新文案会显示英文，往 `internal/lang/zh.go` 里补一行即可。
