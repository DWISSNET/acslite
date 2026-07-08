package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/DWISSNET/acsgo/models"
	"github.com/DWISSNET/acsgo/pkg"
)

// CWMPService handles TR-069 CWMP session logic
type CWMPService struct {
	devices *DeviceService
	rpc     *RPCService
	hub     *WSHub
}

func NewCWMPService(devices *DeviceService, rpc *RPCService, hub *WSHub) *CWMPService {
	return &CWMPService{devices: devices, rpc: rpc, hub: hub}
}

// HandleRequest is the main entry point for POST /cwmp requests.
func (s *CWMPService) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Read raw body (SOAP XML)
	body, err := io.ReadAll(io.LimitReader(r.Body, 10*1024*1024))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Empty body → device polling for pending RPC (HTTP 204)
	if len(strings.TrimSpace(string(body))) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	env, err := pkg.ParseEnvelope(body)
	if err != nil {
		log.Printf("⚠️  CWMP parse error from %s: %v", r.RemoteAddr, err)
		http.Error(w, "invalid SOAP", http.StatusBadRequest)
		return
	}

	// Sanitize the CWMP session ID to prevent XSS/injection in XML responses
	id := sanitizeCWMPID(env.Header.ID.Value)
	if id == "" {
		id = "1"
	}

	ipAddress := extractIP(r)

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")

	b := env.Body
	switch {
	case b.Inform != nil:
		s.handleInform(w, b.Inform, id, ipAddress)

	case b.GetRPCMethodsResponse != nil:
		// Device responded to GetRPCMethods → just acknowledge
		w.Write([]byte(pkg.BuildEmptyResponse(id)))

	case b.SetParameterValuesResponse != nil:
		s.handleSetParamResponse(w, b.SetParameterValuesResponse, id, ipAddress)

	case b.GetParameterValuesResponse != nil:
		w.Write([]byte(pkg.BuildEmptyResponse(id)))

	case b.TransferComplete != nil:
		w.Write([]byte(pkg.BuildEmptyResponse(id)))

	default:
		log.Printf("⚠️  Unhandled CWMP body from %s", ipAddress)
		w.Write([]byte(pkg.BuildEmptyResponse(id)))
	}
}

// --------------------------------------------------------------------------
// Inform handler
// --------------------------------------------------------------------------

func (s *CWMPService) handleInform(w http.ResponseWriter, inform *pkg.CWMPInform, id, ipAddress string) {
	serial := inform.DeviceId.SerialNumber
	if serial == "" {
		serial = "UNKNOWN-" + id
	}

	log.Printf("📡 INFORM from %s (serial=%s)", ipAddress, serial)

	// Build/update device record
	now := time.Now()
	vendor := pkg.DetectVendorFromManufacturer(inform.DeviceId.Manufacturer)
	modem := DetectModel(inform.DeviceId.Manufacturer, inform.DeviceId.ProductClass, serial)

	// Extract parameters from Inform ParameterList
	params := models.JSONMap{}
	for _, pv := range inform.ParameterList.ParameterValueStruct {
		params[pv.Name] = pv.Value
	}

	// Try to get existing device
	existing, _ := s.devices.GetBySerial(serial)
	d := &models.Device{
		SerialNumber:     serial,
		DeviceID:         pkg.BuildDeviceID(inform.DeviceId.OUI, serial),
		Manufacturer:     vendor,
		ModelName:        modem.Model,
		ProductClass:     inform.DeviceId.ProductClass,
		IPAddress:        ipAddress,
		MACAddress:       pkg.ExtractMACFromSerial(serial),
		ConnectionStatus: "online",
		LastInform:       &now,
	}
	if existing != nil {
		d.ID = existing.ID
		d.FirstSeen = existing.FirstSeen
		d.ISP = existing.ISP
		d.CustomerName = existing.CustomerName
		d.CustomerPhone = existing.CustomerPhone
		d.CustomerEmail = existing.CustomerEmail
		d.CustomerPackage = existing.CustomerPackage
		d.LocationCity = existing.LocationCity
		d.LocationProvince = existing.LocationProvince
		// Merge parameters
		for k, v := range existing.Parameters {
			if _, ok := params[k]; !ok {
				params[k] = v
			}
		}
	}
	d.Parameters = params

	if err := s.devices.Upsert(d); err != nil {
		log.Printf("❌ DB upsert error for %s: %v", serial, err)
	}

	// Detect bootstrap / initial registration
	isBootstrap := false
	for _, ev := range inform.Event.EventStruct {
		if ev.EventCode == "0 BOOTSTRAP" || ev.EventCode == "1 BOOT" {
			isBootstrap = true
			break
		}
	}

	if isBootstrap {
		s.devices.LogEvent(serial, "bootstrap", "Device bootstrap/first connection")
		log.Printf("🚀 Bootstrap detected for %s", serial)
	} else {
		s.devices.LogEvent(serial, "inform", "Device sent periodic Inform")
	}

	// Broadcast to WebSocket dashboard
	s.hub.Broadcast(WSMessage{
		Event: "device:online",
		Data: map[string]interface{}{
			"serial":       serial,
			"ip":           ipAddress,
			"manufacturer": vendor,
			"model":        modem.Model,
			"timestamp":    now,
		},
	})

	// Check for pending RPC commands
	pending, err := s.rpc.NextPending(serial)
	if err != nil {
		log.Printf("⚠️  RPC queue error: %v", err)
	}

	if pending != nil {
		log.Printf("📤 Dispatching %s to %s (rpc_id=%s)", pending.CommandType, serial, pending.ID)
		s.rpc.MarkSent(pending.ID)
		response := buildRPCResponse(id, pending)
		w.Write([]byte(response))
		s.devices.LogEvent(serial, "rpc_sent", "Command "+pending.CommandType+" dispatched")
		return
	}

	// No pending RPC → plain InformResponse
	w.Write([]byte(pkg.BuildInformResponse(id)))
}

func (s *CWMPService) handleSetParamResponse(w http.ResponseWriter, resp *pkg.SetParameterValuesResponse, id, ipAddress string) {
	if resp.Status == 0 {
		log.Printf("✅ SetParameterValues success from %s", ipAddress)
	} else {
		log.Printf("⚠️  SetParameterValues failed (status=%d) from %s", resp.Status, ipAddress)
	}
	w.Write([]byte(pkg.BuildEmptyResponse(id)))
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func buildRPCResponse(id string, cmd *models.RPCQueue) string {
	switch cmd.CommandType {
	case "Reboot":
		return pkg.BuildRebootRequest(id, cmd.ID)
	case "FactoryReset":
		return pkg.BuildFactoryResetRequest(id)
	case "SetParameterValues":
		params := make(map[string]string)
		for k, v := range cmd.Parameters {
			params[k] = strings.TrimSpace(strings.Trim(fmt.Sprintf("%v", v), `"`))
		}
		return pkg.BuildSetParameterValuesRequest(id, cmd.ID, params)
	default:
		return pkg.BuildEmptyResponse(id)
	}
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return strings.Trim(ip, "[]")
}

// sanitizeCWMPID ensures the session ID (device-supplied) contains only safe
// characters before it is embedded in XML responses, preventing XSS/injection.
func sanitizeCWMPID(id string) string {
	// Keep only alphanumeric chars and a small set of safe punctuation
	var out strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			out.WriteRune(r)
		}
	}
	return out.String()
}
