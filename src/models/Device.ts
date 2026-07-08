import mongoose, { Document, Schema } from 'mongoose';

export interface IDevice extends Document {
  deviceId: string;
  serialNumber: string;
  productClass: string;
  manufacturer: string;
  modelName: string;
  softwareVersion: string;
  hardwareVersion: string;
  ipAddress: string;
  macAddress: string;
  lastInform: Date;
  firstSeen: Date;
  connectionStatus: 'online' | 'offline' | 'suspended';
  isp: string;
  location: {
    city: string;
    province: string;
    coordinates?: { lat: number; lng: number }
  };
  customer: {
    name: string;
    phone: string;
    email: string;
    package: string;
  };
  parameters: Map<string, any>;
  eventLog: Array<{
    event: string;
    timestamp: Date;
    description: string;
  }>;
  tags: string[];
}

const DeviceSchema: Schema = new Schema({
  deviceId: { type: String, required: true, unique: true, index: true },
  serialNumber: { type: String, required: true, unique: true, index: true },
  productClass: String,
  manufacturer: { type: String, index: true },
  modelName: { type: String, index: true },
  softwareVersion: String,
  hardwareVersion: String,
  ipAddress: String,
  macAddress: { type: String, index: true },
  lastInform: Date,
  firstSeen: { type: Date, default: Date.now },
  connectionStatus: { 
    type: String, 
    enum: ['online', 'offline', 'suspended'], 
    default: 'offline',
    index: true
  },
  isp: { type: String, index: true },
  location: {
    city: String,
    province: String,
    coordinates: { lat: Number, lng: Number }
  },
  customer: {
    name: String,
    phone: String,
    email: String,
    package: String
  },
  parameters: { type: Map, of: Schema.Types.Mixed, default: new Map() },
  eventLog: [{
    event: String,
    timestamp: { type: Date, default: Date.now },
    description: String
  }],
  tags: [String]
}, { timestamps: true });

DeviceSchema.index({ manufacturer: 1, modelName: 1 });
DeviceSchema.index({ lastInform: -1 });

export default mongoose.model<IDevice>('Device', DeviceSchema);