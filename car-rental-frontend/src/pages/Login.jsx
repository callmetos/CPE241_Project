// src/pages/Login.jsx
import React, { useState, useContext } from 'react';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import { login as loginService } from '../services/authService';
import { AuthContext } from '../context/AuthContext';
import './Login.css'; // Keep CSS import
// --- No logo import needed here ---

const Login = () => {
  // ... (state and handlers same as previous 'final' version) ...
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const navigate = useNavigate();
    const location = useLocation();
    const { login: loginContext } = useContext(AuthContext);
    const from = location.state?.from?.pathname || "/"; // Default to home

    const handleLogin = async (e) => {
        e.preventDefault();
        setError(''); setIsSubmitting(true);
        if (!email || !password) { setError('Email and Password are required'); setIsSubmitting(false); return; }
        try {
            const token = await loginService(email, password);
            await loginContext(token);
            navigate(from, { replace: true });
        } catch (err) { setError(err.message || 'Login failed.'); }
        finally { setIsSubmitting(false); }
    };

  return (
    <>
      <div className="background-image"></div>
      {/* --- NO Standalone Logo Header Here --- */}
      <div className="login-container">
        <h2>Customer Login</h2>
        {error && <p className="error">{error}</p>}
        <form onSubmit={handleLogin}>
           <label htmlFor="email">Email</label>
           <input id="email" type="email" placeholder="Enter your email" value={email} onChange={(e) => setEmail(e.target.value)} disabled={isSubmitting} required />
           <label htmlFor="password">Password</label>
           <input id="password" type="password" placeholder="Enter your password" value={password} onChange={(e) => setPassword(e.target.value)} disabled={isSubmitting} required/>
           <div className="forgot-password"><Link to="/forgot-password">forgot password?</Link></div>
           <button type="submit" disabled={isSubmitting}>{isSubmitting ? 'Logging in...' : 'Log in'}</button>
        </form>
        <p>Don't have an account? <Link to="/signup">Sign Up</Link></p>
      </div>
    </>
  );
};

export default Login;