# go-bat
Go bat is a extenadable framework based on echo/v4. Everything is opt-in. It also includes a project generator.
# Contributing
Want to add another extension, open a pull request and i will check it out.
# Install
## Linux
```bash
curl -L $(curl -s https://api.github.com/repos/JensvandeWiel/go-bat/releases/latest | grep "browser_download_url.*Linux_x86_64.tar.gz" | cut -d '"' -f 4) | sudo tar -xz -C /usr/local/bin
```
### Auto completion
```bash
echo 'source <(go-bat completion bash)' >> ~/.bashrc && source ~/.bashrc
```
## Windows
```powershell
$downloadUrl = (Invoke-RestMethod https://api.github.com/repos/JensvandeWiel/go-bat/releases/latest).assets | Where-Object { $_.name -like "*Windows_x86_64.zip" } | Select-Object -ExpandProperty browser_download_url; $destinationPath = "C:\Program Files\go-bat"; if (-Not (Test-Path $destinationPath)) { New-Item -ItemType Directory -Path $destinationPath }; Invoke-WebRequest -Uri $downloadUrl -OutFile "$destinationPath\go-bat.zip"; Expand-Archive -Path "$destinationPath\go-bat.zip" -DestinationPath $destinationPath -Force; Remove-Item "$destinationPath\go-bat.zip"; [System.Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Program Files\go-bat", [System.EnvironmentVariableTarget]::Machine)
```
### Auto completion
```powershell
go-bat completion powershell | Out-String | Invoke-Expression
```
### Go Install
```bash
go install github.com/JensvandeWiel/go-bat
```
### Arm
You can change the `x86_64` to `arm64` in the download url