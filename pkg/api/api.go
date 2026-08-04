package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Ali-Marandi/GoNetworkMonitor/pkg/alert"
	"github.com/Ali-Marandi/GoNetworkMonitor/pkg/capture"
	"github.com/Ali-Marandi/GoNetworkMonitor/pkg/config"
	"github.com/Ali-Marandi/GoNetworkMonitor/pkg/stats"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Server is the main HTTP/WebSocket API server.
type Server struct {
	cfg        *config.Config
	engine     *capture.Engine
	timeSeries *stats.TimeSeries
	connTable  *stats.ConnectionTable
	alertMgr   *alert.Manager
	alertChecker *alert.Checker
	mux        *http.ServeMux

	wsMu      sync.Mutex
	wsClients map[*websocket.Conn]bool
}

// NewServer creates a new API server.
func NewServer(cfg *config.Config) *Server {
	s := &Server{
		cfg:        cfg,
		timeSeries: stats.NewTimeSeries(3600),
		connTable:  stats.NewConnectionTable(10000),
		alertMgr:   alert.NewManager(500),
		wsClients:  make(map[*websocket.Conn]bool),
	}
	s.alertChecker = alert.NewChecker(s.alertMgr)
	s.mux = http.NewServeMux()
	s.registerRoutes()
	return s
}

// SetEngine sets the capture engine (called after creation).
func (s *Server) SetEngine(engine *capture.Engine) {
	s.engine = engine
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	// API routes
	s.mux.HandleFunc("/api/interfaces", s.handleInterfaces)
	s.mux.HandleFunc("/api/capture/start", s.handleCaptureStart)
	s.mux.HandleFunc("/api/capture/stop", s.handleCaptureStop)
	s.mux.HandleFunc("/api/capture/status", s.handleCaptureStatus)
	s.mux.HandleFunc("/api/stats/snapshot", s.handleStatsSnapshot)
	s.mux.HandleFunc("/api/stats/timeseries", s.handleTimeSeries)
	s.mux.HandleFunc("/api/stats/connections", s.handleConnections)
	s.mux.HandleFunc("/api/alerts", s.handleAlerts)
	s.mux.HandleFunc("/api/config", s.handleConfig)
	s.mux.HandleFunc("/api/ws", s.handleWebSocket)

	// Serve embedded web UI
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "./web/index.html")
}

func (s *Server) handleInterfaces(w http.ResponseWriter, r *http.Request) {
	ifaces, err := capture.ListInterfaces()
	if err != nil {
		errorResponse(w, 500, err.Error())
		return
	}
	jsonResponse(w, map[string]interface{}{"interfaces": ifaces})
}

func (s *Server) handleCaptureStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, 405, "method not allowed")
		return
	}

	var req struct {
		Interface string `json:"interface"`
		BPFFilter string `json:"bpf_filter"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	iface := req.Interface
	if iface == "" || iface == "auto" {
		ifaces, err := capture.ListInterfaces()
		if err != nil || len(ifaces) == 0 {
			errorResponse(w, 500, "no interfaces available")
			return
		}
		// Pick first non-loopback
		for _, i := range ifaces {
			if i.Name != "lo" {
				iface = i.Name
				break
			}
		}
		if iface == "" {
			iface = ifaces[0].Name
		}
	}

	cfg := s.cfg.Get()
	engine := capture.NewEngine(iface, cfg.SnapLen, cfg.Promiscuous, req.BPFFilter)
	s.engine = engine

	if err := engine.Start(); err != nil {
		errorResponse(w, 500, fmt.Sprintf("failed to start capture: %v", err))
		return
	}

	// Start time-series collector
	go s.collectTimeSeries()

	jsonResponse(w, map[string]interface{}{
		"status":    "started",
		"interface": iface,
	})
}

func (s *Server) handleCaptureStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, 405, "method not allowed")
		return
	}
	if s.engine != nil {
		s.engine.Stop()
	}
	jsonResponse(w, map[string]string{"status": "stopped"})
}

func (s *Server) handleCaptureStatus(w http.ResponseWriter, r *http.Request) {
	running := false
	iface := ""
	if s.engine != nil {
		running = s.engine.IsRunning()
		if running {
			iface = s.cfg.Get().Interface
		}
	}
	jsonResponse(w, map[string]interface{}{
		"running":   running,
		"interface": iface,
	})
}

func (s *Server) handleStatsSnapshot(w http.ResponseWriter, r *http.Request) {
	if s.engine == nil {
		jsonResponse(w, capture.StatsSnapshot{})
		return
	}
	snap := s.engine.Stats().Snapshot()
	jsonResponse(w, snap)
}

func (s *Server) handleTimeSeries(w http.ResponseWriter, r *http.Request) {
	points := s.timeSeries.GetLast(120)
	jsonResponse(w, map[string]interface{}{"points": points})
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	conns := s.connTable.GetAll()
	jsonResponse(w, map[string]interface{}{
		"connections": conns,
		"count":       len(conns),
	})
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := s.alertMgr.GetRecent(50)
	jsonResponse(w, map[string]interface{}{"alerts": alerts})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.cfg.Get()
		jsonResponse(w, cfg)
	case http.MethodPost:
		var newCfg config.Config
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			errorResponse(w, 400, "invalid JSON")
			return
		}
		s.cfg.Update(newCfg)
		jsonResponse(w, map[string]string{"status": "updated"})
	default:
		errorResponse(w, 405, "method not allowed")
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ws] upgrade error: %v", err)
		return
	}
	s.wsMu.Lock()
	s.wsClients[conn] = true
	s.wsMu.Unlock()

	defer func() {
		s.wsMu.Lock()
		delete(s.wsClients, conn)
		s.wsMu.Unlock()
		conn.Close()
	}()

	// Keep connection alive and send periodic updates
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.engine == nil {
				continue
			}
			snap := s.engine.Stats().Snapshot()
			msg, _ := json.Marshal(map[string]interface{}{
				"type":  "stats",
				"data":  snap,
				"time":  time.Now().Unix(),
			})
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}
}

// collectTimeSeries periodically samples stats into the time series.
func (s *Server) collectTimeSeries() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		if s.engine == nil || !s.engine.IsRunning() {
			return
		}
		<-ticker.C
		snap := s.engine.Stats().Snapshot()
		dp := stats.DataPoint{
			Timestamp:    time.Now(),
			PacketsPerSec: snap.PacketsPerSec,
			BytesPerSec:   snap.BytesPerSec,
			MbpsIn:        snap.BytesPerSec * 8 / 1_000_000,
		}
		s.timeSeries.Add(dp)

		// Check alerts
		cfg := s.cfg.Get()
		if cfg.Alerts.Enabled {
			s.alertChecker.CheckBandwidth(snap.BytesPerSec, cfg.Alerts.BandwidthMbps)
			s.alertChecker.CheckPPS(snap.PacketsPerSec, cfg.Alerts.PacketsPerSecond)
		}
	}
}

// Start begins listening on the configured address.
func (s *Server) Start() error {
	cfg := s.cfg.Get()
	addr := fmt.Sprintf("%s:%d", cfg.ListenAddr, cfg.Port)
	log.Printf("[api] Listening on http://%s", addr)
	return http.ListenAndServe(addr, s)
}
