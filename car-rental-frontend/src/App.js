import React from 'react';
import { Routes, Route } from 'react-router-dom'; // Use Routes and Route without Router
import Login from './components/Login';
import Register from './components/Register';
import CarList from './components/CarList';
import Dashboard from './components/Dashboard';

const App = () => {
  return (
    <div>
      <h1>Car Rental Management</h1>
      <Routes>
        <Route path="/" element={<CarList />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/dashboard" element={<Dashboard />} />
      </Routes>
    </div>
  );
};

export default App;
