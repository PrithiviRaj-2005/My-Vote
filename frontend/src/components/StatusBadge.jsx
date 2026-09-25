import React from 'react';

export default function StatusBadge({ isActive, isConnected }) {
  if (!isActive) {
    return (
      <span className="badge-closed" id="badge-poll-status-closed">
        <span>Closed</span>
      </span>
    );
  }

  return (
    <span className="badge-live" id="badge-poll-status-live">
      <span className="badge-pulse-dot"></span>
      <span>{isConnected ? 'Live & Realtime' : 'Active'}</span>
    </span>
  );
}
