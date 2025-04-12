import React, { createContext, useState, useEffect, useCallback } from 'react';
import { jwtDecode } from 'jwt-decode';
import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

export const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [token, setToken] = useState(() => localStorage.getItem('jwt_token'));
  const [user, setUser] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(!!token);
  const [loading, setLoading] = useState(true);

  const fetchUserProfile = useCallback(async (currentToken) => {
    // ... (fetch profile logic same as before) ...
    if (!currentToken) { setUser(null); setIsAuthenticated(false); return; }
    try {
        const response = await axios.get(`${API_URL}/me/profile`, { headers: { Authorization: `Bearer ${currentToken}` } });
        const decodedToken = jwtDecode(currentToken);
        setUser({ ...response.data, role: decodedToken.role, userType: decodedToken.user_type });
        setIsAuthenticated(true);
    } catch (error) {
        localStorage.removeItem('jwt_token'); setToken(null); setUser(null); setIsAuthenticated(false);
        console.error("fetchUserProfile failed:", error.message);
    }
  }, []);

  useEffect(() => {
    // ... (initialization logic same as before) ...
    const initializeAuth = async () => {
        setLoading(true);
        const storedToken = localStorage.getItem('jwt_token');
        if (storedToken) {
            try {
                const decoded = jwtDecode(storedToken);
                if (decoded.exp * 1000 > Date.now()) { setToken(storedToken); await fetchUserProfile(storedToken); }
                else { localStorage.removeItem('jwt_token'); setIsAuthenticated(false); setUser(null); setToken(null); }
            } catch (e) { localStorage.removeItem('jwt_token'); setIsAuthenticated(false); setUser(null); setToken(null); }
        } else { setIsAuthenticated(false); setUser(null); }
        setLoading(false);
    };
    initializeAuth();
  }, [fetchUserProfile]);

  const login = useCallback(async (newToken) => {
    // ... (login logic same as before) ...
     setLoading(true);
     try {
        const decoded = jwtDecode(newToken);
        if (decoded.exp * 1000 > Date.now()) { localStorage.setItem('jwt_token', newToken); setToken(newToken); await fetchUserProfile(newToken); }
        else { throw new Error("Login failed: Token is expired."); }
     } catch (error) { localStorage.removeItem('jwt_token'); setToken(null); setUser(null); setIsAuthenticated(false); throw error; }
     finally { setLoading(false); }
  }, [fetchUserProfile]);

  const logout = useCallback(() => {
    // ... (logout logic same as before) ...
    localStorage.removeItem('jwt_token'); setToken(null); setUser(null); setIsAuthenticated(false);
  }, []);

  const value = { token, user, isAuthenticated, loading, login, logout };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};