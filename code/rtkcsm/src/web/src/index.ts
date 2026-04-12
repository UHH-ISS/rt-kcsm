import DirectedGraph from "graphology"
import FA2Layout from "graphology-layout-forceatlas2/worker"
import Sigma from "sigma"
import { EdgeCurvedArrowProgram, DEFAULT_EDGE_CURVATURE, indexParallelEdgesIndex } from "@sigma/edge-curve"
import { LoopEdgeProgram } from "./LoopEdge"
import { EdgeArrowProgram } from "sigma/rendering"

const graphContainer = document.querySelector(".graph") as HTMLDivElement
const graphWrapper = document.querySelector(".graph-wrapper") as HTMLDivElement
const graphIdElement = document.querySelector(".graph-id") as HTMLSpanElement
const graphCountElement = document.querySelector(".graph-count") as HTMLSpanElement
const hostModalElement = document.querySelector(".modal") as HTMLDialogElement
const hostListElement = document.querySelector(".modal .hosts") as HTMLDivElement
const newHostAddressElement = document.querySelector(".modal .create-host .address") as HTMLInputElement
const newHostRiskElement = document.querySelector(".modal .create-host .risk-level") as HTMLInputElement
const alertsListElement = document.querySelector(".alerts-list") as HTMLElement
const centerViewButton = document.querySelector(".center-button") as HTMLButtonElement
const alertPreviewElement = document.querySelector(".alert-preview") as HTMLDivElement
const alertStagesElement = document.querySelector(".alert-preview > .stages") as HTMLDivElement
const alertSourceElement = document.querySelector(".alert-preview > .src") as HTMLDivElement
const alertDestinationElement = document.querySelector(".alert-preview > .dst") as HTMLSpanElement
const alertMessageElement = document.querySelector(".alert-preview > .message") as HTMLSpanElement
const exportToOCSFButton = document.querySelector(".header .export-button") as HTMLButtonElement

const stageNamesToSymbols: { [stageName: string]: string } = {
    "Reconnaissance": "R",
    "Delivery Phase 1": "D1",
    "Delivery Phase 2": "D2",
    "Command&Control": "C2",
    "Lateral Movement": "L",
    "Discovery": "S",
    "Pivot": "P",
    "Exfiltration": "E",
    "Objectives": "O",
    "Execution": "X"
}

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

let currentGraphId: number | null = null
let focusedAlertId: string | null = null

const graph = new DirectedGraph({ multi: true, allowSelfLoops: true })
const renderer = new Sigma(graph, graphContainer, {
    edgeProgramClasses: {
        curvedArrow: EdgeCurvedArrowProgram,
        looped: LoopEdgeProgram,
        default: EdgeArrowProgram,
    },
    enableEdgeEvents: true,
    zIndex: true
})

const layout = new FA2Layout(graph, {
    settings: {
        gravity: 0.5,
        scalingRatio: 2,
        barnesHutOptimize: true,
        slowDown: 50,
    }
})

let edgeId: string | null = null

graphWrapper.addEventListener("mousemove", ({ clientX, clientY }) => {
    if (edgeId) {
        const attributes = graph.getEdgeAttributes(edgeId)
        alertPreviewElement.style.left = (clientX - graphWrapper.offsetLeft + 10) + "px"
        alertPreviewElement.style.top = (clientY - graphWrapper.offsetTop + 10) + "px"

        alertStagesElement.innerHTML = ""

        attributes.label.forEach((stageName: string) => {
            const stageElement = document.createElement("span")
            stageElement.className = "stage"
            stageElement.textContent = stageNamesToSymbols[stageName]

            alertStagesElement.appendChild(stageElement)
        })

        const source = graph.source(edgeId)
        const target = graph.target(edgeId)

        alertSourceElement.textContent = source
        alertDestinationElement.textContent = target

        alertMessageElement.textContent = attributes.message
        alertPreviewElement.classList.remove("hide")
    } else {
        alertPreviewElement.classList.add("hide")
    }
})

renderer.on("enterEdge", ({ edge }) => {
    edgeId = edge
    graph.setEdgeAttribute(edge, "highlight", true)
})

renderer.on("leaveEdge", ({ edge }) => {
    if (edge != focusedAlertId) {
        graph.removeEdgeAttribute(edge, "highlight")
    }
    edgeId = null
})

renderer.setSetting("edgeReducer", (_, attributes) => {
    if ("highlight" in attributes) {
        return { ...attributes, color: "#ff0000", zIndex: 10000000, forceLabel: true }
    } else {
        return { ...attributes }
    }
})

renderer.on("clickEdge", (event) => {
    const clickedEdgeId = event.edge

    let alert = document.querySelector(`[data-alert-id="${clickedEdgeId}"]`) as HTMLDivElement
    alertsListElement.scrollTo(0, alert.offsetTop + alert.clientHeight)
    let dateSeparator = (alert.parentElement as HTMLDetailsElement)
    dateSeparator.open = true
    alert.focus()
})

renderer.on("clickNode", (event) => {
    const clickedNodeId = event.node

    let alerts = document.querySelectorAll(`.alert`)

    alerts.forEach((alert) => {
        if (alert.getAttribute("data-alert-from") != clickedNodeId || alert.getAttribute("data-alert-to") != clickedNodeId) {
            alert.classList.add("hide")
        }
    })
})

let layoutIterations: number = 0

function renderEdges() {
    function getCurvature(index: number, maxIndex: number): number {
        if (maxIndex <= 0) throw new Error("Invalid maxIndex")
        if (index < 0) return -getCurvature(-index, maxIndex)
        const amplitude = 3.5
        const maxCurvature = amplitude * (1 - Math.exp(-maxIndex / amplitude)) * DEFAULT_EDGE_CURVATURE
        return (maxCurvature * index) / maxIndex
    }

    // Use dedicated helper to identify parallel edges:
    indexParallelEdgesIndex(graph, {
        edgeIndexAttribute: "parallelIndex",
        edgeMinIndexAttribute: "parallelMinIndex",
        edgeMaxIndexAttribute: "parallelMaxIndex",
    })

    // Adapt types and curvature of parallel edges for rendering:
    graph.forEachEdge(
        (
            edge: any,
            {
                parallelIndex,
                parallelMinIndex,
                parallelMaxIndex,
            }:
                | { parallelIndex: number; parallelMinIndex?: number; parallelMaxIndex: number }
                | { parallelIndex?: null; parallelMinIndex?: null; parallelMaxIndex?: null },
        ) => {
            if(graph.source(edge) === graph.target(edge)) {
                graph.setEdgeAttribute(edge, "type", "looped")
            } else {
                if (typeof parallelMinIndex === "number") {
                    graph.mergeEdgeAttributes(edge, {
                        type: parallelIndex ? "curvedArrow" : "arrow",
                        curvature: getCurvature(parallelIndex, parallelMaxIndex),
                    })
                } else if (typeof parallelIndex === "number") {
                    graph.mergeEdgeAttributes(edge, {
                        type: "curvedArrow",
                        curvature: getCurvature(parallelIndex, parallelMaxIndex),
                    })
                } else {
                    graph.setEdgeAttribute(edge, "type", "arrow")
                }
            }
        },
    )
}

function renderLoop() {
    if (layoutIterations < graph.order) {
        layoutIterations += 1
        renderer.refresh({ skipIndexation: true })
    } else {
        layout.stop()
    }

    requestAnimationFrame(renderLoop)
}
renderLoop()

function centerView() {
    renderer.getCamera().animate(
        { x: 0.5, y: 0.5, ratio: 1 },
        { duration: 300 }
    )
}

function openGraph(id: number) {
    centerView()
    focusedAlertId = null
    let previouslySelectedGraphElement = document.querySelector(`#graph-list-item-${currentGraphId}`)
    if (previouslySelectedGraphElement) {
        previouslySelectedGraphElement.classList.remove("active")
    }
    currentGraphId = id
    exportToOCSFButton.disabled = false

    if ("URLSearchParams" in window) {
        const url = new URL(window.location.href)
        url.searchParams.set("graphId", String(id))
        history.replaceState(null, "", url)
    }

    let selectedGraphElement = document.querySelector(`#graph-list-item-${id}`)
    if (selectedGraphElement) {
        selectedGraphElement.classList.add("active")
    }


    graphIdElement.textContent = String(id)
    alertsListElement.innerHTML = ""

    fetch(`/api/graphs/${encodeURIComponent(String(id))}`)
        .then(response => response.json())
        .then(data => {
            let nodesDict: {[id: string]: any} = {}
            graph.clear()
            layoutIterations = 0
            layout.start()

            let lastTimestamp: Date | null = null
            let alertContainer: HTMLElement | null = null

            data.relations.forEach((edge: { [key: string]: any }) => {
                if (!(edge.from in nodesDict)) {
                    graph.addNode(
                        edge.from,
                        {
                            label: edge.from,
                            color: edge.from_is_internal ? "blue" : "red",
                            size: edge.from_is_internal ? 30 : 10,
                            zIndex: edge.from_is_internal ? 2 : 1,
                            x: Math.random(),
                            y: Math.random(),
                        }
                    )
                    nodesDict[edge.from] = true
                }

                if (!(edge.to in nodesDict)) {
                    graph.addNode(
                        edge.to,
                        {
                            label: edge.to,
                            color: edge.to_is_internal ? "blue" : "red",
                            size: edge.to_is_internal ? 30 : 10,
                            zIndex: edge.to_is_internal ? 2 : 1,
                            x: Math.random(),
                            y: Math.random(),
                        }
                    )
                    nodesDict[edge.to] = true
                }
                    
                graph.addEdgeWithKey(
                    edge.id,
                    edge.from,
                    edge.to,
                    {
                        label: edge.confirmed_ukc_stages,
                        message: edge.cause,
                        size: edge.severity * 3,
                        showLabel: false,
                        zIndex: 1,
                    }
                )

                let timestamp = new Date(edge.timestamp)
                if (lastTimestamp == null || timestamp.toDateString() !== lastTimestamp.toDateString()) {
                    alertContainer = addDateSeparator(timestamp)
                }
                if (alertContainer) {
                    addAlertToList(edge, alertContainer)
                }
                lastTimestamp = timestamp
            })

            renderEdges()
        })
        .catch(error => console.error("Error fetching graph:", error))
}

const list: HTMLDivElement | null = document.querySelector(".sidebar .list")

function refreshGraphs() {
    if(list) {
        list.innerHTML = ""
        loadGraphs(0)
    }
}

function loadGraphs(page: number) {
    if(list) {
        fetch(`/api/graphs?page=${page}`)
        .then(response => {
            if (!response.ok) {
                throw new Error("Network response was not ok")
            }
            return response.json()
        })
        .then(data => {
            graphCountElement.textContent = data.count

            data.graphs.forEach((item: { [key: string]: any }) => {
                const graph = document.createElement("button")
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
                }

                graph.appendChild(id)
                graph.appendChild(edgeCount)


                list.appendChild(graph)
            })

            if (data.graphs.length > 0) {
                const loadMoreGraphs = document.createElement("button")
                loadMoreGraphs.className = "button"
                loadMoreGraphs.textContent = "Load more graphs..."
                loadMoreGraphs.addEventListener("click", () => {
                    loadGraphs(page + 1)
                    loadMoreGraphs.remove()
                })

                const scrollListener = function () {
                    if (list.scrollTop + list.clientHeight > list.scrollHeight - 200) {
                        list.removeEventListener("scroll", scrollListener)
                        loadGraphs(page + 1)
                    }
                }
                list.addEventListener("scroll", scrollListener)
            }

            if ("URLSearchParams" in window && page == 0) {
                const url = new URL(window.location.href)
                let graphId = url.searchParams.get("graphId")
                if (graphId) {
                    openGraph(Number.parseInt(graphId))
                }
            }
        })
    }
}

function addDateSeparator(timestamp: Date): HTMLElement {
    const dateElement = document.createElement("details")
    dateElement.className = "date-separator"
    dateElement.role = "lisitem"

    const dateLabelElement = document.createElement("summary")
    dateLabelElement.textContent = timestamp.toLocaleDateString()
    dateElement.appendChild(dateLabelElement)
    
    const alertsDetailsElement = document.createElement("div")
    alertsDetailsElement.role = "list"
    dateElement.appendChild(alertsDetailsElement)

    alertsListElement.appendChild(dateElement)

    return alertsDetailsElement
}

function addAlertToList(relation: any, parent: HTMLElement) {
    const alertElement = document.createElement("button")
    alertElement.className = "alert"
    alertElement.role = "listitem"
    alertElement.setAttribute("data-alert-id", relation.id)
    alertElement.setAttribute("data-alert-count", relation.count)
    alertElement.setAttribute("data-alert-from", relation.from)
    alertElement.setAttribute("data-alert-to", relation.to)

    const alertStagesElement = document.createElement("span")
    alertStagesElement.className = "stages"

    relation.confirmed_ukc_stages.forEach((stageName: string) => {
        const stageElement = document.createElement("span")
        stageElement.className = "stage"
        stageElement.textContent = stageNamesToSymbols[stageName]

        alertStagesElement.appendChild(stageElement)
    })

    const alertSrcLabelElement = document.createElement("span")
    alertSrcLabelElement.className = "src-label"
    alertSrcLabelElement.textContent = "src:"

    const alertSrcElement = document.createElement("span")
    alertSrcElement.className = "src"
    alertSrcElement.textContent = relation.from
    if (relation.from_is_internal) {
        alertSrcElement.classList.add("internal")
    }

    const alertDstLabelElement = document.createElement("span")
    alertDstLabelElement.className = "dst-label"
    alertDstLabelElement.textContent = "dst:"

    const alertDstElement = document.createElement("span")
    alertDstElement.className = "dst"
    alertDstElement.textContent = relation.to
    if (relation.to_is_internal) {
        alertDstElement.classList.add("internal")
    }

    const alertMessageElement = document.createElement("span")
    alertMessageElement.className = "message"
    alertMessageElement.textContent = relation.cause

    const detailsElement = document.createElement("details")
    detailsElement.className = "details"

    const detailSummaryElement = document.createElement("summary")
    detailSummaryElement.textContent = "Details"
    detailsElement.appendChild(detailSummaryElement)

    const attributes: { [key: string]: any } = {
        "Timestamp": new Date(relation.timestamp).toLocaleString(),
        "Count": relation.count,
        "Relevance": relation.computed_host_relevance.toFixed(2)
    }

    const attributeListItem = document.createElement("ul")
    attributeListItem.className = "attributes"

    for (let key in attributes) {
        const attributeElement = document.createElement("li")
        attributeElement.className = "attribute"

        const keyElement = document.createElement("span")
        keyElement.className = "key"
        keyElement.textContent = key

        const valueElement = document.createElement("span")
        valueElement.className = "value"
        valueElement.textContent = attributes[key]

        attributeElement.appendChild(keyElement)
        attributeElement.appendChild(valueElement)
        attributeListItem.appendChild(attributeElement)
    }
    detailsElement.appendChild(attributeListItem)

    alertElement.appendChild(alertStagesElement)
    alertElement.appendChild(alertSrcLabelElement)
    alertElement.appendChild(alertSrcElement)
    alertElement.appendChild(alertDstLabelElement)
    alertElement.appendChild(alertDstElement)
    alertElement.appendChild(alertMessageElement)
    alertElement.appendChild(detailsElement)

    alertElement.addEventListener("focusin", () => {
        if (focusedAlertId) {
            graph.removeEdgeAttribute(focusedAlertId, "highlight")
        }
        focusedAlertId = relation.id
        graph.setEdgeAttribute(relation.id, "highlight", true)
        centerView()
    })

    parent.appendChild(alertElement)
}

function resetGraphs() {
    fetch("/api/reset").then(() => {
        refreshGraphs()
    })
}

function openHostModal() {
    hostModalElement.showModal()
    listHosts()
}

function createHost(event: Event) {
    event.preventDefault()

    const ipAddress = newHostAddressElement.value
    const riskLevel = newHostRiskElement.value

    fetch("/api/hosts", {
        method: "POST",
        body: JSON.stringify({
            ip_address: ipAddress,
            risk_level: parseFloat(riskLevel)
        })
    }).then((response) => {
        if (response.ok) {
            listHosts()
            refreshGraphs()
            newHostAddressElement.value = ""
        }
    })
}

function deleteHost(ipAddress: string) {
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
        hosts.forEach((host: { [key: string]: any }) => {
            const hostElement = document.createElement("div")
            hostElement.className = "host"
            hostElement.role = "listitem"

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
    hostModalElement.close()
}

function downloadOCSFExport() {
    fetch(`/api/graphs/${currentGraphId}?format=ocsf`).then((response) => {
        return response.text()
    }).then((text) => {
        const blob = new Blob([text], { type: "application/json" })
        const url = URL.createObjectURL(blob)

        const anchor = document.createElement("a")
        anchor.href = url
        anchor.download = `incident_${currentGraphId}.ocsf`

        document.body.appendChild(anchor)
        anchor.click()
        document.body.removeChild(anchor)
    })
}

document.querySelector(".sidebar .controls .refresh")?.addEventListener("click", refreshGraphs)
document.querySelector(".sidebar .controls .reset")?.addEventListener("click", resetGraphs)
document.querySelector(".sidebar .controls .hosts")?.addEventListener("click", openHostModal)
document.querySelector(".modal .close")?.addEventListener("click", closeHostModal)
document.querySelector(".modal .create-host .button")?.addEventListener("click", createHost)
centerViewButton.addEventListener("click", centerView)
exportToOCSFButton.addEventListener("click", downloadOCSFExport)

refreshGraphs()