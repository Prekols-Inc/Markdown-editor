import React from "react";
import "../styles/LogoutConfirmModal.css";

export default function LogoutConfirmModal({ open, onClose, onConfirm }) {
    if (!open) return null;

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal-window" onClick={(e) => e.stopPropagation()}>
                <h2 className="modal-title">Вы действительно хотите выйти?</h2>

                <div className="modal-buttons">
                    <button className="modal-btn cancel" onClick={onClose}>
                        Отмена
                    </button>
                    <button className="modal-btn confirm" onClick={onConfirm}>
                        Выйти
                    </button>
                </div>
            </div>
        </div>
    );
}
