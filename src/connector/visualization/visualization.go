package visualization

import (
	"io/fs"
	"log"
	"net/http"
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Start[T structure.Stage, K structure.Stage](listenAddress string, rtkcsm behaviour.RTKCSM[T, K], fileSystem fs.FS) error {
	server := gin.Default()
	server.Use(cors.Default())

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
				ctx.JSON(http.StatusOK, FromGraphToOCSFIncidentFinding(&preComputedGraph))
			default:
				ctx.JSON(http.StatusOK, preComputedGraph)
			}
		} else {
			ctx.Writer.WriteHeader(http.StatusNotFound)
		}
	})

	log.Printf("Visit the web UI at: http://%s/web/", listenAddress)

	err := server.Run(listenAddress)
	if err != nil {
		return err
	}

	return nil
}
