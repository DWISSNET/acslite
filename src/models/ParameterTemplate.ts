import mongoose, { Document, Schema } from 'mongoose';

export interface IParameter extends Document {
  path: string;
  name: string;
  description: string;
  type: 'string' | 'integer' | 'boolean' | 'unsignedInt' | 'base64' | 'dateTime';
  writable: boolean;
  defaultValue: any;
  minValue?: number;
  maxValue?: number;
  allowedValues?: string[];
  category: string;
  subcategory: string;
  supportedManufacturers: string[];
  supportedModels: string[];
  isStandardTR069: boolean;
  unit?: string;
}

const ParameterSchema: Schema = new Schema({
  path: { type: String, required: true, unique: true, index: true },
  name: { type: String, required: true },
  description: String,
  type: { 
    type: String, 
    required: true,
    enum: ['string', 'integer', 'boolean', 'unsignedInt', 'base64', 'dateTime']
  },
  writable: { type: Boolean, default: true },
  defaultValue: Schema.Types.Mixed,
  minValue: Number,
  maxValue: Number,
  allowedValues: [String],
  category: { type: String, index: true },
  subcategory: String,
  supportedManufacturers: [{ type: String, index: true }],
  supportedModels: [String],
  isStandardTR069: { type: Boolean, default: true },
  unit: String
});

ParameterSchema.index({ category: 1, supportedManufacturers: 1 });

export default mongoose.model<IParameter>('ParameterTemplate', ParameterSchema);