// Internationalization (i18n) Engine for VPN Client
// Supports Russian (RU) and English (EN) with automatic system detection

export const translations = {
  ru: {
    // Navigation
    nav_overview: 'Обзор',
    nav_routes: 'Маршруты',
    nav_nodes: 'Узлы',
    nav_traffic: 'Трафик',
    nav_settings: 'Настройки',

    // Titlebar
    status_disconnected: 'Отключено',
    status_connected: 'Подключено',
    status_connecting: 'Подключение...',
    admin_badge: 'Admin',

    // Dashboard
    vpn_status_disconnected: 'VPN Отключен',
    vpn_status_connected: 'VPN Подключен',
    vpn_status_connecting: 'Подключение...',
    protection_inactive: 'Защита не активна',
    protection_active: 'Защита активна',
    active_node_label: 'Текущий узел',
    ping_label: 'Пинг',
    btn_check_ping: 'Проверить',
    quick_mode_switch: 'Быстрое переключение режима',
    choose_mode: 'Выбрать ▾',
    select_routing_mode: 'Выберите режим туннелирования',
    recommended: 'Рекомендуется',
    click_to_switch: 'Сменить ⟳',
    mode_changed_to: 'Режим изменен',
    mode_whitelist_short: 'Только выбранные',
    mode_whitelist_desc: 'Проксируются только приложения из списка',
    mode_blacklist_short: 'Все кроме выбранных',
    mode_blacklist_desc: 'Проксируется всё, кроме выбранных приложений',
    mode_global_short: 'Весь трафик (Global)',
    mode_global_desc: 'Весь трафик направляется через VPN без исключений',
    session_traffic: 'Трафик сессии',
    downloaded: 'Скачано',
    uploaded: 'Отдано',
    speed: 'Скорость',
    active_connections: 'Активных соединений',
    no_nodes_add: 'Нет узлов (добавьте узел)',

    // Smart Routing (Routes tab)
    routes_title: 'Умная маршрутизация',
    routes_subtitle: 'Разделение туннелирования для приложений и доменов',
    subtab_apps: 'Приложения (Split Tunneling)',
    subtab_geo: 'Сайты и Гео-правила',
    subtab_custom: 'Свои списки (АнтиЗапрет)',
    mode_whitelist_full: 'Проксировать только выбранные',
    mode_blacklist_full: 'Проксировать всё, кроме выбранных',
    search_apps_placeholder: 'Поиск приложений...',
    btn_add_exe: '+ Добавить .exe',
    btn_refresh: 'Обновить',
    smart_bypass_ru_title: 'Обход блокировок (Bypass RU)',
    smart_bypass_ru_desc: 'Автоматическое проксирование заблокированных в РФ сервисов (YouTube, Discord, Twitter, Instagram и др.)',
    smart_adblock_title: 'Блокировка рекламы (AdBlock)',
    smart_adblock_desc: 'Блокировка баннеров, трекеров и вредоносных сайтов на уровне DNS',
    preset_lists_title: 'Предустановленные списки',
    preset_lists_desc: 'Правила из репозитория v2fly / community',
    btn_add_list: 'Добавить список',
    btn_refresh_all: 'Обновить все',
    custom_lists_title: 'Пользовательские списки',
    custom_lists_desc: 'Поддержка списков АнтиЗапрет, .txt, .json и доменов',

    // Nodes tab
    nodes_title: 'Серверы и Узлы',
    nodes_subtitle: 'Управление списком VPN узлов и подписок',
    btn_ping_all: 'Проверить пинг',
    btn_add_link: 'По ссылке',
    btn_qr: 'QR-код',
    btn_subs: 'Подписки',
    btn_warp: 'Создать WARP',
    tab_all_nodes: 'Все узлы',
    tab_manual_nodes: 'Ручные',
    btn_del_node_title: 'Удалить узел',
    empty_nodes_msg: 'Нет узлов в этой категории. Добавьте ссылку или переключитесь на «Все узлы».',

    // Traffic tab
    traffic_title: 'Мониторинг трафика',
    traffic_subtitle: 'Статистика сети и активные сессии в реальном времени',
    live_chart: 'График скорости (Live)',
    current_connections: 'Текущие соединения',
    th_destination: 'Назначение',
    th_rule: 'Правило',
    th_time: 'Время',
    no_active_conn: 'Нет активных соединений',

    // Settings tab
    settings_title: 'Настройки',
    settings_subtitle: 'Конфигурация ядра, сети и поведения клиента',
    btn_save_settings: 'Сохранить',
    settings_section_interface: 'Интерфейс и Язык',
    settings_lang_label: 'Язык интерфейса',
    settings_lang_desc: 'Выберите язык приложения или используйте автоопределение по системе',
    lang_auto: 'Авто (системный)',
    lang_ru: 'Русский (RU)',
    lang_en: 'English (US)',
    settings_section_routing: 'Маршрутизация и Сеть',
    settings_bypass_lan: 'Обход локальных сетей (LAN)',
    settings_bypass_lan_desc: 'Прямое подключение к локальным ресурсам 192.168.x.x, 10.x.x.x',
    settings_strict_route: 'Строгая маршрутизация (Strict Route)',
    settings_strict_route_desc: 'Предотвращает утечки трафика мимо TUN интерфейса',
    settings_section_dns: 'DNS и Параметры адаптера',
    settings_remote_dns: 'Удаленный DNS (DoH/DoT)',
    settings_remote_dns_desc: 'DNS сервер для защищенных запросов через VPN',
    settings_direct_dns: 'Локальный DNS',
    settings_direct_dns_desc: 'DNS сервер для прямых подключений (по умолчанию: local)',
    settings_mtu: 'Размер MTU адаптера',
    settings_mtu_desc: 'Оптимальное значение для стабильного канала (обычно 1400 или 1500)',
    settings_section_system: 'Системные',
    settings_autostart: 'Автозапуск при входе в Windows',
    settings_autostart_desc: 'Автоматически запускать клиент при загрузке системы',
    settings_minimize_tray: 'Сворачивать в системный трей',
    settings_minimize_tray_desc: 'При закрытии окна прятать в область уведомлений',
    settings_privacy_blur: 'Размытие конфиденциальных данных',
    settings_privacy_blur_desc: 'Скрывать IP адреса и ключи на снимках экрана',
    header_update_available: 'Доступно обновление',
    header_update_tooltip: 'Перейти в Настройки и посмотреть',
    header_update_dismiss: 'Скрыть',
    settings_section_update: 'Обновление приложения',
    settings_update_label: 'Текущая версия',
    settings_update_desc: 'Проверка наличия обновлений на GitHub и автозамена файла',
    btn_check_update: 'Проверить обновления',
    btn_checking_update: 'Проверка...',
    update_latest_installed: 'У вас установлена последняя версия',
    update_latest_installed_desc: 'Обновлений не найдено. Вы используете самую свежую версию VibeVPN.',
    modal_update_title: 'Доступно обновление',
    modal_update_subtitle: 'Обнаружен новый релиз VibeVPN на GitHub',
    update_changelog_label: 'Что нового в этом выпуске:',
    btn_update_later: 'Не сейчас',
    btn_update_now: 'Обновить сейчас',
    update_downloading: 'Загрузка обновления...',
    update_restarting: 'Установка и перезапуск приложения...',
    update_error: 'Ошибка при проверке обновлений',
    update_force_reinstall: 'Переустановить',
    settings_section_subs_headers: 'Параметры подписок и Идентификация',
    settings_ua_label: 'User-Agent подписок',
    settings_ua_desc: 'HTTP заголовок клиента при запросе подписок. Некоторые провайдеры требуют v2rayN, ClashMeta или Happ',
    settings_hwid_label: 'HWID устройства (Hardware ID)',
    settings_hwid_desc: 'Идентификатор устройства в заголовках HWID, X-HWID и X-Device-Id для панелей с лимитом устройств (Marzban, 3X-UI)',
    hwid_copied: 'HWID скопирован в буфер обмена',
    hwid_generated: 'Сгенерирован новый случайный HWID',
    btn_copy_hwid: 'Копировать HWID',
    btn_generate_hwid: 'Сгенерировать случайный HWID',

    // Modals - General
    btn_close: 'Закрыть',
    btn_cancel: 'Отмена',
    btn_add: 'Добавить',
    btn_import: 'Импортировать',

    // Modal: QR Code
    modal_qr_title: 'Импорт через QR-код',
    modal_qr_subtitle: 'Добавление узла или подписки через сканирование QR',
    qr_screen_title: 'С экрана',
    qr_screen_desc: 'Снимок экрана и поиск QR',
    qr_paste_title: 'Из буфера (Ctrl+V)',
    qr_paste_desc: 'Скриншот или ссылка',
    qr_file_title: 'Из файла',
    qr_file_desc: 'Выбрать PNG или JPG',
    qr_camera_title: 'Веб-камера',
    qr_camera_desc: 'Навести на QR-код',
    qr_camera_hint: 'Наведите камеру на QR-код конфига или подписки',
    qr_dropzone_lead: 'Перетащите файл сюда или нажмите',
    qr_recognized_badge: 'QR успешно распознан',
    qr_type_node: 'VPN Узел',
    qr_type_sub: 'Подписка',
    qr_type_list: 'Список узлов',
    qr_type_config: 'Конфиг',

    // Modal: Add Manual Node
    modal_add_node_title: 'Добавить узел вручную',
    modal_add_node_subtitle: 'Вставьте ссылку конфигурации (vless://, vmess://, ss://, trojan://, hy2://)',
    input_link_label: 'Ссылка конфигурации',
    input_link_placeholder: 'vless://... или ss://...',

    // Modal: Subscriptions
    modal_subs_title: 'Управление подписками',
    modal_subs_subtitle: 'Автоматическое обновление списков узлов от провайдеров',
    sub_name_label: 'Название подписки',
    sub_name_placeholder: 'Например: Мой провайдер',
    sub_url_label: 'URL подписки',
    sub_url_placeholder: 'https://example.com/api/v1/sub?...',
    btn_add_sub: 'Добавить подписку',
    my_subs: 'Мои подписки',
    btn_update_all: 'Обновить все',

    // Modal: Connection Logs
    btn_logs: 'Логи',
    modal_logs_title: 'Журнал подключения (Логи)',
    modal_logs_subtitle: 'События ядра sing-box, рукопожатия и ошибки подключения',
    logs_filter_all: 'Все',
    logs_filter_error: 'Ошибки',
    logs_filter_warn: 'Предупреждения',
    logs_autoscroll: 'Автоскролл',
    btn_copy: 'Копировать',
    btn_clear: 'Очистить',
    logs_copied: 'Логи скопированы в буфер обмена',
    logs_cleared: 'Журнал логов очищен',
    logs_empty: 'Журнал пуст. Запустите подключение для записи событий.',

    // Modal: WARP
    modal_warp_title: 'Cloudflare WARP & WireGuard',
    modal_warp_subtitle: 'Генератор профиля Cloudflare или импорт готового конфига',
    tab_warp_generate: '⚡ Генератор Cloudflare',
    tab_warp_import: '📄 Импорт файла / текста (.conf)',
    btn_warp_pick_file: 'Выбрать файл .conf',
    warp_import_placeholder: 'Вставьте содержимое файла конфигурации [Interface] / [Peer]...',
    warp_custom_name_label: 'Название узла (необязательно)',
    btn_import_warp_submit: 'Импортировать в список',
    warp_import_success: 'Конфигурация успешно импортирована!',
    warp_import_error: 'Ошибка импорта конфигурации: ',
    warp_proto_label: 'Протокол подключения',
    warp_proto_awg: 'AmneziaWG (Обход DPI через мусорные пакеты)',
    warp_proto_masque: 'MASQUE / HTTP/3 (Маскировка под веб-трафик)',
    warp_awg_lead: 'Параметры обхода DPI (AmneziaWG):',
    btn_generate_warp: 'Сгенерировать узел',
    warp_generating: 'Регистрация ключей и генерация конфигурации...',

    // Modal: Add Geo List
    modal_rule_title: 'Добавить список правил',
    modal_rule_subtitle: 'Импорт списка доменов или IP-адресов по URL',
    rule_name_label: 'Название списка',
    rule_url_label: 'URL списка (.txt / .json)',
    rule_action_label: 'Действие для этого списка',
    rule_action_proxy: 'В туннель (Proxy)',
    rule_action_direct: 'Напрямую (Direct)',

    // Conflicts
    conflict_detected: 'Обнаружены конфликтующие VPN: ',
    btn_kill_conflicts: 'Закрыть все',

    // Toasts
    toast_settings_saved: 'Настройки успешно сохранены',
    toast_node_added: 'Узел успешно добавлен',
    toast_node_deleted: 'Узел удален',
    toast_sub_added: 'Подписка успешно добавлена',
    toast_sub_updated: 'Подписка обновлена',
    toast_sub_deleted: 'Подписка удалена',
    toast_pinging: 'Проверка пинга всех узлов...',
    toast_ping_done: 'Проверка пинга завершена',
    toast_warp_success: 'Узел Cloudflare WARP успешно добавлен',
    toast_vpn_connected: 'VPN подключен',
    toast_vpn_disconnected: 'VPN отключен',
    toast_qr_recognized: 'QR распознан',
    toast_qr_empty: 'QR-код пуст',
    toast_qr_not_found: 'QR-код не найден на изображении',
    toast_qr_screen_not_found: 'QR-код не обнаружен на выбранном экране',
    toast_clipboard_empty: 'Буфер обмена пуст или не содержит изображение/ссылку',
    toast_select_img: 'Пожалуйста, выберите файл изображения',
    toast_screen_select: 'Выберите экран или окно для сканирования...',
    toast_camera_unsupported: 'Камера не поддерживается',
    toast_lang_changed: 'Язык интерфейса изменён',
    country_prefix: 'Страна',
    del_node_title: 'Удалить узел',
    del_list_title: 'Удалить список',
    rules_suffix: 'правил',
    updated_prefix: 'Обновлено',
    vpn_starting: 'VPN запускается...',
    settings_error_load: 'Ошибка загрузки настроек',
    settings_error_save: 'Ошибка сохранения настроек',
    no_custom_lists: 'Нет пользовательских списков правил',
    no_custom_lists_sub: 'Нажмите «+ Добавить список» или выберите быстрый набор выше',
    total_prefix: 'Всего'
  },

  en: {
    // Navigation
    nav_overview: 'Overview',
    nav_routes: 'Routes',
    nav_nodes: 'Nodes',
    nav_traffic: 'Traffic',
    nav_settings: 'Settings',

    // Titlebar
    status_disconnected: 'Disconnected',
    status_connected: 'Connected',
    status_connecting: 'Connecting...',
    admin_badge: 'Admin',

    // Dashboard
    vpn_status_disconnected: 'VPN Disconnected',
    vpn_status_connected: 'VPN Connected',
    vpn_status_connecting: 'Connecting...',
    protection_inactive: 'Protection inactive',
    protection_active: 'Protection active',
    active_node_label: 'Active Node',
    ping_label: 'Ping',
    btn_check_ping: 'Check',
    quick_mode_switch: 'Quick Mode Switch',
    choose_mode: 'Choose ▾',
    select_routing_mode: 'Select routing mode',
    recommended: 'Recommended',
    click_to_switch: 'Switch ⟳',
    mode_changed_to: 'Mode changed',
    mode_whitelist_short: 'Only Selected',
    mode_whitelist_desc: 'Only selected applications are routed through VPN',
    mode_blacklist_short: 'Bypass Selected',
    mode_blacklist_desc: 'All traffic is routed except selected applications',
    mode_global_short: 'All Traffic (Global)',
    mode_global_desc: 'All traffic is routed through the VPN without exceptions',
    session_traffic: 'Session Traffic',
    downloaded: 'Downloaded',
    uploaded: 'Uploaded',
    speed: 'Speed',
    active_connections: 'Active Connections',
    no_nodes_add: 'No nodes (add a node)',

    // Smart Routing (Routes tab)
    routes_title: 'Smart Routing',
    routes_subtitle: 'Split tunneling rules for applications and web domains',
    subtab_apps: 'Applications (Split Tunneling)',
    subtab_geo: 'Sites & Geo-Rules',
    subtab_custom: 'Custom Lists',
    mode_whitelist_full: 'Route only selected applications',
    mode_blacklist_full: 'Route all except selected applications',
    search_apps_placeholder: 'Search applications...',
    btn_add_exe: '+ Add .exe',
    btn_refresh: 'Refresh',
    smart_bypass_ru_title: 'Block Bypass (Bypass RU)',
    smart_bypass_ru_desc: 'Automatically unblock blocked services (YouTube, Discord, Twitter, Instagram, etc.)',
    smart_adblock_title: 'Ad Blocker (AdBlock)',
    smart_adblock_desc: 'Block ads, analytics, trackers and malicious domains at DNS level',
    preset_lists_title: 'Preset Rule Lists',
    preset_lists_desc: 'Community rulesets from v2fly repository',
    btn_add_list: 'Add List',
    btn_refresh_all: 'Refresh All',
    custom_lists_title: 'Custom Lists',
    custom_lists_desc: 'Supports AntiZapret lists, .txt, .json, and plain domain files',

    // Nodes tab
    nodes_title: 'Servers & Nodes',
    nodes_subtitle: 'Manage VPN servers, links and subscription feeds',
    btn_ping_all: 'Check Ping',
    btn_add_link: 'Add Link',
    btn_qr: 'QR Code',
    btn_subs: 'Subscriptions',
    btn_warp: 'Create WARP',
    tab_all_nodes: 'All Nodes',
    tab_manual_nodes: 'Manual',
    btn_del_node_title: 'Delete node',
    empty_nodes_msg: 'No nodes in this category. Add a link or switch to "All Nodes".',

    // Traffic tab
    traffic_title: 'Traffic Monitor',
    traffic_subtitle: 'Real-time bandwidth usage and active socket sessions',
    live_chart: 'Live Bandwidth Chart',
    current_connections: 'Current Connections',
    th_destination: 'Destination',
    th_rule: 'Rule',
    th_time: 'Time',
    no_active_conn: 'No active connections',

    // Settings tab
    settings_title: 'Settings',
    settings_subtitle: 'Core engine, network adapter, and client preferences',
    btn_save_settings: 'Save Settings',
    settings_section_interface: 'Interface & Language',
    settings_lang_label: 'Interface Language',
    settings_lang_desc: 'Choose client language or automatically follow Windows system language',
    lang_auto: 'Auto (System default)',
    lang_ru: 'Русский (RU)',
    lang_en: 'English (US)',
    settings_section_routing: 'Routing & Network',
    settings_bypass_lan: 'Bypass Local Networks (LAN)',
    settings_bypass_lan_desc: 'Direct connection to local network resources (192.168.x.x, 10.x.x.x)',
    settings_strict_route: 'Strict Routing (Strict Route)',
    settings_strict_route_desc: 'Prevents traffic leaks outside the TUN interface',
    settings_section_dns: 'DNS & Adapter Settings',
    settings_remote_dns: 'Remote DNS (DoH/DoT)',
    settings_remote_dns_desc: 'Secure encrypted DNS server for queries routed via VPN',
    settings_direct_dns: 'Direct DNS',
    settings_direct_dns_desc: 'DNS server for non-tunneled connections (default: local)',
    settings_mtu: 'Adapter MTU Size',
    settings_mtu_desc: 'Optimal packet size for stability (typically 1400 or 1500)',
    settings_section_system: 'System Preferences',
    settings_autostart: 'Start with Windows',
    settings_autostart_desc: 'Automatically launch client on system boot',
    settings_minimize_tray: 'Minimize to System Tray',
    settings_minimize_tray_desc: 'Hide window to notification tray instead of closing application',
    settings_privacy_blur: 'Blur Sensitive Data',
    settings_privacy_blur_desc: 'Conceal IP addresses and keys for safe screen capture',
    header_update_available: 'Update Available',
    header_update_tooltip: 'Go to Settings and review',
    header_update_dismiss: 'Dismiss',
    settings_section_update: 'Application Update',
    settings_update_label: 'Current Version',
    settings_update_desc: 'Check for latest releases on GitHub and auto-replace binary',
    btn_check_update: 'Check for Updates',
    btn_checking_update: 'Checking...',
    update_latest_installed: 'You have the latest version installed',
    update_latest_installed_desc: 'No updates found. You are running the newest version of VibeVPN.',
    modal_update_title: 'Update Available',
    modal_update_subtitle: 'A new release of VibeVPN was found on GitHub',
    update_changelog_label: 'What\'s new in this release:',
    btn_update_later: 'Not Now',
    btn_update_now: 'Update Now',
    update_downloading: 'Downloading update...',
    update_restarting: 'Installing and restarting application...',
    update_error: 'Update check error',
    update_force_reinstall: 'Reinstall',
    settings_section_subs_headers: 'Subscriptions & Device Identification',
    settings_ua_label: 'Subscription User-Agent',
    settings_ua_desc: 'Client HTTP User-Agent for subscription queries. Some providers require v2rayN, ClashMeta, or Happ',
    settings_hwid_label: 'Device HWID (Hardware ID)',
    settings_hwid_desc: 'Sent in HWID, X-HWID and X-Device-Id headers for panels with device limits (Marzban, 3X-UI)',
    hwid_copied: 'HWID copied to clipboard',
    hwid_generated: 'Generated new random HWID',
    btn_copy_hwid: 'Copy HWID',
    btn_generate_hwid: 'Generate random HWID',

    // Modals - General
    btn_close: 'Close',
    btn_cancel: 'Cancel',
    btn_add: 'Add',
    btn_import: 'Import',

    // Modal: QR Code
    modal_qr_title: 'Import via QR Code',
    modal_qr_subtitle: 'Add a node or subscription by scanning a QR code',
    qr_screen_title: 'From Screen',
    qr_screen_desc: 'Capture screen and scan QR',
    qr_paste_title: 'Clipboard (Ctrl+V)',
    qr_paste_desc: 'Screenshot or copied link',
    qr_file_title: 'From File',
    qr_file_desc: 'Choose PNG or JPG image',
    qr_camera_title: 'Webcam',
    qr_camera_desc: 'Scan via video feed',
    qr_camera_hint: 'Point your camera at a VPN config or subscription QR code',
    qr_dropzone_lead: 'Drop image file here or press',
    qr_recognized_badge: 'QR Code recognized',
    qr_type_node: 'VPN Node',
    qr_type_sub: 'Subscription',
    qr_type_list: 'Node List',
    qr_type_config: 'Config',

    // Modal: Add Manual Node
    modal_add_node_title: 'Add Node Manually',
    modal_add_node_subtitle: 'Paste configuration link (vless://, vmess://, ss://, trojan://, hy2://)',
    input_link_label: 'Configuration Link',
    input_link_placeholder: 'vless://... or ss://...',

    // Modal: Subscriptions
    modal_subs_title: 'Subscription Management',
    modal_subs_subtitle: 'Automatically fetch and update node feeds from providers',
    sub_name_label: 'Subscription Name',
    sub_name_placeholder: 'e.g.: My VPN Provider',
    sub_url_label: 'Subscription URL',
    sub_url_placeholder: 'https://example.com/api/v1/sub?...',
    btn_add_sub: 'Add Subscription',
    my_subs: 'My Subscriptions',
    btn_update_all: 'Update All',

    // Modal: Connection Logs
    btn_logs: 'Logs',
    modal_logs_title: 'Connection Logs',
    modal_logs_subtitle: 'Sing-box core events, handshakes, and connection errors',
    logs_filter_all: 'All',
    logs_filter_error: 'Errors',
    logs_filter_warn: 'Warnings',
    logs_autoscroll: 'Auto-scroll',
    btn_copy: 'Copy',
    btn_clear: 'Clear',
    logs_copied: 'Logs copied to clipboard',
    logs_cleared: 'Logs cleared',
    logs_empty: 'Log is empty. Start a connection to record events.',

    // Modal: WARP
    modal_warp_title: 'Cloudflare WARP & WireGuard',
    modal_warp_subtitle: 'Generate Cloudflare profile or import existing config',
    tab_warp_generate: '⚡ Cloudflare Generator',
    tab_warp_import: '📄 Import File / Text (.conf)',
    btn_warp_pick_file: 'Choose .conf file',
    warp_import_placeholder: 'Paste config content [Interface] / [Peer]...',
    warp_custom_name_label: 'Node Name (optional)',
    btn_import_warp_submit: 'Import Config',
    warp_import_success: 'Config imported successfully!',
    warp_import_error: 'Config import error: ',
    warp_proto_label: 'Connection Protocol',
    warp_proto_awg: 'AmneziaWG (DPI bypass with junk packets)',
    warp_proto_masque: 'MASQUE / HTTP/3 (Web traffic disguise)',
    warp_awg_lead: 'DPI Bypass Parameters (AmneziaWG):',
    btn_generate_warp: 'Generate Node',
    warp_generating: 'Registering keys and generating configuration...',

    // Modal: Add Geo List
    modal_rule_title: 'Add Rule List',
    modal_rule_subtitle: 'Import domain or IP address list by URL',
    rule_name_label: 'List Name',
    rule_url_label: 'List URL (.txt / .json)',
    rule_action_label: 'Action for this list',
    rule_action_proxy: 'Tunnel (Proxy)',
    rule_action_direct: 'Direct (Bypass)',

    // Conflicts
    conflict_detected: 'Conflicting VPNs detected: ',
    btn_kill_conflicts: 'Close All',

    // Toasts
    toast_settings_saved: 'Settings saved successfully',
    toast_node_added: 'Node added successfully',
    toast_node_deleted: 'Node deleted',
    toast_sub_added: 'Subscription added successfully',
    toast_sub_updated: 'Subscription updated',
    toast_sub_deleted: 'Subscription deleted',
    toast_pinging: 'Testing latency for all nodes...',
    toast_ping_done: 'Latency test complete',
    toast_warp_success: 'Cloudflare WARP node added successfully',
    toast_vpn_connected: 'VPN Connected',
    toast_vpn_disconnected: 'VPN Disconnected',
    toast_qr_recognized: 'QR Code recognized',
    toast_qr_empty: 'QR Code is empty',
    toast_qr_not_found: 'No QR code found in image',
    toast_qr_screen_not_found: 'No QR code found on selected screen',
    toast_clipboard_empty: 'Clipboard is empty or has no image/link',
    toast_select_img: 'Please select an image file',
    toast_screen_select: 'Select a screen or window to scan...',
    toast_camera_unsupported: 'Camera is not supported',
    toast_lang_changed: 'Interface language updated',
    country_prefix: 'Country',
    del_node_title: 'Delete node',
    del_list_title: 'Delete list',
    rules_suffix: 'rules',
    updated_prefix: 'Updated',
    vpn_starting: 'VPN starting...',
    settings_error_load: 'Error loading settings',
    settings_error_save: 'Error saving settings',
    no_custom_lists: 'No custom rule lists',
    no_custom_lists_sub: 'Click "+ Add List" or choose a quick preset above',
    total_prefix: 'Total'
  }
};

let currentLang = 'ru';

/**
 * Detects the system language from browser / OS environment.
 * If Russian, Belarusian, Ukrainian, or Kazakh -> returns 'ru', else 'en'.
 */
export function detectSystemLanguage() {
  try {
    const navLangs = navigator.languages || [navigator.language || ''];
    for (const l of navLangs) {
      if (!l) continue;
      const lower = l.toLowerCase();
      if (lower.startsWith('ru') || lower.startsWith('be') || lower.startsWith('uk') || lower.startsWith('kk')) {
        return 'ru';
      }
    }
  } catch (_) {}
  return 'en';
}

/**
 * Resolves the effective language based on user setting ('auto' | 'ru' | 'en')
 * and detected system language.
 */
export function getEffectiveLanguage(settingValue, systemLangFallback = null) {
  if (!settingValue || settingValue === 'auto') {
    if (systemLangFallback === 'ru' || systemLangFallback === 'en') {
      return systemLangFallback;
    }
    return detectSystemLanguage();
  }
  return settingValue === 'en' ? 'en' : 'ru';
}

export function getCurrentLanguage() {
  return currentLang;
}

/**
 * Translates a given key with optional fallback.
 */
export function t(key, fallback = '') {
  const dict = translations[currentLang] || translations.ru;
  if (dict && dict[key] !== undefined) {
    return dict[key];
  }
  const enDict = translations.en;
  if (enDict && enDict[key] !== undefined) {
    return enDict[key];
  }
  return fallback || key;
}

/**
 * Applies the given language code ('ru' or 'en') to all DOM elements
 * with [data-i18n], [data-i18n-placeholder], and [data-i18n-title].
 */
export function applyLanguage(langCode) {
  currentLang = langCode === 'en' ? 'en' : 'ru';
  document.documentElement.lang = currentLang;

  const dict = translations[currentLang];
  if (!dict) return;

  // Text content
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    if (dict[key] !== undefined) {
      el.textContent = dict[key];
    }
  });

  // Placeholders
  document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
    const key = el.getAttribute('data-i18n-placeholder');
    if (dict[key] !== undefined) {
      el.placeholder = dict[key];
    }
  });

  // Tooltips / Titles
  document.querySelectorAll('[data-i18n-title]').forEach(el => {
    const key = el.getAttribute('data-i18n-title');
    if (dict[key] !== undefined) {
      el.title = dict[key];
    }
  });
}
