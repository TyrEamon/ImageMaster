param(
    [string]$PythonExe = "py",
    [string]$PythonVersionArg = "-3.11",
    [string]$RuntimeVersion = "0.1.0-dev",
    [string]$BuildTime = ""
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$runtimeDir = Join-Path $repoRoot "runtime"
$bridgeScript = Join-Path $runtimeDir "jm_bridge.py"
$buildRoot = Join-Path $runtimeDir ".build-jm-runtime"
$venvDir = Join-Path $buildRoot "venv"
$distDir = Join-Path $buildRoot "dist"
$finalExe = Join-Path $runtimeDir "imagemaster-jm-runtime.exe"
$manifestPath = Join-Path $runtimeDir "runtime-manifest.json"

if (-not (Test-Path $bridgeScript)) {
    throw "Bridge script not found: $bridgeScript"
}

if (Test-Path $buildRoot) {
    Remove-Item -Recurse -Force $buildRoot
}

New-Item -ItemType Directory -Force -Path $buildRoot | Out-Null

& $PythonExe $PythonVersionArg -m venv $venvDir

$venvPython = Join-Path $venvDir "Scripts\python.exe"
& $venvPython -m pip install --upgrade pip
& $venvPython -m pip install pyinstaller jmcomic
& $venvPython -m PyInstaller `
    --onefile `
    --clean `
    --name "imagemaster-jm-runtime" `
    --distpath $distDir `
    $bridgeScript

Copy-Item -Force (Join-Path $distDir "imagemaster-jm-runtime.exe") $finalExe

if (-not $BuildTime) {
    $BuildTime = Get-Date -Format "yyyy-MM-ddTHH:mm:ssK"
}

$manifest = @{
    name = "JM Runtime"
    version = $RuntimeVersion
    engine = "jmcomic"
    upstream = "hect0x7/JMComic-Crawler-Python"
    buildTime = $BuildTime
}

$manifest | ConvertTo-Json | Set-Content -Encoding UTF8 $manifestPath

Write-Host "Built JM runtime:" $finalExe
