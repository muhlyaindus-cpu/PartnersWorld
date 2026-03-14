document.getElementById('loginForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;

    const resultDiv = document.getElementById('result');
    resultDiv.innerHTML = 'Вход...';
    resultDiv.className = 'result info';

    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        const data = await response.json();

        if (response.ok) {
            localStorage.setItem('token', data.token);
            resultDiv.innerHTML = '✅ Успешный вход! Перенаправление...';
            resultDiv.className = 'result success';
            setTimeout(() => window.location.href = '/', 500);
        } else {
            const errorMsg = data.message || data.error || 'Ошибка входа';
            resultDiv.innerHTML = `❌ ${errorMsg}`;
            resultDiv.className = 'result error';
        }
    } catch (error) {
        console.error('Login error:', error);
        resultDiv.innerHTML = '❌ Ошибка соединения с сервером';
        resultDiv.className = 'result error';
    }
});