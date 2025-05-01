import React, { useContext } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
// *** แก้ไข Path ให้ถูกต้อง ***
import { AuthContext } from '../context/AuthContext.jsx';
import LoadingSpinner from '../components/LoadingSpinner.jsx';

/**
 * Route guard สำหรับหน้าเฉพาะ Customer
 * ตรวจสอบว่าผู้ใช้ login แล้ว และเป็น 'customer'
 * Redirect ไปหน้า customer login ถ้าไม่ผ่าน
 */
const PrivateRoute = ({ children }) => {
  const { isAuthenticated, userType, loading } = useContext(AuthContext);
  const location = useLocation(); // เก็บ location ปัจจุบัน

  // แสดง loading spinner ขณะรอตรวจสอบสถานะ authentication
  if (loading) {
    console.log("PrivateRoute: Auth context loading...");
    return <LoadingSpinner />;
  }

  // ตรวจสอบเงื่อนไข
  const isAuthorized = isAuthenticated && userType === 'customer';

  if (!isAuthorized) {
     // Log เหตุผลที่ไม่ได้รับอนุญาต
     console.warn(
        `PrivateRoute: Access Denied to ${location.pathname}. \n` +
        `  - Authenticated: ${isAuthenticated}\n` +
        `  - User Type: ${userType} (Required: 'customer')`
      );
    // Redirect ไปหน้า login พร้อมเก็บ location เดิม
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // ถ้าได้รับอนุญาต แสดง children
  console.log(`PrivateRoute: Access Granted to ${location.pathname}.`);
  return children;
};

export default PrivateRoute;
