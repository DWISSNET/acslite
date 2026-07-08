import express from 'express';
import deviceRoutes from './devices';
import parameterRoutes from './parameters';
import provisioningRoutes from './provisioning';

const router = express.Router();

// API Routes
router.use('/devices', deviceRoutes);
router.use('/parameters', parameterRoutes);
router.use('/provisioning', provisioningRoutes);

// Health check
router.get('/health', (req, res) => {
  res.json({
    status: 'ok',
    timestamp: new Date(),
    service: 'Advanced ACS Server',
    version: '1.0.0'
  });
});

// Get server stats
router.get('/stats', async (req, res) => {
  try {
    const Device = require('../models/Device').default;
    const ParameterTemplate = require('../models/ParameterTemplate').default;
    
    const totalDevices = await Device.countDocuments();
    const onlineDevices = await Device.countDocuments({ connectionStatus: 'online' });
    const totalParameters = await ParameterTemplate.countDocuments();
    const supportedModems = require('../services/cwmp/utils/deviceDetector').DeviceDetector.getAllSupportedModems();

    res.json({
      devices: {
        total: totalDevices,
        online: onlineDevices,
        offline: totalDevices - onlineDevices
      },
      parameters: totalParameters,
      supportedModems: supportedModems.length
    });
  } catch (error) {
    res.status(500).json({ error: 'Failed to fetch stats' });
  }
});

export default router;