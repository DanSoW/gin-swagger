# Скрипт автоматизации развёртывания рабочей среды Go-приложений
param([String]$separator=";", [String]$path=".\requirements.go.txt")

foreach($line in [System.IO.File]::ReadLines($path)){
    & { go get -u $line.Split($separator)[0].Trim() }
}