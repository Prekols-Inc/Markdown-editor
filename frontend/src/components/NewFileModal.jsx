import React, { useState } from "react";
import "../styles/NewFileModal.css";

export default function NewFileModal({ open, onClose, onConfirm }) {
    const [filename, setFilename] = useState("untitled.md");

    if (!open) return null;

    const handleConfirm = () => {
        if (!filename.trim()) return;
        onConfirm(filename.trim());
        setFilename("untitled.md");
    };

    const handleCancel = () => {
        setFilename("untitled.md");
        onClose();
    };

    return (
        <div className="modal-overlay" onClick={handleCancel}>
            <div className="modal-window" onClick={(e) => e.stopPropagation()}>
                <h2 className="modal-title">Создать новый файл</h2>

                <input
                    type="text"
                    value={filename}
                    onChange={(e) => setFilename(e.target.value)}
                    placeholder="Введите имя файла"
                    className="modal-input"
                    autoFocus
                />

                <p className="modal-hint">
                    Допустимые расширения: .md, .txt
                </p>

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
