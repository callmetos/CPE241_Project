// src/pages/SignUp.jsx
import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { signup } from '../services/authService';
import './SignUp.css'; // Keep CSS import
// --- No logo import needed here ---

const SignUp = () => {
   // ... (state and handlers same as previous 'final' version) ...
    const [email, setEmail] = useState('');
    const [name, setName] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const navigate = useNavigate();

    const handleSignUp = async (e) => {
        e.preventDefault();
        setError(''); setIsSubmitting(true);
        if (!email || !name || !password) { setError('All fields are required'); setIsSubmitting(false); return; }
        if (password.length < 6) { setError('Password must be at least 6 characters long.'); setIsSubmitting(false); return; }
        try {
            await signup(email, name, password);
            alert('Sign up successful! Please log in.');
            navigate('/login');
        } catch (err) { setError(err.message || 'Sign Up failed.'); }
        finally { setIsSubmitting(false); }
    };

  return (
    <>
      <div className="background-image"></div>
       {/* --- NO Standalone Logo Header Here --- */}
      <div className="signup-container">
        <h2>Create Your Account</h2>
        {error && <p className="error">{error}</p>}
        <form onSubmit={handleSignUp}>
           <label htmlFor="name">Name</label>
           <input id="name" type="text" placeholder="Enter your full name" value={name} onChange={(e) => setName(e.target.value)} disabled={isSubmitting} required />
           <label htmlFor="email">E-mail</label>
           <input id="email" type="email" placeholder="Enter your email address" value={email} onChange={(e) => setEmail(e.target.value)} disabled={isSubmitting} required />
           <label htmlFor="password">Password (min. 6 characters)</label>
           <input id="password" type="password" placeholder="Create a password" value={password} onChange={(e) => setPassword(e.target.value)} disabled={isSubmitting} required />
           <button type="submit" disabled={isSubmitting}>{isSubmitting ? 'Signing Up...' : 'Sign Up'}</button>
        </form>
        <p>Already have an account? <Link to="/login">Log In</Link></p>
      </div>
    </>
  );
};

export default SignUp;