/* =============================================
   GoNetworkMonitor — Dashboard Application JS
   ============================================= */

'use strict';

// ---- State ----
const state = {
  ws: null,
  capturing: false,
  charts: {},
  trafficHistory: { labels: [], pps: [], mbps: [] },
  maxHistoryPoints: 60,
  alertCount: 0,
};

// ---- DOM Helpers ----
const $ = (id) => document.getElementById(id);
const fmt = (n, d = 0) => Number(n).toLocaleString('en-US', { maximumFractionDigits: d });
const fmtBytes = (b) => {
  if (b >= 1e9) return (b / 1e9).toFixed(2) + ' GB';
  if (b >= 1e6) return (b / 1e6).toFixed(2) + ' MB';
  if (b >= 1e3) return (b / 1e3).toFixed(2) + ' KB';
  return b + ' B';
};

// ---- Toast Notifications ----
function showToast(msg, type = 'info') {
  const container = $('toastContainer');
  const toast = document.createElement('div');
  toast.className = `toast toast--${type}`;
  toast.textContent = msg;
  container.appendChild(toast);
  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transform = 'translateX(100%)';
    toast.style.transition = 'all 0.3s ease';
    setTimeout(() => toast.remove(), 300);
  }, 3500);
}

// ---- Navigation ----
function navigateTo(page) {
  document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
  const pageEl = $(`page-${page}`);
  if (pageEl) pageEl.classList.add('active');
  const navEl = document.querySelector(`[data-page="${page}"]`);
  if (navEl) navEl.classList.add('active');
  $('pageTitle').textContent = page.charAt(0).toUpperCase() + page.slice(1);

  if (page === 'interfaces') loadInterfaces();
  if (page === 'connections') loadConnections();
  if (page === 'alerts') loadAlerts();
  if (page === 'settings') loadConfig();
}

document.querySelectorAll('.nav-item').forEach(item => {
  item.addEventListener('click', (e) => {
    e.preventDefault();
    navigateTo(item.dataset.page);
  });
});

// ---- Mobile sidebar ----
$('menuToggle').addEventListener('click', () => {
  document.getElementById('sidebar').classList.toggle('open');
});

// ---- Charts Initialization ----
function initCharts() {
  // Traffic Line Chart
  const trafficCtx = $('trafficChart').getContext('2d');
  state.charts.traffic = new Chart(trafficCtx, {
    type: 'line',
    data: {
      labels: [],
      datasets: [
        {
          label: 'Packets/s',
          data: [],
          borderColor: '#6366f1',
          backgroundColor: 'rgba(99,102,241,0.1)',
          borderWidth: 2,
          fill: true,
          tension: 0.4,
          pointRadius: 0,
          yAxisID: 'y1',
        },
        {
          label: 'Mbps',
          data: [],
          borderColor: '#10b981',
          backgroundColor: 'rgba(16,185,129,0.1)',
          borderWidth: 2,
          fill: true,
          tension: 0.4,
          pointRadius: 0,
          yAxisID: 'y2',
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: { duration: 0 },
      interaction: { mode: 'index', intersect: false },
      plugins: { legend: { display: false } },
      scales: {
        x: {
          type: 'time',
          time: { unit: 'second', displayFormats: { second: 'HH:mm:ss' } },
          ticks: { color: '#64748b', maxTicksLimit: 8 },
          grid: { color: 'rgba(45,55,72,0.5)' },
        },
        y1: {
          position: 'left',
          ticks: { color: '#6366f1' },
          grid: { color: 'rgba(45,55,72,0.5)' },
          title: { display: true, text: 'Packets/s', color: '#6366f1' },
        },
        y2: {
          position: 'right',
          ticks: { color: '#10b981' },
          grid: { display: false },
          title: { display: true, text: 'Mbps', color: '#10b981' },
        },
      },
    },
  });

  // Protocol Donut Chart
  const protoCtx = $('protocolChart').getContext('2d');
  state.charts.protocol = new Chart(protoCtx, {
    type: 'doughnut',
    data: {
      labels: ['No Data'],
      datasets: [{
        data: [1],
        backgroundColor: ['#2d3748'],
        borderWidth: 0,
        hoverOffset: 4,
      }],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      cutout: '65%',
      plugins: {
        legend: {
          position: 'bottom',
          labels: { color: '#94a3b8', padding: 12, font: { size: 12 } },
        },
      },
    },
  });
}

// ---- Update Traffic Chart ----
function updateTrafficChart(snap) {
  const chart = state.charts.traffic;
  const now = new Date();

  chart.data.labels.push(now);
  chart.data.datasets[0].data.push(snap.packets_per_sec || 0);
  chart.data.datasets[1].data.push((snap.bytes_per_sec || 0) * 8 / 1e6);

  if (chart.data.labels.length > state.maxHistoryPoints) {
    chart.data.labels.shift();
    chart.data.datasets[0].data.shift();
    chart.data.datasets[1].data.shift();
  }

  chart.update('none');
}

// ---- Update Protocol Chart ----
function updateProtocolChart(counts) {
  if (!counts || Object.keys(counts).length === 0) return;

  const colorMap = {
    TCP: '#6366f1', UDP: '#10b981', ICMP: '#f59e0b',
    DNS: '#8b5cf6', HTTP: '#06b6d4', ARP: '#ef4444', Other: '#64748b',
  };

  const labels = Object.keys(counts);
  const data = Object.values(counts);
  const colors = labels.map(l => colorMap[l] || colorMap.Other);

  const chart = state.charts.protocol;
  chart.data.labels = labels;
  chart.data.datasets[0].data = data;
  chart.data.datasets[0].backgroundColor = colors;
  chart.update('none');
}

// ---- Update KPI Cards ----
function updateKPIs(snap) {
  const pps = snap.packets_per_sec || 0;
  const mbps = ((snap.bytes_per_sec || 0) * 8 / 1e6);
  const totalPkts = snap.total_packets || 0;

  $('kpi-pps').textContent = fmt(pps, 0);
  $('kpi-mbps').textContent = mbps.toFixed(2);
  $('kpi-total-packets').textContent = fmt(totalPkts);
  $('kpi-alerts').textContent = state.alertCount;
}

// ---- Update Top IPs Table ----
function updateTopIPs(srcIPs, dstIPs) {
  const renderTable = (bodyId, ipMap) => {
    const tbody = $(bodyId);
    if (!ipMap || Object.keys(ipMap).length === 0) {
      tbody.innerHTML = '<tr><td colspan="3" class="empty-row">No data yet</td></tr>';
      return;
    }
    const sorted = Object.entries(ipMap).sort((a, b) => b[1] - a[1]).slice(0, 10);
    const total = sorted.reduce((s, [, v]) => s + v, 0);
    tbody.innerHTML = sorted.map(([ip, count]) => {
      const pct = total > 0 ? (count / total * 100).toFixed(1) : 0;
      return `<tr>
        <td style="font-family:monospace;color:var(--accent-cyan)">${ip}</td>
        <td>${fmt(count)}</td>
        <td>
          <div style="display:flex;align-items:center;gap:8px">
            <div class="progress-bar" style="flex:1">
              <div class="progress-fill" style="width:${pct}%"></div>
            </div>
            <span style="font-size:11px;color:var(--text-muted);min-width:36px">${pct}%</span>
          </div>
        </td>
      </tr>`;
    }).join('');
  };
  renderTable('topSrcBody', srcIPs);
  renderTable('topDstBody', dstIPs);
}

// ---- WebSocket Connection ----
function connectWebSocket() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  const ws = new WebSocket(`${proto}://${location.host}/api/ws`);

  ws.onopen = () => {
    console.log('[ws] connected');
  };

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);
      if (msg.type === 'stats') {
        const snap = msg.data;
        updateKPIs(snap);
        updateTrafficChart(snap);
        updateProtocolChart(snap.protocol_counts);
        updateTopIPs(snap.top_src_ips, snap.top_dst_ips);
      }
    } catch (e) {
      console.error('[ws] parse error', e);
    }
  };

  ws.onclose = () => {
    console.log('[ws] disconnected, reconnecting in 3s...');
    setTimeout(connectWebSocket, 3000);
  };

  ws.onerror = (err) => {
    console.error('[ws] error', err);
  };

  state.ws = ws;
}

// ---- Load Interfaces ----
async function loadInterfaces() {
  try {
    const res = await fetch('/api/interfaces');
    const data = await res.json();
    const ifaces = data.interfaces || [];

    // Populate select
    const sel = $('ifaceSelect');
    sel.innerHTML = ifaces.map(i => `<option value="${i.name}">${i.name}</option>`).join('');

    // Populate interfaces page
    const grid = $('interfacesList');
    grid.innerHTML = ifaces.map(i => `
      <div class="iface-card">
        <div class="iface-card-name">${i.name}</div>
        <div class="iface-card-detail">Flags: ${i.flags || '—'}</div>
        <div class="iface-card-detail">MTU: ${i.mtu || '—'}</div>
        <div class="iface-card-detail">MAC: <span class="iface-card-addr">${i.hw_addr || '—'}</span></div>
        ${(i.addrs || []).map(a => `<div class="iface-card-addr">${a}</div>`).join('')}
      </div>
    `).join('');
  } catch (e) {
    console.error('Failed to load interfaces', e);
  }
}

// ---- Load Connections ----
async function loadConnections() {
  try {
    const res = await fetch('/api/stats/connections');
    const data = await res.json();
    const conns = data.connections || [];
    $('connCount').textContent = conns.length;

    const tbody = $('connectionsBody');
    if (conns.length === 0) {
      tbody.innerHTML = '<tr><td colspan="8" class="empty-row">No connections tracked yet</td></tr>';
      return;
    }

    const sorted = conns.sort((a, b) => b.packets - a.packets).slice(0, 100);
    tbody.innerHTML = sorted.map(c => {
      const proto = c.protocol || 'Other';
      const cls = `proto-${proto.toLowerCase()}`;
      const lastSeen = c.last_seen ? new Date(c.last_seen).toLocaleTimeString() : '—';
      return `<tr>
        <td><span class="proto-badge ${cls}">${proto}</span></td>
        <td style="font-family:monospace">${c.src_ip || '—'}</td>
        <td>${c.src_port || '—'}</td>
        <td style="font-family:monospace">${c.dst_ip || '—'}</td>
        <td>${c.dst_port || '—'}</td>
        <td>${fmt(c.packets)}</td>
        <td>${fmtBytes(c.bytes || 0)}</td>
        <td>${lastSeen}</td>
      </tr>`;
    }).join('');
  } catch (e) {
    console.error('Failed to load connections', e);
  }
}

// ---- Load Alerts ----
async function loadAlerts() {
  try {
    const res = await fetch('/api/alerts');
    const data = await res.json();
    const alerts = data.alerts || [];
    state.alertCount = alerts.filter(a => !a.resolved).length;
    $('alert-badge').textContent = state.alertCount;
    $('kpi-alerts').textContent = state.alertCount;

    const list = $('alertsList');
    if (alerts.length === 0) {
      list.innerHTML = '<div class="empty-state">No alerts triggered yet</div>';
      return;
    }

    const iconMap = {
      info:     '&#9432;',
      warning:  '&#9888;',
      critical: '&#9888;',
    };

    list.innerHTML = [...alerts].reverse().map(a => `
      <div class="alert-item alert-item--${a.severity}">
        <div class="alert-icon">${iconMap[a.severity] || '!'}</div>
        <div style="flex:1">
          <div class="alert-title">${a.title}</div>
          <div class="alert-message">${a.message}</div>
          <div class="alert-time">${new Date(a.timestamp).toLocaleString()} &bull; Value: ${a.value?.toFixed(2)} / Threshold: ${a.threshold?.toFixed(2)}</div>
        </div>
      </div>
    `).join('');
  } catch (e) {
    console.error('Failed to load alerts', e);
  }
}

// ---- Load Config ----
async function loadConfig() {
  try {
    const res = await fetch('/api/config');
    const cfg = await res.json();
    if ($('cfg-port'))            $('cfg-port').value = cfg.port || 8080;
    if ($('cfg-snaplen'))         $('cfg-snaplen').value = cfg.snap_len || 65535;
    if ($('cfg-bpf'))             $('cfg-bpf').value = cfg.bpf_filter || '';
    if ($('cfg-bw-threshold'))    $('cfg-bw-threshold').value = cfg.alerts?.bandwidth_mbps_threshold || 100;
    if ($('cfg-pps-threshold'))   $('cfg-pps-threshold').value = cfg.alerts?.pps_threshold || 10000;
    if ($('cfg-alerts-enabled'))  $('cfg-alerts-enabled').checked = cfg.alerts?.enabled !== false;
  } catch (e) {
    console.error('Failed to load config', e);
  }
}

// ---- Save Config ----
$('configForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  const cfg = {
    port: parseInt($('cfg-port').value),
    snap_len: parseInt($('cfg-snaplen').value),
    bpf_filter: $('cfg-bpf').value,
    alerts: {
      enabled: $('cfg-alerts-enabled').checked,
      bandwidth_mbps_threshold: parseFloat($('cfg-bw-threshold').value),
      pps_threshold: parseFloat($('cfg-pps-threshold').value),
    },
  };
  try {
    const res = await fetch('/api/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(cfg),
    });
    if (res.ok) {
      showToast('Configuration saved successfully', 'success');
    } else {
      showToast('Failed to save configuration', 'error');
    }
  } catch (e) {
    showToast('Network error saving config', 'error');
  }
});

// ---- Capture Controls ----
$('startBtn').addEventListener('click', async () => {
  const iface = $('ifaceSelect').value;
  try {
    const res = await fetch('/api/capture/start', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ interface: iface }),
    });
    const data = await res.json();
    if (res.ok) {
      state.capturing = true;
      $('startBtn').disabled = true;
      $('stopBtn').disabled = false;
      $('statusDot').classList.add('active');
      $('statusText').textContent = `Capturing on ${data.interface}`;
      showToast(`Capture started on ${data.interface}`, 'success');
      connectWebSocket();
    } else {
      showToast(data.error || 'Failed to start capture', 'error');
    }
  } catch (e) {
    showToast('Failed to start capture: ' + e.message, 'error');
  }
});

$('stopBtn').addEventListener('click', async () => {
  try {
    await fetch('/api/capture/stop', { method: 'POST' });
    state.capturing = false;
    $('startBtn').disabled = false;
    $('stopBtn').disabled = true;
    $('statusDot').classList.remove('active');
    $('statusText').textContent = 'Idle';
    if (state.ws) state.ws.close();
    showToast('Capture stopped', 'info');
  } catch (e) {
    showToast('Failed to stop capture', 'error');
  }
});

// ---- Periodic Refresh ----
setInterval(() => {
  if (document.querySelector('#page-connections.active')) loadConnections();
  if (document.querySelector('#page-alerts.active')) loadAlerts();
}, 5000);

// ---- Initialize ----
(async function init() {
  initCharts();
  await loadInterfaces();

  // Check if already capturing
  try {
    const res = await fetch('/api/capture/status');
    const data = await res.json();
    if (data.running) {
      state.capturing = true;
      $('startBtn').disabled = true;
      $('stopBtn').disabled = false;
      $('statusDot').classList.add('active');
      $('statusText').textContent = `Capturing on ${data.interface}`;
      connectWebSocket();
    }
  } catch (e) {
    console.log('Could not check capture status');
  }
})();
