# VibeVPN — Удобный VPN-клиент для Windows

<div align="center">

<p align="center">
  <b>🇷🇺 Русский</b> | <a href="README_EN.md">🇬🇧 English</a>
</p>

![Platform](https://img.shields.io/badge/Платформа-Windows%2010%20%2F%2011%20x64-0078D6?logo=windows&logoColor=white)
![Version](https://img.shields.io/badge/Версия-v1.0.8-brightgreen.svg)
![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-v2-DF0000?logo=wails&logoColor=white)
![Sing-Box](https://img.shields.io/badge/Ядро-sing--box%201.14-FF6B00?logo=box&logoColor=white)
![License](https://img.shields.io/badge/Лицензия-MIT-green.svg)
![i18n](https://img.shields.io/badge/Язык-RU%20%7C%20EN-blueviolet.svg)

**Простой, красивый и быстрый VPN-клиент для Windows с разделением трафика по приложениям (Split Tunneling), генератором и импортом WARP (AmneziaWG), каскадным туннелированием WARP Detour, гео-маршрутизацией списков IP/доменов, живым журналом логов ядра и защитой от сетевого шума.**

[Скачать релиз v1.0.8 (.exe)](https://github.com/zalupacode228/vibe-vpn/releases/tag/v1.0.8) • [Возможности](#-основные-возможности) • [Скриншоты](#-скриншоты-интерфейса) • [Сборка из исходников](#-сборка-из-исходников)

</div>

---

## 📸 Скриншоты интерфейса

<div align="center">

| **Главный экран (Подключение и трафик)** | **Серверы и Узлы (Списки и пинг)** |
|:---:|:---:|
| <img src="assets/screenshots/dashboard_ru.png?v=1.0.7" width="420" alt="Dashboard"/> | <img src="assets/screenshots/nodes_ru.png?v=1.0.7" width="420" alt="Nodes"/> |

| **Универсальный QR-сканер** | **Параметры и настройки клиента** |
|:---:|:---:|
| <img src="assets/screenshots/modal_qr_ru.png?v=1.0.7" width="420" alt="QR Modal"/> | <img src="assets/screenshots/settings_ru.png?v=1.0.7" width="420" alt="Settings"/> |

</div>

---

## ✨ Основные возможности

- **Split Tunneling** — белый / чёрный список приложений, глобальный и прямой режимы.
- **WARP Detour (Clean IP)** — каскад через Cloudflare WARP поверх любого прокси. ChatGPT, Claude, Netflix без капч и блокировок.
- **Генератор WARP** — AmneziaWG (обход ТСПУ/DPI) и MASQUE/HTTP3 в один клик; импорт `.conf` файлов.
- **Быстрейший узел** — автовыбор сервера с наименьшим пингом.
- **Kill Switch** — блокировка утечек при обрыве туннеля.
- **Bypass RU** — российские сервисы (банки, Госуслуги, VK) идут напрямую мимо VPN.
- **AdBlock** — блокировка рекламы и трекеров на уровне DNS.
- **QR-импорт** — с экрана, из буфера, файла или веб-камеры; одиночные узлы и подписки.
- **Журнал ядра** — логи sing-box в реальном времени с фильтрацией системного шума Windows.
- **Протоколы** — VLESS (Reality, XHTTP, ML-KEM-768), VMess, Shadowsocks 2022, Trojan, Hysteria 2, TUIC v5, WireGuard, AmneziaWG.
- **Телеметрия** — график скорости 60 fps, дерево соединений по процессам, виджет подписки с лимитами.
- **Бэкап профиля** — экспорт/импорт всех настроек в `.json`.
- **Режим приватности** — блюр IP, доменов и ключей для стримов и скриншотов.
- **Material 3 Dark UI** — строгий дизайн, трей с быстрыми действиями.

---

## 🚀 Скачивание и запуск

1. Перейдите на страницу [Релизов](https://github.com/zalupacode228/vibe-vpn/releases/latest).
2. Скачайте свежий файл `VibeVPN-v1.0.8-Windows-x64.exe` (или `VPN.exe`).
3. Запустите приложение.
   > **Примечание**: Требуется Windows 10 или 11 (64-бит). Для создания виртуального сетевого адаптера (`wintun.dll`) требуются права администратора.

---

## 🛠️ Сборка из исходников

### Требования
- **Go**: `1.22+` или `1.23+` ([golang.org](https://go.dev/dl/))
- **Node.js**: `18+` или `20+` ([nodejs.org](https://nodejs.org/))
- **Wails CLI v2**:
  ```powershell
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- **WebView2 Runtime** (установлен по умолчанию в Windows 10 и 11)

### Режим разработки (Live Dev)
Запуск с горячей перезагрузкой изменений фронтенда и Go-бэкенда:
```powershell
git clone https://github.com/zalupacode228/vibe-vpn.git
cd vibe-vpn

wails dev
```

### Сборка production-бинарника
Сборка полностью автономного однофайлового приложения (все ресурсы, иконка и ядро вшиваются внутрь):
```powershell
wails build
```
Готовый исполняемый файл появится по пути:
```
build/bin/VPN.exe
```

---

## 🏗️ Архитектура проекта

```
vibe-go/
├── app.go                 # Контроллер Wails, обработчики диалогов и вызовы Windows API
├── main.go                # Точка входа, синглтон-мьютекс и конфигурация окна Wails
├── assets/
│   ├── core/              # Встроенное ядро (sing-box.exe с поддержкой AWG + wintun.dll)
│   └── screenshots/       # Скриншоты интерфейса для документации
├── build/
│   ├── appicon.png        # Фирменная векторная иконка щита (1024x1024)
│   └── windows/
│       └── icon.ico       # Мультиразмерная Windows-иконка (16–256px) для .exe и таскбара
├── pkg/
│   ├── models/            # Структуры данных (узлы, настройки, гео-правила, подписки)
│   ├── singbox/           # Генератор конфигураций sing-box, роутер и процессный контроллер
│   ├── procmon/           # Диспетчер процессов Windows и извлечение иконок .exe
│   ├── storage/           # Сохранение конфигураций в %APPDATA%\VPN
│   ├── subs/              # Парсер подписок и декодер Base64/Clash
│   ├── pinger/            # Движок параллельного TCP/ICMP пинга
│   ├── warp/              # Генератор и парсер Cloudflare WARP / AmneziaWG
│   └── tray/              # Интеграция с системным треем Windows
└── frontend/              # Быстрый фронтенд на чистом JS + Vite
    ├── src/
    │   ├── main.js        # Логика UI, парсинг логов и обработка событий
    │   ├── i18n.js        # Словарь локализации и автоопределение языка
    │   └── style.css      # Стили Dark Glassmorphism
    └── index.html         # Разметка приложения
```

---

## 📄 Лицензия

Проект распространяется под открытой лицензией **MIT License**. Подробнее см. в файле [`LICENSE`](LICENSE).

<div align="center">
Сделано с ❤️ на Go, Wails v2 и sing-box.
</div>
