package visualization

import (
	"encoding/json"
	"fmt"
	"io"
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

type MCPMessage struct {
	JSONRPCVersion string `json:"jsonrpc"`
	ID             int    `json:"id"`
}

type MCPRequestMethod string

const InitializeMCPRequestMethod MCPRequestMethod = "initialize"
const InitializedMCPRequestMethod MCPRequestMethod = "notifications/initialized"
const ToolsListMCPRequestMethod MCPRequestMethod = "tools/list"
const ToolCallMCPRequestMethod MCPRequestMethod = "tools/call"

type MCPRequest struct {
	MCPMessage
	Method     MCPRequestMethod     `json:"method"`
	Parameters MCPRequestParameters `json:"params"`
}

type MCPRequestParameters struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type MCPResponse struct {
	MCPMessage
	Result interface{} `json:"result"`
}

type MCPDiscovery struct {
	ProtocolVersion string                    `json:"protocolVersion"`
	Capabilities    map[string]map[string]any `json:"capabilities"`
	ServerInfo      MCPServerInfo             `json:"serverInfo"`
}

type MCPServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type MCPTools struct {
	Tools []MCPTool `json:"tools"`
}

type MCPTool struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	InputSchema MCPToolInput `json:"inputSchema"`
}

type MCPToolInput struct {
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
	Required   []string       `json:"required,omitempty"`
}

type MCPToolOutputs struct {
	Content []MCPToolOutput `json:"content"`
}

type MCPToolOutput struct {
	Type string `json:"type"`
	Text string `json:"text"`
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

	server.POST("/api/mcp", func(ctx *gin.Context) {
		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		var request = MCPRequest{}
		err = json.Unmarshal(body, &request)
		if err != nil {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		response := &MCPResponse{
			MCPMessage: MCPMessage{
				JSONRPCVersion: "2.0",
				ID:             request.ID,
			},
		}

		switch request.Method {

		case InitializeMCPRequestMethod:
			response.Result = MCPDiscovery{
				ProtocolVersion: "2024-11-05",
				Capabilities: map[string]map[string]any{
					"tools": map[string]any{
						"listChanged": false,
					},
				},
				ServerInfo: MCPServerInfo{
					Name:    "RT-KCSM",
					Version: "1.0",
				},
			}
		case InitializedMCPRequestMethod:
			ctx.AbortWithStatus(http.StatusOK)
			return
		case ToolsListMCPRequestMethod:
			response.Result = MCPTools{
				Tools: []MCPTool{
					{
						Name:        "get_attack_graph",
						Description: "Returns the graph of an ongoing attack",
						InputSchema: MCPToolInput{
							Type: "object",
							Properties: map[string]any{
								"graphId": map[string]any{
									"type":        "integer",
									"description": "ID of graph",
								},
							},
							Required: []string{"graphId"},
						},
					},
					{
						Name:        "list_attack_graphs",
						Description: "Sorted list of attacks graphs and their relevance score in descending order of relevance score",
						InputSchema: MCPToolInput{
							Type: "object",
							Properties: map[string]any{
								"page": map[string]any{
									"type":        "integer",
									"description": "Page number of paginated attack graphs",
								},
							},
						},
					},
				},
			}
		case ToolCallMCPRequestMethod:
			if request.Parameters.Name == "get_attack_graph" {
				graphId, ok := request.Parameters.Arguments["graphId"].(float64)
				if !ok {
					ctx.AbortWithStatus(http.StatusNotFound)
					return
				}

				graph := rtkcsm.GetGraph(structure.GraphID(graphId))
				if graph == nil {
					ctx.AbortWithStatus(http.StatusNotFound)
					return
				}

				text, err := json.Marshal(graph)
				if err != nil {
					ctx.AbortWithStatus(http.StatusNotFound)
					return
				}

				response.Result = MCPToolOutputs{
					Content: []MCPToolOutput{
						{
							Type: "text",
							Text: string(text),
						},
					},
				}
			} else if request.Parameters.Name == "list_attack_graphs" {
				page, ok := request.Parameters.Arguments["page"].(float64)
				if !ok {
					page = 0.0
				}

				pageNumber := int(page)

				graphs := rtkcsm.GetGraphList(pageNumber)
				text, err := json.Marshal(graphs)
				if err != nil {
					ctx.AbortWithStatus(http.StatusNotFound)
					return
				}

				response.Result = MCPToolOutputs{
					Content: []MCPToolOutput{
						{
							Type: "text",
							Text: string(text),
						},
					},
				}
			}
		default:
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.JSON(http.StatusOK, response)
	})

	log.Printf("Visit the web UI at: http://%s/web/", listenAddress)

	err := server.Run(listenAddress)
	if err != nil {
		return err
	}

	return nil
}
