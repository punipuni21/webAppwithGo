package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/punipuni21/webApplicationWithGolang/chapter1/trace"
)

type room struct {
	forward chan []byte
	join    chan *client     //チャットルームに参加使用としているクライアントの為のチャネル
	leave   chan *client     //チャットルームから退室しようとしているクライアントの為のチャネル
	clients map[*client]bool //在室している全てのクライアントが保持
	tracer  trace.Tracer
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			r.tracer.Trace("新しいクライアントが参加しました．")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("クライアントが退出しました．")
		case msg := <-r.forward:
			r.tracer.Trace("メッセージを送信しました：", string(msg))
			for client := range r.clients {
				select {
				case client.send <- msg:
					//メッセージを送信
					r.tracer.Trace(" -- クライアントに送信されました")
				default:
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace(" -- 送信に失敗しました．クライアントをクリーンアップします")
				}
			}
		}
	}
}

const (
	sockBufferSize    = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  sockBufferSize,
	WriteBufferSize: sockBufferSize,
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client                     //Websocketコネクションの取得が成功したら現在のチャットルームのjoinチャネルに渡す
	defer func() { r.leave <- client }() //クライアント終了時に退室処理が行われる
	go client.write()                    //goroutineとして実行（別スレッドが立ち上がる）
	client.read()
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}
