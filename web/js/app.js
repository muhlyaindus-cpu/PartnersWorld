document.addEventListener('DOMContentLoaded', () => {
    const token = localStorage.getItem('token');
    const currentPath = window.location.pathname;

    // Список страниц, доступных только авторизованным
    const protectedPages = ['/profile.html', '/requests.html'];

    // Если страница защищена и токена нет — редирект на логин
    if (protectedPages.includes(currentPath) && !token) {
        window.location.href = '/login.html';
        return;
    }

    // Обновляем навигацию (показываем/скрываем ссылки)
    updateNavBasedOnAuth(!!token);

    // Загружаем данные, если пользователь авторизован
    if (token) {
    loadUserInfo();
    if (currentPath === '/' || currentPath === '/index.html' || currentPath === '/requests.html') {
        fetchRequests();
    }
    // Скрываем hero-блок для авторизованных
    const hero = document.querySelector('.hero');
    if (hero) hero.style.display = 'none';
} else {
    if (currentPath === '/' || currentPath === '/index.html') {
        showLoginPrompt();
    }
}
});

// Функция обновления навигации
function updateNavBasedOnAuth(isLoggedIn) {
    const requestsLink = document.getElementById('requestsLink');
    if (requestsLink) {
        requestsLink.style.display = isLoggedIn ? 'inline-block' : 'none';
    }
    const loginLink = document.getElementById('loginLink');
    const registerLink = document.getElementById('registerLink');
    const profileLink = document.getElementById('profileLink');
    const logoutBtn = document.getElementById('logoutBtn');

    if (isLoggedIn) {
        if (loginLink) loginLink.style.display = 'none';
        if (registerLink) registerLink.style.display = 'none';
        if (profileLink) profileLink.style.display = 'inline-block';
        if (logoutBtn) logoutBtn.style.display = 'inline-block';
    } else {
        if (loginLink) loginLink.style.display = 'inline-block';
        if (registerLink) registerLink.style.display = 'inline-block';
        if (profileLink) profileLink.style.display = 'none';
        if (logoutBtn) logoutBtn.style.display = 'none';
    }
}

// Функция для отображения приглашения на главной
function showLoginPrompt() {
    const requestsList = document.getElementById('requestsList');
    if (requestsList) {
        requestsList.innerHTML = `
            <div class="login-prompt">
                <p>🔒 Чтобы увидеть запросы, <a href="/login.html">войдите</a> или <a href="/register.html">зарегистрируйтесь</a>.</p>
            </div>
        `;
    }
}

// Загрузка информации о пользователе из токена
function loadUserInfo() {
    const token = localStorage.getItem('token');
    if (!token) return;

    const userInfo = document.getElementById('userInfo');
    if (!userInfo) return;

    try {
        const base64Url = token.split('.')[1];
        const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
        const payload = JSON.parse(atob(base64));

        userInfo.innerHTML = `
            <span class="user-email">${payload.email}</span>
            <span class="user-role">(${payload.role === 'exporter' ? 'Экспортёр' : 'Партнёр'})</span>
        `;

        const createRequestBtn = document.getElementById('createRequestBtn');
        if (createRequestBtn && payload.role === 'exporter') {
            createRequestBtn.style.display = 'block';
        }
    } catch (e) {
        console.error('Ошибка при декодировании токена:', e);
    }
}

// Загрузка списка запросов с сервера
async function fetchRequests(countryFilter = '') {
    const token = localStorage.getItem('token');
    const requestsList = document.getElementById('requestsList');
    if (!requestsList) return;

    // Если нет токена — ничего не делаем (показываем приглашение)
    if (!token) {
        return;
    }

    requestsList.innerHTML = '<div class="loading">Загрузка запросов...</div>';

    try {
        let url = '/api/requests';
        if (countryFilter) {
            url += `?country=${encodeURIComponent(countryFilter)}`;
        }

        const response = await fetch(url, {
            headers: {
                'Authorization': 'Bearer ' + token
            }
        });

        if (!response.ok) {
            if (response.status === 401) {
                localStorage.removeItem('token');
                window.location.href = '/login.html';
                return;
            }
            throw new Error('Ошибка загрузки запросов');
        }

        const requests = await response.json();
        displayRequests(requests);
    } catch (error) {
        console.error('Error:', error);
        requestsList.innerHTML = '<div class="error">❌ Ошибка при загрузке запросов</div>';
    }
}

// Отображение списка запросов на странице
function displayRequests(requests) {
    const requestsList = document.getElementById('requestsList');
    if (!requestsList) return;

    if (!requests || requests.length === 0) {
        requestsList.innerHTML = '<div class="no-requests">📭 Нет открытых запросов</div>';
        return;
    }

    let html = '';

    requests.forEach(req => {
        let typeIcon = '';
        let typeText = '';
        if (req.type === 'need') {
            typeIcon = '🔍';
            typeText = 'Ищет партнёра';
        } else if (req.type === 'offer') {
            typeIcon = '🤝';
            typeText = 'Предлагает помощь';
        }

        const budgetText = req.budget ? `${req.budget} USD` : 'не указан';
        const createdDate = new Date(req.created_at);
        const dateStr = createdDate.toLocaleDateString('ru-RU');

        html += `
            <div class="request-card">
                <div class="request-type">${typeIcon} ${typeText}</div>
                <h3>${escapeHtml(req.title)}</h3>
                <p class="description">${escapeHtml(req.description)}</p>
                <div class="request-details">
                    <span class="country">🌍 Страна: ${escapeHtml(req.country)}</span>
                    <span class="category">📦 Категория: ${escapeHtml(req.category) || 'любая'}</span>
                    <span class="budget">💰 Бюджет: ${budgetText}</span>
                    <span class="date">📅 Создан: ${dateStr}</span>
                </div>
                <button class="respond-btn" onclick="respondToRequest(${req.id})">Откликнуться</button>
            </div>
        `;
    });

    requestsList.innerHTML = html;
}

// Функция для экранирования HTML
function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Фильтрация по стране
function applyFilter() {
    const filter = document.getElementById('countryFilter').value.trim();
    fetchRequests(filter);
}

function clearFilter() {
    document.getElementById('countryFilter').value = '';
    fetchRequests();
}

// Отклик на запрос (заглушка)
async function respondToRequest(requestId) {
    const token = localStorage.getItem('token');
    if (!token) return;

    try {
        const base64Url = token.split('.')[1];
        const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
        const payload = JSON.parse(atob(base64));

        if (payload.role !== 'partner') {
            alert('Только партнёры могут откликаться на запросы');
            return;
        }

        const response = await fetch(`/api/requests/${requestId}/respond`, {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + token,
                'Content-Type': 'application/json'
            }
        });

        if (response.ok) {
            alert('✅ Отклик отправлен!');
        } else {
            const data = await response.json();
            alert('❌ Ошибка: ' + (data.message || 'Не удалось откликнуться'));
        }
    } catch (error) {
        console.error('Error:', error);
        alert('❌ Ошибка соединения с сервером');
    }
}

// Выход
function logout() {
    localStorage.removeItem('token');
    window.location.href = '/login.html';
}