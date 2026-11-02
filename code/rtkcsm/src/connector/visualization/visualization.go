package visualization

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Action int

const (
	SuscribeAction    Action = 1
	UnsubscribeAction Action = 2
)

type WebsocketMessage struct {
	Action  Action            `json:"action"`
	GraphID structure.GraphID `json:"graph_id"`
}

func Start[T structure.Stage, K structure.Stage](listenAddress string, rtkcsm behaviour.RTKCSM[T, K], fileSystem fs.FS) error {
	server := gin.New()
	server.Use(cors.Default())
	server.Use(gin.Recovery())

	server.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusTemporaryRedirect, "/web/")
	})

	fileSystem, _ = fs.Sub(fileSystem, "static")

	server.StaticFS("/web/", http.FS(fileSystem))

	server.GET("/api/reset", func(ctx *gin.Context) {
		rtkcsm.Reset()
	})

	server.GET("/api/hosts", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, rtkcsm.GetHostRisks())
	})

	server.DELETE("/api/hosts/:address", func(ctx *gin.Context) {
		address := structure.ParseIPAddress(ctx.Param("address"))
		rtkcsm.DeleteHostRisk(address)
		ctx.JSON(http.StatusOK, nil)
	})

	server.POST("/api/hosts", func(ctx *gin.Context) {
		var hostRiskBody structure.HostRisk
		err := ctx.BindJSON(&hostRiskBody)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		ipAddress := structure.ParseIPAddress(hostRiskBody.IpAddress)
		if !ipAddress.IsInternal() || ipAddress.IsUnspecified() {
			ctx.JSON(http.StatusBadRequest, nil)
		} else {
			rtkcsm.AddHostRisk(ipAddress, structure.RiskLevel(hostRiskBody.RiskLevel))
			ctx.JSON(http.StatusOK, nil)
		}
	})

	server.GET("/api/graphs", func(ctx *gin.Context) {
		pageString := ctx.Request.URL.Query().Get("page")

		page, err := strconv.Atoi(pageString)
		if err != nil {
			page = 0
		}

		ctx.JSON(http.StatusOK, rtkcsm.GetGraphList(page))
	})

	server.GET("/api/graphs/:id", func(ctx *gin.Context) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			return
		}
		graph := rtkcsm.GetGraph(structure.GraphID(id))

		if graph != nil {
			preComputedGraph := graph.GetPreComputed()
			format := ctx.Request.URL.Query().Get("format")
			switch format {
			case "ocsf":
				finding := FromGraphToOCSFIncidentFinding(&preComputedGraph)
				ctx.JSON(http.StatusOK, finding)
				findingBytes, err := json.Marshal(finding)
				if err == nil {
					fmt.Println(string(findingBytes))
				}
			default:
				ctx.JSON(http.StatusOK, preComputedGraph)
			}
		} else {
			ctx.Writer.WriteHeader(http.StatusNotFound)
		}
	})

	upgrader := websocket.Upgrader{}

	server.GET("/websocket", func(c *gin.Context) {
		connection, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer connection.Close()

		var queue = make(chan *structure.GraphAddEdgeEvent[T])
		var handle *structure.SubscriberHandle

		go func() {
			for {
				event, ok := <-queue
				if !ok {
					break
				}

				message, err := json.Marshal(event)
				if err == nil {
					connection.WriteMessage(websocket.BinaryMessage, message)
				}
			}
		}()

		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				break
			}

			websocketMessage := WebsocketMessage{}
			json.Unmarshal(message, &websocketMessage)
			if websocketMessage.Action == SuscribeAction || websocketMessage.Action == UnsubscribeAction {
				if handle != nil {
					rtkcsm.Unsubscribe(*handle)
				}

				if websocketMessage.Action == SuscribeAction {
					handleInstance := rtkcsm.Subscribe(websocketMessage.GraphID, queue)
					handle = &handleInstance
				}
			}
		}

		if handle != nil {
			rtkcsm.Unsubscribe(*handle)
		}

		close(queue)
	})

	log.Printf("Visit the web UI at: http://%s/web/", listenAddress)

	err := server.Run(listenAddress)
	if err != nil {
		return err
	}

	return nil
}
