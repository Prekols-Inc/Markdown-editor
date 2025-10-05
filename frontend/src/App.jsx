import { Routes, Route, Navigate } from "react-router-dom";
import { useState, useEffect } from "react";
import LoginPage from "./components/LoginPage";
import EditorPage from "./components/EditorPage";
import API from "./API";

export default function App() {
  const [isAuth, setIsAuth] = useState(null);

  useEffect(() => {
    async function checkAuth() {
      try {
        const res = await API.AUTH.get("/v1/check_auth");
        if (res.status === 200) {
          setIsAuth(true);
        } else {
          setIsAuth(false);
        }
      } catch (err) {
        setIsAuth(false);
      }
    }

    checkAuth();
  }, []);

  if (isAuth === null) {
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
            <LoginPage onLogin={() => setIsAuth(true)} />
          )
        }
      />
      <Route
        path="/editor"
        element={
          isAuth ? (
            <EditorPage onLogout={() => setIsAuth(false)} />
          ) : (
            <Navigate to="/login" replace />
          )
        }
      />
      <Route path="*" element={<Navigate to="/login" replace />} />
    </Routes>
  );
}
