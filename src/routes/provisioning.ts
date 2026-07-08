import express from 'express';
import Device from '../models/Device';
import logger from '../utils/logger';
import { v4 as uuidv4 } from 'uuid';

const router = express.Router();

// Bulk provisioning
router.post('/bulk', async (req, res) => {
  try {
    const { devices } = req.body;
    const results: any[] = [];

    for (const dev of devices) {
      const device = new Device({
        deviceId: uuidv4(),
        serialNumber: dev.serial,
        macAddress: dev.mac,
        manufacturer: dev.manufacturer || 'Huawei',
        modelName: dev.model || 'HG8245H',
        isp: dev.isp || 'Indihome',
        customer: {
          name: dev.customerName,
          phone: dev.phone,
          package: dev.package
        },
        location: {
          city: dev.city,
          province: dev.province
        },
        connectionStatus: 'pending'
      });

      await device.save();
      results.push({ serial: dev.serial, status: 'queued' });
    }

    logger.info(`📦 Bulk provisioning: ${results.length} devices queued`);
    res.json({ success: true, processed: results.length, results });
  } catch (error) {
    logger.error('Bulk provisioning error:', error);
    res.status(500).json({ error: 'Failed to process bulk provisioning' });
  }
});

// Generate configuration template untuk ISP
router.get('/template/:isp', async (req, res) => {
  const isp = req.params.isp.toLowerCase();
  
  const templates: Record<string, any> = {
    indihome: {
      name: 'Indihome Default Template',
      isp: 'Indihome',
      wan: {
        pppoe: { username: '%SESSION%@indihome', password: '%PASSWORD%' },
        vlan: 35,
        pppoeVlan: 35
      },
      iptv: {
        vlan: 4000,
        igmp: true,
        multicast: true
      },
      voice: {
        sipServer: 'sip.indihome.co.id',
        port: 5060
      },
      wifi: {
        ssidPrefix: 'INDIHOME-',
        band2g: true,
        band5g: true
      }
    },
    firstmedia: {
      name: 'First Media Template',
      isp: 'First Media',
      wan: {
        dhcp: true,
        vlan: 100
      },
      wifi: {
        ssidPrefix: 'FIRSTMEDIA-',
        band2g: true,
        band5g: true
      }
    },
    biznet: {
      name: 'Biznet Home Template',
      isp: 'Biznet',
      wan: {
        pppoe: { username: '%USER%', password: '%PASS%' },
        vlan: 100
      }
    },
    mncplay: {
      name: 'MNC Play Template',
      isp: 'MNC Play',
      wan: {
        dhcp: true,
        vlan: 200
      }
    },
    xlhome: {
      name: 'XL Home LTE Template',
      isp: 'XL Axiata',
      apn: 'xlhome',
      pincode: '0000'
    }
  };

  if (!templates[isp]) {
    return res.status(404).json({ error: 'Template not found for this ISP' });
  }

  res.json(templates[isp]);
});

// Get provisioning history
router.get('/history', async (req, res) => {
  try {
    const devices = await Device.find({
      firstSeen: { $gte: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000) }
    }).sort({ firstSeen: -1 }).limit(100);
    
    res.json(devices.map(d => ({
      serial: d.serialNumber,
      model: d.modelName,
      provisionedAt: d.firstSeen,
      status: d.connectionStatus
    })));
  } catch (error) {
    res.status(500).json({ error: 'Failed to fetch provisioning history' });
  }
});

export default router;