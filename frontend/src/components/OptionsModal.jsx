import React from "react";
import "../styles/NewFileModal.css";
import OptionsEditor from "./OptionsEditor";

export default function OptionsModal({ open, onClose, value, onChange }) {
  if (!open) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div
        className="modal-window options-modal"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="modal-title">Options</h2>

        <div className="options-modal-body">
          <OptionsEditor value={value} onChange={onChange} />
        </div>

        <div className="modal-buttons">
          <button className="modal-btn confirm" onClick={onClose}>
            Готово
          </button>
        </div>
      </div>
    </div>
  );
}
