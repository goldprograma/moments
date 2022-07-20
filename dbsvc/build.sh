output_linux="./_build/linux/dbsvc"
output_linux_production="./_build/linux_prod/dbsvc"
output_mac="./_build/mac/dbsvc"
output_win="./_build/win/dbsvc.exe"
output="./_build/"

rm -rf "./_build/"*

case $1 in
    linux)
    echo "build for linux_debug"
    CGO_ENABLED=0 GOOS=linux go build -o=$output_linux
    ;;
    linux_prod)
    echo "build for linux_production"
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o=$output_linux_production
    ;;
    mac)
    echo "build for macos"
    CGO_ENABLED=0 GOOS=darwin go build -o=$output_mac
    ;;
    win)
    echo "build for windows"
    CGO_ENABLED=0 GOOS=windows go build -o=$output_win
    ;;
    *)
    echo "build for local"
    go build -o=$output
    ;;
esac

echo "done"