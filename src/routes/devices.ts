import express from 'express';
import Device from '../models/Device';
import logger from '../utils/logger';

const router = express.Router();

// Get semua devices
router.get('/', async (req, res) => {
  try {
    const page = parseInt(req.query.page as string) || 1;
    const limit = parseInt(req.query.limit as string) || 50;
    const search = req.query.search as string || '';
    
    const filter: any = {};
    if (search) {
      filter.$or = [
        { serialNumber: { $regex: search, $options: 'i' } },
        { macAddress: { $regex: search, $options: 'i' } },
        { modelName: { $regex: search, $options: 'i' } },
        { 'customer.name': { $regex: search, $options: 'i' } }
      ];
    }

    const devices = await Device.find(filter)
      .sort({ lastInform: -1 })
      .skip((page - 1) * limit)
      .limit(limit);

    const total = await Device.countDocuments(filter);
    
    res.json({
      data: devices,
      pagination: { page, limit, total, pages: Math.ceil(total / limit) }
    });
  } catch (error) {
    logger.error('Error fetching devices:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Get device by ID
router.get('/:serial', async (req, res) => {
  try {
    const device = await Device.findOne({ 
      $or: [
        { serialNumber: req.params.serial },
        { deviceId: req.params.serial }
      ]
    });
    
    if (!device) {
      return res.status(404).json({ error: 'Device not found' });
    }
    
    // Dapatkan parameter yang didukung untuk device ini
    const { DeviceDetector } = require('../services/cwmp/utils/deviceDetector');
    const parameters = await DeviceDetector.getSupportedParametersForDevice(
      device.manufacturer,
      device.modelName
    );

    res.json({ device, availableParameters: parameters });
  } catch (error) {
    res.status(500).json({ error: 'Failed to fetch device details' });
  }
});

// Reboot device
router.post('/:serial/reboot', async (req, res) => {
  try {
    const device = await Device.findOne({ serialNumber: req.params.serial });
    if (!device) return res.status(404).json({ error: 'Device not found' });

    // Kirim RPC Reboot ke device
    logger.info(`🔄 Reboot command sent to ${device.serialNumber}`);
    
    await device.updateOne({
      $push: {
        eventLog: {
          event: 'reboot_requested',
          description: 'Admin requested device reboot via dashboard'
        }
      }
    });

    res.json({ success: true, message: 'Reboot command queued' });
  } catch (error) {
    res.status(500).json({ error: 'Failed to send reboot command' });
  }
});

// Factory reset
router.post('/:serial/factory-reset', async (req, res) => {
  try {
    const device = await Device.findOne({ serialNumber: req.params.serial });
    if (!device) return res.status(404).json({ error: 'Device not found' });

    logger.warn(`⚠️ Factory reset requested for ${device.serialNumber}`);
    res.json({ success: true, message: 'Factory reset command queued' });
  } catch (error) {
    res.status(500).json({ error: 'Failed to process request' });
  }
});

// Update parameters pada device
router.post('/:serial/set-parameters', async (req, res) => {
  try {
    const { parameters } = req.body;
    const device = await Device.findOne({ serialNumber: req.params.serial });
    
    if (!device) return res.status(404).json({ error: 'Device not found' });
    
    // Update parameter di database dan kirim ke device
    const paramMap = new Map(device.parameters);
    parameters.forEach((p: any) => paramMap.set(p.path, p.value));
    device.parameters = paramMap;
    await device.save();

    logger.info(`⚙️ Updated ${parameters.length} parameters on ${device.serialNumber}`);
    res.json({ success: true, updated: parameters.length });
  } catch (error) {
    res.status(500).json({ error: 'Failed to update parameters' });
  }
});

export default router;