import cytoscape from "cytoscape"
import dagre from "cytoscape-dagre"
import { parseDOTNetwork } from "vis-network/standalone"

cytoscape.use(dagre)

const graphElement = document.querySelector(".graph") as HTMLDivElement
const graphIdElement = document.querySelector(".graph-id")
const graphCountElement = document.querySelector(".graph-count")
const hostModalElement = document.querySelector(".modal")
const hostListElement = document.querySelector(".modal .hosts")
const newHostAddressElement = document.querySelector(".modal .create-host .address") as HTMLInputElement
const newHostRiskElement = document.querySelector(".modal .create-host .riskLevel") as HTMLInputElement
const simplifyCheckboxElement = document.querySelector(".sidebar .controls #simplify-button") as HTMLInputElement

const tenzirBaseUrl = "https://app.tenzir.com/explorer"
const riskLevels = [{
    name: "low",
    value: 0.5
}, {
    name: "normal",
    value: 1
}, {
    name: "high",
    value: 1.5
}]

let parsedData = null

let graph: cytoscape.Core | null = null
let webSocket: WebSocket | null = null

function openGraph(id: Number) {
    graphIdElement.textContent = String(id)
    let webSocketProtocol = "ws"

    if (window.location.protocol == "https:") {
        webSocketProtocol = "wss"
    }

    webSocket = new WebSocket(`${webSocketProtocol}://${document.location.host}/websocket/graphs/${encodeURIComponent(String(id))}`)

    webSocket.addEventListener("message", (event) => {
        const data = JSON.parse(event.data)
        const relation = data.relation

        if (graph) {
            if (graph.nodes(`#${relation.from.address}`).length == 0) {
                graph.add([{
                    group: "nodes",
                    data: {
                        id: relation.from.address,
                        width: relation.from.risk * 100,
                        height: relation.from.risk * 100,
                        label: relation.from.address,
                    },
                    style: {
                        "background-color": relation.from.is_private ? "blue" : "red"
                    }
                }])
            }

            if (graph.nodes(`#${relation.to.address}`).length == 0) {
                graph.add([{
                    group: "nodes",
                    data: {
                        id: relation.to.address,
                        width: relation.to.risk * 100,
                        height: relation.to.risk * 100,
                        label: relation.to.address,
                    },
                    style: {
                        "background-color": relation.to.is_private ? "blue" : "red"
                    }
                }])
            }

            if (graph.edges(`edge[source = "${relation.from.address}"][target = "${relation.to.address}"]`).length == 0) {
                graph.add([{
                    group: "edges",
                    data: {
                        source: relation.from.address,
                        target: relation.to.address,
                        date: new Date(relation.timestamp),
                        width: relation.severity,
                        label: relation.stages,
                    }
                }])
            }
        }

        let graphListElement = document.querySelector(`#graph-list-item-${data.id}`)

        if (graphListElement) {
            graphListElement.querySelector(".relevance").textContent = data.relevance
        }
    })

    fetch(`/api/graphs/${encodeURIComponent(String(id))}?simplify=${simplifyCheckboxElement.checked ? "true" : "false"}`)
        .then(response => response.text())
        .then(data => {
            parsedData = parseDOTNetwork(data)

            let elements = [];

            parsedData.nodes.forEach((node) => {
                elements.push({
                    group: "nodes",
                    data: {
                        id: node.id,
                        width: node.risk * 100,
                        height: node.risk * 100,
                        label: node.id,
                    },
                    style: {
                        "background-color": node.color.background,
                    }
                })
            })

            parsedData.edges.forEach((edge) => {
                elements.push({
                    group: "edges",
                    data: {
                        source: edge.from,
                        target: edge.to,
                        date: new Date(edge.date * 1000),
                        width: edge.weight,
                        label: edge.label,
                        edge: edge.date,
                    }
                })
            })

            graph = cytoscape({
                container: graphElement,
                elements: elements,
                layout: {
                    name: "dagre",
                    rankDir: "LR",
                    rankSep: 110,
                    edgeSep: 2,
                    nodeSep: 110,
                },
                style: [{
                    selector: 'node',
                    css: {
                        'label': 'data(id)',
                        'color': '#fff',
                    },
                },
                {
                    selector: 'edge',
                    css: {
                        'line-color': '#ccc',
                        'target-arrow-color': "#ccc",
                        'target-arrow-shape': 'triangle',
                        'curve-style': 'bezier',
                        'label': 'data(label)',
                        'width': 'data(width)',
                        'color': "grey"
                    }
                }]
            })

            window["graph"] = graph

            graph.on("dblclick", (event) => {
                const address = event.target[0]._private.data.id
                const query = `export | where :ip == ${address}`
                const url = `${tenzirBaseUrl}?pipeline=${encodeURIComponent(btoa(query))}`

                window.open(url, "_blank")
            })

            graph.on("add", (event) => {
                event.cy.layout({
                    name: "dagre",
                    rankDir: "LR",
                    rankSep: 110,
                    edgeSep: 2,
                    nodeSep: 110,
                }).start()
            })
        })
        .catch(error => console.error("Error fetching DOT graph:", error))
}

const list = document.querySelector(".sidebar .list")

function refreshGraphs() {
    list.innerHTML = ""

    fetch(`/api/graphs?page=0`)
        .then(response => {
            if (!response.ok) {
                throw new Error("Network response was not ok")
            }
            return response.json()
        })
        .then(data => {
            graphCountElement.textContent = data.count

            if (data.graphs.length > 0) {
                openGraph(data.graphs[0].id)
            }

            data.graphs.forEach(item => {
                const graph = document.createElement("div")
                graph.className = "graph"
                graph.id = `graph-list-item-${item.id}`

                const id = document.createElement("span")
                id.className = "id"
                id.textContent = item.id

                const edgeCount = document.createElement("span")
                edgeCount.className = "relevance"
                edgeCount.textContent = item.relevance

                graph.onclick = function () {
                    openGraph(item.id)
                };

                graph.appendChild(id)
                graph.appendChild(edgeCount)

                list.appendChild(graph)
            });
        });
}

function resetGraphs() {
    fetch("/api/reset").then(() => {
        refreshGraphs()
        if (graph) {
            graph.destroy()
        }

        graphIdElement.textContent = ""
    })
}

function openHostModal() {
    hostModalElement.classList.remove("closed")
    listHosts()
}

function createHost() {
    const ipAddress = newHostAddressElement.value
    const riskLevel = newHostRiskElement.value

    fetch("/api/hosts", {
        method: "POST",
        body: JSON.stringify({
            ip_address: ipAddress,
            risk_level: parseFloat(riskLevel)
        })
    }).then(() => {
        listHosts()
        refreshGraphs()
    })
}

function deleteHost(ipAddress) {
    return fetch(`/api/hosts/${ipAddress}`, {
        method: "DELETE"
    })
}

function listHosts() {
    hostListElement.innerHTML = ""
    fetch("/api/hosts").then(response => {
        if (!response.ok) {
            throw new Error("Network response was not ok")
        }
        return response.json()
    }).then((hosts) => {
        hosts.forEach((host) => {
            const hostElement = document.createElement("div")
            hostElement.className = "host"

            const hostIpAddressElement = document.createElement("div")
            hostIpAddressElement.className = "address"
            hostIpAddressElement.textContent = host.ip_address

            const hostRiskElement = document.createElement("div")
            hostRiskElement.className = "risk"

            riskLevels.forEach((riskLevel) => {
                if (riskLevel.value == host.risk_level) {
                    hostRiskElement.textContent = `${riskLevel.name} (${riskLevel.value})`
                    return
                }
            })

            const hostDeleteElement = document.createElement("button")
            hostDeleteElement.className = "button delete"
            hostDeleteElement.textContent = "Delete"

            hostDeleteElement.addEventListener("click", () => {
                deleteHost(host.ip_address).then(() => {
                    listHosts()
                    refreshGraphs()
                })
            })

            hostElement.appendChild(hostIpAddressElement)
            hostElement.appendChild(hostRiskElement)
            hostElement.appendChild(hostDeleteElement)

            hostListElement.appendChild(hostElement)
        })
    })
}

function closeHostModal() {
    hostModalElement.classList.add("closed")
}

document.querySelector(".sidebar .controls .refresh").addEventListener("click", refreshGraphs)
document.querySelector(".sidebar .controls .simplify ").addEventListener("change", refreshGraphs)
document.querySelector(".sidebar .controls .reset").addEventListener("click", resetGraphs)
document.querySelector(".sidebar .controls .hosts").addEventListener("click", openHostModal)
document.querySelector(".modal .close").addEventListener("click", closeHostModal)
document.querySelector(".modal .create-host .button").addEventListener("click", createHost)

refreshGraphs()