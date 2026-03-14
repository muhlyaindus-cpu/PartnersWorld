document.getElementById('registerForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const formData = {
        email: document.getElementById('email').value,
        password: document.getElementById('password').value,
        role: document.getElementById('role').value,
        country: document.getElementById('country').value,
        company_name: document.getElementById('company').value || undefined,
        description: document.getElementById('description').value || undefined
    };

    const resultDiv = document.getElementById('result');
    resultDiv.innerHTML = 'Отправка...';
    resultDiv.className = 'result info';

    try {
        const response = await fetch('/api/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(formData)
        });

        const data = await response.json();

        if (response.ok) {
            resultDiv.innerHTML = `<div class="success">✅ Регистрация успешна! Ваш ID: ${data.id}. Перенаправляем на вход...</div>`;
            resultDiv.className = 'result success';
            setTimeout(() => window.location.href = '/login.html', 2000);
        } else {
            let errorMsg = data.message || 'Ошибка регистрации';
            if (response.status === 409) errorMsg = 'Этот email уже зарегистрирован';
            resultDiv.innerHTML = `<div class="error">❌ ${errorMsg}</div>`;
            resultDiv.className = 'result error';
        }
    } catch (error) {
        console.error(error);
        resultDiv.innerHTML = '<div class="error">❌ Ошибка соединения с сервером</div>';
        resultDiv.className = 'result error';
    }
});