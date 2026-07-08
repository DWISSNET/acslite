import express from 'express';
import jwt from 'jsonwebtoken';
import bcrypt from 'bcryptjs';

const router = express.Router();
const JWT_SECRET = process.env.JWT_SECRET || 'acs-lite-secret-key-2026';
const ADMIN_PASSWORD = process.env.ADMIN_PASSWORD || 'admin123';
const ADMIN_EMAIL = process.env.ADMIN_EMAIL || 'admin@acslite.local';

// Pre-hash password untuk hindari hash berulang
const hashedPassword = bcrypt.hashSync(ADMIN_PASSWORD, 10);

// Login endpoint
router.post('/login', async (req, res) => {
  const { email, password } = req.body;
  
  if (email !== ADMIN_EMAIL) {
    return res.status(401).json({ error: 'Email atau password salah' });
  }
  
  const isPasswordValid = await bcrypt.compare(password, hashedPassword);
  if (!isPasswordValid) {
    return res.status(401).json({ error: 'Email atau password salah' });
  }

  const token = jwt.sign({ email }, JWT_SECRET, { expiresIn: '24h' });
  res.json({ 
    success: true,
    token, 
    user: { email: ADMIN_EMAIL } 
  });
});

// Logout endpoint (cukup hapus token di client)
router.post('/logout', (req, res) => {
  res.json({ success: true, message: 'Logout berhasil' });
});

export default router;