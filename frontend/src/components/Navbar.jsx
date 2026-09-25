import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { Activity, PlusCircle, LayoutDashboard, LogOut, LogIn, UserPlus } from 'lucide-react';

export default function Navbar() {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <nav className="navbar">
      <div className="navbar-inner">
        <Link to="/" className="brand-logo" id="nav-brand-logo">
          <div className="brand-icon">
            <Activity size={22} strokeWidth={2.5} />
          </div>
          <span>PulseVote</span>
        </Link>

        <div className="nav-links">
          {isAuthenticated ? (
            <>
              <Link to="/dashboard" className="btn btn-outline btn-sm" id="nav-btn-dashboard">
                <LayoutDashboard size={16} />
                <span>Dashboard</span>
              </Link>

              <Link to="/create" className="btn btn-primary btn-sm" id="nav-btn-create-poll">
                <PlusCircle size={16} />
                <span>Create Poll</span>
              </Link>

              <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', marginLeft: '0.5rem' }}>
                <span style={{ fontSize: '0.88rem', color: 'var(--text-secondary)' }}>
                  Hi, <strong style={{ color: 'var(--text-primary)' }}>{user?.name}</strong>
                </span>
                <button onClick={handleLogout} className="btn btn-secondary btn-sm" id="nav-btn-logout" title="Log Out">
                  <LogOut size={16} />
                </button>
              </div>
            </>
          ) : (
            <>
              <Link to="/login" className="btn btn-secondary btn-sm" id="nav-btn-login">
                <LogIn size={16} />
                <span>Login</span>
              </Link>
              <Link to="/signup" className="btn btn-primary btn-sm" id="nav-btn-signup">
                <UserPlus size={16} />
                <span>Sign Up</span>
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
}
