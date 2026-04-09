param(
    [string]$PythonExe = "py",
    [string]$PythonVersionArg = "-3.11"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$runtimeDir = Join-Path $repoRoot "runtime"
$bridgeScript = Join-Path $runtimeDir "jm_bridge.py"
$buildRoot = Join-Path $runtimeDir ".build-jm-runtime"
$venvDir = Join-Path $buildRoot "venv"
$distDir = Join-Path $buildRoot "dist"
$finalExe = Join-Path $runtimeDir "imagemaster-jm-runtime.exe"

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
Write-Host "Built JM runtime:" $finalExe
