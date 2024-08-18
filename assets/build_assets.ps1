$files = Get-ChildItem -Path "." -Recurse -Filter "*.svg" -File

svg2gio -pkg assets -o gio_assets.go $files 
