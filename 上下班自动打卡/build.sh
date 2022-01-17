cp ./config.xml bin/
go build -ldflags="-H windowsgui" -o bin/auto.exe .