package im

import (
	"fmt"
	"stream_hub/internal/infra"
	"stream_hub/internal/security"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/utils"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

type IMRouter struct {
	router     *gin.Engine
	im         *IMApi
	middleware *Middleware
	port       int
	hub        *Hub
}

func NewIMRouter(base *infra.Base, auth *security.Auth, conf *config.IMConfig, producer sarama.AsyncProducer) *IMRouter {
	hub := NewHub()
	go hub.Run()

	router := &IMRouter{
		port:       conf.Port,
		im:         NewIMApi(base, producer),
		middleware: NewMiddleware(base, auth),
		hub:        hub,
	}

	router.init()

	return router
}

func (r *IMRouter) init() {
	r.router = gin.Default()
	r.router.Use(r.middleware.Cors(), r.middleware.LogToStorage())

	im := r.router.Group("/im")
	{
		im.POST("/send_message", r.middleware.Auth(), r.im.SendMessage)
		im.POST("/get_messages", r.middleware.Auth(), r.im.GetMessages)
		im.POST("/get_conversations", r.middleware.Auth(), r.im.GetConversations)
		im.GET("/ws", r.middleware.Auth(), r.handleWebSocket)
	}
}

func (r *IMRouter) handleWebSocket(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	if userID == "" {
		utils.UnAuthorizationRequest(ctx, "need login")
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		ClientID: userID,
	}

	r.hub.Register <- client

	go client.WritePump()
	go client.ReadPump(r.hub)
}

func (r *IMRouter) Run() error {
	return r.router.Run(fmt.Sprintf(":%d", r.port))
}

func (r *IMRouter) Hub() *Hub {
	return r.hub
}
