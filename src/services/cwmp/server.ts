import { Server } from 'http';
import express from 'express';
import xml2js from 'xml2js';
import logger from '../../utils/logger';
import { handleInform } from './handlers/informHandler';
import { handleGetRPCMethods } from './handlers/rpcHandler';
import { handleSetParameterValues } from './handlers/setParamHandler';
import { DeviceDetector } from './utils/deviceDetector';

const cwmpApp = express();

export async function startCWMPserver(server: Server) {
  // CWMP/TR-069 berjalan di port 7547 (standar)
  const cwmpServer = express();
  
  cwmpServer.use(express.raw({ type: 'text/xml', limit: '10mb' }));
  cwmpServer.use(express.text({ limit: '10mb' }));

  // Endpoint utama TR-069 - /cwmp atau /acs
  cwmpServer.post('/cwmp', async (req, res) => {
    try {
      const xml = req.body.toString();
      logger.debug(`📥 Received CWMP request (${xml.length} bytes)`);
      
      // Parse SOAP/XML CWMP
      const parser = new xml2js.Parser();
      const parsed = await parser.parseStringPromise(xml);
      
      // Deteksi tipe RPC method
      const envelope = parsed['SOAP-ENV:Envelope'];
      if (!envelope || !envelope['SOAP-ENV:Body']) {
        res.status(400).send('Invalid SOAP envelope');
        return;
      }

      const body = envelope['SOAP-ENV:Body'][0];
      const rpcMethod = Object.keys(body)[0];
      const deviceIp = req.ip || req.socket?.remoteAddress || 'unknown';

      logger.info(`📡 CWMP: ${rpcMethod} from ${deviceIp}`);

      // Handle berbagai RPC method standar TR-069
      let response: any;
      
      switch(rpcMethod) {
        case 'cwmp:Inform':
          response = await handleInform(body[rpcMethod], deviceIp);
          break;
          
        case 'cwmp:GetRPCMethods':
          response = await handleGetRPCMethods(body[rpcMethod]);
          break;
          
        case 'cwmp:SetParameterValuesResponse':
          response = await handleSetParameterValues(body[rpcMethod]);
          break;
          
        case 'cwmp:GetParameterValuesResponse':
          response = { status: 'OK', message: 'GetParameterValues acknowledged' };
          break;
          
        case 'cwmp:TransferComplete':
          response = { status: 'OK', message: 'TransferComplete received' };
          break;
          
        default:
          logger.warn(`⚠️ Unhandled CWMP method: ${rpcMethod}`);
          response = { status: 'unhandled', method: rpcMethod };
      }

      // Auto-detect perangkat (model/manufacturer)
      if (body[0]?.DeviceId) {
        DeviceDetector.identifyAndCache(body[0].DeviceId, deviceIp);
      }

      // Generate SOAP response
      const soapResponse = generateSOAPResponse(response);
      res.setHeader('Content-Type', 'text/xml');
      res.send(soapResponse);

    } catch (error) {
      logger.error('❌ CWMP processing error:', error);
      res.status(500).send('Internal server error');
    }
  });

  // Mount cwmp server ke port 7547
  const acsServer = new Server(cwmpServer);
  acsServer.listen(7547, '0.0.0.0', () => {
    logger.info(`🔌 CWMP/TR-069 server listening on 0.0.0.0:7547`);
    logger.info(`📶 All modems can connect to this ACS server`);
  });

  return acsServer;
}

function generateSOAPResponse(data: any): string {
  return `<?xml version="1.0" encoding="utf-8"?>
<SOAP-ENV:Envelope 
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" 
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">${Date.now()}</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:InformResponse>
      <cwmp:MaxEnvelopes>1</cwmp:MaxEnvelopes>
    </cwmp:InformResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`;
}