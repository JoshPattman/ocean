both: windows linux
windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o bin/windows/ocean.exe .
	cp -r ./sprites ./bin/windows/sprites
linux:
	GOOS=linux GOARCH=amd64 go build -o bin/linux/ocean .
	cp -r ./sprites ./bin/linux/sprites