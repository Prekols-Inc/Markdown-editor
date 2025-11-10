import React, { useState } from "react";
import "../styles/NewFileModal.css";
import { validateFilename } from "../utils";

export default function NewFileModal({ open, onClose, onConfirm }) {
  const [filename, setFilename] = useState("untitled.md");
  const [error, setError] = useState(null);

  if (!open) return null;

  const handleConfirm = () => {
    const v = validateFilename(filename.trim());
    if (!v.ok) {
      setError(v.message);
      return;
    }

    onConfirm(filename.trim());
    setFilename("untitled.md");
    setError(null);
  };

  const handleCancel = () => {
    setFilename("untitled.md");
    setError(null);
    onClose();
  };

  return (
    <div className="modal-overlay" onClick={handleCancel}>
      <div className="modal-window" onClick={(e) => e.stopPropagation()}>
        <h2 className="modal-title">Создать новый файл</h2>

        <input
          type="text"
          value={filename}
          onChange={(e) => {
            setFilename(e.target.value);
            if (error) setError(null);
          }}
          placeholder="Введите имя файла"
          className={`modal-input ${error ? "has-error" : ""}`}
          autoFocus
        />

        <p className="modal-hint">
          Допустимые расширения: .md, .markdown
        </p>

        {error && <p className="modal-error">{error}</p>}

        <div className="modal-buttons">
          <button className="modal-btn cancel" onClick={handleCancel}>
            Отмена
          </button>
          <button className="modal-btn confirm" onClick={handleConfirm}>
            Создать
          </button>
        </div>
      </div>
    </div>
  );
}
