import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Activity, PlusCircle, LogIn, UserPlus, Zap, BarChart2, Radio, CheckCircle, ArrowRight } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

export default function Home() {
  const { isAuthenticated } = useAuth();

  // Demo interactive animation state for preview card
  const [demoVotes, setDemoVotes] = useState({
    Go: 42,
    Python: 38,
    JavaScript: 20,
  });

  // Periodically increment demo votes to show live feel on homepage
  useEffect(() => {
    const timer = setInterval(() => {
      const keys = ['Go', 'Python', 'JavaScript'];
      const randomKey = keys[Math.floor(Math.random() * keys.length)];
      setDemoVotes((prev) => ({
        ...prev,
        [randomKey]: prev[randomKey] + 1,
      }));
    }, 3500);

    return () => clearInterval(timer);
  }, []);

  const totalDemo = demoVotes.Go + demoVotes.Python + demoVotes.JavaScript;

  return (
    <div style={{ maxWidth: '1080px', margin: '0 auto', paddingTop: '1rem' }}>
      {/* Hero Section */}
      <section style={{ textAlign: 'center', padding: '3.5rem 1rem 4rem 1rem' }}>
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.5rem',
            padding: '0.4rem 1rem',
            borderRadius: 'var(--radius-full)',
            background: 'rgba(99, 102, 241, 0.1)',
            border: '1px solid rgba(99, 102, 241, 0.3)',
            color: '#a5b4fc',
            fontSize: '0.85rem',
            fontWeight: 600,
            marginBottom: '1.5rem',
          }}
        >
          <span className="badge-pulse-dot" />
          <span>Real-time Live Polling Engine</span>
        </div>

        <h1
          style={{
            fontSize: 'clamp(2.5rem, 5vw, 4.2rem)',
            fontWeight: 800,
            letterSpacing: '-0.03em',
            marginBottom: '1.25rem',
            lineHeight: 1.15,
          }}
        >
          My Vote
          <br />
          <span
            style={{
              background: 'linear-gradient(135deg, #6366f1 0%, #06b6d4 100%)',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
            }}
          >
            Live Polling Tool
          </span>
        </h1>

        <p
          style={{
            fontSize: 'clamp(1.1rem, 2vw, 1.35rem)',
            color: 'var(--text-secondary)',
            maxWidth: '650px',
            margin: '0 auto 2.5rem auto',
            lineHeight: 1.6,
          }}
        >
          Create polls. Share them. Collect votes. See results instantly.
          <br />
          <span style={{ fontSize: '0.95rem', color: 'var(--text-muted)' }}>
            Zero page refreshes. Real-time updates delivered instantly to every connected audience screen.
          </span>
        </p>

        {/* CTA Buttons */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            gap: '1rem',
            flexWrap: 'wrap',
          }}
        >
          <Link
            to={isAuthenticated ? '/create' : '/signup'}
            className="btn btn-primary btn-lg"
            id="home-btn-create-poll"
          >
            <PlusCircle size={20} />
            <span>Create Poll</span>
            <ArrowRight size={18} />
          </Link>

          {!isAuthenticated && (
            <>
              <Link to="/login" className="btn btn-secondary btn-lg" id="home-btn-login">
                <LogIn size={20} />
                <span>Login</span>
              </Link>

              <Link to="/signup" className="btn btn-outline btn-lg" id="home-btn-signup">
                <UserPlus size={20} />
                <span>Sign Up</span>
              </Link>
            </>
          )}
        </div>
      </section>

      {/* Live Interactive Interactive Preview Demo Card */}
      <section style={{ marginBottom: '4.5rem' }}>
        <div
          className="card"
          style={{
            maxWidth: '680px',
            margin: '0 auto',
            border: '1px solid rgba(99, 102, 241, 0.3)',
            boxShadow: '0 20px 40px -10px rgba(99, 102, 241, 0.25)',
            background: 'linear-gradient(180deg, rgba(22, 28, 45, 0.9) 0%, rgba(15, 23, 42, 0.95) 100%)',
          }}
        >
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              borderBottom: '1px solid var(--border-subtle)',
              paddingBottom: '1rem',
              marginBottom: '1.25rem',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
              <div
                style={{
                  width: '10px',
                  height: '10px',
                  borderRadius: '50%',
                  background: '#ef4444',
                }}
              />
              <div
                style={{
                  width: '10px',
                  height: '10px',
                  borderRadius: '50%',
                  background: '#f59e0b',
                }}
              />
              <div
                style={{
                  width: '10px',
                  height: '10px',
                  borderRadius: '50%',
                  background: '#10b981',
                }}
              />
              <span style={{ fontSize: '0.82rem', color: 'var(--text-muted)', marginLeft: '0.5rem' }}>
                LIVE PREVIEW SIMULATOR
              </span>
            </div>
            <div className="badge-live">
              <span className="badge-pulse-dot" />
              <span>WebSocket Stream Active</span>
            </div>
          </div>

          <h3 style={{ fontSize: '1.35rem', marginBottom: '1.25rem' }}>
            Which backend stack excels at high-concurrency real-time systems?
          </h3>

          {Object.entries(demoVotes).map(([key, count]) => {
            const pct = Math.round((count / totalDemo) * 100);
            return (
              <div key={key} style={{ marginBottom: '1.15rem' }}>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    marginBottom: '0.45rem',
                    fontSize: '0.95rem',
                  }}
                >
                  <span style={{ fontWeight: 600 }}>{key}</span>
                  <div style={{ display: 'flex', gap: '0.75rem', color: 'var(--text-secondary)' }}>
                    <span>{count} votes</span>
                    <strong style={{ color: 'var(--primary)', width: '38px', textAlign: 'right' }}>
                      {pct}%
                    </strong>
                  </div>
                </div>
                <div className="progress-track">
                  <div
                    className="progress-fill"
                    style={{
                      width: `${pct}%`,
                      background:
                        key === 'Go'
                          ? 'linear-gradient(90deg, #6366f1 0%, #06b6d4 100%)'
                          : 'linear-gradient(90deg, #4f46e5 0%, #818cf8 100%)',
                    }}
                  />
                </div>
              </div>
            );
          })}

          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              marginTop: '1.5rem',
              paddingTop: '1rem',
              borderTop: '1px solid var(--border-subtle)',
              fontSize: '0.85rem',
              color: 'var(--text-muted)',
            }}
          >
            <span>Total Simulated Votes: {totalDemo}</span>
            <span style={{ color: '#34d399' }}>✓ Live broadcast active</span>
          </div>
        </div>
      </section>

      {/* Feature Highlights Grid */}
      <section style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1.5rem', marginBottom: '4rem' }}>
        <div className="card">
          <div style={{ color: 'var(--primary)', marginBottom: '1rem' }}>
            <Zap size={32} />
          </div>
          <h3 style={{ fontSize: '1.2rem', marginBottom: '0.5rem' }}>Instant Redis Pub/Sub</h3>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.92rem' }}>
            Every cast vote instantly increments Redis live hashes and triggers high-throughput sub-millisecond pub/sub notifications.
          </p>
        </div>

        <div className="card">
          <div style={{ color: 'var(--accent-cyan)', marginBottom: '1rem' }}>
            <Radio size={32} />
          </div>
          <h3 style={{ fontSize: '1.2rem', marginBottom: '0.5rem' }}>Gorilla WebSockets</h3>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.92rem' }}>
            Bidirectional WebSocket channels broadcast live count deltas directly to every connected audience browser without reload.
          </p>
        </div>

        <div className="card">
          <div style={{ color: 'var(--accent-emerald)', marginBottom: '1rem' }}>
            <CheckCircle size={32} />
          </div>
          <h3 style={{ fontSize: '1.2rem', marginBottom: '0.5rem' }}>Permanent MongoDB Storage</h3>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.92rem' }}>
            Users, poll structures, and every single vote are stored persistently with ACID guarantees and complete auditability.
          </p>
        </div>
      </section>
    </div>
  );
}
