import React, { createContext, useState, useEffect, useCallback } from 'react';
import { jwtDecode } from 'jwt-decode';
// Removed direct API calls, they should be imported from apiService if needed here
// import { fetchUserProfile } from '../services/apiService'; // Example if needed

export const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [token, setToken] = useState(() => localStorage.getItem('jwt_token'));
  const [user, setUser] = useState(null); // Can be customer or employee info
  const [userType, setUserType] = useState(null); // 'customer' or 'employee'
  const [role, setRole] = useState(null); // 'customer', 'admin', 'manager'
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true); // Start loading until token processed

  /**
   * Processes a JWT token to update authentication state.
   * Decodes the token, checks expiration, sets user type, role, and basic user info.
   * Clears state and storage if token is invalid or expired.
   * @param {string|null} currentToken The JWT token string or null.
   */
  const processToken = useCallback(async (currentToken) => {
    setLoading(true); // Ensure loading state is true during processing
    if (!currentToken) {
      setUser(null);
      setUserType(null);
      setRole(null);
      setIsAuthenticated(false);
      setToken(null); // Clear token state
      localStorage.removeItem('jwt_token');
      setLoading(false); // Finished processing (no token)
      return;
    }

    try {
      const decoded = jwtDecode(currentToken);

      // Check if token is expired
      if (decoded.exp * 1000 < Date.now()) {
        throw new Error("Token expired");
      }

      // Token is valid and not expired
      localStorage.setItem('jwt_token', currentToken);
      setToken(currentToken);
      setUserType(decoded.user_type);
      setRole(decoded.role);
      setIsAuthenticated(true);

      // Set basic user info directly from token claims
      // More detailed profile fetching can happen in specific components if needed
      if (decoded.user_type === 'customer') {
        setUser({ id: decoded.customer_id, email: decoded.email, name: 'Customer' }); // Use email as placeholder name initially
      } else if (decoded.user_type === 'employee') {
        setUser({ id: decoded.employee_id, email: decoded.email, name: 'Employee' }); // Use email as placeholder name initially
      } else {
        throw new Error("Unknown user type in token");
      }

    } catch (error) {
      console.error("Token processing failed:", error.message);
      // Clear everything if token is invalid or expired
      localStorage.removeItem('jwt_token');
      setToken(null);
      setUser(null);
      setUserType(null);
      setRole(null);
      setIsAuthenticated(false);
    } finally {
        setLoading(false); // Finished processing token (valid or invalid)
    }
  }, []); // No dependencies needed for processToken itself

  // Initialize auth state on component mount
  useEffect(() => {
    const storedToken = localStorage.getItem('jwt_token');
    processToken(storedToken);
    // No need for setLoading(false) here, processToken handles it
  }, [processToken]); // Rerun if processToken definition changes (shouldn't)

  /**
   * Logs in a user by processing a new token.
   * @param {string} newToken The new JWT token received after successful login.
   */
  const login = useCallback(async (newToken) => {
    // processToken already sets loading state
    await processToken(newToken);
    // No need for setLoading(false) here, processToken handles it
  }, [processToken]);

  /**
   * Logs out the current user, clearing state and storage.
   */
  const logout = useCallback(() => {
    localStorage.removeItem('jwt_token');
    setToken(null);
    setUser(null);
    setUserType(null);
    setRole(null);
    setIsAuthenticated(false);
    // Redirect logic should happen in the component calling logout (e.g., Navbar or Logout page)
    console.log("User logged out from context.");
  }, []);

  // Value provided by the context
  const value = {
    token,
    user,
    userType,
    role,
    isAuthenticated,
    loading, // Provide loading state for components
    login,
    logout
  };

  return (
    <AuthContext.Provider value={value}>
      {/* Render children only after initial loading is complete? Optional. */}
      {/* {loading ? <LoadingSpinner /> : children} */}
      {children} {/* Render children immediately, components can check loading state */}
    </AuthContext.Provider>
  );
};
