import React, {
  createContext, useCallback, useContext, useMemo, useState
} from "react";

const ToastCtx = createContext(null);


function normalizeMessage(m) {
  if (m == null) return "Произошла ошибка";
  if (typeof m === "string") return m;
  if (m && typeof m === "object") {
    if (typeof m.message === "string") return m.message;
    try {
      return JSON.stringify(m, (k, v) => (v === undefined ? null : v));
    } catch {
      return String(m);
    }
  }
  return String(m);
}

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);

  const remove = useCallback((id) => {
    setToasts((ts) => ts.filter((t) => t.id !== id));
  }, []);

  const push = useCallback(({ message, type = "info", duration = 3000 }) => {
    const id = (crypto?.randomUUID?.() ?? Math.random().toString(36).slice(2));
    const text = normalizeMessage(message);
    setToasts((ts) => [...ts, { id, message: text, type }]);
    if (duration > 0) setTimeout(() => remove(id), duration);
    return id;
  }, [remove]);

  const api = useMemo(() => ({
    push,
    success: (m, opts = {}) => push({ message: normalizeMessage(m), type: "success", ...opts }),
    error:   (m, opts = {}) => push({ message: normalizeMessage(m), type: "error",   ...opts }),
    warn:    (m, opts = {}) => push({ message: normalizeMessage(m), type: "warn",    ...opts }),
    info:    (m, opts = {}) => push({ message: normalizeMessage(m), type: "info",    ...opts }),
  }), [push]);

  return (
    <ToastCtx.Provider value={api}>
      {children}
      <div className="toast-viewport">
        {toasts.map((t) => (
          <div key={t.id} className={`toast toast-${t.type}`} role="status" aria-live="polite">
            <span className="toast-message">{t.message}</span>
            <button className="toast-close" aria-label="Закрыть" onClick={() => remove(t.id)}>×</button>
          </div>
        ))}
      </div>
    </ToastCtx.Provider>
  );
}

export function useToast() {
  const ctx = useContext(ToastCtx);
  if (ctx) return ctx;
  const stub = (m) => console.error(normalizeMessage(m));
  return { push: stub, error: stub, warn: stub, info: stub, success: stub };
}