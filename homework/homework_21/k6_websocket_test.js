import ws from 'k6/ws';

// 生成指定大小的消息
function generateMessage(size) {
    let chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let message = '';
    for (var i = 0; i < size; i++) {
        message += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return message;
}

export default function () {
    // WebSocket服务器的地址
    const url = 'ws://localhost:8081/ws';

    // 建立WebSocket连接
    const res = ws.connect(url, null, function (socket) {
        // 连接成功时执行的回调函数
        console.log('WebSocket连接已建立');

        // 发送消息给服务器
        socket.send(JSON.stringify({message: generateMessage(8192)}));

        // 监听服务器发送的消息
        socket.on('message', function (message) {
            console.log(`接收到服务器的消息：${message}`);
            socket.close();
        });
    });
}
