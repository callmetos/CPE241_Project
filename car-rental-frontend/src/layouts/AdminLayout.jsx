import React from 'react';
import { Outlet } from 'react-router-dom'; // Import Outlet
import AdminNavbar from '../components/admin/AdminNavbar.jsx'; // Import Admin Navbar
// import AdminSidebar from '../components/admin/AdminSidebar.jsx'; // Optional: ถ้ามี Sidebar

const AdminLayout = () => {
  // --- Styles ---
  const layoutStyle = {
    display: 'flex',
    flexDirection: 'column', // Navbar อยู่บน เนื้อหาอยู่ล่าง
    minHeight: 'calc(100vh)', // ให้เต็มความสูงจอ (อาจปรับตาม Footer)
  };

  const contentStyle = {
    flexGrow: 1, // ให้เนื้อหายืดเต็มพื้นที่ที่เหลือ
    padding: '20px', // ระยะห่างภายในของส่วนเนื้อหา
    // backgroundColor: '#f8f9fa', // สีพื้นหลังอ่อนๆ (Optional)
  };

  // Optional Sidebar Styling (ถ้าใช้)
  /*
  const mainAreaStyle = {
      display: 'flex',
      flexGrow: 1,
  };
  const sidebarStyle = {
      width: '250px', // Fixed width sidebar
      flexShrink: 0,
      backgroundColor: '#e9ecef',
      padding: '15px',
  };
   const mainContentStyle = {
      flexGrow: 1,
      padding: '20px',
      overflowY: 'auto', // Allow content scrolling
  };
  */

  return (
    <div style={layoutStyle}>
      <AdminNavbar />
      {/* Optional: ถ้ามี Sidebar */}
      {/*
      <div style={mainAreaStyle}>
          <aside style={sidebarStyle}>
              <AdminSidebar />
          </aside>
          <main style={mainContentStyle}>
              <Outlet /> // Render nested route's element here
          </main>
      </div>
      */}

      {/* ถ้าไม่มี Sidebar */}
      <main style={contentStyle}>
        <Outlet /> {/* Outlet จะ Render Component ของ Route ที่ซ้อนอยู่ */}
      </main>
      {/* Footer อาจจะไม่จำเป็นในหน้า Admin หรืออาจจะใช้ Footer เดิมก็ได้ */}
      {/* <Footer /> */}
    </div>
  );
};

export default AdminLayout;
