import express from 'express';
import http from 'http';
import cors from 'cors';
import helmet from 'helmet';
import { Server } from 'socket.io';
import { connectDB } from './config/database';
import { startCWMPserver } from './services/cwmp/server';
import { redisClient } from './config/redis';
import logger from './utils/logger';
import routes from './routes';
import './services/scheduler';

const app = express();
const server = http.createServer(app);
const PORT = process.env.PORT || 7547; // Port standar TR-069

// Middleware
app.use(helmet());
app.use(cors());
app.use(express.json({ limit: '50mb' }));
app.use(express.urlencoded({ extended: true }));

// Socket.IO untuk realtime dashboard
const io = new Server(server, {
  cors: { origin: "*" }
});

// API Routes
app.use('/api', routes);

// Initialize semua services
async function initialize() {
  try {
    // Koneksi MongoDB
    await connectDB();
    logger.info('✅ MongoDB connected successfully');

    // Koneksi Redis
    await redisClient.connect();
    logger.info('✅ Redis connected successfully');

    // Start CWMP/TR-069 Server
    await startCWMPserver(server);
    logger.info('✅ CWMP Server started on port 7547');

    // Web server
    server.listen(PORT + 1, () => {
      logger.info(`✅ Web Server running on http://localhost:${PORT + 1}`);
      logger.info(`✅ Dashboard: http://localhost:${PORT + 1}/dashboard`);
      logger.info(`✅ CWMP/TR-069: :7547 (standar ACS port)`);
    });

    // Socket connection
    io.on('connection', (socket) => {
      logger.info(`Dashboard client connected: ${socket.id}`);
      socket.on('disconnect', () => logger.info(`Client disconnected: ${socket.id}`));
    });

    global.io = io;

  } catch (error) {
    logger.error('❌ Failed to initialize server:', error);
    process.exit(1);
  }
}

initialize();