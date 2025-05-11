import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider } from './context/AuthContext.jsx';
import Navbar from './components/Navbar.jsx';
import Footer from './components/Footer.jsx';
import AdminLayout from './layouts/AdminLayout.jsx';
import PrivateRoute from './routes/PrivateRoute.jsx';
import AdminRoute from './routes/AdminRoute.jsx';
import Home from './pages/Home.jsx';
import SignUp from './pages/SignUp.jsx';
import Login from './pages/Login.jsx';
import Logout from './pages/Logout.jsx';
import CarRental from './pages/CarRental.jsx';
import Profile from './pages/Profile.jsx';
import RentalHistory from './pages/RentalHistory.jsx';
import NotFound from './pages/NotFound.jsx';
import CheckoutSummary from './pages/CheckoutSummary.jsx';
import CheckoutUserInfo from './pages/CheckoutUserInfo.jsx';
import CheckoutPaymentUpload from './pages/CheckoutPaymentUpload.jsx';
import EmployeeLogin from './pages/admin/EmployeeLogin.jsx';
import AdminDashboard from './pages/admin/AdminDashboard.jsx';
import BranchManagement from './pages/admin/BranchManagement.jsx';
import CarManagement from './pages/admin/CarManagement.jsx';
import CustomerManagement from './pages/admin/CustomerManagement.jsx';
import RentalManagement from './pages/admin/RentalManagement.jsx';
import UserManagement from './pages/admin/UserManagement.jsx';
import SlipVerification from './pages/admin/SlipVerification.jsx';
import ReviewManagement from './pages/admin/ReviewManagement.jsx'; // Import หน้าใหม่
import Thankyou from './pages/Thankyou.jsx';
import Reports from './pages/admin/Reports.jsx';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

const DefaultLayout = ({ children }) => (
  <div style={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
    <Navbar />
    <main style={{ flex: 1 }}> {/* Ensure main content takes available space */}
      {children}
    </main>
    <Footer />
  </div>
);

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <Router>
          <Routes>
            <Route path="/" element={<DefaultLayout><Home /></DefaultLayout>} />
            <Route path="/signup" element={<DefaultLayout><SignUp /></DefaultLayout>} />
            <Route path="/login" element={<DefaultLayout><Login /></DefaultLayout>} />
            <Route path="/logout" element={<Logout />} />
            <Route path="/thankyou" element={<DefaultLayout><Thankyou /></DefaultLayout>} />
            <Route path="/promotions" element={<DefaultLayout><div>Promotions Placeholder</div></DefaultLayout>} />

            <Route path="/rental/:rentalType" element={<PrivateRoute><DefaultLayout><CarRental /></DefaultLayout></PrivateRoute>} />
            <Route path="/profile" element={<PrivateRoute><DefaultLayout><Profile /></DefaultLayout></PrivateRoute>} />
            <Route path="/rental-history" element={<PrivateRoute><DefaultLayout><RentalHistory /></DefaultLayout></PrivateRoute>} />
            <Route path="/checkout/:rentalId/summary" element={<PrivateRoute><DefaultLayout><CheckoutSummary /></DefaultLayout></PrivateRoute>} />
            <Route path="/checkout/:rentalId/user-info" element={<PrivateRoute><DefaultLayout><CheckoutUserInfo /></DefaultLayout></PrivateRoute>} />
            <Route path="/checkout/:rentalId/payment-upload" element={<PrivateRoute><DefaultLayout><CheckoutPaymentUpload /></DefaultLayout></PrivateRoute>} />

            <Route path="/admin/login" element={<EmployeeLogin />} />

            <Route
              path="/admin"
              element={
                <AdminRoute>
                  <AdminLayout />
                </AdminRoute>
              }
            >
              <Route index element={<Navigate to="dashboard" replace />} />
              <Route path="dashboard" element={<AdminDashboard />} />
              <Route path="branches" element={<BranchManagement />} />
              <Route path="cars" element={<CarManagement />} />
              <Route path="customers" element={<CustomerManagement />} />
              <Route path="rentals" element={<RentalManagement />} />
              <Route path="reviews" element={<ReviewManagement />} /> {/* เพิ่ม Route สำหรับ Review Management */}
              <Route path="verify-slips" element={<SlipVerification />} />
              <Route path="reports" element={<Reports />} />
              <Route
                path="users"
                element={
                  <AdminRoute allowedRoles={['admin']}>
                    <UserManagement />
                  </AdminRoute>
                }
              />
            </Route>
            <Route path="*" element={<DefaultLayout><NotFound /></DefaultLayout>} />
          </Routes>
        </Router>
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;
