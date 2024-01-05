name="downloader"
GOOS=windows GOARCH=amd64 go build -v -ldflags="-w -s -H windowsgui -X main.Debug=true" -o ./$name.exe
echo "Windows编译完成..."
sleep 10
