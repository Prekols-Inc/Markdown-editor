import React from 'react';
import LoginForm from './LoginForm';
import './App.css';

function App() {
  return (
    <div className="App">
      <header className="App-header">
        <h1>Добро пожаловать в систему</h1>
        <p>Пожалуйста, авторизуйтесь для продолжения</p>
      </header>
      <main className="App-main">
        <LoginForm />
      </main>
    </div>
  );
}

export default App;