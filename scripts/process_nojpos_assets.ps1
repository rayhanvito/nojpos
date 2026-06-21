param(
    [int]$MaxDimension = 720,
    [int]$Tolerance = 34
)

$ErrorActionPreference = 'Stop'

Add-Type -AssemblyName System.Drawing

$source = @"
using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Drawing2D;
using System.Drawing.Imaging;
using System.IO;
using System.Runtime.InteropServices;

public static class NojposAssetProcessor
{
    public static void Process(string sourcePath, string destinationPath, int maxDimension, int tolerance)
    {
        using (var source = new Bitmap(sourcePath))
        {
            int width = source.Width;
            int height = source.Height;
            double scale = Math.Min(1.0, (double)maxDimension / Math.Max(width, height));
            int targetWidth = Math.Max(1, (int)Math.Round(width * scale));
            int targetHeight = Math.Max(1, (int)Math.Round(height * scale));

            using (var resized = new Bitmap(targetWidth, targetHeight, PixelFormat.Format32bppArgb))
            {
                using (var graphics = Graphics.FromImage(resized))
                {
                    graphics.Clear(Color.Transparent);
                    graphics.CompositingQuality = CompositingQuality.HighQuality;
                    graphics.InterpolationMode = InterpolationMode.HighQualityBicubic;
                    graphics.SmoothingMode = SmoothingMode.HighQuality;
                    graphics.PixelOffsetMode = PixelOffsetMode.HighQuality;
                    graphics.DrawImage(source, 0, 0, targetWidth, targetHeight);
                }

                RemoveConnectedBackground(resized, tolerance);

                Directory.CreateDirectory(Path.GetDirectoryName(destinationPath));
                resized.Save(destinationPath, ImageFormat.Png);
            }
        }
    }

    private static void RemoveConnectedBackground(Bitmap bitmap, int tolerance)
    {
        int width = bitmap.Width;
        int height = bitmap.Height;
        var rect = new Rectangle(0, 0, width, height);
        var data = bitmap.LockBits(rect, ImageLockMode.ReadWrite, PixelFormat.Format32bppArgb);
        int bytes = Math.Abs(data.Stride) * height;
        byte[] buffer = new byte[bytes];
        Marshal.Copy(data.Scan0, buffer, 0, bytes);

        int sample = Math.Min(12, Math.Min(width, height));
        long r = 0, g = 0, b = 0, count = 0;
        Action<int, int> samplePixel = (x, y) =>
        {
            int idx = y * data.Stride + x * 4;
            b += buffer[idx + 0];
            g += buffer[idx + 1];
            r += buffer[idx + 2];
            count++;
        };

        for (int y = 0; y < sample; y++)
        {
            for (int x = 0; x < sample; x++)
            {
                samplePixel(x, y);
                samplePixel(width - 1 - x, y);
                samplePixel(x, height - 1 - y);
                samplePixel(width - 1 - x, height - 1 - y);
            }
        }

        int bgR = (int)(r / count);
        int bgG = (int)(g / count);
        int bgB = (int)(b / count);
        int toleranceSq = tolerance * tolerance;

        bool[] visited = new bool[width * height];
        var queue = new Queue<int>();

        Action<int, int> enqueueIfBackground = (x, y) =>
        {
            if (x < 0 || y < 0 || x >= width || y >= height) return;
            int pos = y * width + x;
            if (visited[pos]) return;
            int idx = y * data.Stride + x * 4;
            int db = buffer[idx + 0] - bgB;
            int dg = buffer[idx + 1] - bgG;
            int dr = buffer[idx + 2] - bgR;
            if ((dr * dr + dg * dg + db * db) <= toleranceSq)
            {
                visited[pos] = true;
                queue.Enqueue(pos);
            }
        };

        for (int x = 0; x < width; x++)
        {
            enqueueIfBackground(x, 0);
            enqueueIfBackground(x, height - 1);
        }
        for (int y = 0; y < height; y++)
        {
            enqueueIfBackground(0, y);
            enqueueIfBackground(width - 1, y);
        }

        while (queue.Count > 0)
        {
            int pos = queue.Dequeue();
            int x = pos % width;
            int y = pos / width;
            int idx = y * data.Stride + x * 4;
            buffer[idx + 3] = 0;
            enqueueIfBackground(x + 1, y);
            enqueueIfBackground(x - 1, y);
            enqueueIfBackground(x, y + 1);
            enqueueIfBackground(x, y - 1);
        }

        Marshal.Copy(buffer, 0, data.Scan0, bytes);
        bitmap.UnlockBits(data);
    }
}
"@

Add-Type -TypeDefinition $source -ReferencedAssemblies System.Drawing

$root = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$illustrationOut = Join-Path $root 'apps/cashier/assets/illustrations/nojpos'
$logoOut = Join-Path $root 'apps/cashier/assets/logos/nojpos'
New-Item -ItemType Directory -Force -Path $illustrationOut, $logoOut | Out-Null

$mobileAssets = @(
    'payment-success',
    'payment-failed',
    'receipt-ready',
    'refund-success',
    'success-saved',
    'shift-open',
    'shift-closed',
    'empty-products',
    'empty-customers',
    'empty-transactions',
    'empty-reports',
    'empty-notifications',
    'offline',
    'no-internet-cashier',
    'error-occurred',
    'session-expired',
    'permission-denied',
    'access-denied',
    'loading-data',
    'cloud-sync',
    'low-stock-warning',
    'out-of-stock',
    'barcode-scanner',
    'maintenance'
)

$logoAssets = @(
    'logo-app-icon',
    'logo-nojpos',
    'logo-symbol'
)

$report = @()

foreach ($name in $mobileAssets) {
    $src = Join-Path $root "Asset NOJPOS/mobile/$name.png"
    $dst = Join-Path $illustrationOut "$name.png"
    if (Test-Path $src) {
        [NojposAssetProcessor]::Process($src, $dst, $MaxDimension, $Tolerance)
        $report += [pscustomobject]@{
            Type = 'illustration'
            Name = $name
            SourceKB = [math]::Round((Get-Item $src).Length / 1KB, 1)
            OutputKB = [math]::Round((Get-Item $dst).Length / 1KB, 1)
            Output = $dst.Replace($root + [IO.Path]::DirectorySeparatorChar, '')
        }
    }
}

foreach ($name in $logoAssets) {
    $src = Join-Path $root "Asset NOJPOS/logo/$name.png"
    $dst = Join-Path $logoOut "$name.png"
    if (Test-Path $src) {
        [NojposAssetProcessor]::Process($src, $dst, $MaxDimension, $Tolerance)
        $report += [pscustomobject]@{
            Type = 'logo'
            Name = $name
            SourceKB = [math]::Round((Get-Item $src).Length / 1KB, 1)
            OutputKB = [math]::Round((Get-Item $dst).Length / 1KB, 1)
            Output = $dst.Replace($root + [IO.Path]::DirectorySeparatorChar, '')
        }
    }
}

$reportPath = Join-Path $root 'apps/cashier/assets/nojpos_asset_processing_report.csv'
$report | Export-Csv -NoTypeInformation -Path $reportPath
$report | Sort-Object Type, Name | Format-Table -AutoSize
Write-Host "Report: $reportPath"
