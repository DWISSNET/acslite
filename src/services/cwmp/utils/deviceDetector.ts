import Device from '../../../models/Device';
import ParameterTemplate from '../../../models/ParameterTemplate';
import logger from '../../../utils/logger';

// Database semua modem yang didukung
const SUPPORTED_MODEMS = {
  huawei: [
    { model: 'HG8245H', name: 'Huawei EchoLife HG8245H', type: 'GPON', isp: ['Indihome', 'First Media'] },
    { model: 'HG8546M', name: 'Huawei EchoLife HG8546M', type: 'GPON', isp: ['Indihome'] },
    { model: 'EG8141A5', name: 'Huawei EchoLife EG8141A5', type: 'GPON', isp: ['Indihome'] },
    { model: 'EG8145V5', name: 'Huawei EG8145V5', type: 'GPON', isp: ['Indihome', 'MyRepublic'] },
    { model: 'HG659', name: 'Huawei HG659', type: 'VDSL', isp: ['First Media', 'MNC Play'] },
    { model: 'HS8546V5', name: 'Huawei HS8546V5', type: 'GPON', isp: ['Indihome', 'Biznet'] },
    { model: 'B618s-22d', name: 'Huawei B618', type: 'LTE', isp: ['XL', 'Telkomsel', 'Indosat'] },
    { model: 'B525s-65a', name: 'Huawei B525', type: 'LTE', isp: ['Smartfren', 'Tri'] },
  ],
  zte: [
    { model: 'F609', name: 'ZTE ZXHN F609', type: 'GPON', isp: ['Indihome'] },
    { model: 'F670L', name: 'ZTE ZXHN F670L', type: 'GPON', isp: ['Indihome', 'First Media'] },
    { model: 'F670Y', name: 'ZTE F670Y', type: 'GPON', isp: ['Indihome'] },
    { model: 'F680', name: 'ZTE ZXHN F680', type: 'GPON', isp: ['Biznet'] },
    { model: 'ZXHN F601', name: 'ZTE ZXHN F601', type: 'GPON', isp: ['Indihome'] },
    { model: 'MF286R', name: 'ZTE MF286R', type: 'LTE', isp: ['XL', 'Telkomsel'] },
  ],
  mikrotik: [
    { model: 'RB750Gr3', name: 'MikroTik hEX', type: 'RouterOS', isp: ['All'] },
    { model: 'RB4011iGS+', name: 'MikroTik RB4011', type: 'RouterOS', isp: ['All'] },
    { model: 'CCR1009', name: 'MikroTik CCR1009', type: 'RouterOS', isp: ['Biznet'] },
  ],
  tplink: [
    { model: 'XC220-G3v', name: 'TP-Link XC220-G3v', type: 'GPON', isp: ['Indihome'] },
    { model: 'AX1800', name: 'TP-Link Archer AX1800', type: 'WiFi6', isp: ['All'] },
  ],
  dlink: [
    { model: 'DPR-1041', name: 'D-Link DPR-1041', type: 'GPON', isp: ['First Media'] },
    { model: 'DSL-2750U', name: 'D-Link DSL-2750U', type: 'ADSL/VDSL', isp: ['All'] },
  ]
};

export class DeviceDetector {
  private static cache = new Map<string, any>();

  static async identifyAndCache(deviceId: any, ipAddress: string) {
    try {
      const serial = deviceId?.SerialNumber?.[0] || 'unknown';
      const manufacturer = deviceId?.Manufacturer?.[0] || 'unknown';
      const productClass = deviceId?.ProductClass?.[0] || 'unknown';

      // Deteksi model dari serial dan manufaktur
      const detected = this.detectModel(manufacturer, productClass, serial);
      
      // Simpan atau update di database
      await this.updateDeviceInDB({
        deviceId: deviceId?.ID?.[0] || serial,
        serialNumber: serial,
        manufacturer: detected.manufacturer,
        modelName: detected.model,
        productClass,
        ipAddress,
        macAddress: this.extractMACFromSerial(serial),
        isp: detected.isp?.[0] || 'Unknown'
      });

      this.cache.set(serial, detected);
      logger.info(`✅ Detected: ${detected.manufacturer} ${detected.model} (${ipAddress})`);
      
      return detected;
    } catch (error) {
      logger.error('Device detection error:', error);
      return null;
    }
  }

  static detectModel(manufacturer: string, productClass: string, serial: string) {
    const lowerManu = manufacturer.toLowerCase();
    
    // Cari di database modem
    for (const [vendor, models] of Object.entries(SUPPORTED_MODEMS)) {
      if (lowerManu.includes(vendor)) {
        // Cari model yang cocok
        const match = models.find(m => 
          productClass.includes(m.model) || 
          serial.toUpperCase().includes(m.model)
        );
        
        if (match) {
          return { ...match, manufacturer: vendor };
        }
        
        // Return first model jika tidak cocok persis
        return { ...models[0], manufacturer: vendor };
      }
    }

    // Fallback jika tidak terdeteksi
    return {
      model: productClass || 'Unknown Model',
      manufacturer,
      type: 'Unknown',
      isp: ['Unknown']
    };
  }

  static async getSupportedParametersForDevice(manufacturer: string, modelName: string) {
    // Ambil SEMUA parameter yang mendukung device ini
    const params = await ParameterTemplate.find({
      $or: [
        { supportedManufacturers: manufacturer },
        { supportedManufacturers: "All" },
        { supportedModels: { $in: [modelName, "All"] } }
      ]
    });
    
    logger.debug(`📋 Loaded ${params.length} parameters for ${manufacturer} ${modelName}`);
    return params;
  }

  private static extractMACFromSerial(serial: string): string {
    // Ekstrak MAC dari serial jika memungkinkan
    if (serial.length >= 12) {
      const mac = serial.substring(serial.length - 12).match(/.{2}/g)?.join(':');
      return mac || '00:00:00:00:00:00';
    }
    return '00:00:00:00:00:00';
  }

  private static async updateDeviceInDB(deviceData: any) {
    await Device.findOneAndUpdate(
      { serialNumber: deviceData.serialNumber },
      {
        ...deviceData,
        lastInform: new Date(),
        connectionStatus: 'online',
        $setOnInsert: { firstSeen: new Date() }
      },
      { upsert: true, new: true }
    );
  }

  // Get semua modem yang didukung
  static getAllSupportedModems() {
    const all: any[] = [];
    Object.values(SUPPORTED_MODEMS).forEach(models => all.push(...models));
    return all;
  }

  // Dapatkan daftar modem untuk ISP tertentu
  static getModemsForISP(isp: string) {
    const all: any[] = [];
    Object.values(SUPPORTED_MODEMS).forEach(models => {
      models.forEach(m => {
        if (m.isp.includes(isp) || m.isp.includes('All')) {
          all.push(m);
        }
      });
    });
    return all;
  }
}