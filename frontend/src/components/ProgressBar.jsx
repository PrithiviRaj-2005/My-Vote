import React from 'react';

export default function ProgressBar({ option, isUserVoted, isLeader }) {
  const percentage = Math.round(option.percentage || 0);

  return (
    <div className="result-row" id={`result-option-${option.id}`}>
      <div className="result-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <span className="result-option-name">{option.text}</span>
          {isUserVoted && (
            <span
              style={{
                fontSize: '0.72rem',
                padding: '0.15rem 0.5rem',
                borderRadius: '9999px',
                background: 'rgba(99, 102, 241, 0.2)',
                color: '#818cf8',
                border: '1px solid rgba(99, 102, 241, 0.4)',
                fontWeight: 600,
              }}
            >
              Your Vote
            </span>
          )}
        </div>
        <div className="result-stats">
          <span>{option.votes || 0} {option.votes === 1 ? 'vote' : 'votes'}</span>
          <span className="result-percentage">{percentage}%</span>
        </div>
      </div>

      <div className="progress-track">
        <div
          className="progress-fill"
          style={{
            width: `${percentage}%`,
            background: isLeader
              ? 'linear-gradient(90deg, #6366f1 0%, #06b6d4 100%)'
              : 'linear-gradient(90deg, #4f46e5 0%, #6366f1 100%)',
          }}
        />
      </div>
    </div>
  );
}
