set GOARCH=amd64
set GOOS=windows
go build -ldflags="-w -s" -gcflags "-N -l" -trimpath -o blackJack-win-x64.exe
set GOOS=linux
go build -ldflags="-w -s" -gcflags "-N -l" -trimpath -o blackJack-linux-x64
set GOOS=darwin
go build -ldflags="-w -s" -gcflags "-N -l" -trimpath -o blackJack-osx-x64