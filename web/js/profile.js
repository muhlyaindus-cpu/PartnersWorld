// Переключение вкладок
document.querySelectorAll('.tab').forEach(tab => {
    tab.addEventListener('click', () => {
        // Убираем активный класс у всех вкладок и содержимого
        document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));

        // Активируем выбранную вкладку
        tab.classList.add('active');
        const tabId = tab.getAttribute('data-tab');
        document.getElementById(tabId + '-tab').classList.add('active');
    });
});

// Загрузка профиля
async function loadProfile() {
    const token = localStorage.getItem('token');
    if (!token) {
        window.location.href = '/login.html';
        return;
    }

    try {
        const response = await fetch('/api/user/profile', {
            headers: { 'Authorization': 'Bearer ' + token }
        });

        if (!response.ok) {
            if (response.status === 401) {
                localStorage.removeItem('token');
                window.location.href = '/login.html';
            }
            throw new Error('Ошибка загрузки профиля');
        }

        const user = await response.json();

        // Заполняем форму
        document.getElementById('company').value = user.company_name || '';
        document.getElementById('description').value = user.description || '';
        document.getElementById('services').value = user.services || '';
        document.getElementById('portfolio').value = user.portfolio || '';
        document.getElementById('contact_info').value = user.contact_info || '';

    } catch (error) {
        console.error('Error loading profile:', error);
        document.getElementById('profileMessage').innerHTML = '<div class="error">❌ Не удалось загрузить профиль</div>';
    }
}

// Обновление профиля
document.getElementById('profileForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const token = localStorage.getItem('token');
    if (!token) return;

    const formData = {
        company_name: document.getElementById('company').value || undefined,
        description: document.getElementById('description').value || undefined,
        services: document.getElementById('services').value || undefined,
        portfolio: document.getElementById('portfolio').value || undefined,
        contact_info: document.getElementById('contact_info').value || undefined
    };

    // Удаляем undefined поля
    Object.keys(formData).forEach(key => formData[key] === undefined && delete formData[key]);

    const msgDiv = document.getElementById('profileMessage');
    msgDiv.innerHTML = 'Сохранение...';
    msgDiv.className = 'result info';

    try {
        const response = await fetch('/api/user/profile/update', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + token
            },
            body: JSON.stringify(formData)
        });

        const data = await response.json();

        if (response.ok) {
            msgDiv.innerHTML = '<div class="success">✅ Профиль успешно обновлён</div>';
            msgDiv.className = 'result success';
        } else {
            msgDiv.innerHTML = '<div class="error">❌ ' + (data.message || 'Ошибка обновления') + '</div>';
            msgDiv.className = 'result error';
        }
    } catch (error) {
        console.error('Error updating profile:', error);
        msgDiv.innerHTML = '<div class="error">❌ Ошибка соединения с сервером</div>';
        msgDiv.className = 'result error';
    }
});

// Загружаем профиль при загрузке страницы
document.addEventListener('DOMContentLoaded', loadProfile);