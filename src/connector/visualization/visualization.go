package visualization

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func Start(listenAddress string, rtkcsm behaviour.RTKCSM, fileSystem fs.FS) error {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	server := gin.Default()
	server.Use(cors.Default())

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
		ctx.BindJSON(&hostRiskBody)

		ipAddress := structure.ParseIPAddress(hostRiskBody.IpAddress)
		if !ipAddress.IsPrivate() || ipAddress.IsUnspecified() {
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
			page = -1
		}

		ctx.JSON(http.StatusOK, rtkcsm.GetGraphList(page))
	})

	server.GET("/api/graphs/:id", func(ctx *gin.Context) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.AbortWithError(http.StatusNotFound, errors.New("graph not found"))
		}
		graph := rtkcsm.GetGraph(structure.GraphID(id))

		if graph != nil {
			ctx.Writer.WriteString(graph.ExportDot(ctx.Query("simplify") == "true"))
		} else {
			ctx.Writer.WriteHeader(http.StatusNotFound)
		}
	})

	server.GET("/websocket/graphs/:id", func(ctx *gin.Context) {
		connection, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("cannot upgrade websocket"))
		}

		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.AbortWithError(http.StatusNotFound, errors.New("graph not found"))
		}

		callback := func(event structure.Event) {
			newDirectedRelationEventData := event.GetData().(structure.NewDirectedRelationEventData)

			directedRelation := newDirectedRelationEventData.DirectedRelation
			srcNode := directedRelation.SrcNode
			dstNode := directedRelation.DstNode

			data := RelationEvent{
				GraphRelevance: newDirectedRelationEventData.GraphRelevance,
				GraphID:        event.GetGraphID(),
				Relation: RelationEventData{
					From: NodeRelationEventData{
						Address:   srcNode.String(),
						IsPrivate: srcNode.IsPrivate(),
						Risk:      float32(structure.HostManager.GetHostRiskLevel(srcNode)),
					},
					To: NodeRelationEventData{
						Address:   dstNode.String(),
						IsPrivate: dstNode.IsPrivate(),
						Risk:      float32(structure.HostManager.GetHostRiskLevel(dstNode)),
					},
					Stages:    fmt.Sprintf("%v", directedRelation.Stage.ToUKCStages()),
					Timestamp: directedRelation.Timestamp.Format(time.RFC3339),
					Severity:  directedRelation.Severity,
				},
			}
			messageBody, _ := json.Marshal(&data)
			connection.WriteMessage(websocket.TextMessage, messageBody)
		}

		eventManager := rtkcsm.GetEventManager()
		eventManager.Subscribe(structure.GraphID(id), &callback)

		for {
			_, _, err := connection.NextReader()
			if err != nil {
				break
			}
		}

		defer eventManager.Unsubscribe(structure.GraphID(id), &callback)
	})

	server.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusTemporaryRedirect, "/web/")
	})

	log.Printf("Visit the web UI at: http://%s/web/", listenAddress)

	err := server.Run(listenAddress)
	if err != nil {
		return err
	}

	return nil
}
