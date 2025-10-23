// routes/tracking.js - Email Tracking (Opens, Clicks, Submissions)
const express = require('express');
const router = express.Router();
const db = require('../database/db');

// Track email open (1x1 pixel)
router.get('/open/:trackingId', async (req, res) => {
  try {
    const trackingId = req.params.trackingId;
    const ipAddress = req.ip || req.connection.remoteAddress;
    const userAgent = req.get('user-agent');

    // Update recipient record
    const result = await db.run(`
      UPDATE recipients 
      SET opened = 1, opened_at = CURRENT_TIMESTAMP, ip_address = ?, user_agent = ?
      WHERE tracking_id = ? AND opened = 0
    `, [ipAddress, userAgent, trackingId]);

    // Update campaign stats if this was first open
    if (result.changes > 0) {
      const recipient = await db.get(`
        SELECT campaign_id FROM recipients WHERE tracking_id = ?
      `, [trackingId]);

      if (recipient) {
        await db.run(`
          UPDATE campaigns 
          SET opened_count = opened_count + 1
          WHERE id = ?
        `, [recipient.campaign_id]);
      }
    }

    // Return 1x1 transparent pixel
    const pixel = Buffer.from(
      'R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7',
      'base64'
    );
    res.writeHead(200, {
      'Content-Type': 'image/gif',
      'Content-Length': pixel.length,
      'Cache-Control': 'no-cache, no-store, must-revalidate'
    });
    res.end(pixel);
  } catch (err) {
    console.error('Error tracking open:', err);
    res.status(500).send('Error');
  }
});

// Track link click
router.get('/click/:trackingId', async (req, res) => {
  try {
    const trackingId = req.params.trackingId;
    const ipAddress = req.ip || req.connection.remoteAddress;
    const userAgent = req.get('user-agent');

    // Get recipient and campaign info
    const recipient = await db.get(`
      SELECT r.*, c.landing_url 
      FROM recipients r
      JOIN campaigns c ON r.campaign_id = c.id
      WHERE r.tracking_id = ?
    `, [trackingId]);

    if (!recipient) {
      return res.status(404).send('Invalid tracking link');
    }

    // Update recipient record
    const result = await db.run(`
      UPDATE recipients 
      SET clicked = 1, clicked_at = CURRENT_TIMESTAMP, ip_address = ?, user_agent = ?
      WHERE tracking_id = ? AND clicked = 0
    `, [ipAddress, userAgent, trackingId]);

    // Update campaign stats if this was first click
    if (result.changes > 0) {
      await db.run(`
        UPDATE campaigns 
        SET clicked_count = clicked_count + 1
        WHERE id = ?
      `, [recipient.campaign_id]);
    }

    // Redirect to landing page with tracking ID
    const landingUrl = recipient.landing_url || process.env.DEFAULT_LANDING_URL;
    const separator = landingUrl.includes('?') ? '&' : '?';
    res.redirect(`${landingUrl}${separator}tid=${trackingId}`);
  } catch (err) {
    console.error('Error tracking click:', err);
    res.status(500).send('Error');
  }
});

// Track form submission
router.post('/submit/:trackingId', async (req, res) => {
  try {
    const trackingId = req.params.trackingId;
    const submittedData = JSON.stringify(req.body);
    const ipAddress = req.ip || req.connection.remoteAddress;
    const userAgent = req.get('user-agent');

    // Get recipient
    const recipient = await db.get(`
      SELECT campaign_id FROM recipients WHERE tracking_id = ?
    `, [trackingId]);

    if (!recipient) {
      return res.status(404).json({ error: 'Invalid tracking ID' });
    }

    // Update recipient record
    const result = await db.run(`
      UPDATE recipients 
      SET submitted = 1, submitted_at = CURRENT_TIMESTAMP, 
          submitted_data = ?, ip_address = ?, user_agent = ?
      WHERE tracking_id = ? AND submitted = 0
    `, [submittedData, ipAddress, userAgent, trackingId]);

    // Update campaign stats if this was first submission
    if (result.changes > 0) {
      await db.run(`
        UPDATE campaigns 
        SET submitted_count = submitted_count + 1
        WHERE id = ?
      `, [recipient.campaign_id]);
    }

    res.json({ 
      success: true, 
      message: 'Data recorded',
      redirect_url: process.env.TRAINING_URL || 'https://example.com/security-training'
    });
  } catch (err) {
    console.error('Error tracking submission:', err);
    res.status(500).json({ error: 'Failed to record submission' });
  }
});

// Get tracking data for a campaign
router.get('/campaign/:campaignId', async (req, res) => {
  try {
    const recipients = await db.query(`
      SELECT 
        email,
        sent,
        opened,
        clicked,
        submitted,
        sent_at,
        opened_at,
        clicked_at,
        submitted_at,
        ip_address,
        user_agent
      FROM recipients
      WHERE campaign_id = ?
      ORDER BY email
    `, [req.params.campaignId]);

    res.json(recipients);
  } catch (err) {
    console.error('Error fetching tracking data:', err);
    res.status(500).json({ error: 'Failed to fetch tracking data' });
  }
});

// Get detailed submission data
router.get('/submissions/:campaignId', async (req, res) => {
  try {
    const submissions = await db.query(`
      SELECT 
        email,
        submitted_at,
        submitted_data,
        ip_address,
        user_agent
      FROM recipients
      WHERE campaign_id = ? AND submitted = 1
      ORDER BY submitted_at DESC
    `, [req.params.campaignId]);

    // Parse submitted data
    submissions.forEach(sub => {
      try {
        sub.submitted_data = JSON.parse(sub.submitted_data);
      } catch (e) {
        sub.submitted_data = {};
      }
    });

    res.json(submissions);
  } catch (err) {
    console.error('Error fetching submissions:', err);
    res.status(500).json({ error: 'Failed to fetch submissions' });
  }
});

module.exports = router;