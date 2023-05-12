both: windows linux
windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -o bin/windows/evo-sim.exe .
linux:
	GOOS=linux GOARCH=amd64 go build -o bin/linux/evo-sim .