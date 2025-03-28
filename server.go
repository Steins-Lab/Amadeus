package main

import (
	"encoding/json"
	"fmt"
	"github.com/Steins-Lab/Amadeus-SDK/event"
	"github.com/Steins-Lab/Amadeus-SDK/handler"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Server struct {
	port         string
	path         string
	conn         *websocket.Conn
	logger       *slog.Logger
	timeout      int
	messageEvent chan event.Event
	id           int64
}

var upGrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

func (wsc *Server) sendGroupMessage(groupMessage *event.GroupMessage) {
	wsc.sendRequest("send_group_msg", groupMessage)
}

func (wsc *Server) sendRequest(method string, data any) {
	wsc.id += 1
	var request = event.Request{
		Action: method,
		Params: data,
		Echo:   strconv.FormatInt(wsc.id, 10),
	}
	marshal, err := json.Marshal(request)
	if err != nil {
		return
	}
	err = wsc.conn.WriteMessage(websocket.TextMessage, marshal)
	if err != nil {
		wsc.logger.Warn("send request error:", err)
	}
}

func (wsc *Server) readMessage() bool {
	_, p, err := wsc.conn.ReadMessage()
	if err != nil {
		return false
	}
	var e event.Event
	err = json.Unmarshal(p, &e)
	if err != nil {
		return true
	}
	wsc.messageEvent <- e
	return true
}

func (wsc *Server) analyze() {
	go func() {
		for data := range wsc.messageEvent {
			if !strings.EqualFold(data.MessageType, "") {
				switch data.MessageType {
				case "group":
					// TODO 暂时限定群回复
					if data.GroupID == 104967737 {
						wsc.sendGroupMessage(&event.GroupMessage{GroupId: data.GroupID, Message: "hello"})
					}
					wsc.logger.Info(fmt.Sprintf("[%s][%d][%d]%s: %s", wsc.conn.RemoteAddr(), data.GroupID, data.Sender.UserID, data.Sender.Nickname, data.Message))
				case "private":
					wsc.logger.Info(fmt.Sprintf("[%s][%d]%s: %s", wsc.conn.RemoteAddr(), data.Sender.UserID, data.Sender.Nickname, data.Message))
				}

			} else {
				wsc.logger.Debug(fmt.Sprintf("%+v\n", data))
			}
		}
	}()
}

func (wsc *Server) handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		wsc.logger.Error(err.Error())
		return
	}
	wsc.logger.Info(fmt.Sprintf("[%s] 连接到 %s", conn.RemoteAddr(), r.URL.Path))
	wsc.conn = conn
	go func() {
		defer func(conn *websocket.Conn) {
			err := conn.Close()
			if err != nil {
				wsc.logger.Warn(fmt.Sprintf("[%s]连接已断开", wsc.conn.RemoteAddr()))
			}
		}(wsc.conn)
		for {

			if !wsc.readMessage() {
				break
			}
		}
	}()
}

func (wsc *Server) Start() {
	fmt.Println("  ______    _               _____       \n |  ____|  (_)             / ____|      \n | |__ __ _ _ _ __ _   _  | |  __  ___  \n |  __/ _` | | '__| | | | | | |_ |/ _ \\ \n | | | (_| | | |  | |_| | | |__| | (_) |\n |_|  \\__,_|_|_|   \\__, |  \\_____|\\___/ \n                    __/ |               \n                   |___/                ")
	wsc.logger.Info(fmt.Sprintf("服务端已启动, 端口[%s]", wsc.port))
	http.HandleFunc(wsc.path, wsc.handler)
	wsc.analyze()
	err := http.ListenAndServe(fmt.Sprintf(":%s", wsc.port), nil)
	if err != nil {
		wsc.logger.Error(err.Error())
	}
}

func NewWsServerManager(port, path string, timeout int) *Server {
	var conn *websocket.Conn

	var opts = handler.PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	var prettyHandler = handler.NewPrettyHandler(os.Stdout, opts)
	var logger = slog.New(prettyHandler)

	return &Server{
		port:         port,
		path:         path,
		conn:         conn,
		logger:       logger,
		timeout:      timeout,
		messageEvent: make(chan event.Event, 100),
		id:           0,
	}
}

func main() {
	NewWsServerManager("8080", "/onebot/v11", 5).Start()
}
