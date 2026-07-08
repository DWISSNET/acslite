import cron from 'node-cron';
import Device from '../models/Device';
import logger from '../utils/logger';

// Scheduler berjalan setiap 5 menit untuk cek status device
export function startScheduler() {
  // Cek device offline (tidak kirim inform > 15 menit)
  cron.schedule('*/5 * * * *', async () => {
    try {
      const fifteenMinutesAgo = new Date(Date.now() - 15 * 60 * 1000);
      
      const updated = await Device.updateMany(
        { 
          lastInform: { $lt: fifteenMinutesAgo },
          connectionStatus: 'online'
        },
        { connectionStatus: 'offline' }
      );

      if (updated.modifiedCount > 0) {
        logger.info(`🔌 Marked ${updated.modifiedCount} devices as offline`);
        
        // Emit ke dashboard
        if (global.io) {
          global.io.emit('devices:offline', { count: updated.modifiedCount });
        }
      }

    } catch (error) {
      logger.error('Scheduler error:', error);
    }
  });

  // Cleanup log lama setiap hari
  cron.schedule('0 0 * * *', async () => {
    try {
      const thirtyDaysAgo = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
      
      await Device.updateMany(
        {},
        {
          $pull: {
            eventLog: { timestamp: { $lt: thirtyDaysAgo } }
          }
        }
      );
      
      logger.info('🧹 Old event logs cleaned up');
    } catch (error) {
      logger.error('Log cleanup error:', error);
    }
  });

  logger.info('✅ Schedulers started');
}

// Start immediately
startScheduler();