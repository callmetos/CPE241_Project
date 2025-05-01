import React, { useState, useContext } from 'react';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import { AuthContext } from '../../context/AuthContext';
import { loginEmployee } from '../../services/apiService'; // Import employee login function
import ErrorMessage from '../../components/ErrorMessage'; // Assuming ErrorMessage is in components
import LoadingSpinner from '../../components/LoadingSpinner'; // Assuming LoadingSpinner is in components

const EmployeeLogin = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const navigate = useNavigate();
    const location = useLocation();
    const { login: loginContext, loading: authLoading } = useContext(AuthContext);

    // Determine where to redirect after successful login
    // If the user was redirected here, 'from' will contain the original target path
    const from = location.state?.from?.pathname || "/admin/dashboard"; // Default to admin dashboard

    const handleLogin = async (e) => {
        e.preventDefault();
        setError('');
        setIsSubmitting(true);
        try {
            console.log(`Attempting employee login for: ${email}`);
            const token = await loginEmployee(email, password); // Use employee login service
            console.log("Employee login successful, received token.");
            await loginContext(token); // Update auth context
            console.log(`Login context updated. Navigating to: ${from}`);
            navigate(from, { replace: true }); // Redirect to the original target or dashboard
        } catch (err) {
            console.error("Employee login failed:", err);
            setError(err.message || 'Login failed. Please check credentials.');
        } finally {
            setIsSubmitting(false);
        }
    };

    // Basic Form Styling (Consider moving to a CSS file)
    const formContainerStyle = { maxWidth: '400px', margin: '50px auto', padding: '30px', border: '1px solid #ccc', borderRadius: '8px', backgroundColor: '#f9f9f9', boxShadow: '0 4px 8px rgba(0,0,0,0.1)' };
    const inputStyle = { width: '100%', padding: '12px', marginBottom: '15px', border: '1px solid #ccc', borderRadius: '4px', boxSizing: 'border-box' };
    const buttonStyle = { width: '100%', padding: '12px', backgroundColor: '#28a745', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '1rem', fontWeight: 'bold' };
    const disabledButtonStyle = { ...buttonStyle, backgroundColor: '#cccccc', cursor: 'not-allowed' };
    const linkContainerStyle = { textAlign: 'center', marginTop: '20px', fontSize: '0.9em' };

    // Show loading spinner if auth context is still loading
    if (authLoading) {
        return <LoadingSpinner />;
    }

    return (
        <div style={formContainerStyle}>
            <h2 style={{ textAlign: 'center', marginBottom: '25px', color: '#333' }}>Employee Login</h2>
            <ErrorMessage message={error} />
            <form onSubmit={handleLogin}>
                <label htmlFor="email" style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>Email</label>
                <input
                    id="email"
                    type="email"
                    placeholder="Enter employee email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    style={inputStyle}
                    required
                    disabled={isSubmitting}
                />
                <label htmlFor="password" style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>Password</label>
                <input
                    id="password"
                    type="password"
                    placeholder="Enter password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    style={inputStyle}
                    required
                    disabled={isSubmitting}
                />
                <button
                    type="submit"
                    style={isSubmitting ? disabledButtonStyle : buttonStyle}
                    disabled={isSubmitting}
                >
                    {isSubmitting ? 'Logging in...' : 'Log in'}
                </button>
            </form>
            <div style={linkContainerStyle}>
                <Link to="/login">Switch to Customer Login</Link>
            </div>
        </div>
    );
};

export default EmployeeLogin;
