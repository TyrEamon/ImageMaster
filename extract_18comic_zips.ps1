[CmdletBinding(SupportsShouldProcess = $true)]
param(
    [string]$RootPath = '',
    [string]$BandizipPath = 'D:\bandizip\bz.exe',
    [switch]$Force,
    [switch]$DeleteArchive,
    [int]$MaxFiles = 0
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Write-Step {
    param([string]$Message)
    Write-Host "[18comic-unzip] $Message"
}

function Test-HasExtractedContent {
    param(
        [string]$TargetDir,
        [string]$ArchivePath
    )

    if (-not (Test-Path -LiteralPath $TargetDir)) {
        return $false
    }

    $children = Get-ChildItem -LiteralPath $TargetDir -Force -ErrorAction SilentlyContinue
    foreach ($child in $children) {
        if ($child.PSIsContainer) {
            return $true
        }

        if ($child.FullName -ne $ArchivePath -and $child.Extension -notin @('.zip', '.cbz', '.7z', '.rar')) {
            return $true
        }
    }

    return $false
}

function Resolve-ExtractDirectory {
    param(
        [System.IO.FileInfo]$Archive,
        [string]$RootFullPath
    )

    $parentPath = $Archive.Directory.FullName.TrimEnd('\')
    $rootPath = $RootFullPath.TrimEnd('\')

    if ($parentPath -ieq $rootPath) {
        return Join-Path -Path $RootFullPath -ChildPath $Archive.BaseName
    }

    return $Archive.Directory.FullName
}

function Get-DefaultRootPath {
    $configPath = Join-Path -Path $env:APPDATA -ChildPath 'imagemaster'
    if (-not (Test-Path -LiteralPath $configPath)) {
        return ''
    }

    try {
        $config = Get-Content -LiteralPath $configPath -Encoding UTF8 -Raw | ConvertFrom-Json
        $candidates = @()

        if ($config.active_library) {
            $candidates += [string]$config.active_library
        }

        if ($config.output_dir) {
            $candidates += [string]$config.output_dir
        }

        if ($config.libraries) {
            $candidates += @($config.libraries | ForEach-Object { [string]$_ })
        }

        foreach ($candidate in $candidates | Select-Object -Unique) {
            if ($candidate -and (Test-Path -LiteralPath $candidate)) {
                return (Resolve-Path -LiteralPath $candidate).Path
            }
        }
    } catch {
        Write-Step "Failed to read imagemaster config: $($_.Exception.Message)"
    }

    return ''
}

if (-not (Test-Path -LiteralPath $BandizipPath)) {
    throw "Bandizip console tool not found: $BandizipPath"
}

$resolvedRootPath = if ($RootPath) { $RootPath } else { Get-DefaultRootPath }
if (-not $resolvedRootPath) {
    throw 'Root path not found. Pass -RootPath explicitly or make sure imagemaster config has a valid library path.'
}

if (-not (Test-Path -LiteralPath $resolvedRootPath)) {
    throw "Root path not found: $resolvedRootPath"
}

$rootFullPath = (Resolve-Path -LiteralPath $resolvedRootPath).Path
$patterns = @('*.zip', '*.cbz', '*.7z', '*.rar')

$archives = foreach ($pattern in $patterns) {
    Get-ChildItem -LiteralPath $rootFullPath -Filter $pattern -File -Recurse -ErrorAction SilentlyContinue
}

$archives = $archives |
    Sort-Object FullName -Unique

if ($MaxFiles -gt 0) {
    $archives = $archives | Select-Object -First $MaxFiles
}

if (-not $archives) {
    Write-Step "No archives found under $rootFullPath"
    exit 0
}

$summary = [ordered]@{
    Found     = @($archives).Count
    Extracted = 0
    Skipped   = 0
    Failed    = 0
    Deleted   = 0
}

Write-Step "Root: $rootFullPath"
Write-Step "Bandizip: $BandizipPath"
Write-Step "Archives found: $($summary.Found)"

foreach ($archive in $archives) {
    $targetDir = Resolve-ExtractDirectory -Archive $archive -RootFullPath $rootFullPath
    $archiveLabel = $archive.FullName
    $targetLabel = $targetDir

    try {
        $alreadyExtracted = Test-HasExtractedContent -TargetDir $targetDir -ArchivePath $archive.FullName
        if ($alreadyExtracted -and -not $Force) {
            Write-Step "Skip (content exists): $archiveLabel"
            $summary.Skipped++
            continue
        }

        if (-not (Test-Path -LiteralPath $targetDir)) {
            if ($PSCmdlet.ShouldProcess($targetDir, 'Create extract folder')) {
                New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
            }
        }

        if ($PSCmdlet.ShouldProcess($archiveLabel, "Extract to $targetLabel")) {
            Write-Step "Extract: $archiveLabel -> $targetLabel"
            & $BandizipPath x -y $archive.FullName "-o:$targetDir"
            $exitCode = $LASTEXITCODE
            if ($exitCode -ne 0) {
                throw "Bandizip exited with code $exitCode"
            }
        } else {
            Write-Step "Plan extract: $archiveLabel -> $targetLabel"
            continue
        }

        $summary.Extracted++

        if ($DeleteArchive) {
            if ($PSCmdlet.ShouldProcess($archiveLabel, 'Delete archive after extract')) {
                Remove-Item -LiteralPath $archive.FullName -Force
                Write-Step "Deleted archive: $archiveLabel"
                $summary.Deleted++
            }
        }
    } catch {
        $summary.Failed++
        Write-Step "Failed: $archiveLabel | $($_.Exception.Message)"
    }
}

Write-Host ''
Write-Host 'Summary'
Write-Host "  Found:     $($summary.Found)"
Write-Host "  Extracted: $($summary.Extracted)"
Write-Host "  Skipped:   $($summary.Skipped)"
Write-Host "  Failed:    $($summary.Failed)"
Write-Host "  Deleted:   $($summary.Deleted)"
