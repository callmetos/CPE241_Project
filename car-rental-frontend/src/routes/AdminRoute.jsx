import React, { useContext } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
// *** แก้ไข Path ให้ถูกต้อง ***
import { AuthContext } from '../context/AuthContext.jsx';
import LoadingSpinner from '../components/LoadingSpinner.jsx';

/**
 * Route guard สำหรับหน้า Admin/Manager
 * ตรวจสอบว่าผู้ใช้ login แล้ว, เป็น 'employee', และมี role ที่อนุญาต
 * Redirect ไปหน้า employee login ถ้าไม่ผ่าน
 */
const AdminRoute = ({ children, allowedRoles = ['admin', 'manager'] }) => {
  const { isAuthenticated, userType, role, loading } = useContext(AuthContext);
  const location = useLocation(); // เก็บ location ปัจจุบันเพื่อ redirect กลับหลัง login

  // แสดง loading spinner ขณะรอตรวจสอบสถานะ authentication
  if (loading) {
    console.log("AdminRoute: Auth context loading...");
    return <LoadingSpinner />;
  }

  // ตรวจสอบเงื่อนไข
  const isAuthorized = isAuthenticated && userType === 'employee' && allowedRoles.includes(role);

  if (!isAuthorized) {
    // Log เหตุผลที่ไม่ได้รับอนุญาต
    console.warn(
      `AdminRoute: Access Denied to ${location.pathname}. \n` +
      `  - Authenticated: ${isAuthenticated}\n` +
      `  - User Type: ${userType} (Required: 'employee')\n` +
      `  - Role: ${role} (Required one of: [${allowedRoles.join(', ')}])`
    );

    // Redirect ไปหน้า employee login พร้อมเก็บ location เดิม
    return <Navigate to="/admin/login" state={{ from: location }} replace />;
  }

  // ถ้าได้รับอนุญาต แสดง child components
  console.log(`AdminRoute: Access Granted to ${location.pathname}. Role: ${role}`);
  return children;
};

export default AdminRoute;
