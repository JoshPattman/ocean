both: windows linux
windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o bin/windows/evo-sim.exe .
	cp -r ./sprites ./bin/windows/sprites
linux:
	GOOS=linux GOARCH=amd64 go build -o bin/linux/evo-sim .
	cp -r ./sprites ./bin/linux/sprites