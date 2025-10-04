import { Routes, Route, Navigate } from 'react-router-dom';
import { useState } from 'react';
import LoginPage from './components/LoginPage';
import MarkdownApp from './components/MarkdownApp';

export default function App() {
  const [isAuth, setIsAuth] = useState(false);

  return (
    <Routes>
      <Route path="/login" element={<LoginPage onLogin={() => setIsAuth(true)} />} />
      <Route
        path="/editor"
        element={isAuth ? <MarkdownApp /> : <Navigate to="/login" replace />}
      />
      <Route path="*" element={<Navigate to="/login" replace />} />
    </Routes>
  );
}
