import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AiOutlineEye, AiOutlineEyeInvisible } from 'react-icons/ai';
import '../styles/LoginPage.css';
import API from '../API';

export default function SignupPage() {
  const [formData, setFormData] = useState({ username: '', email: '', password: '', confirmPassword: '' });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showPw, setShowPw] = useState(false);
  const [showPw2, setShowPw2] = useState(false);
  const navigate = useNavigate();

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    if (!formData.username.trim() || !formData.password) {
      setError('Заполните имя пользователя и пароль');
      return;
    }
    if (formData.password.length < 6) {
      setError('Пароль должен быть не менее 6 символов');
      return;
    }
    if (formData.password !== formData.confirmPassword) {
      setError('Пароли не совпадают');
      return;
    }

    setLoading(true);
    try {
      const resp = await API.AUTH.post('/v1/register', {
        username: formData.username.trim(),
        password: formData.password,
        email: formData.email.trim()
      });
      if (resp.status >= 200 && resp.status < 300) {
        navigate('/login', { replace: true });
      } else {
        setError(resp?.data?.message || 'Ошибка регистрации');
      }
    } catch (err) {
      setError(err?.response?.data?.message || err?.message || 'Ошибка регистрации');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-page">
      <form className="login-form" onSubmit={handleSubmit}>
        <h2>Регистрация</h2>

        <input
          type="text"
          name="username"
          value={formData.username}
          onChange={handleChange}
          placeholder="Имя пользователя"
          required
        />

        <input
          type="email"
          name="email"
          value={formData.email}
          onChange={handleChange}
          placeholder="Email (необязательно)"
        />

        <div className="password-field">
          <input
            type={showPw ? 'text' : 'password'}
            name="password"
            value={formData.password}
            onChange={handleChange}
            placeholder="Пароль"
            required
            minLength={6}
          />
          <span
            className="toggle-password"
            onClick={() => setShowPw(p => !p)}
            aria-label={showPw ? 'Скрыть пароль' : 'Показать пароль'}
            role="button"
          >
            {showPw ? <AiOutlineEyeInvisible /> : <AiOutlineEye />}
          </span>
        </div>

        <div className="password-field">
          <input
            type={showPw2 ? 'text' : 'password'}
            name="confirmPassword"
            value={formData.confirmPassword}
            onChange={handleChange}
            placeholder="Подтвердите пароль"
            required
            minLength={6}
          />
          <span
            className="toggle-password"
            onClick={() => setShowPw2(p => !p)}
            aria-label={showPw2 ? 'Скрыть пароль' : 'Показать пароль'}
            role="button"
          >
            {showPw2 ? <AiOutlineEyeInvisible /> : <AiOutlineEye />}
          </span>
        </div>

        {error && <p className="error">{error}</p>}

        <button type="submit" disabled={loading}>
          {loading ? 'Регистрируем...' : 'Зарегистрироваться'}
        </button>

        <p style={{ marginTop: '0.75rem', fontSize: '0.9rem', textAlign: 'center' }}>
          Уже есть аккаунт? <Link to="/login">Войти</Link>
        </p>
      </form>
    </div>
  );
}
