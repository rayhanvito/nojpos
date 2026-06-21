param(
    [int]$WebpQuality = 86,
    [switch]$CreateRasterBackedSvg
)

$ErrorActionPreference = 'Stop'

function Resolve-ToolPath {
    param([string]$Name)
    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if (-not $command) {
        throw "Required tool '$Name' was not found in PATH."
    }
    return $command.Source
}

function Get-PngSize {
    param([string]$Path)
    Add-Type -AssemblyName System.Drawing
    $image = [System.Drawing.Image]::FromFile($Path)
    try {
        return [pscustomobject]@{
            Width = $image.Width
            Height = $image.Height
        }
    }
    finally {
        $image.Dispose()
    }
}

function Convert-PngToWebp {
    param(
        [string]$Ffmpeg,
        [string]$Source,
        [string]$Destination,
        [int]$Quality
    )

    $destinationDirectory = Split-Path -Parent $Destination
    if ($destinationDirectory) {
        New-Item -ItemType Directory -Force -Path $destinationDirectory | Out-Null
    }

    & $Ffmpeg -hide_banner -loglevel error -y -i $Source -c:v libwebp -q:v $Quality -lossless 0 -compression_level 6 -preset picture $Destination

    if ($LASTEXITCODE -ne 0 -or -not (Test-Path $Destination)) {
        throw "Failed to convert '$Source' to WebP."
    }
}

function New-RasterBackedSvg {
    param(
        [string]$Source,
        [string]$Destination
    )

    $size = Get-PngSize -Path $Source
    $bytes = [System.IO.File]::ReadAllBytes($Source)
    $base64 = [Convert]::ToBase64String($bytes)
    $svg = @"
<svg xmlns="http://www.w3.org/2000/svg" width="$($size.Width)" height="$($size.Height)" viewBox="0 0 $($size.Width) $($size.Height)" role="img">
  <image width="$($size.Width)" height="$($size.Height)" href="data:image/png;base64,$base64"/>
</svg>
"@
    $destinationDirectory = Split-Path -Parent $Destination
    if ($destinationDirectory) {
        New-Item -ItemType Directory -Force -Path $destinationDirectory | Out-Null
    }
    Set-Content -Path $Destination -Value $svg -Encoding UTF8
}

$root = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$ffmpeg = Resolve-ToolPath -Name 'ffmpeg'

$assetRoots = @(
    [pscustomobject]@{
        Type = 'illustration'
        Directory = Join-Path $root 'apps/cashier/assets/illustrations/nojpos'
        Svg = $false
    },
    [pscustomobject]@{
        Type = 'logo'
        Directory = Join-Path $root 'apps/cashier/assets/logos/nojpos'
        Svg = $true
    }
)

$report = @()

foreach ($assetRoot in $assetRoots) {
    if (-not (Test-Path $assetRoot.Directory)) {
        continue
    }

    Get-ChildItem -Path $assetRoot.Directory -Filter '*.png' | Sort-Object Name | ForEach-Object {
        $png = $_.FullName
        $name = $_.BaseName
        $webp = Join-Path $assetRoot.Directory "$name.webp"
        Convert-PngToWebp -Ffmpeg $ffmpeg -Source $png -Destination $webp -Quality $WebpQuality

        $svgPath = ''
        $svgKb = $null
        $svgKind = ''
        if ($assetRoot.Svg -and $CreateRasterBackedSvg) {
            $svgPath = Join-Path $assetRoot.Directory "$name.svg"
            New-RasterBackedSvg -Source $png -Destination $svgPath
            $svgKb = [math]::Round((Get-Item $svgPath).Length / 1KB, 1)
            $svgKind = 'raster-backed-svg-not-true-vector'
        }

        $report += [pscustomobject]@{
            Type = $assetRoot.Type
            Name = $name
            PngKB = [math]::Round((Get-Item $png).Length / 1KB, 1)
            WebpKB = [math]::Round((Get-Item $webp).Length / 1KB, 1)
            WebpSavingsPercent = [math]::Round((1 - ((Get-Item $webp).Length / [double](Get-Item $png).Length)) * 100, 1)
            SvgKB = $svgKb
            SvgKind = $svgKind
            Png = $png.Replace($root + [IO.Path]::DirectorySeparatorChar, '')
            Webp = $webp.Replace($root + [IO.Path]::DirectorySeparatorChar, '')
            Svg = if ($svgPath) { $svgPath.Replace($root + [IO.Path]::DirectorySeparatorChar, '') } else { '' }
        }
    }
}

$reportPath = Join-Path $root 'apps/cashier/assets/nojpos_webp_svg_report.csv'
$report | Export-Csv -NoTypeInformation -Path $reportPath
$report | Sort-Object Type, Name | Format-Table -AutoSize
Write-Host "Report: $reportPath"
Write-Host "Note: SVG files are created only when -CreateRasterBackedSvg is passed and are not true vector assets. Prefer real vector source for production SVG."
