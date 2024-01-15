package redismock

// source需要指定本地文件，并将用户权限设置为完全控制
// 这里仅仅用于Windows开发环境，这个奇怪的玩意
// 如果包的版本变了，可能需要重新指定路径和生成mock
//go:generate mockgen -source=E:/workspace/GoLandProject/pkg/mod/github.com/redis/go-redis/v9@v9.3.1/commands.go -package=redismock -destination=./redis.mock.go
