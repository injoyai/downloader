name="downloader"
GOOS=windows GOARCH=amd64 go build -v -ldflags="-w -s -H windowsgui" -o ./$name.exe
echo "Windows编译完成..."
sleep 5
