package main

import (
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

func main() {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			panic(err)
		}

		ws := &Ws{conn: conn}
		go func() {
			ws.ReadCycle()
		}()

		go func() {
			// 模拟写回响应
			err = ws.Write("响应：" + makeResponse(8192))
			if err != nil {
				log.Println(err)
				return
			}
		}()

	})

	err := http.ListenAndServe("127.0.0.1:8081", nil)
	if err != nil {
		panic(err)
	}
}

func makeResponse(size int) string {
	var data []byte
	for i := 0; i < size; i++ {
		data = append(data, 'a')
	}
	return string(data)
}

type Ws struct {
	conn *websocket.Conn
}

func (w *Ws) ReadCycle() {
	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("收到数据", string(message))
	}
}

func (w *Ws) Write(data string) error {

	return w.conn.WriteMessage(websocket.TextMessage, []byte(data))
}
