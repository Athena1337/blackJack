set GOARCH=amd64
set GOOS=windows
go build -ldflags="-w -s" -gcflags "-N -l" -trimpath -o blackJack-windows-x64.exe
set GOOS=darwin
go build -ldflags="-w -s" -gcflags "-N -l" -trimpath -o blackJack-osx-x64
set GOOS=linux
go build -ldflags="-w -s" -gcflags "-N -l" -trimpath -o blackJack-linux-x64