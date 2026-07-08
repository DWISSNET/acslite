import express from 'express';
import ParameterTemplate from '../models/ParameterTemplate';
import logger from '../utils/logger';

const router = express.Router();

// Get semua parameter
router.get('/', async (req, res) => {
  try {
    const category = req.query.category as string;
    const manufacturer = req.query.manufacturer as string;
    
    const filter: any = {};
    if (category) filter.category = category;
    if (manufacturer) {
      filter.$or = [
        { supportedManufacturers: manufacturer },
        { supportedManufacturers: "All" }
      ];
    }

    const parameters = await ParameterTemplate.find(filter).sort({ path: 1 });
    
    // Group by category
    const grouped = parameters.reduce((acc, p) => {
      if (!acc[p.category]) acc[p.category] = [];
      acc[p.category].push(p);
      return acc;
    }, {} as Record<string, any[]>);

    res.json({
      total: parameters.length,
      categories: Object.keys(grouped),
      parameters: grouped
    });
  } catch (error) {
    logger.error('Error fetching parameters:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Get parameter by path
router.get('/:path(*)', async (req, res) => {
  try {
    const param = await ParameterTemplate.findOne({ path: req.params.path });
    if (!param) return res.status(404).json({ error: 'Parameter not found' });
    res.json(param);
  } catch (error) {
    res.status(500).json({ error: 'Failed to fetch parameter' });
  }
});

// Get supported manufacturers
router.get('/manufacturers/list', (req, res) => {
  res.json({
    manufacturers: [
      { id: 'All', name: 'Semua Vendor' },
      { id: 'Huawei', name: 'Huawei (Semua model GPON/EPON/LTE)' },
      { id: 'ZTE', name: 'ZTE (F609, F670L, dst)' },
      { id: 'TP-Link', name: 'TP-Link' },
      { id: 'MikroTik', name: 'MikroTik RouterOS' },
      { id: 'D-Link', name: 'D-Link' },
      { id: 'Cisco', name: 'Cisco' }
    ]
  });
});

// Get supported models by manufacturer
router.get('/models/:manufacturer', (req, res) => {
  const { DeviceDetector } = require('../services/cwmp/utils/deviceDetector');
  const modems = DeviceDetector.getModemsForISP(req.params.manufacturer);
  res.json({ models: modems });
});

export default router;