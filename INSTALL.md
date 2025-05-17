# SprayShark instalation procedures 

## Linux


### Installing Chrome and Chrome Driver

```
wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
apt install ./google-chrome-stable_current_amd64.deb
apt install chromium-driver
```

### Installing SprayShark

```
apt install curl jq


url=$(curl -s https://api.github.com/repos/helviojunior/sprayshark/releases | jq -r '[ .[] | {id: .id, tag_name: .tag_name, assets: [ .assets[] | select(.name|match("linux-amd64.tar.gz$")) | {name: .name, browser_download_url: .browser_download_url} ]} | select(.assets != []) ] | sort_by(.id) | reverse | first(.[].assets[]) | .browser_download_url')

cd /tmp
rm -rf sprayshark-latest.tar.gz sprayshark
wget -nv -O sprayshark-latest.tar.gz "$url"
tar -xzf sprayshark-latest.tar.gz

rsync -av sprayshark /usr/local/sbin/
chmod +x /usr/local/sbin/sprayshark

sprayshark version
```

## MacOS

### Installing HomeBrew

Note: Just run this command if you need to install HomeBrew to first time

```
/bin/bash -c "$(curl -fsSL raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### Installing SprayShark

```
brew install curl jq

arch=$(if [[ "$(uname -m)" -eq "x86_64" ]]; then echo "amd64"; else echo "arm64"; fi)

url=$(curl -s https://api.github.com/repos/helviojunior/sprayshark/releases | jq -r --arg filename "darwin-${arch}.tar.gz\$" '[ .[] | {id: .id, tag_name: .tag_name, assets: [ .assets[] | select(.name|match($filename)) | {name: .name, browser_download_url: .browser_download_url} ]} | select(.assets != []) ] | sort_by(.id) | reverse | first(.[].assets[]) | .browser_download_url')

cd /tmp
rm -rf sprayshark-latest.tar.gz sprayshark
curl -sS -L -o sprayshark-latest.tar.gz "$url"
tar -xzf sprayshark-latest.tar.gz

rsync -av sprayshark /usr/local/sbin/
chmod +x /usr/local/sbin/sprayshark

sprayshark version
```

## Windows

Just run the following powershell script

```
# Download latest helviojunior/sprayshark release from github
function Invoke-Downloadsprayshark {

    $repo = "helviojunior/sprayshark"
    
    # Determine OS and Architecture
    $osPlatform = [System.Runtime.InteropServices.RuntimeInformation]::OSDescription
    $architecture = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture

    if ($architecture -eq $null -or $architecture -eq "") {
        $architecture = $Env:PROCESSOR_ARCHITECTURE
    }
        
    if ($osPlatform -eq $null -or $osPlatform -eq "") {
        $osPlatform = $Env:OS
    }

    # Adjust the platform and architecture for the API call
    $platform = switch -Wildcard ($osPlatform) {
        "*Windows*" { "windows" }
        "*Linux*"   { "linux" }
        "*Darwin*"  { "darwin" } # MacOS is identified as Darwin
        Default     { "unknown" }
    }
    $arch = switch ($architecture) {
        "X64"  { "amd64" }
        "AMD64"  { "amd64" }
        "X86"  { "386" }
        "Arm"  { "arm" }
        "Arm64" { "arm64" }
        Default { "unknown" }
    }

    if ($platform -eq "unknown" -or $arch -eq "unknown") {
        Write-Error "Cannot get OS Platform and Architecture"
        Return
    }

    Write-Host Getting release list
    $releases = "https://api.github.com/repos/$repo/releases"

    $asset = Invoke-WebRequest $releases | ConvertFrom-Json | Sort-Object -Descending -Property "Id" | ForEach-Object -Process { Get-AssetData -Release $_ -OSPlatform $platform -OSArchitecture $arch } | Select-Object -First 1

    if ($asset -eq $null -or $asset.browser_download_url -eq $null){
        Write-Error "Cannot find a valid URL"
        Return
    }

    $tmpPath = $Env:Temp
    if ($tmpPath -eq $null -or $tmpPath -eq "") {
        $tmpPath = $Env:TMPDIR
    }
    if ($tmpPath -eq $null -or $tmpPath -eq "") {
        $tmpPath = switch ($platform) {
            "windows" { "c:\windows\temp\" }
            "linux"   { "/tmp" }
            "darwin"  { "/tmp" }
        }
    }

    $extension = switch ($platform) {
        "windows" { ".zip" }
        "linux"   { ".tar.gz" }
        "darwin"  { ".tar.gz" } # MacOS is identified as Darwin
        Default     { "unknown" }
    }

    $file = "sprayshark-latest$extension"

    Write-Host Dowloading latest release
    $zip = Join-Path -Path $tmpPath -ChildPath $file
    Remove-Item $zip -Force -ErrorAction SilentlyContinue 
    Invoke-WebRequest $asset.browser_download_url -Out $zip

    Write-Host Extracting release files
    if ($extension -eq ".zip") {
        Expand-Archive $zip -Force -DestinationPath $tmpPath
    }else{
        . tar -xzf "$zip" -C "$tmpPath"
    }

    $exeFilename = switch ($platform) {
        "windows" { "sprayshark.exe" }
        "linux"   { "sprayshark" }
        "darwin"  { "sprayshark" } 
    }

    try {
        $dstPath = (New-Object -ComObject Shell.Application).NameSpace('shell:Downloads').Self.Path
    } catch {
        $dstPath = switch ($platform) {
            "windows" { "~\Downloads\" }
            "linux"   { "/usr/local/sbin/" }
            "darwin"  { "/usr/local/sbin/" } 
        }
    }

    try {
        $name = Join-Path -Path $dstPath -ChildPath $exeFilename

        # Cleaning up target dir
        Remove-Item $name -Recurse -Force -ErrorAction SilentlyContinue 

        # Moving from temp dir to target dir
        Move-Item $(Join-Path -Path $tmpPath -ChildPath $exeFilename) -Destination $name -Force

        # Removing temp files
        Remove-Item $zip -Force
    } catch {
        $name = Join-Path -Path $tmpPath -ChildPath $exeFilename
    }
    
    Write-Host "SprayShark saved at $name" -ForegroundColor DarkYellow

    Write-Host "Getting sprayshark version banner"
    . $name version 
}

Function Get-AssetData {
    [CmdletBinding(SupportsShouldProcess = $False)]
    [OutputType([object])]
    Param (
        [Parameter(Mandatory = $True, Position = 0)]
        [object]$Release,
        [Parameter(Mandatory = $True, Position = 1)]
        [string]$OSPlatform,
        [Parameter(Mandatory = $True, Position = 2)]
        [string]$OSArchitecture
    )

    if($Release -is [system.array]){
        $Release = $Release[0]
    }
    
    if (Get-Member -inputobject $Release -name "assets" -Membertype Properties) {
        
        $extension = switch ($OSPlatform) {
            "windows" { ".zip" }
            "linux"   { ".tar.gz" }
            "darwin"  { ".tar.gz" } # MacOS is identified as Darwin
            Default     { "unknown" }
        }

        foreach ($asset in $Release.assets)
        {
            If ($asset.name.Contains("sprayshark-") -and $asset.name.Contains("$platform-$arch$extension")) { Return $asset }
        }

    }
    Return $null
} 

Invoke-Downloadsprayshark  

```


# Build from source

## Linux environment

Follows the suggest commands to install linux environment

### Installing Go v1.23.5

```
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz
rm -rf /usr/bin/go && ln -s /usr/local/go/bin/go /usr/bin/go
```

## Build sprayshark

Clone the repository and build the project with Golang:

```
git clone https://github.com/helviojunior/sprayshark.git
cd sprayshark
go get ./...
go build
```

If you want to update go.sum file just run the command `go mod tidy`.

## Installing system wide

After build run the commands bellow

```
go install .
ln -s /root/go/bin/sprayshark /usr/bin/sprayshark
```
