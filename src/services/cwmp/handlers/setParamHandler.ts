import logger from '../../../utils/logger';
import Device from '../../../models/Device';

export async function handleSetParameterValues(request: any) {
  try {
    const deviceId = request?.DeviceId?.[0];
    const status = request?.Status?.[0];
    const startTime = request?.StartTime?.[0];
    
    logger.debug(`SetParameterValuesResponse received: Device ${deviceId}, Status: ${status}`);

    if (deviceId?.SerialNumber?.[0]) {
      const serial = deviceId.SerialNumber[0];
      await Device.findOneAndUpdate(
        { serialNumber: serial },
        {
          $push: {
            eventLog: {
              event: 'parameters_updated',
              timestamp: new Date(),
              description: `SetParameterValues completed with status: ${status}`
            }
          }
        }
      );
    }

    return { status: 'acknowledged' };
  } catch (error) {
    logger.error('Error handling SetParameterValues:', error);
    throw error;
  }
}