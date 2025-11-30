import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AiOutlineEye, AiOutlineEyeInvisible } from 'react-icons/ai';
import { toast, Toaster } from 'react-hot-toast';
import '../styles/LoginPage.css';
import API from '../API';

export default function AuthPage({ onLogin }) {
    const [mode, setMode] = useState('login'); // 'login' | 'signup'
    const [formData, setFormData] = useState({
        username: '',
        password: '',
        confirmPassword: ''
    });
    const [loading, setLoading] = useState(false);
    const [showPasswords, setShowPasswords] = useState(false);
    const navigate = useNavigate();

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();

        if (mode === 'signup') {
            if (!formData.username.trim() || !formData.password) {
                toast.error('Заполните имя пользователя и пароль');
                return;
            }
            if (formData.password.length < 6) {
                toast.error('Пароль должен быть не менее 6 символов');
                return;
            }
            if (formData.password !== formData.confirmPassword) {
                toast.error('Пароли не совпадают');
                return;
            }
        }

        setLoading(true);

        try {
            if (mode === 'login') {
                const response = await API.AUTH.post('/v1/login', {
                    username: formData.username.trim(),
                    password: formData.password
                });

                if (response.status === 200) {
                    toast.success('Добро пожаловать!');
                    onLogin?.();
                    navigate('/editor');
                } else {
                    toast.error(response.data.message || 'Неверный логин или пароль');
                }
            } else {
                const resp = await API.AUTH.post('/v1/register', {
                    username: formData.username.trim(),
                    password: formData.password
                });

                if (resp.status === 201) {
                    toast.success('Регистрация прошла успешно! Теперь вы можете войти.');
                    setMode('login');
                    setFormData({ username: '', password: '', confirmPassword: '' });
                } else {
                    toast.error(resp?.data?.message || 'Ошибка регистрации');
                }
            }
        } catch (err) {
            switch (err?.response?.status) {
                case 409:
                    toast.error('Пользователь с таким именем уже существует');
                    break;
                case 500:
                default:
                    toast.error('Произошла ошибка. Попробуйте еще раз.');
                    break;
            }
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="login-page">
            <Toaster position="top-right" reverseOrder={false} />

            <form onSubmit={handleSubmit} className="login-form">
                <h2>{mode === 'login' ? 'Вход в систему' : 'Регистрация'}</h2>

                <input
                    type="text"
                    name="username"
                    value={formData.username}
                    onChange={handleChange}
                    placeholder="Имя пользователя"
                    required
                />

                <div className="password-field">
                    <input
                        type={showPasswords ? 'text' : 'password'}
                        name="password"
                        value={formData.password}
                        onChange={handleChange}
                        placeholder="Пароль"
                        required
                        minLength={6}
                    />
                    <span
                        className="toggle-password"
                        onClick={() => setShowPasswords(p => !p)}
                        aria-label={showPasswords ? 'Скрыть пароль' : 'Показать пароль'}
                        role="button"
                    >
                        {showPasswords ? <AiOutlineEyeInvisible /> : <AiOutlineEye />}
                    </span>
                </div>

                {mode === 'signup' && (
                    <div className="password-field">
                        <input
                            type={showPasswords ? 'text' : 'password'}
                            name="confirmPassword"
                            value={formData.confirmPassword}
                            onChange={handleChange}
                            placeholder="Подтвердите пароль"
                            required
                            minLength={6}
                        />
                        <span
                            className="toggle-password"
                            onClick={() => setShowPasswords(p => !p)}
                            aria-label={showPasswords ? 'Скрыть пароль' : 'Показать пароль'}
                            role="button"
                        >
                            {showPasswords ? <AiOutlineEyeInvisible /> : <AiOutlineEye />}
                        </span>
                    </div>
                )}

                <button type="submit" disabled={loading}>
                    {loading
                        ? mode === 'login'
                            ? 'Входим...'
                            : 'Регистрируем...'
                        : mode === 'login'
                            ? 'Войти'
                            : 'Зарегистрироваться'}
                </button>

                <p style={{ marginTop: '0.75rem', fontSize: '0.9rem', textAlign: 'center' }}>
                    {mode === 'login' ? (
                        <>
                            Нет аккаунта?{' '}
                            <span
                                className="switch-mode"
                                onClick={() => setMode('signup')}
                                role="button"
                                style={{ color: 'blue' }}
                            >
                                Зарегистрироваться
                            </span>
                        </>
                    ) : (
                        <>
                            Уже есть аккаунт?{' '}
                            <span
                                className="switch-mode"
                                onClick={() => setMode('login')}
                                role="button"
                                style={{ color: 'blue' }}
                            >
                                Войти
                            </span>
                        </>
                    )}
                </p>
            </form>
        </div >
    );
}
