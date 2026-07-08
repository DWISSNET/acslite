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
import authRoutes from './routes/auth';
import { authenticate } from './middleware/auth';
import './services/scheduler';

// Declare global io type untuk TypeScript
declare global {
  var io: Server;
}

const app = express();
const server = http.createServer(app);
const CWMP_PORT = Number(process.env.CWMP_PORT) || 7547;  // Port TR-069 standar
const WEB_PORT = Number(process.env.PORT) || 7548;        // Port web/dashboard

// Middleware
app.use(helmet());
app.use(cors());
app.use(express.json({ limit: '50mb' }));
app.use(express.urlencoded({ extended: true }));

// Serve static files untuk login & dashboard
app.use(express.static('public'));
app.use('/login', express.static('public/login.html'));
app.use('/dashboard', express.static('public/dashboard'));

// Socket.IO untuk realtime dashboard
const io = new Server(server, {
  cors: { origin: "*" }
});

// Auth routes (tidak diproteksi)
app.use('/api/auth', authRoutes);
// Protek semua API lainnya dengan auth middleware
app.use('/api', authenticate, routes);

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
    logger.info('✅ CWMP Server started on port ' + CWMP_PORT);

    // Web server listen di SEMUA interface (bukan cuma localhost!)
    server.listen(WEB_PORT, '0.0.0.0', () => {
      logger.info('✅ Web Server running on http://0.0.0.0:' + WEB_PORT);
      logger.info('✅ Login page: http://0.0.0.0:' + WEB_PORT + '/login');
      logger.info('✅ Dashboard: http://0.0.0.0:' + WEB_PORT + '/dashboard');
      logger.info('✅ CWMP/TR-069: :' + CWMP_PORT + ' (standar ACS port)');
    });

    // Socket connection
    io.on('connection', (socket) => {
      logger.info('Dashboard client connected: ' + socket.id);
      socket.on('disconnect', () => logger.info('Client disconnected: ' + socket.id));
    });

    global.io = io;

  } catch (error) {
    logger.error('❌ Failed to initialize server:', error);
    process.exit(1);
  }
}

initialize();
