import React from 'react';
import { Activity } from 'lucide-react';

export default function Footer() {
  return (
    <footer className="footer">
      <div style={{ maxWidth: '1200px', margin: '0 auto' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
          <Activity size={18} color="var(--primary)" />
          <strong style={{ color: 'var(--text-primary)' }}>My Vote</strong>
          <span>– Real-time Live Polling Engine</span>
        </div>
        <p style={{ color: 'var(--text-muted)', fontSize: '0.82rem' }}>
          Engineered with Go &bull; Gin &bull; Gorilla WebSocket &bull; Redis Pub/Sub &bull; MongoDB &bull; React
        </p>
      </div>
    </footer>
  );
}
