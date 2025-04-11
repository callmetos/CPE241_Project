import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { login } from '../services/authService';
import './Login.css'; // Make sure to import your CSS
import loginlogo from '../assets/logo.png'; // Replace with your actual logo pathX

const Login = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleLogin = async (e) => {
    e.preventDefault();

    if (!email || !password) {
      setError('Email and Password are required');
      return;
    }

    try {
      const token = await login(email, password);
      localStorage.setItem('jwt_token', token);
      navigate('/car-rental');
    } catch (err) {
      setError('Invalid email or password');
    }
  };

  return (
    <>
      <div className="background-image"></div>
      <div className="login-container">
        <img src={loginlogo} alt="GannatRat a Car Logo" className="loginlogo" />
        
        {error && <p className="error">{error}</p>}
        
        <form onSubmit={handleLogin}>
          <label htmlFor="email">Email</label>
          <input
            id="email"
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          
          <label htmlFor="password">Password</label>
          <input
            id="password"
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          
          <div className="forgot-password">
            <a href="/forgot-password">forgot password</a>
          </div>
          
          <button type="submit">Log in</button>
        </form>
        <p>Don't have an account? <a href="/signup">Sign Up</a></p>
      </div>
    </>
  );
};

export default Login;