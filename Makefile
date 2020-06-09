#Build Dll
all:
	set CGO_ENABLED=1 
	go build -buildmode=c-shared -ldflags="-s -w" -o .\main.dll .\dllmain.go 
