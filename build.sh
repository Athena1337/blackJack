GOARCH=amd64
GOOS=windows go build -o blackJack-windows-x64.exe -ldflags="-w -s" -gcflags "-N -l" -trimpath
GOOS=linux go build -o blackJack-linux-x64 -ldflags="-w -s" -gcflags "-N -l" -trimpath
GOOS=darwin go build -o blackJack-osx-x64 -ldflags="-w -s" -gcflags "-N -l" -trimpath