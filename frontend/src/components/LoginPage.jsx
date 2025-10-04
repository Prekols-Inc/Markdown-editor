import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AiOutlineEye, AiOutlineEyeInvisible } from 'react-icons/ai';
import '../styles/LoginPage.css';
import API from '../API';

export default function LoginPage({ onLogin }) {
    const [formData, setFormData] = useState({ username: '', password: '' });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [showPw, setShowPw] = useState(false);
    const navigate = useNavigate();

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData((prev) => ({ ...prev, [name]: value }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setLoading(true);
        setError('');

        try {
            const response = await API.AUTH.post('/v1/login', formData);
            if (response.status == 200 && response.status < 300) {
                onLogin();
                navigate('/editor');
                return;
            } else {
                console.log("error");
                setError(response.data.message || 'Неверный логин или пароль');
            }
        } catch (err) {
            if (err.response) {
                setError(err.response.data.message || `Ошибка: ${err.response.status}`);
            } else {
                setError(err.message || 'Ошибка сети');
            }
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="login-page">
            <form onSubmit={handleSubmit} className="login-form">
                <h2>Вход в систему</h2>

                <input
                    type="text"
                    name="username"
                    value={formData.username}
                    onChange={handleChange}
                    placeholder="Логин"
                    required
                />
                <div className="password-field">
                    <input
                        type={showPw ? 'text' : 'password'}
                        name="password"
                        value={formData.password}
                        onChange={handleChange}
                        placeholder="Пароль"
                        required
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

                {error && <p className="error">{error}</p>}

                <button type="submit" disabled={loading}>
                    {loading ? 'Входим...' : 'Войти'}
                </button>
                <p style={{ marginTop: '0.75rem', fontSize: '0.9rem', textAlign: 'center' }}>
                    Нет аккаунта? <Link to="/signup">Зарегистрируйтесь</Link>
                </p>
            </form>
        </div>
    );
}
