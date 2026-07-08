import { createClient } from 'redis';
import logger from '../utils/logger';
import dotenv from 'dotenv';

dotenv.config();

export const redisClient = createClient({
  url: process.env.REDIS_URI || 'redis://localhost:6379'
});

redisClient.on('error', (err) => logger.error('Redis error:', err));
redisClient.on('connect', () => logger.debug('Redis connecting...'));
redisClient.on('ready', () => logger.debug('Redis ready'));