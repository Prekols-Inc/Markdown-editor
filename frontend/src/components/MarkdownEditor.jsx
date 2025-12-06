import { useState, useCallback, useRef } from 'react';
import { Upload } from 'lucide-react';
import { toast, Toaster } from 'react-hot-toast';

export default function MarkdownEditor({ value, onChange, onFileUpload }) {
  const [isDragging, setIsDragging] = useState(false);
  const dragCounter = useRef(0);
  const fileInputRef = useRef(null);

  const handleFile = useCallback((file) => {
    if (!file) return;

    const validExtensions = ['.md', '.markdown', '.txt'];
    const fileName = file.name.toLowerCase();
    const isValid = validExtensions.some(ext => fileName.endsWith(ext));

    if (!isValid) {
      toast.error('Пожалуйста, загрузите файл .md, .markdown или .txt');
      return;
    }

    const reader = new FileReader();
    reader.onload = (e) => {
      const content = e.target.result;
      // Instead of replacing current content, create a new file
      if (onFileUpload) {
        onFileUpload(content, file.name);
      } else {
        // Fallback to old behavior if onFileUpload is not provided
        onChange(content);
      }
    };
    reader.onerror = () => {
      toast.error('Ошибка чтения файла');
    };
    reader.readAsText(file);
  }, [onChange, onFileUpload]);

  const handleDragEnter = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    dragCounter.current++;
    if (e.dataTransfer.items && e.dataTransfer.items.length > 0) {
      setIsDragging(true);
    }
  }, []);

  const handleDragLeave = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    dragCounter.current--;
    if (dragCounter.current === 0) {
      setIsDragging(false);
    }
  }, []);

  const handleDragOver = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  const handleDrop = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    dragCounter.current = 0;

    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      handleFile(files[0]);
    }
  }, [handleFile]);

  const handleFileSelect = useCallback((e) => {
    const file = e.target.files?.[0];
    if (file) {
      handleFile(file);
    }
    e.target.value = '';
  }, [handleFile]);

  const openFileDialog = useCallback(() => {
    fileInputRef.current?.click();
  }, []);

  return (
    <div
      className="editor-container"
      onDragEnter={handleDragEnter}
      onDragLeave={handleDragLeave}
      onDragOver={handleDragOver}
      onDrop={handleDrop}
    >
      <Toaster position="top-right" reverseOrder={false} />
      <textarea
        className="editor"
        value={value}
        onChange={e => onChange(e.target.value)}
        spellCheck={false}
      />

      {/* Скрытый input для выбора файла */}
      <input
        ref={fileInputRef}
        type="file"
        accept=".md,.markdown,.txt"
        onChange={handleFileSelect}
        style={{ display: 'none' }}
      />

      {/* Оверлей при перетаскивании */}
      {isDragging && (
        <div className="drop-overlay">
          <div className="drop-zone">
            <Upload size={64} strokeWidth={1.5} />
            <p className="drop-title">Перетащите файл сюда</p>
            <p className="drop-hint">.md, .markdown или .txt</p>
          </div>
        </div>
      )}

      {/* Кнопка загрузки файла */}
      <button
        className="upload-btn"
        onClick={openFileDialog}
        title="Загрузить файл"
        type="button"
      >
        <Upload size={18} />
      </button>
    </div>
  );
}
