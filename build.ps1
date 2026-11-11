# 构建中文 + 代理版 Meiro（Windows）
#
#   powershell -ExecutionPolicy Bypass -File .\build.ps1
#
# 必须走 mygo build：它会注入 packageIdentifier 等 ldflags，
# 少这些的话 Windows 凭据管理器里存的 cookie 解不开。
# 普通 go build 编出来的 exe 会报 "not in the system credential store"。

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

function Find-Go {
    $cmd = Get-Command go -ErrorAction SilentlyContinue
    if ($cmd) { return $cmd.Source }
    foreach ($p in @('D:\DSH\.work\go\bin\go.exe', 'C:\Go\bin\go.exe', "$env:LOCALAPPDATA\Programs\Go\bin\go.exe")) {
        if (Test-Path $p) { return $p }
    }
    throw '找不到 go.exe。请安装 Go 1.27+ 并加进 PATH，或把路径加到本脚本的 Find-Go 里。'
}

$go = Find-Go
$goRoot = Split-Path -Parent (Split-Path -Parent $go)

# 构建缓存和临时目录都放在仓库内，避免 %TEMP% 权限问题。
$env:GOROOT    = $goRoot
$env:GOPATH    = Join-Path $root '.gopath'
$env:GOCACHE   = Join-Path $root '.gocache'
$env:GOTMPDIR  = Join-Path $root '.gotmp'
$env:TMPDIR    = $env:GOTMPDIR
$env:TEMP      = $env:GOTMPDIR
$env:TMP       = $env:GOTMPDIR
$env:GOPROXY   = 'https://goproxy.cn,direct'
$env:GOSUMDB   = 'off'
$env:PATH      = "$goRoot\bin;$env:PATH"

New-Item -ItemType Directory -Force -Path $env:GOPATH, $env:GOCACHE, $env:GOTMPDIR | Out-Null

$out = Join-Path $root 'dist'
if (Test-Path $out) { Remove-Item $out -Recurse -Force }

Write-Host "go      : $go"
Write-Host "out     : $out"
Write-Host 'building...'

& $go run github.com/egoist/mygo/cmd/mygo build -platform windows/amd64 -skip-installer -o $out
if ($LASTEXITCODE -ne 0) { throw "构建失败，退出码 $LASTEXITCODE" }

$exe = Join-Path $out 'windows-amd64\Meiro.exe'
if (-not (Test-Path $exe)) { throw "没有找到产物：$exe" }

$size = (Get-Item $exe).Length
Write-Host ''
Write-Host "ok: $exe ($size bytes)"
Write-Host ''
Write-Host '安装（先备份）：'
Write-Host '  copy "D:\Meiro\Meiro.exe" "D:\Meiro\Meiro.exe.bak"'
Write-Host "  copy `"$exe`" `"D:\Meiro\Meiro.exe`""
