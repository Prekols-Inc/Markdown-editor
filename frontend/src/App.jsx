import { Routes, Route, Navigate } from 'react-router-dom';
import { useState } from 'react';
import LoginPage from './components/LoginPage';
import EditorPage from './components/EditorPage';
import SignupPage from './components/SignupPage';

export default function App() {
  const [isAuth, setIsAuth] = useState(false);

  return (
    <Routes>
      <Route path="/login" element={<LoginPage onLogin={() => setIsAuth(true)} />} />
      <Route path="/signup" element={<SignupPage />} />
      <Route
        path="/editor"
        element={isAuth ? <EditorPage /> : <Navigate to="/login" replace />}
      />
      <Route path="*" element={<Navigate to="/login" replace />} />
    </Routes>
  );
}
