$ErrorActionPreference = 'Stop'

$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
$name = "nosub_windows_$arch.exe"
$base = 'https://github.com/ZXlinxk/nosub/releases/latest/download'
$dir = Join-Path $env:LOCALAPPDATA 'nosub'
$exe = Join-Path $dir 'nosub.exe'

New-Item -ItemType Directory -Force $dir | Out-Null
Invoke-WebRequest "$base/$name" -OutFile $exe -UseBasicParsing
$sumsFile = Join-Path $dir 'checksums.txt'
Invoke-WebRequest "$base/checksums.txt" -OutFile $sumsFile -UseBasicParsing
$want = ((Get-Content $sumsFile | Where-Object { $_ -match "\s$name\s*$" }) -split '\s+')[0]
Remove-Item $sumsFile
if ((Get-FileHash $exe -Algorithm SHA256).Hash -ne "$want".ToUpper()) {
    Remove-Item $exe
    throw 'checksum mismatch'
}

$path = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($path -notlike "*$dir*") {
    [Environment]::SetEnvironmentVariable('Path', "$path;$dir", 'User')
}
Write-Host "installed to $exe, open a new terminal to use it"
