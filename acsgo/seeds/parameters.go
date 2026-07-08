package seeds

import (
	"log"

	"github.com/DWISSNET/acsgo/models"
	"github.com/DWISSNET/acsgo/services"
)

// AllParameters is the complete list of TR-069 standard + vendor-specific parameters.
var AllParameters = []models.Parameter{
	// ============================================================
	// Device Information (Standard TR-069)
	// ============================================================
	{Path: "Device.DeviceInfo.Manufacturer", Name: "Manufacturer", Description: "Device manufacturer name", Type: "string", Category: "Device Information", Subcategory: "Basic Info", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.DeviceInfo.ModelName", Name: "Model Name", Description: "Device model name", Type: "string", Category: "Device Information", Subcategory: "Basic Info", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.DeviceInfo.SerialNumber", Name: "Serial Number", Description: "Device serial number", Type: "string", Category: "Device Information", Subcategory: "Basic Info", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.DeviceInfo.HardwareVersion", Name: "Hardware Version", Description: "Hardware version", Type: "string", Category: "Device Information", Subcategory: "Version", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.DeviceInfo.SoftwareVersion", Name: "Software Version", Description: "Firmware/Software version", Type: "string", Category: "Device Information", Subcategory: "Version", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.DeviceInfo.ProvisioningCode", Name: "Provisioning Code", Description: "ISP provisioning code", Type: "string", Category: "Device Information", Subcategory: "Provisioning", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.DeviceInfo.UpTime", Name: "Uptime", Description: "Device uptime in seconds", Type: "unsignedInt", Category: "Device Information", Subcategory: "Status", Writable: false, DefaultValue: "0", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},

	// ============================================================
	// WAN Configuration (Standard TR-069)
	// ============================================================
	{Path: "Device.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Enable", Name: "WAN Enable", Description: "Enable/disable WAN connection", Type: "boolean", Category: "WAN Configuration", Subcategory: "Connection", Writable: true, DefaultValue: "true", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Username", Name: "PPP Username", Description: "PPPoE username for internet", Type: "string", Category: "WAN Configuration", Subcategory: "Credentials", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Password", Name: "PPP Password", Description: "PPPoE password", Type: "string", Category: "WAN Configuration", Subcategory: "Credentials", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.ExternalIPAddress", Name: "Public IP Address", Description: "WAN public IP address", Type: "string", Category: "WAN Configuration", Subcategory: "Status", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.SubnetMask", Name: "Subnet Mask", Description: "WAN subnet mask", Type: "string", Category: "WAN Configuration", Subcategory: "Status", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.DefaultGateway", Name: "Default Gateway", Description: "WAN default gateway", Type: "string", Category: "WAN Configuration", Subcategory: "Status", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},

	// ============================================================
	// WiFi 2.4GHz Configuration
	// ============================================================
	{Path: "Device.WiFi.SSID.1.SSID", Name: "WiFi 2.4GHz SSID", Description: "2.4GHz WiFi network name", Type: "string", Category: "WiFi Configuration", Subcategory: "2.4GHz", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WiFi.SSID.1.BSSID", Name: "WiFi 2.4GHz BSSID", Description: "2.4GHz BSSID/MAC address", Type: "string", Category: "WiFi Configuration", Subcategory: "2.4GHz", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WiFi.SSID.1.KeyPassphrase", Name: "WiFi 2.4GHz Password", Description: "2.4GHz WiFi password", Type: "string", Category: "WiFi Configuration", Subcategory: "2.4GHz", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WiFi.SSID.1.Enable", Name: "WiFi 2.4GHz Enable", Description: "Enable/disable 2.4GHz radio", Type: "boolean", Category: "WiFi Configuration", Subcategory: "2.4GHz", Writable: true, DefaultValue: "true", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WiFi.Radio.1.Channel", Name: "WiFi 2.4GHz Channel", Description: "2.4GHz channel number (1-14)", Type: "integer", Category: "WiFi Configuration", Subcategory: "2.4GHz", Writable: true, DefaultValue: "6", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WiFi.Radio.1.TransmitPower", Name: "WiFi 2.4GHz TX Power", Description: "2.4GHz transmit power (%)", Type: "integer", Category: "WiFi Configuration", Subcategory: "2.4GHz", Writable: true, DefaultValue: "100", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},

	// ============================================================
	// WiFi 5GHz Configuration
	// ============================================================
	{Path: "Device.WiFi.SSID.2.SSID", Name: "WiFi 5GHz SSID", Description: "5GHz WiFi network name", Type: "string", Category: "WiFi Configuration", Subcategory: "5GHz", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WiFi.SSID.2.KeyPassphrase", Name: "WiFi 5GHz Password", Description: "5GHz WiFi password", Type: "string", Category: "WiFi Configuration", Subcategory: "5GHz", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WiFi.SSID.2.Enable", Name: "WiFi 5GHz Enable", Description: "Enable/disable 5GHz radio", Type: "boolean", Category: "WiFi Configuration", Subcategory: "5GHz", Writable: true, DefaultValue: "true", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WiFi.Radio.2.Channel", Name: "WiFi 5GHz Channel", Description: "5GHz channel number", Type: "integer", Category: "WiFi Configuration", Subcategory: "5GHz", Writable: true, DefaultValue: "36", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.WiFi.Radio.2.TransmitPower", Name: "WiFi 5GHz TX Power", Description: "5GHz transmit power (%)", Type: "integer", Category: "WiFi Configuration", Subcategory: "5GHz", Writable: true, DefaultValue: "100", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},

	// ============================================================
	// GPON/Optical Status (Huawei/ZTE)
	// ============================================================
	{Path: "X_HW_GPON.OpticalPower.RXPower", Name: "GPON RX Power", Description: "GPON receive optical power (dBm)", Type: "string", Category: "GPON Status", Subcategory: "Optical", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"huawei", "zte"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: false},
	{Path: "X_HW_GPON.OpticalPower.TXPower", Name: "GPON TX Power", Description: "GPON transmit optical power (dBm)", Type: "string", Category: "GPON Status", Subcategory: "Optical", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"huawei", "zte"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: false},
	{Path: "X_HW_GPON.OpticalPower.Temperature", Name: "GPON Temperature", Description: "Optical module temperature (°C)", Type: "string", Category: "GPON Status", Subcategory: "Optical", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"huawei", "zte"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: false},
	{Path: "X_HW_GPON.OpticalPower.Voltage", Name: "GPON Voltage", Description: "Optical module voltage (V)", Type: "string", Category: "GPON Status", Subcategory: "Optical", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"huawei", "zte"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: false},
	{Path: "X_ZTE_GPON.Pon.RxPower", Name: "ZTE GPON RX Power", Description: "ZTE GPON receive power (dBm)", Type: "string", Category: "GPON Status", Subcategory: "Optical", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"zte"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: false},

	// ============================================================
	// LAN/DHCP Configuration
	// ============================================================
	{Path: "Device.LANDevice.1.LANHostConfigManagement.IPInterface.1.IPInterfaceIPAddress", Name: "LAN IP Address", Description: "LAN/gateway IP address", Type: "string", Category: "LAN Configuration", Subcategory: "IP", Writable: true, DefaultValue: "192.168.1.1", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.LANDevice.1.LANHostConfigManagement.IPInterface.1.IPInterfaceSubnetMask", Name: "LAN Subnet Mask", Description: "LAN subnet mask", Type: "string", Category: "LAN Configuration", Subcategory: "IP", Writable: true, DefaultValue: "255.255.255.0", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.LANDevice.1.LANHostConfigManagement.DHCPServerEnable", Name: "DHCP Server Enable", Description: "Enable/disable DHCP server", Type: "boolean", Category: "LAN Configuration", Subcategory: "DHCP", Writable: true, DefaultValue: "true", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.LANDevice.1.LANHostConfigManagement.MinAddress", Name: "DHCP Start IP", Description: "DHCP pool start address", Type: "string", Category: "LAN Configuration", Subcategory: "DHCP", Writable: true, DefaultValue: "192.168.1.100", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.LANDevice.1.LANHostConfigManagement.MaxAddress", Name: "DHCP End IP", Description: "DHCP pool end address", Type: "string", Category: "LAN Configuration", Subcategory: "DHCP", Writable: true, DefaultValue: "192.168.1.200", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},

	// ============================================================
	// VoIP/SIP Configuration
	// ============================================================
	{Path: "Device.Services.VoiceService.1.VoiceProfile.1.SIP.ProxyServer", Name: "SIP Proxy Server", Description: "SIP proxy server address", Type: "string", Category: "VoIP Configuration", Subcategory: "SIP", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.Services.VoiceService.1.VoiceProfile.1.SIP.ProxyServerPort", Name: "SIP Port", Description: "SIP proxy server port", Type: "integer", Category: "VoIP Configuration", Subcategory: "SIP", Writable: true, DefaultValue: "5060", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.Services.VoiceService.1.VoiceProfile.1.Line.1.SIP.AuthUserName", Name: "SIP Username", Description: "SIP authentication username", Type: "string", Category: "VoIP Configuration", Subcategory: "Auth", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.Services.VoiceService.1.VoiceProfile.1.Line.1.SIP.AuthPassword", Name: "SIP Password", Description: "SIP authentication password", Type: "string", Category: "VoIP Configuration", Subcategory: "Auth", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},

	// ============================================================
	// IPTV/Multicast Configuration
	// ============================================================
	{Path: "Device.X_HUAWEI_IPTV.Enable", Name: "IPTV Enable", Description: "Enable/disable IPTV", Type: "boolean", Category: "IPTV Configuration", Subcategory: "General", Writable: true, DefaultValue: "false", SupportedVendors: models.JSONStrings{"huawei"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: false},
	{Path: "Device.X_HUAWEI_IPTV.VLANId", Name: "IPTV VLAN ID", Description: "IPTV VLAN identifier", Type: "integer", Category: "IPTV Configuration", Subcategory: "VLAN", Writable: true, DefaultValue: "4000", SupportedVendors: models.JSONStrings{"huawei", "zte"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: false},
	{Path: "Device.Routing.Router.1.IPv4Forwarding.1.Enable", Name: "IPv4 Forwarding", Description: "Enable IP forwarding/routing", Type: "boolean", Category: "Network Configuration", Subcategory: "Routing", Writable: true, DefaultValue: "true", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},

	// ============================================================
	// Management/TR-069
	// ============================================================
	{Path: "Device.ManagementServer.URL", Name: "ACS URL", Description: "TR-069 ACS server URL", Type: "string", Category: "Management", Subcategory: "TR-069", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.ManagementServer.Username", Name: "ACS Username", Description: "TR-069 authentication username", Type: "string", Category: "Management", Subcategory: "TR-069", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.ManagementServer.Password", Name: "ACS Password", Description: "TR-069 authentication password", Type: "string", Category: "Management", Subcategory: "TR-069", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.ManagementServer.PeriodicInformEnable", Name: "Periodic Inform Enable", Description: "Enable periodic Inform messages", Type: "boolean", Category: "Management", Subcategory: "TR-069", Writable: true, DefaultValue: "true", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.ManagementServer.PeriodicInformInterval", Name: "Inform Interval", Description: "Periodic Inform interval (seconds)", Type: "unsignedInt", Category: "Management", Subcategory: "TR-069", Writable: true, DefaultValue: "300", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},

	// ============================================================
	// Firmware Update
	// ============================================================
	{Path: "Device.DeviceInfo.X_HW_FirmwareURL", Name: "Firmware URL (Huawei)", Description: "Huawei firmware update URL", Type: "string", Category: "Firmware", Subcategory: "Update", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"huawei"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: false},
	{Path: "Device.DeviceInfo.X_ZTE_COM_FirmwareURL", Name: "Firmware URL (ZTE)", Description: "ZTE firmware update URL", Type: "string", Category: "Firmware", Subcategory: "Update", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"zte"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: false},

	// ============================================================
	// NAT/Port Forwarding
	// ============================================================
	{Path: "Device.NAT.PortMapping.1.Enable", Name: "Port Mapping Enable", Description: "Enable port forwarding entry", Type: "boolean", Category: "NAT Configuration", Subcategory: "Port Mapping", Writable: true, DefaultValue: "false", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.NAT.PortMapping.1.ExternalPort", Name: "External Port", Description: "External port for NAT mapping", Type: "integer", Category: "NAT Configuration", Subcategory: "Port Mapping", Writable: true, DefaultValue: "0", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.NAT.PortMapping.1.InternalClient", Name: "Internal Client IP", Description: "Internal client IP for NAT mapping", Type: "string", Category: "NAT Configuration", Subcategory: "Port Mapping", Writable: true, DefaultValue: "", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},
	{Path: "Device.NAT.PortMapping.1.InternalPort", Name: "Internal Port", Description: "Internal port for NAT mapping", Type: "integer", Category: "NAT Configuration", Subcategory: "Port Mapping", Writable: true, DefaultValue: "0", SupportedVendors: models.JSONStrings{"All"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: true},

	// ============================================================
	// MikroTik-specific
	// ============================================================
	{Path: "Device.X_MIKROTIK.RouterOS.Version", Name: "RouterOS Version", Description: "MikroTik RouterOS version", Type: "string", Category: "Device Information", Subcategory: "Version", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"mikrotik"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: false},
	{Path: "Device.X_MIKROTIK.System.Uptime", Name: "System Uptime", Description: "MikroTik system uptime", Type: "string", Category: "Device Information", Subcategory: "Status", Writable: false, DefaultValue: "", SupportedVendors: models.JSONStrings{"mikrotik"}, SupportedModels: models.JSONStrings{"All"}, IsStandardTR069: false},
}

// SeedParameters inserts all parameter templates into the database.
func SeedParameters(svc *services.ParameterService) error {
	existing, err := svc.Count()
	if err != nil {
		return err
	}
	if existing > 0 {
		log.Printf("📋 Parameters already seeded (%d entries), skipping", existing)
		return nil
	}

	for _, p := range AllParameters {
		if err := svc.Upsert(&p); err != nil {
			log.Printf("⚠️  Failed to seed parameter %s: %v", p.Path, err)
		}
	}
	log.Printf("✅ Seeded %d TR-069 parameters", len(AllParameters))
	return nil
}
