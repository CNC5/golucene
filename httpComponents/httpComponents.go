package httpComponents

import (
	"encoding/json"
	"golucene/indexer"
	"io"
	"log"
	"net/http"
	"time"
)

func logRequest(request *http.Request, statusCode int) {
	log.Printf("%d - %s on %s", statusCode, request.Method, request.URL.Path)
}

func httpFinishResponse(code int, httpWriter http.ResponseWriter, request *http.Request) {
	marshaledJson, _ := json.Marshal(map[string]any{"status": http.StatusText(code)})
	io.WriteString(httpWriter, string(marshaledJson))
	logRequest(request, code)
}

type Metrics struct {
	// Server response metrics
	ResponseTime  map[string]float32
	RequestsCount map[string]int
}

func (metrics Metrics) asJson() []byte {
	metricsMap := map[string]any{
		"ResponseTime":  metrics.ResponseTime,
		"RequestsCount": metrics.RequestsCount}
	json, _ := json.Marshal(metricsMap)
	return json
}

type IndicesServer struct {
	DocumentIndices         map[string]indexer.DocumentIndex
	IndicesServerMuxPointer *http.ServeMux
	MetricsServerMuxPointer *http.ServeMux
	Metrics                 Metrics
}

func (server IndicesServer) updateResponseTime(route string, lastResponseTime time.Duration) {
	responseTimeAverage := server.Metrics.ResponseTime[route]
	requestCount := server.Metrics.RequestsCount[route]
	newResponseTimeAverage := // Include new response time in the average
		(responseTimeAverage*float32(requestCount) + float32(lastResponseTime)) /
			float32(requestCount+1)
	server.Metrics.RequestsCount[route] += 1
	server.Metrics.ResponseTime[route] = newResponseTimeAverage
}

func (server IndicesServer) getSearch(httpWriter http.ResponseWriter, request *http.Request) {
	startTime := time.Now()
	request.ParseForm()
	query := request.Form.Get("q")
	documentIndexName := request.Form.Get("index")
	if len(query) == 0 {
		httpFinishResponse(http.StatusNotFound, httpWriter, request)
		return
	}
	if len(documentIndexName) == 0 {
		httpFinishResponse(http.StatusBadRequest, httpWriter, request)
		return
	}
	documentIndex, doesExist := server.DocumentIndices[documentIndexName]
	if !doesExist {
		httpFinishResponse(http.StatusNotFound, httpWriter, request)
		return
	}
	data, _ := documentIndex.FindWord(query)
	marshaledData, _ := json.Marshal(map[string]any{"data": data, "status": http.StatusText(http.StatusOK)})
	io.WriteString(httpWriter, string(marshaledData))
	logRequest(request, 200)
	server.updateResponseTime("search", time.Since(startTime))
}

func (server IndicesServer) Search(httpWriter http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		server.getSearch(httpWriter, request)
	} else {
		io.WriteString(httpWriter, http.StatusText(http.StatusMethodNotAllowed))
		logRequest(request, http.StatusMethodNotAllowed)
	}
}

func (server IndicesServer) postLoad(httpWriter http.ResponseWriter, request *http.Request) {
	startTime := time.Now()
	request.ParseForm()
	documentIndexName := request.Form.Get("index")
	documentName := request.Form.Get("name")
	documentTextByteBuffer, _ := io.ReadAll(request.Body)
	documentText := string(documentTextByteBuffer)
	if len(documentIndexName) == 0 {
		httpFinishResponse(http.StatusBadRequest, httpWriter, request)
		return
	}
	if len(documentName) == 0 {
		httpFinishResponse(http.StatusBadRequest, httpWriter, request)
		return
	}
	if _, doesExist := server.DocumentIndices[documentIndexName]; !doesExist {
		server.DocumentIndices[documentIndexName] = indexer.CreateReverseIndex(documentIndexName)
	}
	server.DocumentIndices[documentIndexName].LoadDocument(documentName, documentText)
	httpFinishResponse(http.StatusOK, httpWriter, request)
	server.updateResponseTime("load", time.Since(startTime))
}

func (server IndicesServer) Load(httpWriter http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		server.postLoad(httpWriter, request)
	} else {
		io.WriteString(httpWriter, http.StatusText(http.StatusMethodNotAllowed))
		logRequest(request, http.StatusMethodNotAllowed)
	}
}

func (server IndicesServer) getHealthz(httpWriter http.ResponseWriter, request *http.Request) {
	startTime := time.Now()
	request.ParseForm()
	format := request.Form.Get("format")
	if len(format) == 0 {
		format = "json"
	}
	if format == "json" {
		io.WriteString(httpWriter, string(server.Metrics.asJson()))
	} else {
		httpFinishResponse(http.StatusBadRequest, httpWriter, request)
		return
	}
	server.updateResponseTime("healthz", time.Since(startTime))
}

func (server IndicesServer) Healthz(httpWriter http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		server.getHealthz(httpWriter, request)
	} else {
		httpFinishResponse(http.StatusMethodNotAllowed, httpWriter, request)
	}
}

func NewIndicesServer() IndicesServer {
	// Create a new instance of an indices server
	indicesServerMuxPointer := http.NewServeMux()
	metricsServerMuxPointer := http.NewServeMux()
	indicesServer := IndicesServer{
		DocumentIndices:         make(map[string]indexer.DocumentIndex),
		IndicesServerMuxPointer: indicesServerMuxPointer,
		MetricsServerMuxPointer: metricsServerMuxPointer,
		Metrics: Metrics{
			ResponseTime:  make(map[string]float32),
			RequestsCount: make(map[string]int)},
	}
	indicesServerMuxPointer.HandleFunc("/search", indicesServer.Search)
	indicesServerMuxPointer.HandleFunc("/load", indicesServer.Load)

	metricsServerMuxPointer.HandleFunc("/healthz", indicesServer.Healthz)
	return indicesServer
}
