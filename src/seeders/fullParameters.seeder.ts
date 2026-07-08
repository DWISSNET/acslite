import ParameterTemplate from '../models/ParameterTemplate';

// SEMUA PARAMETER STANDAR TR-069 + VENDOR-SPECIFIC
export const ALL_PARAMETERS = [
  // ============================================
  // STANDAR TR-069 - Device Info
  // ============================================
  {
    path: "Device.DeviceInfo.Manufacturer",
    name: "Manufacturer",
    description: "Device manufacturer name",
    type: "string",
    writable: false,
    defaultValue: "",
    category: "Device Information",
    subcategory: "Basic Info",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.DeviceInfo.ModelName",
    name: "Model Name",
    description: "Device model name",
    type: "string",
    writable: false,
    defaultValue: "",
    category: "Device Information",
    subcategory: "Basic Info",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.DeviceInfo.SerialNumber",
    name: "Serial Number",
    description: "Device serial number",
    type: "string",
    writable: false,
    defaultValue: "",
    category: "Device Information",
    subcategory: "Basic Info",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.DeviceInfo.HardwareVersion",
    name: "Hardware Version",
    description: "Hardware version",
    type: "string",
    writable: false,
    defaultValue: "",
    category: "Device Information",
    subcategory: "Version",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.DeviceInfo.SoftwareVersion",
    name: "Software Version",
    description: "Firmware/Software version",
    type: "string",
    writable: false,
    defaultValue: "",
    category: "Device Information",
    subcategory: "Version",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },

  // ============================================
  // WAN Settings - Semua Modem
  // ============================================
  {
    path: "Device.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Enable",
    name: "WAN Enable",
    description: "Enable/disable WAN connection",
    type: "boolean",
    writable: true,
    defaultValue: true,
    category: "WAN Configuration",
    subcategory: "Connection",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Username",
    name: "PPP Username",
    description: "PPPoe username for internet",
    type: "string",
    writable: true,
    defaultValue: "",
    category: "WAN Configuration",
    subcategory: "Credentials",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Password",
    name: "PPP Password",
    description: "PPPoe password",
    type: "string",
    writable: true,
    defaultValue: "",
    category: "WAN Configuration",
    subcategory: "Credentials",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.ExternalIPAddress",
    name: "Public IP Address",
    description: "WAN/External IP address",
    type: "string",
    writable: false,
    defaultValue: "",
    category: "WAN Configuration",
    subcategory: "Status",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },

  // ============================================
  // WiFi Settings - 2.4GHz & 5GHz
  // ============================================
  // WiFi 2.4GHz
  {
    path: "Device.WiFi.SSID.1.SSID",
    name: "WiFi 2.4GHz SSID",
    description: "WiFi network name 2.4GHz",
    type: "string",
    writable: true,
    defaultValue: "INDIHOME-2G",
    category: "WiFi Configuration",
    subcategory: "2.4GHz",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.WiFi.SSID.1.BSSID",
    name: "WiFi 2.4GHz BSSID",
    description: "WiFi MAC address 2.4GHz",
    type: "string",
    writable: false,
    defaultValue: "",
    category: "WiFi Configuration",
    subcategory: "2.4GHz",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.WiFi.SSID.1.KeyPassphrase",
    name: "WiFi 2.4GHz Password",
    description: "WiFi password 2.4GHz",
    type: "string",
    writable: true,
    defaultValue: "",
    category: "WiFi Configuration",
    subcategory: "2.4GHz",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.WiFi.SSID.1.Enable",
    name: "WiFi 2.4GHz Enable",
    description: "Enable/disable 2.4GHz WiFi",
    type: "boolean",
    writable: true,
    defaultValue: true,
    category: "WiFi Configuration",
    subcategory: "2.4GHz",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.WiFi.Radio.1.Channel",
    name: "WiFi 2.4GHz Channel",
    description: "WiFi channel 2.4GHz (1-14)",
    type: "integer",
    writable: true,
    defaultValue: 6,
    minValue: 1,
    maxValue: 14,
    category: "WiFi Configuration",
    subcategory: "2.4GHz",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true,
    unit: "channel"
  },
  {
    path: "Device.WiFi.Radio.1.TransmitPower",
    name: "WiFi 2.4GHz TX Power",
    description: "Transmit power 2.4GHz",
    type: "integer",
    writable: true,
    defaultValue: 100,
    minValue: 0,
    maxValue: 100,
    category: "WiFi Configuration",
    subcategory: "2.4GHz",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true,
    unit: "%"
  },

  // WiFi 5GHz
  {
    path: "Device.WiFi.SSID.2.SSID",
    name: "WiFi 5GHz SSID",
    description: "WiFi network name 5GHz",
    type: "string",
    writable: true,
    defaultValue: "INDIHOME-5G",
    category: "WiFi Configuration",
    subcategory: "5GHz",
    supportedManufacturers: ["All"],
    supportedModels: ["HG8245H, HG8546M, EG8141A5, F609, F670L"],
    isStandardTR069: true
  },
  {
    path: "Device.WiFi.SSID.2.KeyPassphrase",
    name: "WiFi 5GHz Password",
    description: "WiFi password 5GHz",
    type: "string",
    writable: true,
    defaultValue: "",
    category: "WiFi Configuration",
    subcategory: "5GHz",
    supportedManufacturers: ["All"],
    supportedModels: ["HG8245H, HG8546M, EG8141A5, F609, F670L"],
    isStandardTR069: true
  },
  {
    path: "Device.WiFi.SSID.2.Enable",
    name: "WiFi 5GHz Enable",
    description: "Enable/disable 5GHz WiFi",
    type: "boolean",
    writable: true,
    defaultValue: true,
    category: "WiFi Configuration",
    subcategory: "5GHz",
    supportedManufacturers: ["All"],
    supportedModels: ["HG8245H, HG8546M, EG8141A5, F609, F670L"],
    isStandardTR069: true
  },

  // ============================================
  // HUAWEI SPECIFIC PARAMETERS (SEMUA MODEL)
  // ============================================
  // Huawei GPON ONT Status
  {
    path: "Device.HUAWEI.GPON.ONT.Status",
    name: "GPON ONT Status",
    description: "GPON optical line status (O5 = operating)",
    type: "string",
    writable: false,
    defaultValue: "",
    category: "Optical Status",
    subcategory: "GPON",
    supportedManufacturers: ["Huawei"],
    supportedModels: ["HG8245H, HG8546M, EG8141A5, EG8145V5, HG659", "All"],
    isStandardTR069: false
  },
  {
    path: "Device.HUAWEI.GPON.ONT.RXPower",
    name: "Optical RX Power",
    description: "Received optical power in dBm",
    type: "string",
    writable: false,
    defaultValue: "",
    category: "Optical Status",
    subcategory: "GPON",
    supportedManufacturers: ["Huawei"],
    supportedModels: ["All"],
    isStandardTR069: false,
    unit: "dBm"
  },
  {
    path: "Device.HUAWEI.GPON.ONT.TXPower",
    name: "Optical TX Power",
    description: "Transmitted optical power",
    type: "string",
    writable: false,
    defaultValue: "",
    category: "Optical Status",
    subcategory: "GPON",
    supportedManufacturers: ["Huawei"],
    supportedModels: ["All"],
    isStandardTR069: false,
    unit: "dBm"
  },
  {
    path: "Device.HUAWEI.GPON.ONT.Temperature",
    name: "SFP Temperature",
    description: "ONT internal temperature",
    type: "string",
    writable: false,
    defaultValue: "",
    category: "Optical Status",
    subcategory: "GPON",
    supportedManufacturers: ["Huawei"],
    supportedModels: ["All"],
    isStandardTR069: false,
    unit: "°C"
  },
  // Huawei TR-069 settings
  {
    path: "Device.HUAWEI.ManagementServer.URL",
    name: "ACS URL",
    description: "TR-069 ACS server URL",
    type: "string",
    writable: true,
    defaultValue: "http://acs.indihome.co.id:7547",
    category: "ACS Configuration",
    subcategory: "TR-069",
    supportedManufacturers: ["Huawei"],
    supportedModels: ["All"],
    isStandardTR069: false
  },
  // Huawei Reboot
  {
    path: "Device.HUAWEI.Device.Reboot",
    name: "Reboot Device",
    description: "Trigger device reboot",
    type: "string",
    writable: true,
    defaultValue: "",
    category: "Device Control",
    subcategory: "Actions",
    supportedManufacturers: ["Huawei"],
    supportedModels: ["All"],
    isStandardTR069: false
  },
  // Huawei Factory Reset
  {
    path: "Device.HUAWEI.Device.FactoryReset",
    name: "Factory Reset",
    description: "Reset device to factory defaults",
    type: "string",
    writable: true,
    defaultValue: "",
    category: "Device Control",
    subcategory: "Actions",
    supportedManufacturers: ["Huawei"],
    supportedModels: ["All"],
    isStandardTR069: false
  },

  // ============================================
  // ZTE SPECIFIC PARAMETERS (SEMUA MODEL)
  // ============================================
  // ZTE GPON
  {
    path: "Device.ZTE.GPON.OMCI.PONRxPower",
    name: "ZTE Optical RX Power",
    description: "Received optical power ZTE",
    type: "string",
    writable: false,
    defaultValue: "",
    category: "Optical Status",
    subcategory: "GPON",
    supportedManufacturers: ["ZTE"],
    supportedModels: ["F609, F670L, F670Y, ZXHN F670, F680"],
    isStandardTR069: false,
    unit: "dBm"
  },
  {
    path: "Device.ZTE.GPON.OMCI.PONTxPower",
    name: "ZTE Optical TX Power",
    description: "Transmitted optical power ZTE",
    type: "string",
    writable: false,
    defaultValue: "",
    category: "Optical Status",
    subcategory: "GPON",
    supportedManufacturers: ["ZTE"],
    supportedModels: ["All"],
    isStandardTR069: false,
    unit: "dBm"
  },
  {
    path: "Device.ZTE.System.Device.Reboot",
    name: "Reboot ZTE Device",
    description: "Reboot ZTE modem",
    type: "string",
    writable: true,
    defaultValue: "",
    category: "Device Control",
    subcategory: "Actions",
    supportedManufacturers: ["ZTE"],
    supportedModels: ["All"],
    isStandardTR069: false
  },

  // ============================================
  // LAN Settings
  // ============================================
  {
    path: "Device.LANDevice.1.LANHostConfigManagement.IPInterface.1.IPAddress",
    name: "LAN Gateway IP",
    description: "Router LAN IP address (default gateway)",
    type: "string",
    writable: true,
    defaultValue: "192.168.100.1",
    category: "LAN Configuration",
    subcategory: "IP Settings",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.LANDevice.1.LANHostConfigManagement.DHCPServerEnable",
    name: "DHCP Server Enable",
    description: "Enable/disable DHCP server",
    type: "boolean",
    writable: true,
    defaultValue: true,
    category: "LAN Configuration",
    subcategory: "DHCP",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.LANDevice.1.LANHostConfigManagement.DHCPRangeStart",
    name: "DHCP Start IP",
    description: "DHCP pool start address",
    type: "string",
    writable: true,
    defaultValue: "192.168.100.2",
    category: "LAN Configuration",
    subcategory: "DHCP",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.LANDevice.1.LANHostConfigManagement.DHCPRangeEnd",
    name: "DHCP End IP",
    description: "DHCP pool end address",
    type: "string",
    writable: true,
    defaultValue: "192.168.100.254",
    category: "LAN Configuration",
    subcategory: "DHCP",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },

  // ============================================
  // VOIP/PHONE Settings (Untuk PSTN di modem)
  // ============================================
  {
    path: "Device.Services.VoiceService.1.VoiceProfile.1.SIP.ProxyServerAddress",
    name: "SIP Proxy Server",
    description: "VOIP SIP proxy server address",
    type: "string",
    writable: true,
    defaultValue: "",
    category: "VOIP Configuration",
    subcategory: "SIP",
    supportedManufacturers: ["All"],
    supportedModels: ["HG8245H, HG8546M, F609, F670L"],
    isStandardTR069: true
  },
  {
    path: "Device.Services.VoiceService.1.VoiceProfile.1.SIP.Username",
    name: "SIP Username",
    description: "VOIP account username",
    type: "string",
    writable: true,
    defaultValue: "",
    category: "VOIP Configuration",
    subcategory: "SIP",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.Services.VoiceService.1.VoiceProfile.1.SIP.Password",
    name: "SIP Password",
    description: "VOIP account password",
    type: "string",
    writable: true,
    defaultValue: "",
    category: "VOIP Configuration",
    subcategory: "SIP",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },

  // ============================================
  // IPTV Settings
  // ============================================
  {
    path: "Device.IPTV.Enable",
    name: "IPTV Enable",
    description: "Enable IPTV service",
    type: "boolean",
    writable: true,
    defaultValue: true,
    category: "IPTV Configuration",
    subcategory: "General",
    supportedManufacturers: ["Huawei", "ZTE"],
    supportedModels: ["HG8245H, HG8546M, EG8141A5, F609, F670L"],
    isStandardTR069: false
  },
  {
    path: "Device.IPTV.VLAN.ID",
    name: "IPTV VLAN ID",
    description: "VLAN ID for IPTV traffic",
    type: "integer",
    writable: true,
    defaultValue: 4000,
    category: "IPTV Configuration",
    subcategory: "VLAN",
    supportedManufacturers: ["Huawei", "ZTE"],
    supportedModels: ["All"],
    isStandardTR069: false
  },

  // ============================================
  // PORT FORWARDING / NAT
  // ============================================
  {
    path: "Device.NAT.Router.1.PortMappingNumberOfEntries",
    name: "Port Mappings Count",
    description: "Number of active port forwards",
    type: "unsignedInt",
    writable: false,
    defaultValue: 0,
    category: "NAT/Forwarding",
    subcategory: "General",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },

  // ============================================
  // FIREWALL SETTINGS
  // ============================================
  {
    path: "Device.Security.Firewall.Enable",
    name: "Firewall Enable",
    description: "Enable/disable built-in firewall",
    type: "boolean",
    writable: true,
    defaultValue: true,
    category: "Security",
    subcategory: "Firewall",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.Security.Firewall.Level",
    name: "Firewall Level",
    description: "Firewall security level",
    type: "string",
    writable: true,
    defaultValue: "medium",
    allowedValues: ["low", "medium", "high", "extreme"],
    category: "Security",
    subcategory: "Firewall",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },

  // ============================================
  // UPNP, DLNA
  // ============================================
  {
    path: "Device.Services.UPnP.Enable",
    name: "UPnP Enable",
    description: "Enable Universal Plug and Play",
    type: "boolean",
    writable: true,
    defaultValue: true,
    category: "Services",
    subcategory: "UPnP",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.Services.DLNA.Enable",
    name: "DLNA Enable",
    description: "Enable DLNA media server",
    type: "boolean",
    writable: true,
    defaultValue: true,
    category: "Services",
    subcategory: "DLNA",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: false
  },

  // ============================================
  // FIRMWARE UPDATE
  // ============================================
  {
    path: "Device.SoftwareUpdate.Server.URL",
    name: "Firmware Server URL",
    description: "URL untuk download firmware update",
    type: "string",
    writable: true,
    defaultValue: "",
    category: "Firmware",
    subcategory: "Update",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  },
  {
    path: "Device.SoftwareUpdate.UpdateStartTime",
    name: "Scheduled Update Time",
    description: "Jadwal update firmware",
    type: "dateTime",
    writable: true,
    defaultValue: null,
    category: "Firmware",
    subcategory: "Schedule",
    supportedManufacturers: ["All"],
    supportedModels: ["All"],
    isStandardTR069: true
  }
];

// Seed ke database
export async function seedAllParameters() {
  try {
    await ParameterTemplate.deleteMany({});
    const inserted = await ParameterTemplate.insertMany(ALL_PARAMETERS);
    console.log(`✅ ${inserted.length} parameters seeded successfully!`);
    console.log(`📊 Mencakup: Semua modem Huawei, ZTE, dan vendor lain`);
    console.log(`📡 Support: HG8245H, HG8546M, EG8141A5, F609, F670L, dll.`);
    return inserted;
  } catch (error) {
    console.error('❌ Error seeding parameters:', error);
    throw error;
  }
}