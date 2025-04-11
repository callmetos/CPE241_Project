import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { signup } from '../services/authService'; // Assuming you have authService for handling API
import './SignUp.css';

const SignUp = () => {
  const [email, setEmail] = useState('');
  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSignUp = async (e) => {
    e.preventDefault();

    if (!email || !name || !password) {
      setError('All fields are required');
      return;
    }

    try {
      await signup(email, name, password); // Send to backend to create user
      navigate('/login'); // Redirect to login after successful signup
    } catch (err) {
      setError('Sign Up failed. Please try again');
    }
  };

  return (
    <div className="page-wrapper">
      <div className="background-image"></div>
      <div className="signup-container">
        {error && <p className="error">{error}</p>}
        <form onSubmit={handleSignUp}>
          {/* Form fields */}
          <label htmlFor="username">Username</label>
          <input
            id="username"
            type="text"
            placeholder="Username"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          
          <label htmlFor="email">E-mail</label>
          <input
            id="email"
            type="email"
            placeholder="id@email.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          
          <label htmlFor="password">Password</label>
          <input
            id="password"
            type="password"
            placeholder="************"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          
          <button type="submit">Sign in</button>
      </form>
      <p>Already have an account? <a href="/login">Log In</a></p>
    </div>
  </div>
  );
};

export default SignUp;
