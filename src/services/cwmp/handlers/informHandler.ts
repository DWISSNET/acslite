import logger from '../../../utils/logger';
import Device from '../../../models/Device';
import { DeviceDetector } from '../utils/deviceDetector';
import { createSetParameterValuesRequest } from '../utils/messageBuilder';

export async function handleInform(informData: any, ipAddress: string) {
  try {
    logger.info(`📥 INFORM received from ${ipAddress}`);
    
    // Extract data dari INFORM
    const deviceId = informData?.DeviceId?.[0];
    const event = informData?.Event?.[0];
    const currentTime = informData?.CurrentTime?.[0];
    
    if (!deviceId) {
      logger.warn('⚠️ INFORM missing DeviceId');
      return { status: 'invalid_data' };
    }

    const serial = deviceId.SerialNumber?.[0] || 'unknown';
    logger.info(`📡 Device ${serial} sent INFORM from ${ipAddress}`);

    // Deteksi perangkat
    const deviceInfo = await DeviceDetector.identifyAndCache(deviceId, ipAddress);
    
    // Dapatkan semua parameter yang didukung untuk device ini
    const params = await DeviceDetector.getSupportedParametersForDevice(
      deviceInfo?.manufacturer || 'Unknown',
      deviceInfo?.model || 'Unknown'
    );

    // Generate daftar parameter untuk di-set
    const parametersToSet = params.filter(p => p.writable).map(p => ({
      name: p.path,
      value: p.defaultValue
    }));

    // Event logging
    await logDeviceEvent(serial, 'inform_received', `Device sent INFORM event: ${JSON.stringify(event)}`);
    
    // Jika ini adalah BOOTSTRAP (perangkat pertama kali connect)
    const isBootstrap = event?.EventCode?.some((e: any) => e.$.code === '1 BOOT');
    if (isBootstrap) {
      logger.info(`🚀 New device ${serial} is bootstrapping!`);
      await logDeviceEvent(serial, 'bootstrap', 'Device first connection - starting provisioning');
      
      // Kirim SetParameterValues untuk konfigurasi awal
      if (parametersToSet.length > 0) {
        logger.info(`⚙️ Sending ${parametersToSet.length} parameters to ${serial}`);
        // Disini akan dikirim RPC ke device untuk set semua parameter
      }
    }

    // Update status online di database
    await Device.findOneAndUpdate(
      { serialNumber: serial },
      { 
        lastInform: new Date(), 
        connectionStatus: 'online',
        ipAddress: ipAddress
      }
    );

    // Emit ke dashboard realtime
    if (global.io) {
      global.io.emit('device:online', {
        serial,
        ip: ipAddress,
        deviceInfo,
        timestamp: new Date()
      });
    }

    logger.info(`✅ INFORM processed successfully for ${serial}`);
    return { 
      status: 'success', 
      maxEnvelopes: 1,
      parametersToSet: parametersToSet.length
    };

  } catch (error) {
    logger.error('❌ Error handling INFORM:', error);
    throw error;
  }
}

async function logDeviceEvent(serial: string, event: string, description: string) {
  await Device.findOneAndUpdate(
    { serialNumber: serial },
    {
      $push: {
        eventLog: {
          event,
          timestamp: new Date(),
          description
        }
      }
    }
  );
}