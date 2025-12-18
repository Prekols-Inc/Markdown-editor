import { useState } from "react";
import {
  FilePlus2,
  Save,
  Download,
  Settings2,
  LogOut,
  ChevronsLeft,
  ChevronsRight,
  PanelRightClose,
  PanelRightOpen,
} from "lucide-react";

import LogoutConfirmModal from "./LogoutConfirmModal";
import OptionsModal from "./OptionsModal";

export default function AppTopBar({
  isUnauth,
  sidebarOpen,
  onToggleSidebar,
  showPreview,
  onTogglePreview,
  current,
  unsaved,
  onNewFile,
  onSave,
  onDownload,
  onLogout,
  options,
  onOptionsChange,
}) {
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false);
  const [showOptions, setShowOptions] = useState(false);

  const fileLabel = current?.name ? `${current.name}${unsaved ? " ●" : ""}` : "Файл не выбран";

  return (
    <>
      <header className="app-topbar">
        <div className="topbar-left">
          <button
            className="btn btn-sm btn-icon ghost"
            onClick={onToggleSidebar}
            title={sidebarOpen ? "Скрыть список файлов" : "Показать список файлов"}
            type="button"
          >
            {sidebarOpen ? (
              <ChevronsLeft size={18} strokeWidth={1.75} />
            ) : (
              <ChevronsRight size={18} strokeWidth={1.75} />
            )}
          </button>

          <div className="topbar-file" title={fileLabel}>
            {fileLabel}
          </div>

          <div className="topbar-actions">
            <button
              className="btn btn-sm btn-icon"
              onClick={onNewFile}
              title="Новый файл"
              type="button"
            >
              <FilePlus2 size={18} strokeWidth={1.75} />
            </button>

            <button
              className="btn btn-sm btn-icon"
              disabled={!current && !unsaved}
              onClick={() => onSave?.()}
              title="Сохранить"
              type="button"
            >
              <Save size={18} strokeWidth={1.75} />
            </button>

            <button
              className="btn btn-sm btn-icon"
              disabled={!current}
              onClick={() => onDownload?.()}
              title="Скачать"
              type="button"
            >
              <Download size={18} strokeWidth={1.75} />
            </button>

            <button
              className="btn btn-sm btn-icon"
              onClick={() => setShowOptions(true)}
              title="Options"
              type="button"
            >
              <Settings2 size={18} strokeWidth={1.75} />
            </button>

          </div>
        </div>

        <div className="topbar-right">
          <button
            className="btn btn-sm btn-icon ghost"
            onClick={onTogglePreview}
            title={showPreview ? "Скрыть превью" : "Показать превью"}
            type="button"
          >
            {showPreview ? (
              <PanelRightClose size={18} strokeWidth={1.75} />
            ) : (
              <PanelRightOpen size={18} strokeWidth={1.75} />
            )}
          </button>

          <button
            className="btn btn-sm btn-icon danger"
            onClick={() => setShowLogoutConfirm(true)}
            title="Выйти"
            type="button"
          >
            <LogOut size={18} strokeWidth={1.75} />
          </button>
        </div>
      </header>

      <OptionsModal
        open={showOptions}
        onClose={() => setShowOptions(false)}
        value={options}
        onChange={onOptionsChange}
      />

      <LogoutConfirmModal
        open={showLogoutConfirm}
        onClose={() => setShowLogoutConfirm(false)}
        onConfirm={onLogout}
      />
    </>
  );
}
