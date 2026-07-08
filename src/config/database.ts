import mongoose from 'mongoose';
import logger from '../utils/logger';
import dotenv from 'dotenv';

dotenv.config();

export async function connectDB() {
  const mongoUri = process.env.MONGODB_URI || 'mongodb://localhost:27017/advanced_acs';
  
  try {
    await mongoose.connect(mongoUri);
    logger.info('MongoDB connected successfully');
    
    // Seed parameter setelah connect DB
    const { seedAllParameters } = require('../seeders/fullParameters.seeder');
    await seedAllParameters();
    
  } catch (error) {
    logger.error('MongoDB connection error:', error);
    process.exit(1);
  }
}