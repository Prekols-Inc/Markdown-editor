import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import '../styles/LoginPage.css';
import API from '../API';

export default function LoginPage({ onLogin }) {
    const [formData, setFormData] = useState({ username: '', password: '' });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
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
            if (response.status == 200) {
                onLogin();
                navigate('/editor');
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
                <input
                    type="password"
                    name="password"
                    value={formData.password}
                    onChange={handleChange}
                    placeholder="Пароль"
                    required
                />

                {error && <p className="error">{error}</p>}

                <button type="submit" disabled={loading}>
                    {loading ? 'Входим...' : 'Войти'}
                </button>
            </form>
        </div>
    );
}
