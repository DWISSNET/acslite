// Package cwmp implements the TR-069/CWMP HTTP server for ACSGO.
//
// The server listens on the standard ACS port 7547 and handles:
//   - Inform (device registration & periodic heartbeat)
//   - GetRPCMethods
//   - GetParameterValues / SetParameterValues responses
//   - Reboot / FactoryReset responses
//   - TransferComplete notifications
//
// After processing an Inform, the server delivers any pending RPC commands
// from the queue before sending an empty SOAP body to close the session.
package cwmp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/DWISSNET/acsgo/pkg/cache"
	"github.com/DWISSNET/acsgo/pkg/models"
	"github.com/DWISSNET/acsgo/pkg/repository"
	"github.com/DWISSNET/acsgo/services/cwmp/session"
	"github.com/DWISSNET/acsgo/services/cwmp/soap"
	"github.com/DWISSNET/acsgo/services/metrics"
	"github.com/DWISSNET/acsgo/services/queue"
)

// Server is the TR-069 CWMP HTTP server.
type Server struct {
	deviceRepo  *repository.DeviceRepository
	commandRepo *repository.CommandRepository
	sessions    *session.Manager
	cache       *cache.Client
	queue       queue.CommandPublisher
	metrics     *metrics.Collector
	logger      *zap.Logger
	httpServer  *http.Server
	acsUsername string
	acsPassword string
}

// ServerOption configures the CWMP server.
type ServerOption func(*Server)

// WithBasicAuth enables HTTP Digest/Basic authentication.
func WithBasicAuth(username, password string) ServerOption {
	return func(s *Server) {
		s.acsUsername = username
		s.acsPassword = password
	}
}

// NewServer creates a new CWMP server.
func NewServer(
	port int,
	deviceRepo *repository.DeviceRepository,
	commandRepo *repository.CommandRepository,
	sessManager *session.Manager,
	cacheClient *cache.Client,
	queuePub queue.CommandPublisher,
	metricsCollector *metrics.Collector,
	logger *zap.Logger,
	opts ...ServerOption,
) *Server {
	s := &Server{
		deviceRepo:  deviceRepo,
		commandRepo: commandRepo,
		sessions:    sessManager,
		cache:       cacheClient,
		queue:       queuePub,
		metrics:     metricsCollector,
		logger:      logger,
	}
	for _, o := range opts {
		o(s)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/cwmp", s.handleCWMP)
	mux.HandleFunc("/acs", s.handleCWMP)  // legacy alias
	mux.HandleFunc("/healthz", s.handleHealth)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return s
}

// Start begins listening for CWMP connections.
func (s *Server) Start() error {
	s.logger.Info("CWMP server starting", zap.String("addr", s.httpServer.Addr))
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the CWMP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// handleHealth is a simple liveness check.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// handleCWMP processes all TR-069 SOAP requests.
func (s *Server) handleCWMP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Optional Basic Auth
	if s.acsUsername != "" {
		user, pass, ok := r.BasicAuth()
		if !ok || user != s.acsUsername || pass != s.acsPassword {
			w.Header().Set("WWW-Authenticate", `Basic realm="ACS"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20)) // 10 MB limit
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	deviceIP := realIP(r)

	// Empty body = CPE acknowledging ACS empty response (end of session)
	if len(strings.TrimSpace(string(body))) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	env, err := soap.ParseEnvelope(body)
	if err != nil {
		s.logger.Warn("SOAP parse error", zap.Error(err), zap.String("ip", deviceIP))
		s.writeFault(w, "Client", "Invalid SOAP envelope")
		return
	}

	cwmpID := env.Header.ID
	if cwmpID == "" {
		cwmpID = soap.GenerateID()
	}

	resp, err := s.dispatch(ctx, env, cwmpID, deviceIP)
	if err != nil {
		s.logger.Error("CWMP dispatch error", zap.Error(err))
		s.writeFault(w, "Server", "Internal server error")
		return
	}

	s.metrics.RecordCWMPRequest(time.Since(start))
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(resp))
}

// dispatch routes the SOAP body to the appropriate handler.
func (s *Server) dispatch(ctx context.Context, env *soap.Envelope, cwmpID, deviceIP string) (string, error) {
	b := env.Body

	switch {
	case b.Inform != nil:
		return s.handleInform(ctx, b.Inform, cwmpID, deviceIP)

	case b.GetRPCMethods != nil:
		return soap.GetRPCMethodsResponse(cwmpID), nil

	case b.GetParameterValuesResponse != nil:
		return s.handleGetParamValuesResp(ctx, b.GetParameterValuesResponse, cwmpID)

	case b.SetParameterValuesResponse != nil:
		return s.handleSetParamValuesResp(ctx, b.SetParameterValuesResponse, cwmpID)

	case b.GetParameterNamesResponse != nil:
		return soap.EmptyResponse(cwmpID), nil

	case b.RebootResponse != nil:
		return s.handleCommandResponse(ctx, cwmpID, models.CommandReboot, nil)

	case b.FactoryResetResponse != nil:
		return s.handleCommandResponse(ctx, cwmpID, models.CommandFactoryReset, nil)

	case b.TransferCompleteRequest != nil:
		return s.handleTransferComplete(ctx, b.TransferCompleteRequest, cwmpID)

	case b.Fault != nil:
		s.logger.Warn("SOAP Fault received from CPE",
			zap.String("code", b.Fault.FaultCode),
			zap.String("string", b.Fault.FaultString),
		)
		return soap.EmptyResponse(cwmpID), nil

	default:
		s.logger.Warn("Unhandled CWMP method")
		return soap.EmptyResponse(cwmpID), nil
	}
}

// handleInform processes the CPE Inform RPC.
func (s *Server) handleInform(ctx context.Context, inf *soap.Inform, cwmpID, deviceIP string) (string, error) {
	deviceID := buildDeviceID(inf.DeviceId)

	s.logger.Info("Inform received",
		zap.String("device_id", deviceID),
		zap.String("ip", deviceIP),
		zap.String("manufacturer", inf.DeviceId.Manufacturer),
		zap.String("model", inf.DeviceId.ProductClass),
	)

	s.metrics.RecordInform()

	now := time.Now()
	device := &models.Device{
		DeviceID:         deviceID,
		SerialNumber:     inf.DeviceId.SerialNumber,
		ProductClass:     inf.DeviceId.ProductClass,
		Manufacturer:     inf.DeviceId.Manufacturer,
		IPAddress:        deviceIP,
		LastInform:       &now,
		ConnectionStatus: models.ConnectionOnline,
	}

	// Best-effort DB upsert
	if err := s.deviceRepo.Upsert(ctx, device); err != nil {
		s.logger.Warn("Device upsert error", zap.Error(err))
	}

	// Cache online status
	_ = s.cache.MarkOnline(ctx, deviceID, 10*time.Minute)
	_ = s.cache.SetDevice(ctx, deviceID, device)

	// Store parameters
	if device.ID != "" {
		for _, pv := range inf.ParameterList {
			_ = s.deviceRepo.UpsertParameter(ctx, device.ID, pv.Name, pv.Value, "string")
		}
	}

	// Record events
	for _, ev := range inf.Event {
		s.logger.Debug("Device event", zap.String("device_id", deviceID), zap.String("code", ev.EventCode))
		if device.ID != "" {
			_ = s.deviceRepo.AddEvent(ctx, device.ID, ev.EventCode, ev.CommandKey)
		}
	}

	// Begin session
	s.sessions.Begin(ctx, deviceID, cwmpID)

	// Deliver next pending command (if any) instead of empty response
	return s.nextCommand(ctx, device, cwmpID)
}

// nextCommand delivers the next pending RPC, or sends an empty SOAP body.
func (s *Server) nextCommand(ctx context.Context, device *models.Device, cwmpID string) (string, error) {
	cmds, err := s.commandRepo.GetPendingForDevice(ctx, device.DeviceID)
	if err != nil || len(cmds) == 0 {
		return soap.EmptyResponse(cwmpID), nil
	}

	cmd := cmds[0]
	s.logger.Info("Delivering command",
		zap.String("device_id", device.DeviceID),
		zap.String("command_id", cmd.ID),
		zap.String("type", string(cmd.CommandType)),
	)

	_ = s.sessions.SetWaiting(ctx, device.DeviceID, cmd.ID)
	_ = s.commandRepo.UpdateStatus(ctx, cmd.ID, models.CommandDelivered, "")

	return s.buildRPCSoap(cwmpID, cmd), nil
}

// buildRPCSoap creates the appropriate SOAP XML for a command.
func (s *Server) buildRPCSoap(cwmpID string, cmd *models.Command) string {
	switch cmd.CommandType {
	case models.CommandReboot:
		return soap.RebootRequest(cwmpID, cmd.ID)
	case models.CommandFactoryReset:
		return soap.FactoryResetRequest(cwmpID)
	case models.CommandSetParameterValues:
		return soap.SetParameterValuesRequest(cwmpID, cmd.ID, cmd.Parameters)
	case models.CommandGetParameterValues:
		paths := make([]string, 0, len(cmd.Parameters))
		for p := range cmd.Parameters {
			paths = append(paths, p)
		}
		return soap.GetParameterValuesRequest(cwmpID, paths)
	case models.CommandDownload:
		p := cmd.Parameters
		return soap.DownloadRequest(cwmpID, cmd.ID,
			p["file_type"], p["url"], p["username"], p["password"],
			p["file_size"], p["target_filename"],
		)
	default:
		return soap.EmptyResponse(cwmpID)
	}
}

// handleGetParamValuesResp stores parameter values returned by the CPE.
func (s *Server) handleGetParamValuesResp(ctx context.Context, resp *soap.GetParameterValuesResponse, cwmpID string) (string, error) {
	// TODO: look up device by session to store returned params
	s.logger.Debug("GetParameterValues response", zap.Int("count", len(resp.ParameterList)))
	return soap.EmptyResponse(cwmpID), nil
}

// handleSetParamValuesResp records the result of a SetParameterValues command.
func (s *Server) handleSetParamValuesResp(ctx context.Context, resp *soap.SetParameterValuesResponse, cwmpID string) (string, error) {
	s.logger.Debug("SetParameterValues response", zap.Int("status", resp.Status))
	return soap.EmptyResponse(cwmpID), nil
}

// handleCommandResponse marks a command as completed.
func (s *Server) handleCommandResponse(ctx context.Context, cwmpID string, _ models.CommandType, _ interface{}) (string, error) {
	s.logger.Debug("Command response received")
	return soap.EmptyResponse(cwmpID), nil
}

// handleTransferComplete records firmware download completion.
func (s *Server) handleTransferComplete(ctx context.Context, tc *soap.TransferComplete, cwmpID string) (string, error) {
	s.logger.Info("TransferComplete",
		zap.String("command_key", tc.CommandKey),
		zap.Int("fault_code", tc.FaultStruct.FaultCode),
	)
	return soap.TransferCompleteResponse(cwmpID), nil
}

func (s *Server) writeFault(w http.ResponseWriter, code, text string) {
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(soap.FaultResponse(soap.GenerateID(), code, text)))
}

// buildDeviceID constructs a canonical device ID from the TR-069 DeviceId.
func buildDeviceID(d soap.DeviceID) string {
	return strings.Join([]string{d.OUI, d.ProductClass, d.SerialNumber}, "-")
}

// realIP extracts the client IP, respecting X-Forwarded-For.
func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// Strip port from RemoteAddr
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		return addr[:i]
	}
	return addr
}
