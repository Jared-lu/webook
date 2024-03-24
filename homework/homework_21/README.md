# 测试脚本

```shell
k6 run --duration 10s --vus 1000 k6_websocket_test.js
```

## 1. 调整 WebSocket 的 read buffer 和 write buffer 设置
大小512

![img.png](img.png)

大小1024

![img_1.png](img_1.png)

大小2048

![img_2.png](img_2.png)

## 2. 调整请求和响应的大小。
大小512

![img_3.png](img_3.png)

大小1024

![img_4.png](img_4.png)

大小2048

![img_5.png](img_5.png)


## 3. 调整并发数

200

![img_6.png](img_6.png)

500

![img_7.png](img_7.png)

100

![img_8.png](img_8.png)

2000

![img_9.png](img_9.png)

1000

![img_10.png](img_10.png)



