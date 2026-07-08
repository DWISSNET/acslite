import logger from '../../../utils/logger';

// Semua RPC Method yang didukung ACS (standar TR-069)
const SUPPORTED_RPC_METHODS = [
  'GetRPCMethods',
  'SetParameterValues',
  'GetParameterValues',
  'GetParameterNames',
  'SetParameterAttributes',
  'GetParameterAttributes',
  'AddObject',
  'DeleteObject',
  'Reboot',
  'FactoryReset',
  'Download',
  'Upload',
  'GetQueuedTransfers',
  'ScheduleInform',
  'SetVouchers',
  'GetOptions'
];

export async function handleGetRPCMethods(request: any) {
  logger.info('📡 GetRPCMethods requested');
  return {
    methods: SUPPORTED_RPC_METHODS,
    status: 'success'
  };
}