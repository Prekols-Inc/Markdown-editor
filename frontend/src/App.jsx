import { Routes, Route, Navigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import LoginPage from './components/LoginPage';
import MarkdownApp from './components/MarkdownApp';
import API from './API';

export default function App() {
  const [isAuth, setIsAuth] = useState(null);
  const [editorMode, setEditorMode] = useState(
    localStorage.getItem('editorMode')
  );

  useEffect(() => {
    async function checkAuth() {
      try {
        const res = await API.AUTH.get("/v1/check_auth");
        if (res.status === 200) {
          setIsAuth(true);
          localStorage.setItem('editorMode', 'auth');
          setEditorMode('auth');
        } else {
          setIsAuth(false);
        }
      } catch (err) {
        setIsAuth(false);
      }
    }

    if (editorMode === 'unauth') {
      setIsAuth(false);
      return;
    }

    checkAuth();
  }, [editorMode]);


  if (isAuth === null && editorMode !== 'unauth') {
    return <div>Проверка авторизации...</div>;
  }

  return (
    <Routes>
      <Route
        path="/login"
        element={
          isAuth ? (
            <Navigate to="/editor" replace />
          ) : (
            <LoginPage
              onLogin={() => {
                localStorage.setItem('editorMode', 'auth');
                setEditorMode('auth');
                setIsAuth(true);
              }}
            />
          )
        }
      />
      <Route
        path="/editor"
        element={
          isAuth || editorMode === 'unauth' ? (
            <MarkdownApp />
          ) : (
            <Navigate to="/login" replace />
          )
        }
      />
      <Route path="*" element={<Navigate to="/login" replace />} />
    </Routes>
  );
}
