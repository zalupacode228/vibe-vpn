import './style.css';
import jsQR from 'jsqr';

import {
  WindowMin,
  WindowToggleMax,
  WindowClose,
  GetSettings,
  SaveSettings,
  GetNodes,
  GetActiveNodeID,
  SetActiveNodeID,
  GenerateWarp,
  AddManualNode,
  DeleteNode,
  PingAllNodes,
  GetSubscriptions,
  AddSubscription,
  RefreshSubscription,
  DeleteSubscription,
  GetRunningApps,
  ToggleAppRule,
  AddCustomApp,
  SelectExeFile,
  GetCustomLists,
  AddCustomList,
  AddPresetList,
  RefreshAllCustomLists,
  ToggleCustomList,
  DeleteCustomList,
  ToggleSmartRule,
  ToggleVPN,
  GetTrafficStats,
  GetConnections,
  IsAdmin,
  RestartAsAdmin,
  CheckVPNConflicts,
  KillConflictingVPNs,
  GetSystemLanguage,
  SelectWarpConfigFile,
  ImportWarpConfig,
  SelectGeoRuleFile,
  GetCoreLogs,
  ClearCoreLogs,
  CheckUpdate,
  DownloadAndApplyUpdate,
  GetAppVersion,
  GenerateRandomHWID,
  GetDefaultHWID,
  TestWebLatency,
  SelectFastestNode,
  ToggleWarpDetour,
  ToggleKillSwitch,
  ExportProfile,
  ImportProfile
} from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

import {
  applyLanguage,
  getEffectiveLanguage,
  detectSystemLanguage,
  getCurrentLanguage,
  t
} from './i18n.js';

let detectedSystemLanguage = 'ru';

// Global state
let state = {
  isConnected: false,
  activePage: 'dash',
  activeRoutingSubtab: 'apps',
  settings: {},
  nodes: [],
  subscriptions: [],
  activeNodeId: '',
  activeNodeCategory: 'all',
  apps: [],
  customLists: [],
  runningAppsFilter: '',
  pollTimer: null,
  speedHistory: {
    down: new Array(40).fill(0),
    up: new Array(40).fill(0)
  }
};

// Utilities
function formatSpeed(bytesPerSec) {
  if (!bytesPerSec || bytesPerSec <= 0) return '0.0 KB/s';
  if (bytesPerSec < 1024 * 1024) {
    return `${(bytesPerSec / 1024).toFixed(1)} KB/s`;
  }
  return `${(bytesPerSec / (1024 * 1024)).toFixed(2)} MB/s`;
}

function formatBytes(bytes) {
  if (!bytes || bytes <= 0) return '0 KB';
  if (bytes < 1024 * 1024) {
    return `${Math.round(bytes / 1024)} KB`;
  }
  const mb = bytes / (1024 * 1024);
  if (mb < 1024) return `${mb.toFixed(1)} MB`;
  return `${(mb / 1024).toFixed(2)} GB`;
}

function showToast(message) {
  const toast = document.getElementById('toast');
  if (!toast) return;
  toast.innerText = message;
  toast.classList.add('show');
  setTimeout(() => {
    toast.classList.remove('show');
  }, 2500);
}

// ==================== COUNTRY FLAGS & GEO DETECTION ====================
function extractFlagEmojiCode(text) {
  if (!text) return null;
  const chars = Array.from(text);
  const indicators = chars.filter(c => {
    const code = c.codePointAt(0);
    return code >= 0x1F1E6 && code <= 0x1F1FF;
  });
  if (indicators.length >= 2) {
    const c1 = String.fromCharCode(indicators[0].codePointAt(0) - 0x1F1E6 + 65);
    const c2 = String.fromCharCode(indicators[1].codePointAt(0) - 0x1F1E6 + 65);
    return (c1 + c2).toUpperCase();
  }
  return null;
}

function cleanNodeDisplayName(rawName) {
  if (!rawName) return "Proxy Node";
  let cleaned = Array.from(rawName).filter(c => {
    const code = c.codePointAt(0);
    return !(code >= 0x1F1E6 && code <= 0x1F1FF);
  }).join('').trim();
  cleaned = cleaned.replace(/^[-_\s|:]+|[-_\s|:]+$/g, '').trim();
  return cleaned || rawName;
}

function detectNodeCountry(node) {
  if (!node) return 'VPN';
  const name = (node.name || '').trim();
  const server = (node.server || '').trim().toLowerCase();
  const proto = (node.protocol || '').toLowerCase();

  if (name.toLowerCase().includes('warp') || name.toLowerCase().includes('cloudflare') || proto === 'masque' || server.includes('162.159.192')) {
    return 'CF';
  }

  const emojiCC = extractFlagEmojiCode(name);
  if (emojiCC) return emojiCC;

  const m = name.match(/\[([A-Za-z]{2})\]|\(([A-Za-z]{2})\)|\b([A-Za-z]{2})[-_ ]/);
  if (m) {
    const code = (m[1] || m[2] || m[3]).toUpperCase();
    const valid = new Set(['DE', 'NL', 'RU', 'US', 'GB', 'UK', 'FI', 'SE', 'FR', 'IT', 'PL', 'UA', 'TR', 'KZ', 'JP', 'CH', 'AT', 'ES', 'SG', 'CA', 'CZ', 'EE', 'LV', 'LT', 'HK']);
    if (valid.has(code)) {
      return code === 'UK' ? 'GB' : code;
    }
  }

  const lower = name.toLowerCase();
  const keywords = [
    ['DE', ['германи', 'франкфурт', 'берлин', 'germany', 'frankfurt', 'berlin']],
    ['NL', ['нидерланд', 'амстердам', 'голланди', 'netherlands', 'amsterdam', 'holland']],
    ['RU', ['росси', 'москв', 'питер', 'russia', 'moscow', 'saint petersburg']],
    ['US', ['сша', 'америк', 'нью-йорк', 'usa', 'united states', 'america', 'new york']],
    ['FI', ['финлянд', 'хельсинки', 'finland', 'helsinki']],
    ['SE', ['швеци', 'стокгольм', 'sweden', 'stockholm']],
    ['GB', ['великобритан', 'лондон', 'англия', 'uk', 'london', 'britain', 'england']],
    ['FR', ['франци', 'париж', 'france', 'paris']],
    ['PL', ['польш', 'варшав', 'poland', 'warsaw']],
    ['TR', ['турци', 'стамбул', 'turkey', 'istanbul']],
    ['KZ', ['казахстан', 'алмат', 'астан', 'kazakhstan', 'almaty', 'astana']],
    ['UA', ['украин', 'киев', 'ukraine', 'kyiv']],
    ['CH', ['швейцари', 'цюрих', 'switzerland', 'zurich']],
    ['JP', ['япони', 'токио', 'japan', 'tokyo']],
    ['SG', ['сингапур', 'singapore']],
    ['CA', ['канад', 'canada']],
    ['IT', ['итали', 'рим', 'милан', 'italy', 'rome', 'milan']],
    ['ES', ['испани', 'мадрид', 'spain', 'madrid']],
    ['AT', ['австри', 'вена', 'austria', 'vienna']]
  ];
  for (const [cc, words] of keywords) {
    if (words.some(w => lower.includes(w))) {
      return cc;
    }
  }

  return 'VPN';
}

function renderFlagSVG(cc, width = 24, height = 16) {
  cc = (cc || '').toUpperCase();
  const r = 3;
  let inner = '';

  switch (cc) {
    case 'DE':
      inner = `
        <rect width="24" height="5.33" fill="#1A1A1A"/>
        <rect y="5.33" width="24" height="5.33" fill="#D00000"/>
        <rect y="10.66" width="24" height="5.34" fill="#FFCE00"/>
      `;
      break;
    case 'NL':
      inner = `
        <rect width="24" height="5.33" fill="#AE1C28"/>
        <rect y="5.33" width="24" height="5.33" fill="#FFFFFF"/>
        <rect y="10.66" width="24" height="5.34" fill="#21468B"/>
      `;
      break;
    case 'RU':
      inner = `
        <rect width="24" height="5.33" fill="#FFFFFF"/>
        <rect y="5.33" width="24" height="5.33" fill="#0039A6"/>
        <rect y="10.66" width="24" height="5.34" fill="#D52B1E"/>
      `;
      break;
    case 'FR':
      inner = `
        <rect width="8" height="16" fill="#002395"/>
        <rect x="8" width="8" height="16" fill="#FFFFFF"/>
        <rect x="16" width="8" height="16" fill="#ED2939"/>
      `;
      break;
    case 'IT':
      inner = `
        <rect width="8" height="16" fill="#009246"/>
        <rect x="8" width="8" height="16" fill="#FFFFFF"/>
        <rect x="16" width="8" height="16" fill="#CE2B37"/>
      `;
      break;
    case 'PL':
      inner = `
        <rect width="24" height="8" fill="#FFFFFF"/>
        <rect y="8" width="24" height="8" fill="#DC143C"/>
      `;
      break;
    case 'UA':
      inner = `
        <rect width="24" height="8" fill="#0057B7"/>
        <rect y="8" width="24" height="8" fill="#FFD700"/>
      `;
      break;
    case 'FI':
      inner = `
        <rect width="24" height="16" fill="#FFFFFF"/>
        <rect y="5.5" width="24" height="5" fill="#002F6C"/>
        <rect x="7" width="5" height="16" fill="#002F6C"/>
      `;
      break;
    case 'SE':
      inner = `
        <rect width="24" height="16" fill="#006AA7"/>
        <rect y="5.5" width="24" height="5" fill="#FECC00"/>
        <rect x="7" width="5" height="16" fill="#FECC00"/>
      `;
      break;
    case 'US':
      inner = `
        <rect width="24" height="16" fill="#FFFFFF"/>
        <rect y="0" width="24" height="2.28" fill="#B22234"/>
        <rect y="4.56" width="24" height="2.28" fill="#B22234"/>
        <rect y="9.12" width="24" height="2.28" fill="#B22234"/>
        <rect y="13.68" width="24" height="2.32" fill="#B22234"/>
        <rect width="10" height="9.12" fill="#3C3B6E"/>
        <circle cx="3" cy="3" r="0.75" fill="#FFFFFF"/>
        <circle cx="7" cy="3" r="0.75" fill="#FFFFFF"/>
        <circle cx="5" cy="6" r="0.75" fill="#FFFFFF"/>
      `;
      break;
    case 'GB':
    case 'UK':
      inner = `
        <rect width="24" height="16" fill="#012169"/>
        <path d="M0,0 L24,16 M24,0 L0,16" stroke="#FFFFFF" stroke-width="3"/>
        <path d="M0,0 L24,16 M24,0 L0,16" stroke="#C8102E" stroke-width="1.5"/>
        <path d="M12,0 V16 M0,8 H24" stroke="#FFFFFF" stroke-width="5"/>
        <path d="M12,0 V16 M0,8 H24" stroke="#C8102E" stroke-width="3"/>
      `;
      break;
    case 'JP':
      inner = `
        <rect width="24" height="16" fill="#FFFFFF"/>
        <circle cx="12" cy="8" r="4.8" fill="#BC002D"/>
      `;
      break;
    case 'CH':
      inner = `
        <rect width="24" height="16" fill="#DA291C"/>
        <rect x="9" y="3.5" width="6" height="9" fill="#FFFFFF"/>
        <rect x="6" y="5" width="12" height="6" fill="#FFFFFF"/>
      `;
      break;
    case 'AT':
      inner = `
        <rect width="24" height="5.33" fill="#ED2939"/>
        <rect y="5.33" width="24" height="5.33" fill="#FFFFFF"/>
        <rect y="10.66" width="24" height="5.34" fill="#ED2939"/>
      `;
      break;
    case 'TR':
      inner = `
        <rect width="24" height="16" fill="#E30A17"/>
        <circle cx="9" cy="8" r="4.5" fill="#FFFFFF"/>
        <circle cx="10.2" cy="8" r="3.6" fill="#E30A17"/>
        <polygon points="14,8 16,8.6 15.3,7 16,5.4 14,6" fill="#FFFFFF"/>
      `;
      break;
    case 'KZ':
      inner = `
        <rect width="24" height="16" fill="#00AFCA"/>
        <circle cx="12" cy="8" r="3.2" fill="#FEC50C"/>
      `;
      break;
    case 'CF':
    case 'WARP':
      inner = `
        <rect width="24" height="16" fill="#F38020"/>
        <path d="M17.3 12.3c.1-.3.2-.6.2-.9 0-1.7-1.3-3-3-3-.4 0-.7.1-1 .2C13 7.1 11.6 6 10 6c-2.2 0-4 1.8-4 4 0 .3 0 .5.1.8-1.2.5-2.1 1.7-2.1 3.2 0 1.9 1.6 3.5 3.5 3.5h9.8c1.5 0 2.7-1.2 2.7-2.7 0-1.2-.8-2.2-2-2.5z" fill="#FFFFFF" transform="matrix(0.55 0 0 0.55 4 2)"/>
      `;
      break;
    default:
      inner = `
        <rect width="24" height="16" fill="#2B2930"/>
        <text x="12" y="11" font-size="8" font-weight="700" fill="#D0BCFF" text-anchor="middle" font-family="sans-serif">VPN</text>
      `;
      break;
  }

  return `
    <svg width="${width}" height="${height}" viewBox="0 0 24 16" fill="none" xmlns="http://www.w3.org/2000/svg" style="border-radius: ${r}px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.3); vertical-align: middle; flex-shrink: 0;">
      <clipPath id="clip-${cc}-${width}">
        <rect width="24" height="16" rx="${r}" fill="#000"/>
      </clipPath>
      <g clip-path="url(#clip-${cc}-${width})">
        ${inner}
      </g>
    </svg>
  `;
}

// ==================== PRIVACY MODE (BLUR) ====================
function updatePrivacyBlurUI() {
  const isBlurred = !!state.settings.privacy_blur;

  // Toggle blur simultaneously on all privacy targets: server IPs, node names, hero node name, subscription title/name
  document.querySelectorAll('.privacy-target, .node-name, .node-server-ip, #sub-traffic-name, .sub-traffic-name, .node-sub-name, .node-tab-chip-name').forEach(el => {
    el.classList.toggle('privacy-blurred', isBlurred);
  });

  const eyeIconOpen = `<svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/></svg>`;
  const eyeIconSlashed = `<svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 7c2.76 0 5 2.24 5 5 0 .65-.13 1.26-.36 1.83l2.92 2.92c1.51-1.26 2.7-2.89 3.44-4.75-1.73-4.39-6-7.5-11-7.5-1.4 0-2.74.25-3.98.7l2.16 2.16C10.74 7.13 11.35 7 12 7zM2 4.27l2.28 2.28.46.46C3.08 8.3 1.78 10.02 1 12c1.73 4.39 6 7.5 11 7.5 1.55 0 3.03-.3 4.38-.84l.42.42L19.73 22 21 20.73 3.27 3 2 4.27zM7.53 9.8l1.55 1.55c-.05.21-.08.43-.08.65 0 1.66 1.34 3 3 3 .22 0 .44-.03.65-.08l1.55 1.55c-.67.33-1.41.53-2.2.53-2.76 0-5-2.24-5-5 0-.79.2-1.53.53-2.2zm4.31-.78l3.15 3.15.02-.16c0-1.66-1.34-3-3-3l-.17.01z"/></svg>`;

  ['dash-eye-btn', 'nodes-eye-btn'].forEach(id => {
    const btn = document.getElementById(id);
    if (btn) {
      btn.innerHTML = isBlurred ? eyeIconSlashed : eyeIconOpen;
      btn.classList.toggle('active', isBlurred);
      btn.title = isBlurred ? 'Режим анонимности включен (нажмите для отображения)' : 'Включить режим анонимности (скрыть IP и сервер)';
    }
  });

  const setPrivacy = document.getElementById('set-privacy-blur');
  if (setPrivacy) {
    setPrivacy.checked = isBlurred;
  }
}

async function togglePrivacyBlur() {
  state.settings.privacy_blur = !state.settings.privacy_blur;
  updatePrivacyBlurUI();
  try {
    await ToggleSmartRule('privacy_blur', state.settings.privacy_blur);
    showToast(state.settings.privacy_blur ? 'Режим анонимности включен' : 'Режим анонимности выключен');
  } catch (err) {
    console.error('Toggle privacy blur error:', err);
  }
}

// Bind Privacy Listeners
document.getElementById('dash-eye-btn')?.addEventListener('click', togglePrivacyBlur);
document.getElementById('nodes-eye-btn')?.addEventListener('click', togglePrivacyBlur);
document.getElementById('stat-server-ip')?.addEventListener('click', togglePrivacyBlur);
document.getElementById('sub-traffic-name')?.addEventListener('click', togglePrivacyBlur);
document.getElementById('set-privacy-blur')?.addEventListener('change', async (e) => {
  state.settings.privacy_blur = e.target.checked;
  updatePrivacyBlurUI();
  try {
    await ToggleSmartRule('privacy_blur', state.settings.privacy_blur);
  } catch (err) {
    console.error('Save privacy blur error:', err);
  }
});

// Bind Web Latency Test & Hero WARP Detour
document.getElementById('btn-web-test')?.addEventListener('click', handleWebTest);
document.getElementById('btn-hero-warp-detour')?.addEventListener('click', handleWarpDetourToggle);

// Bind Backup & Restore Profile
document.getElementById('btn-export-backup')?.addEventListener('click', handleExportBackup);
document.getElementById('btn-import-backup')?.addEventListener('click', handleImportBackup);

// Window Controls
document.getElementById('btn-win-min').addEventListener('click', () => WindowMin());
document.getElementById('btn-win-max').addEventListener('click', () => WindowToggleMax());
document.getElementById('btn-win-close').addEventListener('click', () => WindowClose());

// Navigation
function switchPage(pageName) {
  state.activePage = pageName;
  document.querySelectorAll('.nav-item').forEach(item => {
    item.classList.toggle('active', item.dataset.page === pageName);
  });
  document.querySelectorAll('.page').forEach(page => {
    page.classList.toggle('active', page.id === `page-${pageName}`);
  });

  if (pageName === 'routing') {
    if (state.activeRoutingSubtab === 'apps') {
      loadRunningApps();
    } else {
      loadGeoRules();
    }
  } else if (pageName === 'nodes') {
    loadNodes();
  } else if (pageName === 'traffic') {
    loadConnections();
  } else if (pageName === 'settings') {
    loadSettingsPage();
  }
}

document.querySelectorAll('.nav-item').forEach(item => {
  item.addEventListener('click', () => {
    const page = item.dataset.page;
    if (page) switchPage(page);
  });
});

// ==================== TIER-1 POWER FEATURES ====================

// 1. Select Fastest Node
async function handleSelectFastestNode() {
  showToast(t('select_fastest') || 'Поиск лучшего сервера...');
  try {
    const best = await SelectFastestNode();
    if (!best || !best.id) {
      showToast('Нет доступных узлов');
      return;
    }
    state.activeNodeId = best.id;
    const cc = detectNodeCountry(best);
    const displayName = cleanNodeDisplayName(best.name);
    const protoLabel = best.protocol_label || (best.protocol ? best.protocol.toUpperCase() : 'VLESS');

    // Update Hero immediately
    const heroNodeName = document.getElementById('hero-node-name');
    if (heroNodeName) heroNodeName.innerText = displayName;
    const heroProto = document.getElementById('hero-protocol-badge');
    if (heroProto) heroProto.innerText = protoLabel;
    const heroFlagEl = document.getElementById('hero-country-flag');
    if (heroFlagEl) {
      heroFlagEl.innerHTML = renderFlagSVG(cc, 22, 14);
      heroFlagEl.title = `Страна: ${cc}`;
    }

    renderNodePickerDropdown();
    renderNodesList();
    renderNodeCategoryTabs();
    updateDashboardSubWidget();
    const pingStr = best.ping > 0 ? ` (${best.ping} ms)` : '';
    showToast(`${t('fastest_selected') || 'Выбран лучший узел'}: ${displayName}${pingStr}`);
  } catch (err) {
    showToast(`Ошибка выбора: ${err}`);
  }
}

// 2. Real Web Latency Test (HTTP 204)
async function handleWebTest() {
  const btn = document.getElementById('btn-web-test');
  if (btn) btn.classList.add('loading');
  showToast(t('web_test_running') || 'Проверка веб-соединения (HTTP 204)...');
  try {
    const res = await TestWebLatency();
    if (res && res.ok) {
      const ms = Math.round(res.latency || 0);
      const okMsg = (t('web_test_ok') || 'Веб: {latency} ms (204 OK)').replace('{latency}', ms);
      showToast(okMsg);
      const pingEl = document.getElementById('stat-ping');
      if (pingEl) {
        pingEl.innerText = `${ms} ms`;
      }
    } else {
      const errMsg = res && res.error ? res.error : '';
      showToast(`${t('web_test_fail') || 'Тест не пройден'}${errMsg ? ': ' + errMsg : ''}`);
    }
  } catch (err) {
    showToast(`${t('web_test_fail')}: ${err}`);
  } finally {
    if (btn) btn.classList.remove('loading');
  }
}

// 3. WARP Detour Toggle & Badge
function updateWarpHeroBadge() {
  const btn = document.getElementById('btn-hero-warp-detour');
  if (!btn) return;
  const isWarp = !!state.settings.warp_detour;
  if (isWarp) {
    btn.classList.add('active');
    btn.setAttribute('aria-pressed', 'true');
    btn.title = 'WARP Detour: ВКЛ (чистый IP Cloudflare для зарубежных сервисов)';
  } else {
    btn.classList.remove('active');
    btn.setAttribute('aria-pressed', 'false');
    btn.title = 'WARP Detour: ВЫКЛ (нажмите для включения маскировки через Cloudflare)';
  }
}

let isTogglingWarp = false;
async function handleWarpDetourToggle() {
  if (isTogglingWarp) return;
  isTogglingWarp = true;
  const btn = document.getElementById('btn-hero-warp-detour');
  if (btn) {
    btn.style.pointerEvents = 'none';
    btn.style.opacity = '0.6';
  }
  try {
    const enabled = await ToggleWarpDetour();
    state.settings.warp_detour = enabled;
    updateWarpHeroBadge();
    const setWarp = document.getElementById('set-warp-detour');
    if (setWarp) setWarp.checked = enabled;
    showToast(enabled ? 'WARP Detour: ВКЛ (трафик идет через чистый IP Cloudflare)' : 'WARP Detour: ВЫКЛ');
  } catch (err) {
    showToast(`Ошибка переключения WARP Detour: ${err}`);
  } finally {
    isTogglingWarp = false;
    if (btn) {
      btn.style.pointerEvents = 'auto';
      btn.style.opacity = '1';
    }
  }
}

// 4. Backup & Restore Profile
async function handleExportBackup() {
  try {
    const path = await ExportProfile();
    if (path) {
      showToast(t('backup_export_success') || 'Резервная копия успешно экспортирована');
    }
  } catch (err) {
    showToast(`Ошибка экспорта: ${err}`);
  }
}

async function handleImportBackup() {
  try {
    const path = await ImportProfile();
    if (path) {
      showToast(t('backup_import_success') || 'Профиль успешно восстановлен');
      state.settings = await GetSettings();
      state.nodes = (await GetNodes()) || [];
      state.subscriptions = (await GetSubscriptions()) || [];
      try {
        state.activeNodeId = await GetActiveNodeID();
      } catch (_) {}
      updateStatusUI();
      updateHeroNodeDisplay();
      updateWarpHeroBadge();
      renderNodeCategoryTabs();
      renderNodesList();
      loadSettingsPage();
    }
  } catch (err) {
    showToast(`${t('backup_import_error') || 'Ошибка импорта'}: ${err}`);
  }
}

// ==================== NODE PICKER DROPDOWN ====================
const nodePickerDropdown = document.getElementById('node-picker-dropdown');
const heroNodeBtn = document.getElementById('hero-node-btn');

function renderNodePickerDropdown() {
  if (!nodePickerDropdown) return;

  const curId = state.activeNodeId;
  const isBlurred = !!state.settings.privacy_blur;
  const nodes = state.nodes || [];

  if (nodes.length === 0) {
    nodePickerDropdown.innerHTML = `
      <div class="node-picker-header">Выбор сервера</div>
      <div class="node-picker-divider"></div>
      <div style="padding: 16px; text-align: center; color: var(--on-surface-muted); font-size: 13px;">
        Нет доступных узлов
      </div>
      <div class="node-picker-footer">
        <a href="#" class="node-picker-manage-link" id="node-picker-open-nodes">Все узлы и подписки →</a>
      </div>
    `;
    document.getElementById('node-picker-open-nodes')?.addEventListener('click', (e) => {
      e.preventDefault();
      closeNodePicker();
      switchPage('nodes');
    });
    return;
  }

  // Group nodes: Cloudflare WARP, then subscriptions, then Пользовательские
  const groups = {};
  nodes.forEach(n => {
    let groupName = 'Пользовательские';
    const proto = (n.protocol || '').toLowerCase();
    const nameLower = (n.name || '').toLowerCase();
    if (nameLower.includes('warp') || nameLower.includes('cloudflare') || proto === 'wireguard' || proto === 'warp' || proto === 'masque') {
      groupName = 'Cloudflare WARP';
    } else if (n.subscription_name) {
      groupName = n.subscription_name;
    }
    if (!groups[groupName]) groups[groupName] = [];
    groups[groupName].push(n);
  });

  let html = `
    <div class="node-picker-header" style="display: flex; justify-content: space-between; align-items: center;">
      <span>${t('hero_server_selection') || 'Выбор сервера'}</span>
      <button id="btn-select-fastest-drop" class="node-picker-fastest-pill" title="${t('select_fastest') || 'Выбрать самый быстрый узел'}">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><path d="M7 2v11h3v9l7-12h-4l4-8z"/></svg>
        <span>${t('btn_select_fastest') || 'Лучший'}</span>
      </button>
    </div>
    <div class="node-picker-divider"></div>
  `;

  // Order groups: Cloudflare WARP first, then Пользовательские, then subscriptions
  const orderedGroupNames = Object.keys(groups).sort((a, b) => {
    if (a === 'Cloudflare WARP') return -1;
    if (b === 'Cloudflare WARP') return 1;
    if (a === 'Пользовательские') return -1;
    if (b === 'Пользовательские') return 1;
    return a.localeCompare(b);
  });

  orderedGroupNames.forEach(gName => {
    html += `<div class="node-picker-group-title">── ${gName} ──</div>`;
    const sortedGroupNodes = [...groups[gName]].sort((a, b) => {
      if (a.id === curId) return -1;
      if (b.id === curId) return 1;
      return 0;
    });
    sortedGroupNodes.forEach(n => {
      const isActive = n.id === curId;
      const cc = detectNodeCountry(n);
      const flagSvg = renderFlagSVG(cc, 20, 13);
      const displayName = cleanNodeDisplayName(n.name);
      const protoLabel = n.protocol_label || (n.protocol ? n.protocol.toUpperCase() : 'VLESS');
      const pingText = n.ping > 0 ? `• ${n.ping} ms` : '';

      html += `
        <div class="node-picker-item ${isActive ? 'active' : ''}" data-id="${n.id}">
          <div class="node-picker-item-left">
            <span class="node-flag-badge" title="Страна: ${cc}">${flagSvg}</span>
            <span class="node-picker-name">${displayName}</span>
            <span class="node-picker-proto">[${protoLabel}]</span>
          </div>
          <div class="node-picker-item-right tabular">
            <span class="node-picker-ping">${pingText}</span>
            ${isActive ? '<span class="node-picker-check">✓</span>' : ''}
          </div>
        </div>
      `;
    });
  });

  html += `
    <div class="node-picker-divider"></div>
    <div id="btn-node-picker-all" class="node-picker-all-btn">Все узлы и подписки →</div>
  `;

  nodePickerDropdown.innerHTML = html;

  // Add click handlers for node items in dropdown
  nodePickerDropdown.querySelectorAll('.node-picker-item').forEach(item => {
    item.addEventListener('click', async () => {
      const nodeId = item.dataset.id;
      const selectedNode = state.nodes.find(n => n.id === nodeId);
      if (!selectedNode) return;

      closeNodePicker();
      state.activeNodeId = nodeId;

      const cc = detectNodeCountry(selectedNode);
      const displayName = cleanNodeDisplayName(selectedNode.name);
      const protoLabel = selectedNode.protocol_label || (selectedNode.protocol ? selectedNode.protocol.toUpperCase() : 'VLESS');

      // Update Hero immediately
      document.getElementById('hero-node-name').innerText = displayName;
      document.getElementById('hero-protocol-badge').innerText = protoLabel;
      const heroFlagEl = document.getElementById('hero-country-flag');
      if (heroFlagEl) {
        heroFlagEl.innerHTML = renderFlagSVG(cc, 22, 14);
        heroFlagEl.title = `Страна: ${cc}`;
      }

      showToast(state.isConnected ? `Переключение узла: ${displayName}...` : `Выбран узел: ${displayName}`);

      try {
        await SetActiveNodeID(nodeId);
        state.activeNodeId = nodeId;
        renderNodeCategoryTabs();
        updateDashboardSubWidget();
        if (state.isConnected) {
          showToast(`Подключено: ${displayName}`);
        }
      } catch (err) {
        showToast(`Ошибка переключения: ${err}`);
      } finally {
        setTimeout(pollTelemetry, 500);
      }
    });
  });

  // "Выбрать лучший" button in dropdown header
  document.getElementById('btn-select-fastest-drop')?.addEventListener('click', async (e) => {
    e.stopPropagation();
    closeNodePicker();
    await handleSelectFastestNode();
  });

  // "Все узлы и подписки →" link
  document.getElementById('btn-node-picker-all')?.addEventListener('click', () => {
    closeNodePicker();
    switchPage('nodes');
  });
}

function openNodePicker() {
  renderNodePickerDropdown();
  nodePickerDropdown.style.display = 'block';
  // Unblur hero node name while picking so user clearly sees what they select
  document.getElementById('hero-node-name')?.classList.remove('privacy-blurred');
}

function closeNodePicker() {
  if (nodePickerDropdown) {
    nodePickerDropdown.style.display = 'none';
  }
  // Restore privacy blur if it was enabled
  if (state.settings.privacy_blur) {
    document.getElementById('hero-node-name')?.classList.add('privacy-blurred');
  }
}

heroNodeBtn?.addEventListener('click', (e) => {
  e.stopPropagation();
  if (nodePickerDropdown.style.display === 'block') {
    closeNodePicker();
  } else {
    openNodePicker();
  }
});

// Close dropdown on outside click
document.addEventListener('click', (e) => {
  if (nodePickerDropdown && nodePickerDropdown.style.display === 'block') {
    if (!nodePickerDropdown.contains(e.target) && !heroNodeBtn.contains(e.target)) {
      closeNodePicker();
    }
  }
});

// Admin Status and Elevation
async function updateAdminStatus() {
  try {
    const admin = await IsAdmin();
    const badge = document.getElementById('btn-admin-status');
    if (!badge) return;
    badge.style.display = 'inline-flex';
    if (admin) {
      badge.className = 'titlebar-admin-badge is-admin';
      badge.innerHTML = `<svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm-2 16l-4-4 1.41-1.41L10 14.17l6.59-6.59L18 9l-8 8z"/></svg> Admin`;
      badge.title = 'Приложение запущено с правами администратора';
      badge.onclick = null;
    } else {
      badge.className = 'titlebar-admin-badge not-admin';
      badge.innerHTML = `<svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm1 14h-2v-2h2v2zm0-4h-2V7h2v4z"/></svg> Запустить как Admin`;
      badge.title = 'Нажмите для перезапуска с правами администратора (UAC)';
      badge.onclick = async () => {
        try {
          showToast('Запрос прав администратора...');
          await RestartAsAdmin();
        } catch (err) {
          showToast(`Ошибка: ${err}`);
        }
      };
    }
  } catch (_) {}
}

// VPN Conflict Detection
async function checkAndDisplayConflicts() {
  try {
    const conflicts = await CheckVPNConflicts();
    const banner = document.getElementById('vpn-conflict-banner');
    const textEl = document.getElementById('conflict-banner-text');
    if (!banner || !textEl) return;
    if (conflicts && conflicts.length > 0) {
      textEl.innerText = `Конфликт: работает сторонний VPN (${conflicts.join(', ')})`;
      banner.style.display = 'flex';
    } else {
      banner.style.display = 'none';
    }
  } catch (_) {}
}

document.getElementById('btn-kill-conflicts')?.addEventListener('click', async () => {
  try {
    showToast('Закрытие конфликтующих процессов...');
    await KillConflictingVPNs();
    showToast('Конфликтующие процессы закрыты');
    setTimeout(checkAndDisplayConflicts, 600);
  } catch (err) {
    showToast(`Ошибка: ${err}`);
  }
});

// VPN Toggle
const btnPower = document.getElementById('btn-power');
btnPower.addEventListener('click', async () => {
  try {
    btnPower.style.opacity = '0.5';
    btnPower.style.pointerEvents = 'none';

    const willConnect = !state.isConnected;
    if (willConnect) {
      const conflicts = await CheckVPNConflicts();
      if (conflicts && conflicts.length > 0) {
        showToast(`Внимание: работает сторонний VPN (${conflicts.join(', ')}). Возможно падение соединения.`);
      }
    }

    await ToggleVPN(willConnect);
    state.isConnected = willConnect;
    updateStatusUI();
    showToast(willConnect ? t('vpn_starting') : t('toast_vpn_disconnected'));
  } catch (err) {
    showToast(`Ошибка: ${err}`);
    document.getElementById('btn-hero-logs')?.classList.add('has-error');
  } finally {
    btnPower.style.opacity = '1';
    btnPower.style.pointerEvents = 'auto';
    setTimeout(pollTelemetry, 400);
  }
});

function updateStatusUI() {
  const dot = document.getElementById('hero-status-dot');
  const titleBadge = document.getElementById('title-status-badge');
  const heroText = document.getElementById('hero-status-text');

  if (state.isConnected) {
    dot.classList.add('connected');
    btnPower.classList.add('connected');
    titleBadge.classList.add('connected');
    titleBadge.innerText = t('status_connected');
    heroText.innerText = t('status_connected');
  } else {
    dot.classList.remove('connected');
    btnPower.classList.remove('connected');
    titleBadge.classList.remove('connected');
    titleBadge.innerText = t('status_disconnected');
    heroText.innerText = t('status_disconnected');
  }
  updateWarpHeroBadge();
}

// Telemetry & Graph
async function pollTelemetry() {
  try {
    const stats = await GetTrafficStats();
    state.isConnected = stats.connected;
    updateStatusUI();

    document.getElementById('stat-down-speed').innerText = formatSpeed(stats.down_speed);
    document.getElementById('stat-up-speed').innerText = formatSpeed(stats.up_speed);
    document.getElementById('stat-down-total').innerText = `${t('total_prefix', 'Всего')}: ${formatBytes(stats.total_down)}`;
    document.getElementById('stat-up-total').innerText = `${t('total_prefix', 'Всего')}: ${formatBytes(stats.total_up)}`;
    document.getElementById('stat-ping').innerText = stats.ping > 0 ? `${stats.ping} ms` : '— ms';
    document.getElementById('stat-server-ip').innerText = stats.ip || '—';
    updateQuickModeCardUI(state.settings.split_mode || stats.mode);

    if (stats.node_name) {
      document.getElementById('hero-node-name').innerText = cleanNodeDisplayName(stats.node_name);
    }
    updateDashboardSubWidget();

    // Push to chart
    state.speedHistory.down.shift();
    state.speedHistory.down.push(stats.down_speed || 0);
    state.speedHistory.up.shift();
    state.speedHistory.up.push(stats.up_speed || 0);
    drawCanvasGraph();

    if (state.activePage === 'traffic') {
      loadConnections();
    }
  } catch (e) {
    // Polling error catch
  }
}

function drawCanvasGraph() {
  const canvas = document.getElementById('speed-canvas');
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  const dpr = window.devicePixelRatio || 1;
  const rect = canvas.getBoundingClientRect();

  if (canvas.width !== rect.width * dpr || canvas.height !== rect.height * dpr) {
    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
  }

  ctx.clearRect(0, 0, canvas.width, canvas.height);

  const w = canvas.width;
  const h = canvas.height;
  const paddingBottom = 8;
  const usableH = h - paddingBottom - 10;

  const maxDown = Math.max(...state.speedHistory.down, 1024 * 50);
  const maxUp = Math.max(...state.speedHistory.up, 1024 * 50);
  const maxVal = Math.max(maxDown, maxUp);

  const pointsCount = state.speedHistory.down.length;
  const step = w / (pointsCount - 1);

  function renderLine(data, strokeColor, fillColor) {
    ctx.beginPath();
    ctx.moveTo(0, h);

    data.forEach((val, idx) => {
      const x = idx * step;
      const y = h - paddingBottom - (val / maxVal) * usableH;
      if (idx === 0) {
        ctx.lineTo(x, y);
      } else {
        const prevX = (idx - 1) * step;
        const prevVal = data[idx - 1];
        const prevY = h - paddingBottom - (prevVal / maxVal) * usableH;
        const cpX1 = prevX + (x - prevX) / 2;
        const cpX2 = prevX + (x - prevX) / 2;
        ctx.bezierCurveTo(cpX1, prevY, cpX2, y, x, y);
      }
    });

    ctx.lineTo(w, h);
    ctx.closePath();
    ctx.fillStyle = fillColor;
    ctx.fill();

    ctx.beginPath();
    data.forEach((val, idx) => {
      const x = idx * step;
      const y = h - paddingBottom - (val / maxVal) * usableH;
      if (idx === 0) {
        ctx.moveTo(x, y);
      } else {
        const prevX = (idx - 1) * step;
        const prevVal = data[idx - 1];
        const prevY = h - paddingBottom - (prevVal / maxVal) * usableH;
        const cpX1 = prevX + (x - prevX) / 2;
        const cpX2 = prevX + (x - prevX) / 2;
        ctx.bezierCurveTo(cpX1, prevY, cpX2, y, x, y);
      }
    });
    ctx.strokeStyle = strokeColor;
    ctx.lineWidth = 2 * dpr;
    ctx.stroke();
  }

  renderLine(state.speedHistory.down, '#00C853', 'rgba(0, 200, 83, 0.15)');
  renderLine(state.speedHistory.up, '#D0BCFF', 'rgba(208, 188, 255, 0.12)');
}

// ==================== ROUTING SUBTABS ====================
const btnSubtabApps = document.getElementById('btn-subtab-apps');
const btnSubtabGeo = document.getElementById('btn-subtab-geo');
const paneSubtabApps = document.getElementById('subtab-apps');
const paneSubtabGeo = document.getElementById('subtab-geo');

btnSubtabApps.addEventListener('click', () => {
  state.activeRoutingSubtab = 'apps';
  btnSubtabApps.classList.add('active');
  btnSubtabGeo.classList.remove('active');
  paneSubtabApps.classList.add('active');
  paneSubtabGeo.classList.remove('active');
  loadRunningApps();
});

btnSubtabGeo.addEventListener('click', () => {
  state.activeRoutingSubtab = 'geo';
  btnSubtabGeo.classList.add('active');
  btnSubtabApps.classList.remove('active');
  paneSubtabGeo.classList.add('active');
  paneSubtabApps.classList.remove('active');
  loadGeoRules();
});

// Routing: Apps (Pinned Enabled at Top)
async function loadRunningApps() {
  try {
    const listEl = document.getElementById('process-list');
    listEl.innerHTML = `<div style="padding: 20px; text-align: center; color: var(--on-surface-muted);">Сканирование запущенных процессов...</div>`;
    state.apps = await GetRunningApps();
    renderAppList();
  } catch (err) {
    showToast(`Ошибка загрузки процессов: ${err}`);
  }
}

function renderAppList() {
  const listEl = document.getElementById('process-list');
  const filter = (state.runningAppsFilter || '').toLowerCase();
  let filtered = state.apps.filter(a => a.name.toLowerCase().includes(filter) || a.path.toLowerCase().includes(filter));

  // Pinned Enabled Apps Always at the Top
  filtered.sort((a, b) => {
    const aEn = !!(a.enabled || a.is_enabled);
    const bEn = !!(b.enabled || b.is_enabled);
    if (aEn !== bEn) {
      return aEn ? -1 : 1;
    }
    return (a.name || '').localeCompare(b.name || '', undefined, { sensitivity: 'base' });
  });

  if (filtered.length === 0) {
    listEl.innerHTML = `<div style="padding: 30px; text-align: center; color: var(--on-surface-muted);">Нет подходящих приложений</div>`;
    return;
  }

  listEl.innerHTML = '';
  filtered.forEach(app => {
    const item = document.createElement('div');
    item.className = 'process-item';

    const firstLetter = (app.name.charAt(0) || 'A').toUpperCase();
    const avatarContent = app.icon
      ? `<img src="${app.icon}" class="process-avatar-img" alt=""/>`
      : firstLetter;
    const isChecked = !!(app.enabled || app.is_enabled);

    item.innerHTML = `
      <div class="process-info">
        <div class="process-avatar">${avatarContent}</div>
        <div class="process-texts">
          <div class="process-name-row">
            <span class="process-name">${app.name}</span>
            <span class="process-badge">${app.category || 'Приложение'}</span>
          </div>
          <div class="process-path" title="${app.path}">${app.path}</div>
        </div>
      </div>
      <label class="m3-switch">
        <input type="checkbox" ${isChecked ? 'checked' : ''}/>
        <span class="switch-slider"></span>
      </label>
    `;

    const toggle = item.querySelector('input');
    toggle.addEventListener('change', async (e) => {
      const checked = e.target.checked;
      app.enabled = checked;
      app.is_enabled = checked;

      // Re-sort immediately so it dynamically moves to the top or bottom
      renderAppList();

      try {
        await ToggleAppRule(app.path, app.name, checked);
      } catch (err) {
        showToast(`Ошибка правила: ${err}`);
      }
    });

    listEl.appendChild(item);
  });
}

document.getElementById('input-search-apps').addEventListener('input', (e) => {
  state.runningAppsFilter = e.target.value;
  renderAppList();
});

document.getElementById('btn-refresh-apps').addEventListener('click', () => {
  loadRunningApps();
});

document.getElementById('btn-add-exe').addEventListener('click', async () => {
  try {
    const path = await SelectExeFile();
    if (!path) return;
    await AddCustomApp(path);
    showToast('Приложение добавлено в правила');
    loadRunningApps();
  } catch (err) {
    showToast(`Ошибка выбора файла: ${err}`);
  }
});

// Routing: Quick Mode Selector Dropdown
const quickModeDropdown = document.getElementById('quick-mode-dropdown');
const quickModeCard = document.getElementById('card-quick-mode');

function updateQuickModeCardUI(mode) {
  const currentMode = mode || state.settings.split_mode || 'whitelist';
  const statModeEl = document.getElementById('stat-mode');
  const iconBoxEl = document.getElementById('quick-mode-icon-box');

  let modeKey = 'mode_whitelist_short';
  let colorClass = 'blue';
  let iconSvg = `<svg class="telemetry-icon" viewBox="0 0 24 24" fill="currentColor"><path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm-2 16l-4-4 1.41-1.41L10 14.17l6.59-6.59L18 9l-8 8z"/></svg>`;

  if (currentMode === 'blacklist') {
    modeKey = 'mode_blacklist_short';
    colorClass = 'orange';
    iconSvg = `<svg class="telemetry-icon" viewBox="0 0 24 24" fill="currentColor"><path d="M10 18h4v-2h-4v2zM3 6v2h18V6H3zm3 7h12v-2H6v2z"/></svg>`;
  } else if (currentMode === 'global') {
    modeKey = 'mode_global_short';
    colorClass = 'purple';
    iconSvg = `<svg class="telemetry-icon" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"/></svg>`;
  }

  if (statModeEl) {
    statModeEl.innerText = t(modeKey);
  }
  if (iconBoxEl) {
    iconBoxEl.className = `telemetry-icon-box ${colorClass}`;
    iconBoxEl.innerHTML = iconSvg;
  }

  // Update active item in quick-mode-dropdown
  document.querySelectorAll('.mode-dropdown-item').forEach(item => {
    item.classList.toggle('active', item.dataset.mode === currentMode);
  });

  // Update active item in routing tab mode-cards
  document.querySelectorAll('.mode-card').forEach(c => {
    c.classList.toggle('active', c.dataset.mode === currentMode);
  });
}

function openQuickModeDropdown() {
  if (!quickModeDropdown) return;
  updateQuickModeCardUI(state.settings.split_mode);
  quickModeDropdown.style.display = 'block';
}

function closeQuickModeDropdown() {
  if (quickModeDropdown) {
    quickModeDropdown.style.display = 'none';
  }
}

if (quickModeCard) {
  quickModeCard.addEventListener('click', (e) => {
    if (e.target.closest('.mode-dropdown-item') || e.target.closest('.mode-dropdown-header')) {
      return;
    }
    const isOpen = quickModeDropdown && quickModeDropdown.style.display === 'block';
    if (isOpen) {
      closeQuickModeDropdown();
    } else {
      openQuickModeDropdown();
    }
  });

  quickModeCard.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      const isOpen = quickModeDropdown && quickModeDropdown.style.display === 'block';
      if (isOpen) closeQuickModeDropdown();
      else openQuickModeDropdown();
    } else if (e.key === 'Escape') {
      closeQuickModeDropdown();
    }
  });
}

// Close when clicking outside
document.addEventListener('click', (e) => {
  if (quickModeCard && !quickModeCard.contains(e.target)) {
    closeQuickModeDropdown();
  }
});

async function selectSplitMode(targetMode) {
  closeQuickModeDropdown();
  if (!targetMode) return;
  state.settings.split_mode = targetMode;
  updateQuickModeCardUI(targetMode);

  try {
    await SaveSettings(state.settings);

    // Hot reload tunnel if connected
    if (state.isConnected) {
      try {
        await ToggleVPN(true);
      } catch (e) {
        console.warn('Hot reload mode error:', e);
      }
    }

    let modeName = t('mode_whitelist_short');
    if (targetMode === 'blacklist') modeName = t('mode_blacklist_short');
    else if (targetMode === 'global') modeName = t('mode_global_short');

    showToast(`${t('mode_changed_to')}: ${modeName}`);
  } catch (err) {
    showToast(`${t('settings_error_save')}: ${err}`);
  }
}

document.querySelectorAll('.mode-dropdown-item').forEach(item => {
  item.addEventListener('click', (e) => {
    e.stopPropagation();
    const mode = item.dataset.mode;
    selectSplitMode(mode);
  });
});

document.querySelectorAll('.mode-card').forEach(card => {
  card.addEventListener('click', async () => {
    const mode = card.dataset.mode;
    selectSplitMode(mode);
  });
});

// ==================== GEO-RULES & CUSTOM LISTS ====================
async function loadGeoRules() {
  try {
    state.settings = await GetSettings();
    document.getElementById('geo-bypass-ru').checked = !!state.settings.bypass_ru;
    document.getElementById('geo-adblock').checked = !!state.settings.adblock;

    state.customLists = await GetCustomLists();
    renderCustomLists();
  } catch (err) {
    showToast(`Ошибка загрузки гео-правил: ${err}`);
  }
}

document.getElementById('geo-bypass-ru').addEventListener('change', async (e) => {
  const val = e.target.checked;
  state.settings.bypass_ru = val;
  try {
    await ToggleSmartRule('bypass_ru', val);
    showToast(val ? 'Обход сайтов РФ включен' : 'Обход сайтов РФ отключен');
  } catch (err) {
    showToast(`Ошибка: ${err}`);
  }
});

document.getElementById('geo-adblock').addEventListener('change', async (e) => {
  const val = e.target.checked;
  state.settings.adblock = val;
  try {
    await ToggleSmartRule('adblock', val);
    showToast(val ? 'Блокировка рекламы включена' : 'Блокировка рекламы отключена');
  } catch (err) {
    showToast(`Ошибка: ${err}`);
  }
});

function renderCustomLists() {
  const container = document.getElementById('geo-lists-container');
  if (!state.customLists || state.customLists.length === 0) {
    container.innerHTML = `
      <div style="background-color: rgba(255,255,255,0.02); border: 1px dashed rgba(255,255,255,0.08); border-radius: 12px; padding: 24px; text-align: center;">
        <div style="font-size: 13px; font-weight: 700; color: #E6E1E5;">${t('no_custom_lists')}</div>
        <div style="font-size: 11px; color: #948F99; margin-top: 4px;">${t('no_custom_lists_sub')}</div>
      </div>
    `;
    return;
  }

  container.innerHTML = '';
  state.customLists.forEach(cl => {
    const isProxy = (cl.action === 'proxy');
    const badgeText = isProxy ? (getCurrentLanguage() === 'en' ? 'VPN' : 'В VPN') : (getCurrentLanguage() === 'en' ? 'Direct' : 'Напрямую');
    const badgeClass = isProxy ? 'proxy' : 'direct';

    const countText = `${cl.count || 0} ${t('rules_suffix')}`;
    const timeText = cl.last_updated || cl.updated_at || '—';

    const row = document.createElement('div');
    row.className = 'geo-list-item';
    row.innerHTML = `
      <div class="geo-badge ${badgeClass}">${badgeText}</div>
      <div class="geo-list-info">
        <div class="geo-list-name">${cl.name}</div>
        <div class="geo-list-sub">${countText} • ${t('updated_prefix')}: ${timeText}</div>
      </div>
      <div class="geo-list-actions">
        <button class="btn-icon-only danger btn-del-geo-list" title="${t('del_list_title')}">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
        </button>
        <label class="m3-switch">
          <input type="checkbox" class="toggle-geo-list" ${cl.enabled ? 'checked' : ''}/>
          <span class="switch-slider"></span>
        </label>
      </div>
    `;

    row.querySelector('.toggle-geo-list').addEventListener('change', async (e) => {
      const checked = e.target.checked;
      cl.enabled = checked;
      try {
        await ToggleCustomList(cl.id, checked);
      } catch (err) {
        showToast(`Ошибка: ${err}`);
      }
    });

    row.querySelector('.btn-del-geo-list').addEventListener('click', async () => {
      try {
        await DeleteCustomList(cl.id);
        showToast(`Список «${cl.name}» удален`);
        loadGeoRules();
      } catch (err) {
        showToast(`Ошибка удаления: ${err}`);
      }
    });

    container.appendChild(row);
  });
}

// Quick Presets
document.querySelectorAll('.preset-chip').forEach(chip => {
  chip.addEventListener('click', async () => {
    const preset = chip.dataset.preset;
    try {
      showToast('Добавление набора...');
      const added = await AddPresetList(preset);
      showToast(`Набор «${added.name}» добавлен!`);
      loadGeoRules();
    } catch (err) {
      showToast(`Ошибка набора: ${err}`);
    }
  });
});

// Refresh All Lists
document.getElementById('btn-refresh-geo-lists').addEventListener('click', async () => {
  const btn = document.getElementById('btn-refresh-geo-lists');
  btn.style.opacity = '0.5';
  btn.style.pointerEvents = 'none';
  showToast('Обновление всех списков...');
  try {
    const count = await RefreshAllCustomLists();
    showToast(`Обновлено списков: ${count}`);
    loadGeoRules();
  } catch (err) {
    showToast(`Ошибка обновления: ${err}`);
  } finally {
    btn.style.opacity = '1';
    btn.style.pointerEvents = 'auto';
  }
});

// Add Custom List Modal
const modalAddGeo = document.getElementById('modal-add-geo-list');
document.getElementById('btn-open-add-geo-list').addEventListener('click', () => {
  modalAddGeo.classList.add('open');
  document.getElementById('input-geo-name').value = '';
  document.getElementById('input-geo-url').value = '';
});

function closeAddGeoModal() {
  modalAddGeo.classList.remove('open');
}

document.getElementById('btn-browse-geo-file')?.addEventListener('click', async () => {
  try {
    const filePath = await SelectGeoRuleFile();
    if (filePath) {
      document.getElementById('input-geo-url').value = filePath;
      const nameInput = document.getElementById('input-geo-name');
      if (!nameInput.value.trim()) {
        const baseName = filePath.split(/[\\/]/).pop().replace(/\.[^/.]+$/, '');
        nameInput.value = baseName;
      }
    }
  } catch (err) {
    showToast(`Ошибка выбора файла: ${err}`);
  }
});

document.getElementById('btn-close-geo-modal-x').addEventListener('click', closeAddGeoModal);
document.getElementById('btn-cancel-geo-list').addEventListener('click', closeAddGeoModal);

document.getElementById('btn-confirm-add-geo-list').addEventListener('click', async () => {
  const name = document.getElementById('input-geo-name').value.trim();
  const url = document.getElementById('input-geo-url').value.trim();
  const actionRadio = document.querySelector('input[name="geo-action"]:checked');
  const action = actionRadio ? actionRadio.value : 'proxy';

  if (!name || !url) {
    showToast('Введите название и ссылку');
    return;
  }

  try {
    showToast('Добавление списка...');
    closeAddGeoModal();
    const added = await AddCustomList(name, url, action);
    showToast(`Список «${added.name}» добавлен (${added.count} правил)!`);
    loadGeoRules();
  } catch (err) {
    showToast(`Ошибка добавления: ${err}`);
  }
});

// ==================== NODES & SUBSCRIPTIONS ====================
async function loadNodes() {
  try {
    state.nodes = (await GetNodes()) || [];
    state.subscriptions = (await GetSubscriptions()) || [];
    try {
      state.activeNodeId = await GetActiveNodeID();
    } catch (_) {
      if (state.nodes.length > 0) state.activeNodeId = state.nodes[0].id;
      else state.activeNodeId = '';
    }
    renderNodeCategoryTabs();
    renderNodesList();
    updateHeroNodeDisplay();
  } catch (err) {
    showToast(`Ошибка загрузки узлов: ${err}`);
  }
}

// Render horizontal category filter chips above nodes list
function renderNodeCategoryTabs() {
  const tabsBar = document.getElementById('nodes-tabs-bar');
  if (!tabsBar) return;

  const nodes = state.nodes || [];
  const subs = state.subscriptions || [];

  const allCount = nodes.length;
  const customCount = nodes.filter(n => !n.subscription_id && !n.subscription_name).length;

  let tabsHtml = `
    <button class="node-tab-chip ${state.activeNodeCategory === 'all' ? 'active' : ''}" data-cat="all">
      ${t('tab_all_nodes')} (${allCount})
    </button>
  `;

  subs.forEach(sub => {
    const subCount = nodes.filter(n => n.subscription_id === sub.id || n.subscription_name === sub.name).length;
    const isActive = state.activeNodeCategory === sub.id;
    tabsHtml += `
      <button class="node-tab-chip ${isActive ? 'active' : ''}" data-cat="${sub.id}">
        <span class="node-tab-chip-name privacy-target ${state.settings.privacy_blur ? 'privacy-blurred' : ''}">${sub.name}</span> (${subCount})
      </button>
    `;
  });

  if (customCount > 0) {
    tabsHtml += `
      <button class="node-tab-chip ${state.activeNodeCategory === 'custom' ? 'active' : ''}" data-cat="custom">
        ${t('tab_manual_nodes')} (${customCount})
      </button>
    `;
  }

  tabsBar.innerHTML = tabsHtml;

  // Add click listeners to category chips
  tabsBar.querySelectorAll('.node-tab-chip').forEach(chip => {
    chip.addEventListener('click', () => {
      state.activeNodeCategory = chip.dataset.cat;
      renderNodeCategoryTabs();
      renderNodesList();
    });
  });

  // Update Subscription Traffic Card in Nodes
  const subCard = document.getElementById('sub-traffic-card');
  let activeSub = subs.find(s => s.id === state.activeNodeCategory);
  if (!activeSub && state.activeNodeCategory === 'all') {
    const activeNode = nodes.find(n => n.id === state.activeNodeId);
    if (activeNode && (activeNode.subscription_id || activeNode.subscription_name)) {
      activeSub = subs.find(s => s.id === activeNode.subscription_id || s.name === activeNode.subscription_name);
    }
  }
  if (activeSub && subCard) {
    subCard.style.display = 'flex';
    const subCount = nodes.filter(n => n.subscription_id === activeSub.id || n.subscription_name === activeSub.name).length;
    
    // Name
    const nameEl = document.getElementById('sub-traffic-name');
    if (nameEl) {
      nameEl.innerText = activeSub.name;
      nameEl.classList.toggle('privacy-blurred', !!state.settings.privacy_blur);
    }

    // Status badge
    const badgeEl = document.getElementById('sub-status-badge');
    const badgeTextEl = document.getElementById('sub-status-text');
    let isExpired = false;
    if (activeSub.expire > 0 && activeSub.expire * 1000 < Date.now()) {
      isExpired = true;
    }
    if (badgeEl && badgeTextEl) {
      badgeEl.className = `sub-status-badge ${isExpired ? 'expired' : 'active'}`;
      badgeTextEl.innerText = isExpired ? 'Истекла' : 'Активна';
    }

    const quotaSection = document.getElementById('sub-traffic-quota-section');
    const fallbackSection = document.getElementById('sub-traffic-fallback-section');

    const hasUserInfo = !!(activeSub.has_userinfo || (activeSub.total > 0) || (activeSub.expire > 0) || ((activeSub.download || 0) > 0) || ((activeSub.upload || 0) > 0));
    if (hasUserInfo) {
      if (quotaSection) quotaSection.style.display = 'block';
      if (fallbackSection) fallbackSection.style.display = 'none';

      const remEl = document.getElementById('sub-traffic-rem');
      const totEl = document.getElementById('sub-traffic-tot');
      const pctEl = document.getElementById('sub-traffic-pct');
      const fillEl = document.getElementById('sub-traffic-fill');
      const expEl = document.getElementById('sub-meta-expire');
      const downEl = document.getElementById('sub-meta-down');
      const upEl = document.getElementById('sub-meta-up');

      const used = (activeSub.upload || 0) + (activeSub.download || 0);

      if (activeSub.total > 0) {
        const remaining = Math.max(0, activeSub.total - used);
        const pct = Math.min(100, Math.max(0, (used / activeSub.total) * 100));

        if (remEl) remEl.innerText = `Осталось ${formatBytes(remaining)}`;
        if (totEl) totEl.innerText = `из ${formatBytes(activeSub.total)}`;
        if (pctEl) pctEl.innerText = `${pct.toFixed(1)}% потрачено`;
        if (fillEl) {
          fillEl.style.width = `${pct.toFixed(1)}%`;
          fillEl.className = 'sub-traffic-fill';
          if (pct > 90) fillEl.classList.add('danger');
          else if (pct > 75) fillEl.classList.add('warning');
        }
      } else {
        if (remEl) remEl.innerText = `Трафик: Безлимитный`;
        if (totEl) totEl.innerText = ``;
        if (pctEl) pctEl.innerText = `Использовано: ${formatBytes(used)}`;
        if (fillEl) {
          fillEl.style.width = '100%';
          fillEl.className = 'sub-traffic-fill unlimited';
        }
      }

      if (expEl) {
        if (activeSub.expire > 0) {
          const expDate = new Date(activeSub.expire * 1000);
          const now = new Date();
          const diffDays = Math.ceil((expDate - now) / (1000 * 60 * 60 * 24));
          const dateFormatted = expDate.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' });
          expEl.innerHTML = `Срок: <b>до ${dateFormatted}</b> ${diffDays > 0 ? `(${diffDays} дн.)` : '(истекла)'}`;
        } else {
          expEl.innerHTML = `Срок: <b>Бессрочно</b>`;
        }
      }

      if (downEl) downEl.innerHTML = `Скачано: <b>${formatBytes(activeSub.download || 0)}</b>`;
      if (upEl) upEl.innerHTML = `Отдано: <b>${formatBytes(activeSub.upload || 0)}</b>`;

    } else {
      // Fallback: no quota headers returned by server
      if (quotaSection) quotaSection.style.display = 'none';
      if (fallbackSection) fallbackSection.style.display = 'block';

      const countEl = document.getElementById('sub-fallback-nodes-count');
      const updatedEl = document.getElementById('sub-fallback-updated');
      if (countEl) countEl.innerText = `Узлов: ${subCount}`;
      if (updatedEl) updatedEl.innerText = `Обновлено: ${activeSub.updated_at || '—'}`;
    }
  } else if (subCard) {
    subCard.style.display = 'none';
  }
}

function updateDashboardSubWidget() {
  const widget = document.getElementById('dash-sub-widget');
  if (!widget) return;

  const nodes = state.nodes || [];
  const subs = state.subscriptions || [];
  const activeNode = nodes.find(n => n.id === state.activeNodeId);

  if (!activeNode || (!activeNode.subscription_id && !activeNode.subscription_name)) {
    widget.style.display = 'none';
    return;
  }

  const activeSub = subs.find(s => s.id === activeNode.subscription_id || s.name === activeNode.subscription_name);
  if (!activeSub) {
    widget.style.display = 'none';
    return;
  }

  const hasUserInfo = !!(activeSub.has_userinfo || (activeSub.total > 0) || (activeSub.expire > 0) || ((activeSub.download || 0) > 0) || ((activeSub.upload || 0) > 0));
  if (!hasUserInfo) {
    widget.style.display = 'none';
    return;
  }

  widget.style.display = 'flex';

  // Subtitle: expiration / status (without leaking provider name)
  let expStr = 'Активна';
  if (activeSub.expire > 0) {
    const expDate = new Date(activeSub.expire * 1000);
    const now = new Date();
    const diffDays = Math.ceil((expDate - now) / (1000 * 60 * 60 * 24));
    const dateFormatted = expDate.toLocaleDateString(state.language === 'en' ? 'en-US' : 'ru-RU', { day: 'numeric', month: 'short', year: 'numeric' });
    if (diffDays > 0) {
      expStr = `Активна • до ${dateFormatted} (${diffDays} дн.)`;
    } else {
      expStr = `Истекла ${dateFormatted}`;
    }
  } else {
    expStr = 'Активна • Бессрочно';
  }
  const expEl = document.getElementById('dash-sub-exp');
  if (expEl) expEl.innerText = expStr;

  // Stat & progress
  const statEl = document.getElementById('dash-sub-stat');
  const fillEl = document.getElementById('dash-sub-mini-fill');
  const used = (activeSub.upload || 0) + (activeSub.download || 0);
  
  if (activeSub.total > 0) {
    const remaining = Math.max(0, activeSub.total - used);
    const pct = Math.min(100, Math.max(0, (used / activeSub.total) * 100));
    
    if (statEl) {
      statEl.innerHTML = `Осталось <b>${formatBytes(remaining)}</b> из ${formatBytes(activeSub.total)}`;
      statEl.classList.toggle('privacy-blurred', !!state.settings.privacy_blur);
    }
    if (fillEl) {
      fillEl.style.width = `${pct.toFixed(1)}%`;
      fillEl.className = 'dash-sub-mini-fill';
      if (pct > 90) fillEl.classList.add('danger');
      else if (pct > 75) fillEl.classList.add('warning');
    }
  } else {
    // Unlimited traffic
    if (statEl) {
      statEl.innerHTML = `Трафик: <b>Безлимитный</b> (исп. ${formatBytes(used)})`;
      statEl.classList.toggle('privacy-blurred', !!state.settings.privacy_blur);
    }
    if (fillEl) {
      fillEl.style.width = '100%';
      fillEl.className = 'dash-sub-mini-fill unlimited';
    }
  }
}

// Click on Dashboard sub widget navigates to nodes view
document.getElementById('dash-sub-widget')?.addEventListener('click', () => {
  const nodes = state.nodes || [];
  const subs = state.subscriptions || [];
  const activeNode = nodes.find(n => n.id === state.activeNodeId);
  if (activeNode) {
    const activeSub = subs.find(s => s.id === activeNode.subscription_id || s.name === activeNode.subscription_name);
    if (activeSub) {
      state.activeNodeCategory = activeSub.id;
    }
  }
  switchPage('nodes');
});

// Update Active Subscription Banner Button
document.getElementById('btn-update-active-sub')?.addEventListener('click', async () => {
  let activeSub = (state.subscriptions || []).find(s => s.id === state.activeNodeCategory);
  if (!activeSub && state.activeNodeCategory === 'all') {
    const activeNode = (state.nodes || []).find(n => n.id === state.activeNodeId);
    if (activeNode && (activeNode.subscription_id || activeNode.subscription_name)) {
      activeSub = (state.subscriptions || []).find(s => s.id === activeNode.subscription_id || s.name === activeNode.subscription_name);
    }
  }
  if (!activeSub) return;
  try {
    showToast(`Обновление «${activeSub.name}»...`);
    await RefreshSubscription(activeSub.id);
    showToast(`Подписка «${activeSub.name}» успешно обновлена`);
    await loadNodes();
  } catch (err) {
    showToast(`Ошибка обновления: ${err}`);
  }
});

function getShortProtocolBadge(node) {
  const raw = (node.protocol_label || node.protocol || '').toUpperCase();
  if (raw.includes('AWG') || raw.includes('AMNEZIA')) return 'AWG';
  if (raw.includes('MASQUE')) return 'MASQUE';
  if (raw.includes('WARP') || raw.includes('WIREGUARD')) return 'WARP';
  if (raw.includes('HYSTERIA') || raw.includes('HY2')) return 'HY2';
  if (raw.includes('SHADOWSOCKS') || raw === 'SS') return 'SS';
  if (raw.includes('TROJAN')) return 'TROJAN';
  if (raw.includes('TUIC')) return 'TUIC';
  if (raw.includes('VLESS')) return 'VLESS';
  if (raw.includes('VMESS')) return 'VMESS';
  return (node.protocol || 'VLESS').toUpperCase();
}

function renderNodesList() {
  const container = document.getElementById('nodes-list');
  if (!container) return;

  const nodes = state.nodes || [];

  // Filter nodes by active category
  let visibleNodes = nodes;
  if (state.activeNodeCategory === 'custom') {
    visibleNodes = nodes.filter(n => !n.subscription_id && !n.subscription_name);
  } else if (state.activeNodeCategory !== 'all') {
    visibleNodes = nodes.filter(n => n.subscription_id === state.activeNodeCategory || n.subscription_name === state.activeNodeCategory);
  }

  // Active selected node always pinned at the very top of the list
  if (state.activeNodeId) {
    visibleNodes = [...visibleNodes].sort((a, b) => {
      if (a.id === state.activeNodeId) return -1;
      if (b.id === state.activeNodeId) return 1;
      return 0;
    });
  }

  if (visibleNodes.length === 0) {
    container.innerHTML = `<div style="padding: 30px; text-align: center; color: var(--on-surface-muted);">${t('empty_nodes_msg')}</div>`;
    updatePrivacyBlurUI();
    return;
  }

  container.innerHTML = '';
  visibleNodes.forEach(node => {
    const isSelected = node.id === state.activeNodeId;
    const card = document.createElement('div');
    card.className = `node-item ${isSelected ? 'selected' : ''}`;

    let pingClass = 'fast';
    if (node.ping > 250) pingClass = 'slow';
    else if (node.ping > 120) pingClass = 'medium';

    const pingText = node.ping > 0 ? `${node.ping} ms` : '—';
    const protoLabel = node.protocol_label || (node.protocol ? node.protocol.toUpperCase() : 'VLESS');
    const isWarp = (node.protocol === 'wireguard' || node.protocol === 'warp' || protoLabel.includes('WARP'));
    const shortBadge = getShortProtocolBadge(node);

    const cc = detectNodeCountry(node);
    const flagSvg = renderFlagSVG(cc, 22, 14);
    const displayName = cleanNodeDisplayName(node.name);
    const isBlurred = !!state.settings.privacy_blur;

    card.innerHTML = `
      <div class="node-left">
        <div class="node-radio">
          <div class="node-radio-dot"></div>
        </div>
        <div class="node-flag-badge" title="${t('country_prefix', 'Страна')}: ${cc}">${flagSvg}</div>
        <div class="node-details">
          <div class="node-name privacy-target ${isBlurred ? 'privacy-blurred' : ''}">${displayName}</div>
          <div class="node-subtext">
            <span class="badge-protocol ${isWarp ? 'warp' : ''}" title="${protoLabel}">${shortBadge}</span>
            <span class="node-server-ip tabular privacy-target ${isBlurred ? 'privacy-blurred' : ''}">${node.server}:${node.port}</span>
            ${node.subscription_name ? `<span class="node-sub-name privacy-target ${isBlurred ? 'privacy-blurred' : ''}">• ${node.subscription_name}</span>` : ''}
          </div>
        </div>
      </div>
      <div class="node-right">
        <span class="ping-badge tabular ${pingClass}">${pingText}</span>
        <button class="btn-icon-only danger btn-del-node" title="${t('del_node_title', 'Удалить узел')}">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
        </button>
      </div>
    `;

    // Selecting node
    card.addEventListener('click', async (e) => {
      if (e.target.closest('.btn-del-node')) return;
      state.activeNodeId = node.id;
      renderNodesList();

      if (state.activeNodeCategory === 'all') {
        renderNodeCategoryTabs();
      }
      updateDashboardSubWidget();

      // Update Dashboard Hero
      document.getElementById('hero-node-name').innerText = displayName;
      document.getElementById('hero-protocol-badge').innerText = protoLabel;
      const heroFlagEl = document.getElementById('hero-country-flag');
      if (heroFlagEl) {
        heroFlagEl.innerHTML = renderFlagSVG(cc, 22, 14);
        heroFlagEl.title = `Страна: ${cc}`;
      }

      showToast(state.isConnected ? `Переключение узла: ${displayName}...` : `Выбран узел: ${displayName}`);

      try {
        await SetActiveNodeID(node.id);
        if (state.isConnected) {
          showToast(`Подключено: ${displayName}`);
        }
      } catch (err) {
        showToast(`Ошибка переключения: ${err}`);
      } finally {
        setTimeout(pollTelemetry, 500);
      }
    });

    const delBtn = card.querySelector('.btn-del-node');
    delBtn.addEventListener('click', async (e) => {
      e.stopPropagation();
      try {
        await DeleteNode(node.id);
        showToast('Узел удален');
        loadNodes();
      } catch (err) {
        showToast(`Ошибка удаления: ${err}`);
      }
    });

    container.appendChild(card);
  });

  updatePrivacyBlurUI();
}

// Ping all nodes
document.getElementById('btn-ping-all').addEventListener('click', async () => {
  const btn = document.getElementById('btn-ping-all');
  btn.style.opacity = '0.5';
  btn.style.pointerEvents = 'none';
  showToast('Проверка пинга всех серверов...');

  try {
    const pingMap = await PingAllNodes();
    state.nodes.forEach(n => {
      if (pingMap[n.id] !== undefined) {
        n.ping = pingMap[n.id];
      }
    });
    renderNodesList();
    showToast('Пинг обновлен');
  } catch (err) {
    showToast(`Ошибка пинга: ${err}`);
  } finally {
    btn.style.opacity = '1';
    btn.style.pointerEvents = 'auto';
  }
});

// Select fastest node from nodes page
document.getElementById('btn-select-fastest-nodes')?.addEventListener('click', handleSelectFastestNode);

// WARP Modal
const modalWarp = document.getElementById('modal-warp');
const btnOpenWarp = document.getElementById('btn-open-warp');
const btnCloseWarp = document.getElementById('btn-close-warp');
const btnCancelWarp = document.getElementById('btn-cancel-warp');
const optWarpAwg = document.getElementById('opt-warp-awg');
const optWarpMasque = document.getElementById('opt-warp-masque');

// WARP Subtabs & Import
const tabWarpGen = document.getElementById('tab-warp-gen');
const tabWarpImport = document.getElementById('tab-warp-import');
const paneWarpGen = document.getElementById('pane-warp-gen');
const paneWarpImport = document.getElementById('pane-warp-import');
const btnWarpPickFile = document.getElementById('btn-warp-pick-file');
const inputWarpConfText = document.getElementById('input-warp-conf-text');
const inputWarpConfName = document.getElementById('input-warp-conf-name');
const btnWarpImportSubmit = document.getElementById('btn-warp-import-submit');

tabWarpGen?.addEventListener('click', () => {
  tabWarpGen.classList.add('active');
  tabWarpImport?.classList.remove('active');
  if (paneWarpGen) paneWarpGen.style.display = 'block';
  if (paneWarpImport) paneWarpImport.style.display = 'none';
});

tabWarpImport?.addEventListener('click', () => {
  tabWarpImport?.classList.add('active');
  tabWarpGen?.classList.remove('active');
  if (paneWarpImport) paneWarpImport.style.display = 'block';
  if (paneWarpGen) paneWarpGen.style.display = 'none';
});

btnWarpPickFile?.addEventListener('click', async () => {
  try {
    const res = await SelectWarpConfigFile();
    if (res && res.content) {
      if (inputWarpConfText) inputWarpConfText.value = res.content;
      if (inputWarpConfName && (!inputWarpConfName.value || inputWarpConfName.value.trim() === '')) {
        let n = res.file_name || '';
        n = n.replace(/\.(conf|wireguard|txt)$/i, '');
        inputWarpConfName.value = n || 'WARP Custom';
      }
      showToast(`Файл ${res.file_name || ''} загружен`);
    }
  } catch (err) {
    showToast(`Ошибка выбора файла: ${err}`);
  }
});

btnWarpImportSubmit?.addEventListener('click', async () => {
  const text = inputWarpConfText?.value?.trim();
  if (!text) {
    showToast('Вставьте текст конфигурации или выберите файл .conf');
    return;
  }
  const customName = inputWarpConfName?.value?.trim() || '';

  btnWarpImportSubmit.style.opacity = '0.5';
  btnWarpImportSubmit.style.pointerEvents = 'none';

  try {
    const node = await ImportWarpConfig(text, customName);
    closeWarpModal();
    showToast(t('warp_import_success', `Узел ${node.name} успешно импортирован!`));
    if (inputWarpConfText) inputWarpConfText.value = '';
    if (inputWarpConfName) inputWarpConfName.value = '';
    await loadNodes();
  } catch (err) {
    showToast(`${t('warp_import_error', 'Ошибка импорта: ')}${err}`);
  } finally {
    btnWarpImportSubmit.style.opacity = '1';
    btnWarpImportSubmit.style.pointerEvents = 'auto';
  }
});

btnOpenWarp.addEventListener('click', () => {
  modalWarp.classList.add('open');
});

function closeWarpModal() {
  modalWarp.classList.remove('open');
}

btnCloseWarp.addEventListener('click', closeWarpModal);
btnCancelWarp.addEventListener('click', closeWarpModal);

async function handleGenerateWarp(mode) {
  closeWarpModal();
  showToast(mode === 'masque' ? 'Генерация WARP (MASQUE)...' : 'Генерация WARP (AmneziaWG)...');
  btnOpenWarp.style.opacity = '0.5';
  btnOpenWarp.style.pointerEvents = 'none';

  try {
    const node = await GenerateWarp(mode);
    showToast(`Узел ${node.name} успешно создан!`);
    await loadNodes();
  } catch (err) {
    showToast(`Ошибка генерации: ${err}`);
  } finally {
    btnOpenWarp.style.opacity = '1';
    btnOpenWarp.style.pointerEvents = 'auto';
  }
}

optWarpAwg.addEventListener('click', () => handleGenerateWarp('awg'));
optWarpMasque.addEventListener('click', () => handleGenerateWarp('masque'));

// ==========================================
// CONNECTION LOGS VIEWER MODAL
// ==========================================
const modalLogs = document.getElementById('modal-logs');
const btnHeroLogs = document.getElementById('btn-hero-logs');
const btnCloseLogs = document.getElementById('btn-close-logs');
const logsTerminal = document.getElementById('logs-terminal');
const btnCopyLogs = document.getElementById('btn-copy-logs');
const btnClearLogs = document.getElementById('btn-clear-logs');
const chkLogsAutoscroll = document.getElementById('chk-logs-autoscroll');
const filterBtns = document.querySelectorAll('.logs-filter-pill');

let currentLogFilter = 'all';
let rawLogData = '';
let logPollTimer = null;

function escapeHtml(str) {
  if (!str) return '';
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

function stripAnsi(str) {
  if (!str) return '';
  return str
    .replace(/\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])/g, '')
    .replace(/[\u001b\x1b]0?\[[0-9;]*[a-zA-Z]?/g, '')
    .replace(/[\u001b\x1b]/g, '');
}

function isNoiseLogLine(line) {
  if (!line) return true;
  // Local Windows DNS probe & malformed packet noise
  if (line.includes('router: process DNS packet: unpack request') ||
      line.includes('dns: bad question') ||
      line.includes('dns: overflow unpacking') ||
      line.includes('dns: bad rdata')) {
    return true;
  }
  // Elevated run notice from AppData
  if (line.includes('lxd: running elevated from') ||
      line.includes('must not be writable by a non-administrative principal') ||
      line.includes('starting anyway; the service would refuse this binary') ||
      line.includes('sing-box lxd --service=install')) {
    return true;
  }
  // High-frequency per-packet routing traces (games, background Windows telemetry, torrents)
  if (line.includes('router: found process path:') ||
      line.includes('inbound packet connection') ||
      line.includes('inbound DNS packet') ||
      line.includes('outbound packet connection') ||
      line.includes('inbound/tun[tun-in]: inbound connection') ||
      line.includes('outbound/direct[direct]: outbound connection')) {
    return true;
  }
  return false;
}

function parseSingboxLogLine(raw) {
  let clean = stripAnsi(raw).trim();
  if (!clean || isNoiseLogLine(clean)) return null;

  let level = 'INFO';
  const upper = clean.toUpperCase();
  if (upper.includes('ERROR') || upper.includes('FATAL') || upper.includes('PANIC')) {
    level = 'ERROR';
  } else if (upper.includes('WARN')) {
    level = 'WARN';
  }

  // 1. Extract timestamp (HH:mm:ss) from start if present
  let timeStr = '';
  const timeMatch = clean.match(/^(?:[+-]\d{4}\s+)?(?:\d{4}-\d{2}-\d{2}\s+)?(\d{2}:\d{2}:\d{2})/);
  if (timeMatch) {
    timeStr = timeMatch[1];
    clean = clean.substring(timeMatch[0].length).trim();
  }

  // 2. Remove level word if at start
  clean = clean.replace(/^(?:ERROR|WARN|INFO|DEBUG|FATAL)\s*:?\s*/i, '').trim();

  // 3. Strip internal connection tracker IDs like [1092837411 14.99s] or [1092837411]
  clean = clean.replace(/^\[\d+(\s+[\d\.]+s)?\]\s*/i, '');
  clean = clean.replace(/\s*\[\d+(\s+[\d\.]+s)?\]\s*/g, ' ');

  // 4. Clean up component names: "[tun-in] " -> "tun-in: "
  clean = clean.replace(/^\[([a-zA-Z0-9_-]+)\]\s*/, '$1: ');

  // 5. Clean up "using outbound/hysteria2[proxy-out]" -> "(hysteria2)"
  clean = clean.replace(/using outbound\/([a-zA-Z0-9_-]+)(?:\[[^\]]*\])?/g, '($1)');
  clean = clean.replace(/outbound\/([a-zA-Z0-9_-]+)\[proxy-out\]/g, '$1');

  clean = clean.trim();

  return {
    level,
    timeStr: timeStr || '--:--:--',
    msg: clean
  };
}

function openLogsModal() {
  btnHeroLogs?.classList.remove('has-error');
  if (modalLogs) {
    modalLogs.classList.add('open');
  }
  fetchAndRenderLogs();
  if (logPollTimer) clearInterval(logPollTimer);
  logPollTimer = setInterval(fetchAndRenderLogs, 1500);
}

function closeLogsModal() {
  if (modalLogs) {
    modalLogs.classList.remove('open');
  }
  if (logPollTimer) {
    clearInterval(logPollTimer);
    logPollTimer = null;
  }
}

btnHeroLogs?.addEventListener('click', openLogsModal);
btnCloseLogs?.addEventListener('click', closeLogsModal);

filterBtns.forEach(btn => {
  btn.addEventListener('click', () => {
    filterBtns.forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    currentLogFilter = btn.dataset.filter || 'all';
    renderFormattedLogs(rawLogData);
  });
});

async function fetchAndRenderLogs() {
  try {
    const logs = await GetCoreLogs(300);
    rawLogData = logs || '';
    renderFormattedLogs(rawLogData);
  } catch (err) {
    console.error('Failed to get core logs:', err);
  }
}

function renderFormattedLogs(text) {
  if (!logsTerminal) return;
  const lines = text ? text.split('\n') : [];

  let errorCount = 0;
  let warnCount = 0;
  const parsedLines = [];

  for (const rawLine of lines) {
    const parsed = parseSingboxLogLine(rawLine);
    if (!parsed) continue;

    if (parsed.level === 'ERROR') {
      errorCount++;
    } else if (parsed.level === 'WARN') {
      warnCount++;
    }

    if (currentLogFilter === 'error' && parsed.level !== 'ERROR') continue;
    if (currentLogFilter === 'warn' && parsed.level !== 'WARN') continue;

    parsedLines.push(parsed);
  }

  const badgeErr = document.getElementById('logs-count-error');
  if (badgeErr) badgeErr.innerText = String(errorCount);
  const badgeWarn = document.getElementById('logs-count-warn');
  if (badgeWarn) badgeWarn.innerText = String(warnCount);

  if (parsedLines.length === 0) {
    logsTerminal.innerHTML = `<div class="log-empty-msg">${t('logs_empty', 'Журнал пуст. Запустите подключение для записи событий.')}</div>`;
    return;
  }

  const html = parsedLines.map(item => {
    const timeHtml = `<span class="log-time">${escapeHtml(item.timeStr || '--:--:--')}</span>`;
    const levelHtml = `<span class="log-level">${escapeHtml(item.level)}</span>`;
    const msgHtml = `<span class="log-msg">${escapeHtml(item.msg)}</span>`;
    const lvlClass = item.level.toLowerCase();

    return `<div class="log-row log-${lvlClass}">${timeHtml}${levelHtml}${msgHtml}</div>`;
  }).join('');

  const shouldScroll = chkLogsAutoscroll?.checked;
  logsTerminal.innerHTML = html;
  if (shouldScroll) {
    logsTerminal.scrollTop = logsTerminal.scrollHeight;
  }
}

btnCopyLogs?.addEventListener('click', async () => {
  if (!rawLogData) {
    showToast(t('logs_empty', 'Логи пусты'));
    return;
  }
  try {
    await navigator.clipboard.writeText(rawLogData);
    showToast(t('logs_copied', 'Логи скопированы в буфер обмена'));
  } catch (e) {
    showToast('Не удалось скопировать логи');
  }
});

btnClearLogs?.addEventListener('click', async () => {
  try {
    await ClearCoreLogs();
    rawLogData = '';
    renderFormattedLogs('');
    showToast(t('logs_cleared', 'Журнал логов очищен'));
  } catch (err) {
    showToast(`Ошибка очистки: ${err}`);
  }
});

// Add Node Modal
const modalAddNode = document.getElementById('modal-add-node');
document.getElementById('btn-open-add-node').addEventListener('click', () => {
  modalAddNode.classList.add('open');
  document.getElementById('input-node-link').value = '';
  document.getElementById('input-node-link').focus();
});

document.getElementById('btn-close-add-node').addEventListener('click', () => {
  modalAddNode.classList.remove('open');
});

document.getElementById('btn-cancel-add-node').addEventListener('click', () => {
  modalAddNode.classList.remove('open');
});

document.getElementById('btn-confirm-add-node').addEventListener('click', async () => {
  const link = document.getElementById('input-node-link').value.trim();
  if (!link) return;
  try {
    await AddManualNode(link);
    modalAddNode.classList.remove('open');
    showToast('Узел успешно добавлен');
    loadNodes();
  } catch (err) {
    showToast(`Ошибка добавления: ${err}`);
  }
});

// Subscriptions Modal
const modalSubs = document.getElementById('modal-subs');
document.getElementById('btn-open-subs').addEventListener('click', async () => {
  modalSubs.classList.add('open');
  loadSubsList();
});

document.getElementById('btn-close-subs-x').addEventListener('click', () => {
  modalSubs.classList.remove('open');
});

document.getElementById('btn-close-subs').addEventListener('click', () => {
  modalSubs.classList.remove('open');
});

async function loadSubsList() {
  const container = document.getElementById('subs-list');
  const countBadge = document.getElementById('subs-count-badge');
  try {
    const subs = await GetSubscriptions();
    state.subscriptions = subs;
    if (countBadge) {
      countBadge.textContent = subs.length;
    }
    if (subs.length === 0) {
      container.innerHTML = `<div style="padding: 18px 12px; text-align: center; color: #7d778a; font-size: 12px;">Нет добавленных подписок</div>`;
      return;
    }
    container.innerHTML = '';

    const getNodesWord = (n) => {
      const abs = Math.abs(n) % 100;
      const last = abs % 10;
      if (abs > 10 && abs < 20) return 'узлов';
      if (last > 1 && last < 5) return 'узла';
      if (last === 1) return 'узел';
      return 'узлов';
    };

    subs.forEach(s => {
      const row = document.createElement('div');
      row.className = 'sub-card';
      row.innerHTML = `
        <div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 3px;">
          <div style="display: flex; align-items: center; gap: 8px;">
            <span style="font-weight: 600; color: #f0ecf6; font-size: 13px;">${s.name}</span>
            <span class="badge-nodes">${s.count} ${getNodesWord(s.count)}</span>
          </div>
          <div style="font-size: 11px; color: #8e8a99; font-family: monospace; text-overflow: ellipsis; overflow: hidden; white-space: nowrap;">${s.url}</div>
        </div>
        <div style="display: flex; gap: 6px; align-items: center; margin-left: 10px;">
          <button class="sub-action-btn btn-refresh-sub" title="Обновить подписку">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor"><path d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/></svg>
          </button>
          <button class="sub-action-btn danger btn-del-sub" title="Удалить подписку">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
          </button>
        </div>
      `;

      row.querySelector('.btn-refresh-sub').addEventListener('click', async () => {
        try {
          showToast('Обновление подписки...');
          await RefreshSubscription(s.id);
          showToast('Подписка обновлена');
          loadSubsList();
          loadNodes();
        } catch (err) {
          showToast(`Ошибка обновления: ${err}`);
        }
      });

      row.querySelector('.btn-del-sub').addEventListener('click', async () => {
        try {
          await DeleteSubscription(s.id);
          showToast('Подписка удалена');
          loadSubsList();
          loadNodes();
        } catch (err) {
          showToast(`Ошибка удаления: ${err}`);
        }
      });

      container.appendChild(row);
    });
  } catch (err) {
    showToast(`Ошибка: ${err}`);
  }
}

document.getElementById('btn-add-sub').addEventListener('click', async () => {
  const name = document.getElementById('input-sub-name').value.trim();
  const url = document.getElementById('input-sub-url').value.trim();
  if (!name || !url) {
    showToast('Введите название и ссылку');
    return;
  }
  try {
    showToast('Загрузка подписки...');
    await AddSubscription(name, url);
    document.getElementById('input-sub-name').value = '';
    document.getElementById('input-sub-url').value = '';
    showToast('Подписка успешно добавлена');
    loadSubsList();
    loadNodes();
  } catch (err) {
    showToast(`Ошибка подписки: ${err}`);
  }
});

// ==================== TRAFFIC PAGE ====================
let isUpdatingConnections = false;
async function loadConnections() {
  if (isUpdatingConnections) return;
  isUpdatingConnections = true;
  const container = document.getElementById('connections-tree');
  try {
    const groups = await GetConnections();
    if (!groups || groups.length === 0) {
      container.innerHTML = `<div style="padding: 40px; text-align: center; color: var(--on-surface-muted);">${t('no_active_conn')}</div>`;
      return;
    }

    // Remember currently open processes so dynamic update doesn't collapse them
    const openProcs = new Set(Array.from(container.querySelectorAll('.conn-group.open .conn-proc-title')).map(el => el.innerText.trim()));

    container.innerHTML = '';
    groups.forEach(g => {
      const groupEl = document.createElement('div');
      const isOpen = openProcs.has(g.process.trim());
      groupEl.className = `conn-group ${isOpen ? 'open' : ''}`;

      groupEl.innerHTML = `
        <div class="conn-group-header">
          <div class="conn-group-left">
            <span class="conn-proc-title">${g.process}</span>
            <span class="conn-count-badge tabular">${g.count}</span>
          </div>
          <div class="conn-group-right tabular">
            <span style="color: var(--success); margin-right: 12px;">↓ ${g.down_speed || '0 B'}</span>
            <span style="color: var(--accent-purple); margin-right: 8px;">↑ ${g.up_speed || '0 B'}</span>
            <svg class="conn-arrow" width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6-6-6 1.41-1.41z"/></svg>
          </div>
        </div>
        <div class="conn-children">
          ${g.children.map(c => `
            <div class="conn-row">
              <span class="conn-dest" title="${c.destination}">${c.destination}</span>
              <span class="conn-rule">${c.rule || 'proxy'}</span>
              <div class="tabular" style="font-size: 11px;">
                <span style="color: var(--success); margin-right: 8px;">↓ ${c.down_speed}</span>
                <span style="color: var(--accent-purple);">↑ ${c.up_speed}</span>
              </div>
            </div>
          `).join('')}
        </div>
      `;

      groupEl.querySelector('.conn-group-header').addEventListener('click', () => {
        groupEl.classList.toggle('open');
      });

      container.appendChild(groupEl);
    });
  } catch (err) {
    container.innerHTML = `<div style="padding: 30px; text-align: center; color: var(--error);">Ошибка получения соединений: ${err}</div>`;
  } finally {
    isUpdatingConnections = false;
  }
}

document.getElementById('btn-refresh-conns').addEventListener('click', () => {
  loadConnections();
});

// ==================== SETTINGS PAGE ====================
function updateUAPresetActive(val) {
  document.querySelectorAll('.btn-preset-pill').forEach(btn => {
    btn.classList.toggle('active', btn.dataset.ua === val);
  });
}

document.querySelectorAll('.btn-preset-pill').forEach(btn => {
  btn.addEventListener('click', () => {
    const ua = btn.dataset.ua;
    const uaInput = document.getElementById('set-user-agent');
    if (uaInput && ua) {
      uaInput.value = ua;
      updateUAPresetActive(ua);
    }
  });
});

const uaInputEl = document.getElementById('set-user-agent');
if (uaInputEl) {
  uaInputEl.addEventListener('input', () => {
    updateUAPresetActive(uaInputEl.value);
  });
}

const btnCopyHwid = document.getElementById('btn-copy-hwid');
if (btnCopyHwid) {
  btnCopyHwid.addEventListener('click', async () => {
    const hwidInput = document.getElementById('set-hwid');
    const val = hwidInput ? hwidInput.value.trim() : '';
    if (!val) return;
    try {
      await navigator.clipboard.writeText(val);
      showToast(t('hwid_copied', 'HWID скопирован в буфер обмена'));
    } catch (_) {
      if (hwidInput) {
        hwidInput.select();
        document.execCommand('copy');
        showToast(t('hwid_copied', 'HWID скопирован в буфер обмена'));
      }
    }
  });
}

const btnGenHwid = document.getElementById('btn-generate-hwid');
if (btnGenHwid) {
  btnGenHwid.addEventListener('click', async () => {
    let newHwid = '';
    try {
      if (typeof GenerateRandomHWID === 'function') {
        newHwid = await GenerateRandomHWID();
      }
    } catch (_) {}
    if (!newHwid && window.crypto && window.crypto.randomUUID) {
      newHwid = window.crypto.randomUUID();
    }
    const hwidInput = document.getElementById('set-hwid');
    if (hwidInput && newHwid) {
      hwidInput.value = newHwid;
      showToast(t('hwid_generated', 'Сгенерирован новый случайный HWID'));
    }
  });
}

async function loadSettingsPage() {
  try {
    state.settings = await GetSettings();
    const setLanguageEl = document.getElementById('set-language');
    if (setLanguageEl) {
      setLanguageEl.value = state.settings.language || 'auto';
    }
    document.getElementById('set-bypass-lan').checked = !!state.settings.bypass_lan;
    document.getElementById('set-strict-route').checked = !!state.settings.strict_route;
    const ksEl = document.getElementById('set-killswitch');
    if (ksEl) ksEl.checked = !!state.settings.kill_switch;
    const wdEl = document.getElementById('set-warp-detour');
    if (wdEl) wdEl.checked = !!state.settings.warp_detour;
    const asEl = document.getElementById('set-auto-subs');
    if (asEl) asEl.checked = state.settings.auto_update_subs !== false;
    document.getElementById('set-remote-dns').value = state.settings.remote_dns || 'https://1.1.1.1/dns-query';
    document.getElementById('set-direct-dns').value = state.settings.direct_dns || 'local';
    document.getElementById('set-mtu').value = state.settings.mtu || 1400;
    document.getElementById('set-autostart').checked = !!state.settings.auto_start;
    document.getElementById('set-minimize-tray').checked = !!state.settings.minimize_to_tray;
    document.getElementById('set-privacy-blur').checked = !!state.settings.privacy_blur;

    const uaInput = document.getElementById('set-user-agent');
    if (uaInput) {
      uaInput.value = state.settings.user_agent || 'v2rayNG/1.8.5';
      updateUAPresetActive(uaInput.value);
    }
    const hwidInput = document.getElementById('set-hwid');
    if (hwidInput) {
      hwidInput.value = state.settings.hwid || '';
    }

    try {
      const ver = await GetAppVersion();
      const verEl = document.getElementById('settings-app-version');
      if (verEl && ver) verEl.innerText = ver;
    } catch (_) {}
  } catch (err) {
    showToast(`${t('settings_error_load')}: ${err}`);
  }
}

document.getElementById('btn-save-settings').addEventListener('click', async () => {
  try {
    const setLanguageEl = document.getElementById('set-language');
    if (setLanguageEl) {
      state.settings.language = setLanguageEl.value;
    }
    state.settings.bypass_lan = document.getElementById('set-bypass-lan').checked;
    state.settings.strict_route = document.getElementById('set-strict-route').checked;
    const ksEl = document.getElementById('set-killswitch');
    if (ksEl) state.settings.kill_switch = ksEl.checked;
    const wdEl = document.getElementById('set-warp-detour');
    if (wdEl) state.settings.warp_detour = wdEl.checked;
    const asEl = document.getElementById('set-auto-subs');
    if (asEl) state.settings.auto_update_subs = asEl.checked;
    state.settings.remote_dns = document.getElementById('set-remote-dns').value.trim();
    state.settings.direct_dns = document.getElementById('set-direct-dns').value.trim();
    state.settings.mtu = parseInt(document.getElementById('set-mtu').value, 10) || 1400;
    state.settings.auto_start = document.getElementById('set-autostart').checked;
    state.settings.minimize_to_tray = document.getElementById('set-minimize-tray').checked;
    state.settings.privacy_blur = document.getElementById('set-privacy-blur').checked;

    const uaInput = document.getElementById('set-user-agent');
    if (uaInput) {
      state.settings.user_agent = uaInput.value.trim() || 'v2rayNG/1.8.5';
    }
    const hwidInput = document.getElementById('set-hwid');
    if (hwidInput) {
      state.settings.hwid = hwidInput.value.trim();
    }

    await SaveSettings(state.settings);
    const eff = getEffectiveLanguage(state.settings.language, detectedSystemLanguage);
    applyLanguage(eff);
    updateStatusUI();
    updateHeroNodeDisplay();
    renderNodeCategoryTabs();
    renderNodesList();
    updatePrivacyBlurUI();
    updateWarpHeroBadge();
    showToast(t('toast_settings_saved'));
  } catch (err) {
    showToast(`${t('settings_error_save')}: ${err}`);
  }
});

function updateHeroNodeDisplay() {
  const nodes = state.nodes || [];
  const heroFlagEl = document.getElementById('hero-country-flag');
  const heroProto = document.getElementById('hero-protocol-badge');
  const heroName = document.getElementById('hero-node-name');

  if (nodes.length > 0) {
    let activeNode = nodes.find(n => n.id === state.activeNodeId) || nodes[0];
    state.activeNodeId = activeNode.id;

    const cc = detectNodeCountry(activeNode);
    if (heroFlagEl) {
      heroFlagEl.innerHTML = renderFlagSVG(cc, 22, 14);
      heroFlagEl.title = `${t('country_prefix', 'Страна')}: ${cc}`;
    }

    const protoLabel = activeNode.protocol_label || (activeNode.protocol ? activeNode.protocol.toUpperCase() : 'VLESS');
    if (heroProto) heroProto.innerText = protoLabel;

    if (heroName) {
      heroName.innerText = cleanNodeDisplayName(activeNode.name);
      heroName.classList.toggle('privacy-blurred', !!state.settings.privacy_blur);
    }
  } else {
    state.activeNodeId = '';
    if (heroFlagEl) heroFlagEl.innerHTML = '';
    if (heroProto) heroProto.innerText = '—';
    if (heroName) {
      heroName.innerText = t('no_nodes_add');
      heroName.classList.remove('privacy-blurred');
    }
  }
}

// ==================== QR Code Import Modal ====================
let qrCameraStream = null;
let qrCameraAnimId = null;
let lastDecodedQRText = null;

const modalQR = document.getElementById('modal-qr');
const btnOpenQR = document.getElementById('btn-open-qr');
const btnCloseQRX = document.getElementById('btn-close-qr-x');
const btnCloseQR = document.getElementById('btn-close-qr');
const btnQRScreen = document.getElementById('btn-qr-screen');
const btnQRPaste = document.getElementById('btn-qr-paste');
const btnQRFileTrigger = document.getElementById('btn-qr-file-trigger');
const inputQRFile = document.getElementById('input-qr-file');
const btnQRCameraToggle = document.getElementById('btn-qr-camera-toggle');
const qrCameraBox = document.getElementById('qr-camera-box');
const qrVideo = document.getElementById('qr-video');
const qrDropzone = document.getElementById('qr-dropzone');
const qrResultBox = document.getElementById('qr-result-box');
const qrResultType = document.getElementById('qr-result-type');
const qrResultText = document.getElementById('qr-result-text');
const btnConfirmImportQR = document.getElementById('btn-confirm-import-qr');

function openQRModal() {
  lastDecodedQRText = null;
  if (qrResultBox) qrResultBox.style.display = 'none';
  if (btnConfirmImportQR) btnConfirmImportQR.style.display = 'none';
  stopQRCamera();
  if (modalQR) modalQR.classList.add('open');
}

function closeQRModal() {
  stopQRCamera();
  if (modalQR) modalQR.classList.remove('open');
}

if (btnOpenQR) btnOpenQR.addEventListener('click', openQRModal);
if (btnCloseQRX) btnCloseQRX.addEventListener('click', closeQRModal);
if (btnCloseQR) btnCloseQR.addEventListener('click', closeQRModal);

function scanCanvasOrImage(source) {
  let canvas;
  const w = source.naturalWidth || source.videoWidth || source.width;
  const h = source.naturalHeight || source.videoHeight || source.height;
  if (!w || !h) return null;

  if (source instanceof HTMLCanvasElement) {
    canvas = source;
  } else {
    canvas = document.createElement('canvas');
    canvas.width = w;
    canvas.height = h;
    const ctx = canvas.getContext('2d');
    ctx.drawImage(source, 0, 0, w, h);
  }

  const ctx = canvas.getContext('2d');
  let imgData = ctx.getImageData(0, 0, canvas.width, canvas.height);
  let res = jsQR(imgData.data, canvas.width, canvas.height, {
    inversionAttempts: 'attemptBoth'
  });
  if (res && res.data) return res.data;

  // Scale down if large (e.g. 4K or 1080p display captures)
  if (w > 1200 || h > 1200) {
    const scale = Math.min(1200 / w, 1200 / h);
    const sw = Math.round(w * scale);
    const sh = Math.round(h * scale);
    const sCanvas = document.createElement('canvas');
    sCanvas.width = sw;
    sCanvas.height = sh;
    const sCtx = sCanvas.getContext('2d');
    sCtx.drawImage(canvas, 0, 0, sw, sh);
    const sData = sCtx.getImageData(0, 0, sw, sh);
    res = jsQR(sData.data, sw, sh, { inversionAttempts: 'attemptBoth' });
    if (res && res.data) return res.data;
  }

  return null;
}

function processDecodedQR(rawText) {
  if (!rawText || !rawText.trim()) {
    showToast(t('toast_qr_empty'));
    return;
  }
  const text = rawText.trim();
  lastDecodedQRText = text;

  let typeKey = 'qr_type_config';
  const lower = text.toLowerCase();
  if (lower.startsWith('http://') || lower.startsWith('https://')) {
    typeKey = 'qr_type_sub';
  } else if (lower.startsWith('vless://') || lower.startsWith('vmess://') || lower.startsWith('ss://') ||
             lower.startsWith('trojan://') || lower.startsWith('hy2://') || lower.startsWith('hysteria2://') ||
             lower.startsWith('tuic://') || lower.startsWith('warp://') || lower.startsWith('wireguard://')) {
    typeKey = 'qr_type_node';
  } else if (text.includes('\n') && (text.includes('://') || text.includes('='))) {
    typeKey = 'qr_type_list';
  }
  const typeLabel = t(typeKey);

  if (qrResultBox && qrResultText && qrResultType) {
    qrResultType.textContent = typeLabel;
    qrResultText.textContent = text;
    qrResultBox.style.display = 'flex';
  }
  if (btnConfirmImportQR) {
    btnConfirmImportQR.style.display = 'inline-flex';
    btnConfirmImportQR.textContent = `${t('btn_import')} (${typeLabel})`;
  }
  showToast(`${t('toast_qr_recognized')}: ${typeLabel}`);
}

async function executeQRImport(content) {
  if (!content) return;
  const text = content.trim();
  const lower = text.toLowerCase();

  try {
    if (lower.startsWith('http://') || lower.startsWith('https://')) {
      let name = 'QR Подписка';
      try {
        const u = new URL(text);
        name = u.hostname || 'Подписка';
      } catch (e) {}
      showToast('Добавление подписки...');
      await AddSubscription(name, text);
      showToast('Подписка успешно добавлена');
      closeQRModal();
      await loadSubsList();
      await loadNodes();
    } else if (text.includes('\n')) {
      const lines = text.split(/\r?\n/).map(l => l.trim()).filter(l => l.length > 0);
      let count = 0;
      for (const line of lines) {
        if (line.includes('://')) {
          try {
            await AddManualNode(line);
            count++;
          } catch (e) {
            console.warn('Failed line import:', line, e);
          }
        }
      }
      if (count > 0) {
        showToast(getCurrentLanguage() === 'en' ? `Imported nodes: ${count}` : `Импортировано узлов: ${count}`);
        closeQRModal();
        await loadNodes();
      } else {
        await AddManualNode(text);
        showToast(t('toast_node_added'));
        closeQRModal();
        await loadNodes();
      }
    } else {
      await AddManualNode(text);
      showToast(t('toast_node_added'));
      closeQRModal();
      await loadNodes();
    }
  } catch (err) {
    console.error('QR import error:', err);
    showToast(`Ошибка импорта: ${err}`);
  }
}

if (btnConfirmImportQR) {
  btnConfirmImportQR.addEventListener('click', () => {
    if (lastDecodedQRText) executeQRImport(lastDecodedQRText);
  });
}

// 1. File Upload & Drag-and-Drop
function handleQRFile(file) {
  if (!file || !file.type.startsWith('image/')) {
    showToast(t('toast_select_img'));
    return;
  }
  const reader = new FileReader();
  reader.onload = (e) => {
    const img = new Image();
    img.onload = () => {
      const decoded = scanCanvasOrImage(img);
      if (decoded) {
        processDecodedQR(decoded);
      } else {
        showToast(t('toast_qr_not_found'));
      }
    };
    img.src = e.target.result;
  };
  reader.readAsDataURL(file);
}

if (btnQRFileTrigger && inputQRFile) {
  btnQRFileTrigger.addEventListener('click', () => {
    inputQRFile.click();
  });
  inputQRFile.addEventListener('change', (e) => {
    if (e.target.files && e.target.files[0]) {
      handleQRFile(e.target.files[0]);
      e.target.value = '';
    }
  });
}

if (qrDropzone) {
  qrDropzone.addEventListener('click', () => {
    if (inputQRFile) inputQRFile.click();
  });
  qrDropzone.addEventListener('dragover', (e) => {
    e.preventDefault();
    qrDropzone.classList.add('drag-active');
  });
  qrDropzone.addEventListener('dragleave', () => {
    qrDropzone.classList.remove('drag-active');
  });
  qrDropzone.addEventListener('drop', (e) => {
    e.preventDefault();
    qrDropzone.classList.remove('drag-active');
    if (e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleQRFile(e.dataTransfer.files[0]);
    }
  });
}

// 2. Clipboard Paste (Ctrl+V and Button)
window.addEventListener('paste', (e) => {
  if (!modalQR || !modalQR.classList.contains('open')) return;

  const items = e.clipboardData && e.clipboardData.items;
  if (items) {
    for (let i = 0; i < items.length; i++) {
      if (items[i].type.indexOf('image') !== -1) {
        const file = items[i].getAsFile();
        if (file) {
          handleQRFile(file);
          return;
        }
      }
    }
  }

  const pastedText = e.clipboardData && e.clipboardData.getData('text');
  if (pastedText && pastedText.trim()) {
    processDecodedQR(pastedText.trim());
  }
});

if (btnQRPaste) {
  btnQRPaste.addEventListener('click', async () => {
    try {
      if (navigator.clipboard && navigator.clipboard.read) {
        const items = await navigator.clipboard.read();
        for (const item of items) {
          const imgType = item.types.find(t => t.startsWith('image/'));
          if (imgType) {
            const blob = await item.getType(imgType);
            handleQRFile(blob);
            return;
          }
        }
      }
    } catch (err) {
      console.log('Clipboard.read() error:', err);
    }

    try {
      if (navigator.clipboard && navigator.clipboard.readText) {
        const text = await navigator.clipboard.readText();
        if (text && text.trim()) {
          processDecodedQR(text.trim());
          return;
        }
      }
    } catch (err) {
      console.warn('Clipboard readText error:', err);
    }

    showToast(t('toast_clipboard_empty'));
  });
}

// 3. Screen Capture ("С экрана")
if (btnQRScreen) {
  btnQRScreen.addEventListener('click', async () => {
    try {
      if (!navigator.mediaDevices || !navigator.mediaDevices.getDisplayMedia) {
        showToast('Захват экрана не поддерживается в данной среде');
        return;
      }
      showToast(t('toast_screen_select'));
      const stream = await navigator.mediaDevices.getDisplayMedia({
        video: { cursor: 'never' },
        audio: false
      });

      const video = document.createElement('video');
      video.srcObject = stream;
      video.muted = true;
      video.playsInline = true;
      await video.play();

      // Wait a frame for video dimensions to populate
      await new Promise(r => setTimeout(r, 350));

      const canvas = document.createElement('canvas');
      canvas.width = video.videoWidth || 1920;
      canvas.height = video.videoHeight || 1080;
      const ctx = canvas.getContext('2d');
      ctx.drawImage(video, 0, 0, canvas.width, canvas.height);

      // Stop stream immediately
      stream.getTracks().forEach(t => t.stop());
      video.srcObject = null;

      const decoded = scanCanvasOrImage(canvas);
      if (decoded) {
        processDecodedQR(decoded);
      } else {
        showToast(t('toast_qr_screen_not_found'));
      }
    } catch (err) {
      console.warn('Screen capture error:', err);
      if (err.name !== 'NotAllowedError') {
        showToast('Не удалось захватить экран: ' + err.message);
      }
    }
  });
}

// 4. Webcam Live Scanner
async function startQRCamera() {
  try {
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      showToast(t('toast_camera_unsupported'));
      return;
    }
    qrCameraStream = await navigator.mediaDevices.getUserMedia({
      video: { facingMode: 'environment' }
    });
    if (qrVideo) {
      qrVideo.srcObject = qrCameraStream;
      qrVideo.setAttribute('playsinline', 'true');
      await qrVideo.play();
    }
    if (qrCameraBox) qrCameraBox.style.display = 'flex';
    if (btnQRCameraToggle) btnQRCameraToggle.classList.add('active');

    function tick() {
      if (qrVideo && qrVideo.readyState === qrVideo.HAVE_ENOUGH_DATA) {
        const decoded = scanCanvasOrImage(qrVideo);
        if (decoded) {
          stopQRCamera();
          processDecodedQR(decoded);
          return;
        }
      }
      qrCameraAnimId = requestAnimationFrame(tick);
    }
    qrCameraAnimId = requestAnimationFrame(tick);
  } catch (err) {
    console.warn('Camera error:', err);
    showToast('Не удалось получить доступ к камере: ' + err.message);
    stopQRCamera();
  }
}

function stopQRCamera() {
  if (qrCameraAnimId) {
    cancelAnimationFrame(qrCameraAnimId);
    qrCameraAnimId = null;
  }
  if (qrCameraStream) {
    qrCameraStream.getTracks().forEach(t => t.stop());
    qrCameraStream = null;
  }
  if (qrCameraBox) qrCameraBox.style.display = 'none';
  if (btnQRCameraToggle) btnQRCameraToggle.classList.remove('active');
}

if (btnQRCameraToggle) {
  btnQRCameraToggle.addEventListener('click', () => {
    if (qrCameraStream) {
      stopQRCamera();
    } else {
      startQRCamera();
    }
  });
}

// Close any open modal on overlay click or Escape
document.querySelectorAll('.modal-overlay').forEach(overlay => {
  overlay.addEventListener('click', (e) => {
    if (e.target === overlay) {
      overlay.classList.remove('open');
      if (overlay.id === 'modal-qr') stopQRCamera();
    }
  });
});

window.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') {
    document.querySelectorAll('.modal-overlay.open').forEach(m => {
      m.classList.remove('open');
      if (m.id === 'modal-qr') stopQRCamera();
    });
  }
});

// App Initialization
async function initApp() {
  try {
    state.settings = await GetSettings();
    updatePrivacyBlurUI();

    try {
      const rawVer = await GetAppVersion();
      const digitsOnly = (rawVer || '').replace(/^v/i, '').trim();
      const railVerEl = document.getElementById('nav-rail-app-version');
      if (railVerEl && digitsOnly) {
        railVerEl.innerText = digitsOnly;
      }
    } catch (_) {}

    // Detect system language and apply effective language
    try {
      detectedSystemLanguage = await GetSystemLanguage();
    } catch (_) {
      detectedSystemLanguage = detectSystemLanguage();
    }
    const effectiveLang = getEffectiveLanguage(state.settings.language, detectedSystemLanguage);
    applyLanguage(effectiveLang);

    // Bind language selector in settings
    const setLanguageEl = document.getElementById('set-language');
    if (setLanguageEl) {
      setLanguageEl.value = state.settings.language || 'auto';
      setLanguageEl.addEventListener('change', async () => {
        state.settings.language = setLanguageEl.value;
        const eff = getEffectiveLanguage(state.settings.language, detectedSystemLanguage);
        applyLanguage(eff);
        updateStatusUI();
        updateHeroNodeDisplay();
        renderNodeCategoryTabs();
        renderNodesList();
        try {
          await SaveSettings(state.settings);
          showToast(t('toast_lang_changed'));
        } catch (e) {
          console.warn('Auto-save language failed:', e);
        }
      });
    }

    const mode = state.settings.split_mode || 'whitelist';
    document.querySelectorAll('.mode-card').forEach(c => {
      c.classList.toggle('active', c.dataset.mode === mode);
    });
    updateQuickModeCardUI(mode);

    state.nodes = (await GetNodes()) || [];
    state.subscriptions = (await GetSubscriptions()) || [];
    try {
      state.activeNodeId = await GetActiveNodeID();
    } catch (_) {
      if (state.nodes.length > 0) state.activeNodeId = state.nodes[0].id;
      else state.activeNodeId = '';
    }

    updateStatusUI();
    updateHeroNodeDisplay();
    updateAdminStatus();
    checkAndDisplayConflicts();

    pollTelemetry();
    state.pollTimer = setInterval(pollTelemetry, 1000);

    // Background check for application updates on startup
    setTimeout(async () => {
      try {
        const info = await CheckUpdate();
        if (info) {
          pendingUpdateInfo = info;
          if (info.available) {
            showHeaderUpdateBadge(info);
          }
        }
      } catch (e) {
        console.debug('Background update check:', e);
      }
    }, 1500);
  } catch (err) {
    console.error('Init error:', err);
  }
}

// ==========================================
// APPLICATION AUTO-UPDATE LOGIC
// ==========================================
const headerUpdateBadge = document.getElementById('header-update-badge');
const headerUpdateText = document.getElementById('header-update-text');
const btnDismissHeaderUpdate = document.getElementById('btn-dismiss-header-update');
const modalAppUpdate = document.getElementById('modal-app-update');
const updateModalActions = document.getElementById('update-modal-actions');
const btnCheckUpdate = document.getElementById('btn-check-update');
const btnCloseAppUpdate = document.getElementById('btn-close-app-update');
const btnCancelUpdate = document.getElementById('btn-cancel-update');
const btnConfirmUpdate = document.getElementById('btn-confirm-update');
const updateModalBadge = document.getElementById('update-modal-badge');
const updateModalVerDiff = document.getElementById('update-modal-ver-diff');
const updateModalSize = document.getElementById('update-modal-size');
const updateModalNotes = document.getElementById('update-modal-notes');
const updateProgressSection = document.getElementById('update-progress-section');
const updateProgressLabel = document.getElementById('update-progress-label');
const updateProgressPercent = document.getElementById('update-progress-percent');
const updateProgressBar = document.getElementById('update-progress-bar');
const settingsUpdateDesc = document.getElementById('settings-update-desc');

let pendingUpdateInfo = null;
let isUpdating = false;

function showHeaderUpdateBadge(info) {
  if (!info || !info.available) return;
  const dismissed = sessionStorage.getItem(`dismissed_update_${info.latest_version}`);
  if (dismissed === 'true') return;

  if (headerUpdateText) {
    headerUpdateText.innerText = `${t('header_update_available')} ${info.latest_version}`;
  }
  if (headerUpdateBadge) {
    headerUpdateBadge.style.display = 'inline-flex';
  }
}

// Clicking the header update badge navigates to Settings and highlights the update card
headerUpdateBadge?.addEventListener('click', (e) => {
  if (e.target.closest('#btn-dismiss-header-update')) return;

  switchPage('settings');

  setTimeout(() => {
    const updateCard = document.getElementById('btn-check-update')?.closest('.card');
    if (updateCard) {
      updateCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
      updateCard.style.transition = 'box-shadow 0.3s ease, border-color 0.3s ease';
      updateCard.style.borderColor = 'rgba(167, 139, 250, 0.7)';
      updateCard.style.boxShadow = '0 0 20px rgba(167, 139, 250, 0.35)';
      setTimeout(() => {
        updateCard.style.borderColor = '';
        updateCard.style.boxShadow = '';
      }, 2000);
    }
  }, 100);

  if (pendingUpdateInfo) {
    openUpdateModalWithInfo(pendingUpdateInfo);
  }
});

btnDismissHeaderUpdate?.addEventListener('click', (e) => {
  e.stopPropagation();
  if (headerUpdateBadge) headerUpdateBadge.style.display = 'none';
  if (pendingUpdateInfo && pendingUpdateInfo.latest_version) {
    sessionStorage.setItem(`dismissed_update_${pendingUpdateInfo.latest_version}`, 'true');
  }
});

function formatReleaseNotes(raw, lang) {
  if (!raw) return '';
  let text = raw;

  // Split language if bilingual markers present
  const isEn = lang === 'en';
  if (text.includes('---')) {
    const parts = text.split('---');
    if (isEn && parts.length > 1) {
      text = parts[1];
    } else {
      text = parts[0];
    }
  }

  // Strip headers like "### 🇷🇺 Что нового:" or "### What's New:"
  text = text.replace(/^#+.*$/gm, '').trim();

  const lines = text.split('\n');
  const items = [];

  for (let line of lines) {
    line = line.trim();
    if (!line || line.startsWith('---')) continue;

    if (line.startsWith('- ') || line.startsWith('* ')) {
      const content = line.substring(2).trim();
      const boldMatch = content.match(/^\*\*(.*?)\*\*:?\s*(.*)$/);
      if (boldMatch) {
        const title = boldMatch[1].trim();
        let desc = boldMatch[2].trim();
        desc = desc.replace(/`([^`]+)`/g, '<code class="update-code">$1</code>');
        items.push(`
          <li class="update-note-item">
            <span class="update-note-bullet"></span>
            <div class="update-note-content">
              <strong class="update-note-strong">${title}</strong>
              <span class="update-note-text">: ${desc}</span>
            </div>
          </li>
        `);
      } else {
        items.push(`
          <li class="update-note-item">
            <span class="update-note-bullet"></span>
            <div class="update-note-content">
              <span class="update-note-text">${content}</span>
            </div>
          </li>
        `);
      }
    }
  }

  if (items.length > 0) {
    return `<ul class="update-notes-list">${items.join('')}</ul>`;
  }

  return `<div style="font-size: 13px; color: #cbd5e1; line-height: 1.5;">${escapeHtml(text).replace(/\n/g, '<br/>')}</div>`;
}

function openUpdateModalWithInfo(info) {
  if (!info) return;
  pendingUpdateInfo = info;
  if (updateModalBadge) updateModalBadge.innerText = info.latest_version;
  if (updateModalVerDiff) updateModalVerDiff.innerText = `${info.current_version} → ${info.latest_version}`;
  
  const sizeMb = info.asset_size > 0 ? `${(info.asset_size / (1024 * 1024)).toFixed(1)} МБ` : '';
  if (updateModalSize) updateModalSize.innerText = sizeMb;

  if (updateModalNotes) {
    const currentLang = typeof getCurrentLanguage === 'function' ? getCurrentLanguage() : 'ru';
    updateModalNotes.innerHTML = formatReleaseNotes(info.release_notes || t('modal_update_subtitle'), currentLang);
  }

  if (updateProgressSection) updateProgressSection.style.display = 'none';
  if (updateProgressBar) updateProgressBar.style.width = '0%';
  if (updateModalActions) updateModalActions.style.display = 'flex';
  if (btnConfirmUpdate) {
    btnConfirmUpdate.disabled = false;
  }
  if (btnCancelUpdate) btnCancelUpdate.style.display = 'inline-flex';
  if (btnCloseAppUpdate) btnCloseAppUpdate.style.display = 'inline-flex';

  modalAppUpdate?.classList.add('open');
}

btnCheckUpdate?.addEventListener('click', async () => {
  if (isUpdating) return;
  const originalHtml = btnCheckUpdate.innerHTML;
  btnCheckUpdate.disabled = true;
  btnCheckUpdate.style.opacity = '0.6';
  btnCheckUpdate.innerHTML = `
    <svg class="btn-icon spinning" viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 4V1L8 5l4 4V6c3.31 0 6 2.69 6 6 0 1.01-.25 1.97-.7 2.8l1.46 1.46C19.54 15.03 20 13.57 20 12c0-4.42-3.58-8-8-8zm0 14c-3.31 0-6-2.69-6-6 0-1.01.25-1.97.7-2.8L5.24 7.74C4.46 8.97 4 10.43 4 12c0 4.42 3.58 8 8 8v3l4-4-4-4v3z"/>
    </svg>
    <span>${t('btn_checking_update')}</span>
  `;

  try {
    const info = await CheckUpdate();
    pendingUpdateInfo = info;

    if (info && info.available) {
      showHeaderUpdateBadge(info);
      openUpdateModalWithInfo(info);
      if (settingsUpdateDesc) {
        settingsUpdateDesc.innerHTML = `<span style="color: #a78bfa; font-weight: 600;">${t('update_new_available')} ${info.latest_version}</span>`;
      }
    } else {
      showToast(`${t('update_latest_installed')} (${info ? info.current_version : ''}) ✅`);
      if (settingsUpdateDesc) {
        settingsUpdateDesc.innerHTML = `
          <span style="color: #4ade80;">✓ ${t('update_latest_installed')} (${info ? info.current_version : ''})</span>
          <button id="btn-force-open-update" type="button" style="background: none; border: none; color: #a78bfa; text-decoration: underline; cursor: pointer; font-size: 11px; margin-left: 8px;">
            ${t('update_force_reinstall')}
          </button>
        `;
        document.getElementById('btn-force-open-update')?.addEventListener('click', () => {
          openUpdateModalWithInfo(info);
        });
      }
    }
  } catch (err) {
    showToast(`${t('update_error')}: ${err}`);
  } finally {
    btnCheckUpdate.disabled = false;
    btnCheckUpdate.style.opacity = '1';
    btnCheckUpdate.innerHTML = originalHtml;
  }
});

function closeUpdateModal() {
  if (isUpdating) return; // Don't close while downloading
  modalAppUpdate?.classList.remove('open');
}

btnCloseAppUpdate?.addEventListener('click', closeUpdateModal);
btnCancelUpdate?.addEventListener('click', closeUpdateModal);

btnConfirmUpdate?.addEventListener('click', async () => {
  if (!pendingUpdateInfo || !pendingUpdateInfo.download_url) {
    showToast('Ссылка на обновление не найдена');
    return;
  }

  isUpdating = true;
  if (updateModalActions) updateModalActions.style.display = 'none';
  if (btnCloseAppUpdate) btnCloseAppUpdate.style.display = 'none';
  if (updateProgressSection) updateProgressSection.style.display = 'block';
  if (updateProgressLabel) updateProgressLabel.innerText = t('update_downloading');
  if (updateProgressPercent) updateProgressPercent.innerText = '0%';
  if (updateProgressBar) updateProgressBar.style.width = '0%';

  try {
    await DownloadAndApplyUpdate(pendingUpdateInfo.download_url);
    if (updateProgressLabel) updateProgressLabel.innerText = t('update_restarting');
  } catch (err) {
    isUpdating = false;
    if (updateModalActions) updateModalActions.style.display = 'flex';
    if (btnConfirmUpdate) btnConfirmUpdate.disabled = false;
    if (btnCancelUpdate) btnCancelUpdate.style.display = 'inline-flex';
    if (btnCloseAppUpdate) btnCloseAppUpdate.style.display = 'inline-flex';
    if (updateProgressSection) updateProgressSection.style.display = 'none';
    showToast(`Ошибка установки: ${err}`);
  }
});

window.addEventListener('DOMContentLoaded', initApp);
