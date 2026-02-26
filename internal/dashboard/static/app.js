/* ═══════════════════════════════════════════════════════════
   OCM Dashboard — Application Logic
   ═══════════════════════════════════════════════════════════ */

const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);

/* ─── Metric definitions ─── */
const METRICS = [
  { key: 'CSA', label: 'Configuration Surface Area', color: '#635bff', desc: 'Configurable parameters, env vars, resources, replicas' },
  { key: 'DD',  label: 'Dependency Depth',           color: '#0ea5e9', desc: 'Longest path in the service dependency graph' },
  { key: 'DB',  label: 'Dependency Breadth',          color: '#8b5cf6', desc: 'Direct upstream + downstream dependencies' },
  { key: 'CV',  label: 'Change Volatility',           color: '#f59e0b', desc: 'Commits affecting config in the last 30 days' },
  { key: 'FE',  label: 'Failure Exposure',             color: '#ef4444', desc: 'Exposed endpoints + external integrations' },
  { key: 'CDR', label: 'Configuration Drift Risk',     color: '#10b981', desc: 'Environment-specific overrides (dev/staging/prod)' },
];

const METRIC_COLOR = Object.fromEntries(METRICS.map(m => [m.key, m.color]));

/* ─── State ─── */
let state = {
  services: [],        // from /api/services
  overview: null,      // from /api/overview
  selectedService: null, // id — null means repo-wide view
  view: 'overview',
};

/* ─── API helper ─── */
async function api(path) {
  const res = await fetch(`/api${path}`, { headers: { Accept: 'application/json' } });
  if (!res.ok) throw new Error(`${res.status} ${await res.text()}`);
  return res.json();
}

/* ─── Format helpers ─── */
function fmt(n, decimals = 3) {
  if (n === null || n === undefined) return '--';
  if (Number.isInteger(n)) return String(n);
  return n.toFixed(decimals);
}

function fmtShort(n) {
  if (n === null || n === undefined) return '--';
  if (Number.isInteger(n)) return String(n);
  if (Math.abs(n) < 0.01) return '0';
  return n.toFixed(2);
}

function escapeHTML(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

/* ─── Context helpers ─── */
function isRepoView() {
  return state.selectedService === null;
}

function getSelectedServiceData() {
  if (!state.overview || !state.overview.services) return null;
  return state.overview.services.find(s => s.id === state.selectedService) || null;
}

function getSelectedServiceName() {
  const svc = state.services.find(s => s.id === state.selectedService);
  return svc?.name ?? null;
}

/* ═══════════════════════════════════════════
   Data loading
   ═══════════════════════════════════════════ */
async function loadAll() {
  const dot = $('#statusDot');
  const txt = $('#statusText');

  try {
    await api('/healthz');
    dot.className = 'status-dot status-dot--ok';
    txt.textContent = 'API connected';
  } catch {
    dot.className = 'status-dot status-dot--err';
    txt.textContent = 'API unreachable';
    return;
  }

  try {
    const [services, overview] = await Promise.all([
      api('/services'),
      api('/overview'),
    ]);
    state.services = services || [];
    state.overview = overview;

    populateServiceList();
    populateServiceSelect();

    renderOverview();
    renderServicesTable();

    if (state.selectedService) {
      await loadServiceDetail(state.selectedService);
    }
  } catch (e) {
    console.error('Load failed:', e);
    $('#statusText').textContent = 'Load error: ' + e.message;
  }
}

/* ─── Populate sidebar service list ─── */
function populateServiceList() {
  const section = $('#serviceListSection');
  const container = $('#serviceList');
  container.innerHTML = '';

  if (state.services.length === 0) {
    section.style.display = 'none';
    return;
  }
  section.style.display = '';

  // "All Services" button at top
  const allBtn = document.createElement('button');
  allBtn.className = 'sidebar-svc sidebar-svc--all' + (isRepoView() ? ' active' : '');
  allBtn.textContent = 'All Services';
  allBtn.addEventListener('click', () => selectService(null));
  container.appendChild(allBtn);

  for (const s of state.services) {
    const btn = document.createElement('button');
    btn.className = 'sidebar-svc' + (state.selectedService === s.id ? ' active' : '');
    btn.textContent = s.name;
    btn.addEventListener('click', () => selectService(s.id));
    container.appendChild(btn);
  }
}

/* ─── Populate service select ─── */
function populateServiceSelect() {
  const sel = $('#serviceSelect');
  sel.innerHTML = '';

  // "All Services" option
  const allOpt = document.createElement('option');
  allOpt.value = '';
  allOpt.textContent = 'All Services (Repo)';
  if (isRepoView()) allOpt.selected = true;
  sel.appendChild(allOpt);

  for (const s of state.services) {
    const opt = document.createElement('option');
    opt.value = s.id;
    opt.textContent = s.name;
    if (state.selectedService === s.id) opt.selected = true;
    sel.appendChild(opt);
  }
  if (state.services.length > 0) {
    sel.style.display = '';
  }
}

function selectService(id) {
  state.selectedService = id;

  // Update sidebar highlights
  $$('.sidebar-svc').forEach(el => {
    el.classList.remove('active');
    if (id === null && el.classList.contains('sidebar-svc--all')) {
      el.classList.add('active');
    } else {
      const svc = state.services.find(s => s.name === el.textContent);
      if (svc && svc.id === id) el.classList.add('active');
    }
  });

  // Update select
  $('#serviceSelect').value = id === null ? '' : id;

  // Re-render overview with correct context
  renderOverview();

  // Load service detail charts (or clear them for repo view)
  if (id !== null) {
    loadServiceDetail(id);
  } else {
    clearServiceCharts();
  }
}

function clearServiceCharts() {
  // Clear per-service charts when in repo-wide view
  const trendCanvas = $('#trendChart');
  const radarCanvas = $('#radarChart');
  const barCanvas = $('#barChart');

  for (const canvas of [trendCanvas, radarCanvas, barCanvas]) {
    if (!canvas) continue;
    const ctx = canvas.getContext('2d');
    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.parentElement.getBoundingClientRect();
    const w = rect.width;
    const h = canvas === trendCanvas ? 200 : 280;
    canvas.width = w * dpr;
    canvas.height = h * dpr;
    canvas.style.width = w + 'px';
    canvas.style.height = h + 'px';
    ctx.scale(dpr, dpr);
    ctx.clearRect(0, 0, w, h);
    ctx.fillStyle = '#6b727d';
    ctx.font = '13px Inter, sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('Select a service to see detailed charts', w / 2, h / 2);
  }
}

/* ═══════════════════════════════════════════
   Render: Overview
   ═══════════════════════════════════════════ */
function renderOverview() {
  const ov = state.overview;
  if (!ov) return;

  if (isRepoView()) {
    renderRepoOverview(ov);
  } else {
    renderServiceOverview(ov);
  }
}

/* ─── Repo-wide overview ─── */
function renderRepoOverview(ov) {
  // Page title
  $('#pageTitle').textContent = 'Repository Overview';

  // Hero cards
  const ocmAvg = ov.ocmAvg;
  $('#heroOCM').textContent = ocmAvg != null ? fmt(ocmAvg) : '--';
  $('#heroOCMSub').textContent = `Average across ${ov.servicesWithOCM || 0} services`;
  $('#heroCountLabel').textContent = 'Services Analyzed';
  $('#heroCount').textContent = ov.serviceCount || 0;

  // Section header
  $('.section-header .section-title').textContent = 'Metric Breakdown';
  $('.section-header .section-sub').textContent = 'Aggregated across all services. Click a tile for details.';

  // Metric tiles — repo-wide aggregates
  const grid = $('#metricGrid');
  grid.innerHTML = '';

  for (const m of METRICS) {
    const agg = ov.metrics?.[m.key];
    const sum = agg?.sum ?? 0;
    const avg = agg?.avg;
    const count = agg?.count ?? 0;

    const tile = document.createElement('div');
    tile.className = 'metric-tile';
    tile.dataset.metric = m.key;
    tile.innerHTML = `
      <div class="metric-tile-indicator" style="background:${m.color}"></div>
      <div class="metric-tile-header">
        <span class="metric-tile-label">${m.key}</span>
        <span class="metric-tile-badge">${count} svc</span>
      </div>
      <div class="metric-tile-value">${fmtShort(avg)}</div>
      <div class="metric-tile-sub">${m.label}</div>
      <div class="metric-tile-detail">avg across repo &middot; total: ${fmtShort(sum)}</div>
      <div class="metric-tile-sparkline"><canvas height="32" data-metric="${m.key}"></canvas></div>
    `;
    // In repo view, clicking a tile is informational (no single service to show evidence for)
    // We'll show aggregate info or prompt to select a service
    tile.addEventListener('click', () => {
      // Switch to services view so user can pick a service
      switchView('services');
    });
    grid.appendChild(tile);
  }

  // Update chart section headers
  updateChartHeaders(null);
}

/* ─── Service-specific overview ─── */
function renderServiceOverview(ov) {
  const svcData = getSelectedServiceData();
  const svcName = getSelectedServiceName();

  // Page title
  $('#pageTitle').textContent = svcName || 'Service Overview';

  // Hero cards — service-specific
  const ocm = svcData?.ocm;
  $('#heroOCM').textContent = ocm != null ? fmt(ocm) : '--';
  $('#heroOCMSub').textContent = svcName ? `Composite score for ${svcName}` : 'Service composite score';
  // Show metric count for this service
  const metricCount = svcData?.metrics ? Object.keys(svcData.metrics).length : 0;
  $('#heroCountLabel').textContent = 'Metrics Active';
  $('#heroCount').textContent = metricCount;

  // Section header
  $('.section-header .section-title').textContent = 'Metric Breakdown';
  $('.section-header .section-sub').textContent = `Individual metric values for ${svcName || 'this service'}. Click a tile for evidence.`;

  // Metric tiles — service-specific values
  const grid = $('#metricGrid');
  grid.innerHTML = '';

  for (const m of METRICS) {
    const val = svcData?.metrics?.[m.key];

    const tile = document.createElement('div');
    tile.className = 'metric-tile';
    tile.dataset.metric = m.key;
    tile.innerHTML = `
      <div class="metric-tile-indicator" style="background:${m.color}"></div>
      <div class="metric-tile-header">
        <span class="metric-tile-label">${m.key}</span>
        <span class="metric-tile-badge">${svcName || 'svc'}</span>
      </div>
      <div class="metric-tile-value">${fmtShort(val)}</div>
      <div class="metric-tile-sub">${m.label}</div>
      <div class="metric-tile-detail">${m.desc}</div>
      <div class="metric-tile-sparkline"><canvas height="32" data-metric="${m.key}"></canvas></div>
    `;
    tile.addEventListener('click', () => openEvidence(m.key));
    grid.appendChild(tile);
  }

  // Update chart section headers
  updateChartHeaders(svcName);
}

function updateChartHeaders(svcName) {
  const trendCard = $('#trendCard');
  if (trendCard) {
    const sub = trendCard.querySelector('.card-sub');
    if (sub) {
      sub.textContent = svcName
        ? `Composite score over time for ${svcName}`
        : 'Select a service to see score trends';
    }
  }
}

/* ═══════════════════════════════════════════
   Render: Services Table
   ═══════════════════════════════════════════ */
function renderServicesTable() {
  const ov = state.overview;
  if (!ov || !ov.services) return;

  const tbody = $('#servicesTableBody');
  tbody.innerHTML = '';

  for (const s of ov.services) {
    const tr = document.createElement('tr');
    tr.style.cursor = 'pointer';
    if (state.selectedService === s.id) {
      tr.classList.add('selected');
    }
    tr.addEventListener('click', () => {
      switchView('overview');
      selectService(s.id);
    });

    const ocm = s.ocm != null ? s.ocm : null;
    const ocmDisplay = ocm != null ? fmt(ocm) : '--';
    const ocmPct = ocm != null ? Math.min(ocm * 100, 100) : 0;

    tr.innerHTML = `
      <td class="svc-name">${escapeHTML(s.name)}</td>
      <td class="num">${fmtShort(s.metrics?.CSA)}</td>
      <td class="num">${fmtShort(s.metrics?.DD)}</td>
      <td class="num">${fmtShort(s.metrics?.DB)}</td>
      <td class="num">${fmtShort(s.metrics?.CV)}</td>
      <td class="num">${fmtShort(s.metrics?.FE)}</td>
      <td class="num">${fmtShort(s.metrics?.CDR)}</td>
      <td class="num ocm-cell">
        <div class="ocm-bar-wrap">
          <span>${ocmDisplay}</span>
          <div class="ocm-bar"><div class="ocm-bar-fill" style="width:${ocmPct}%"></div></div>
        </div>
      </td>
    `;
    tbody.appendChild(tr);
  }
}

/* ═══════════════════════════════════════════
   Load per-service detail (charts)
   ═══════════════════════════════════════════ */
async function loadServiceDetail(serviceId) {
  try {
    // Load all metric series + scores in parallel
    const [scores, ...metricSeries] = await Promise.all([
      api(`/services/${serviceId}/scores`),
      ...METRICS.map(m => api(`/services/${serviceId}/metrics/${m.key}`)),
    ]);

    const metricData = {};
    METRICS.forEach((m, i) => { metricData[m.key] = metricSeries[i]; });

    drawTrendChart(scores);
    drawRadarChart(serviceId, metricData);
    drawBarChart(serviceId, metricData);
    drawSparklines(metricData);
  } catch (e) {
    console.error('Service detail load failed:', e);
  }
}

/* ═══════════════════════════════════════════
   Charts — Trend line (OCM over time)
   ═══════════════════════════════════════════ */
function drawTrendChart(scores) {
  const canvas = $('#trendChart');
  const ctx = canvas.getContext('2d');
  const dpr = window.devicePixelRatio || 1;
  const rect = canvas.parentElement.getBoundingClientRect();
  const w = rect.width;
  const h = 200;
  canvas.width = w * dpr;
  canvas.height = h * dpr;
  canvas.style.width = w + 'px';
  canvas.style.height = h + 'px';
  ctx.scale(dpr, dpr);
  ctx.clearRect(0, 0, w, h);

  if (!scores || scores.length === 0) {
    ctx.fillStyle = '#6b727d';
    ctx.font = '13px Inter, sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('No score history yet. Run analysis multiple times to see trends.', w / 2, h / 2);
    return;
  }

  const pad = { top: 20, right: 20, bottom: 32, left: 48 };
  const cw = w - pad.left - pad.right;
  const ch = h - pad.top - pad.bottom;

  const vals = scores.map(s => s.score);
  const yMin = 0;
  const yMax = Math.max(1, ...vals) * 1.05;

  const xScale = (i) => pad.left + (i / Math.max(scores.length - 1, 1)) * cw;
  const yScale = (v) => pad.top + ch - ((v - yMin) / (yMax - yMin)) * ch;

  // Grid lines
  ctx.strokeStyle = 'rgba(255,255,255,0.06)';
  ctx.lineWidth = 1;
  for (let i = 0; i <= 4; i++) {
    const y = pad.top + (i / 4) * ch;
    ctx.beginPath();
    ctx.moveTo(pad.left, y);
    ctx.lineTo(w - pad.right, y);
    ctx.stroke();

    const label = (yMax - (i / 4) * (yMax - yMin)).toFixed(2);
    ctx.fillStyle = '#6b727d';
    ctx.font = '10px Inter, sans-serif';
    ctx.textAlign = 'right';
    ctx.fillText(label, pad.left - 8, y + 3);
  }

  // X-axis labels
  ctx.fillStyle = '#6b727d';
  ctx.font = '10px Inter, sans-serif';
  ctx.textAlign = 'center';
  const labelCount = Math.min(scores.length, 6);
  for (let i = 0; i < labelCount; i++) {
    const idx = Math.round((i / Math.max(labelCount - 1, 1)) * (scores.length - 1));
    const d = new Date(scores[idx].timestamp);
    const label = d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
    ctx.fillText(label, xScale(idx), h - 8);
  }

  // Area fill
  const grad = ctx.createLinearGradient(0, pad.top, 0, pad.top + ch);
  grad.addColorStop(0, 'rgba(99, 91, 255, 0.15)');
  grad.addColorStop(1, 'rgba(99, 91, 255, 0)');
  ctx.beginPath();
  ctx.moveTo(xScale(0), yScale(0));
  for (let i = 0; i < scores.length; i++) {
    ctx.lineTo(xScale(i), yScale(vals[i]));
  }
  ctx.lineTo(xScale(scores.length - 1), yScale(0));
  ctx.closePath();
  ctx.fillStyle = grad;
  ctx.fill();

  // Line
  ctx.beginPath();
  for (let i = 0; i < scores.length; i++) {
    const x = xScale(i);
    const y = yScale(vals[i]);
    if (i === 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  }
  ctx.strokeStyle = '#635bff';
  ctx.lineWidth = 2;
  ctx.lineJoin = 'round';
  ctx.stroke();

  // Dots
  for (let i = 0; i < scores.length; i++) {
    ctx.beginPath();
    ctx.arc(xScale(i), yScale(vals[i]), 3, 0, Math.PI * 2);
    ctx.fillStyle = '#635bff';
    ctx.fill();
    ctx.strokeStyle = '#111318';
    ctx.lineWidth = 1.5;
    ctx.stroke();
  }
}

/* ═══════════════════════════════════════════
   Charts — Radar (normalized metrics)
   ═══════════════════════════════════════════ */
function drawRadarChart(serviceId, metricData) {
  const canvas = $('#radarChart');
  const ctx = canvas.getContext('2d');
  const dpr = window.devicePixelRatio || 1;
  const rect = canvas.parentElement.getBoundingClientRect();
  const w = rect.width;
  const h = 280;
  canvas.width = w * dpr;
  canvas.height = h * dpr;
  canvas.style.width = w + 'px';
  canvas.style.height = h + 'px';
  ctx.scale(dpr, dpr);
  ctx.clearRect(0, 0, w, h);

  const cx = w / 2;
  const cy = h / 2 + 4;
  const r = Math.min(cx, cy) - 40;

  const n = METRICS.length;
  const angleStep = (Math.PI * 2) / n;

  // Get latest value for each metric
  const svc = state.overview?.services?.find(s => s.id === serviceId);
  const rawVals = METRICS.map(m => {
    if (svc && svc.metrics && svc.metrics[m.key] != null) return svc.metrics[m.key];
    const series = metricData[m.key];
    return series && series.length > 0 ? series[series.length - 1].value : 0;
  });

  // Normalize to 0-1 range for the radar
  const maxVal = Math.max(...rawVals, 1);
  const normVals = rawVals.map(v => v / maxVal);

  // Draw concentric circles
  for (let ring = 1; ring <= 4; ring++) {
    const rr = (ring / 4) * r;
    ctx.beginPath();
    ctx.arc(cx, cy, rr, 0, Math.PI * 2);
    ctx.strokeStyle = 'rgba(255,255,255,0.06)';
    ctx.lineWidth = 1;
    ctx.stroke();
  }

  // Draw axes & labels
  for (let i = 0; i < n; i++) {
    const angle = -Math.PI / 2 + i * angleStep;
    const x = cx + r * Math.cos(angle);
    const y = cy + r * Math.sin(angle);
    ctx.beginPath();
    ctx.moveTo(cx, cy);
    ctx.lineTo(x, y);
    ctx.strokeStyle = 'rgba(255,255,255,0.06)';
    ctx.lineWidth = 1;
    ctx.stroke();

    // Label
    const lx = cx + (r + 20) * Math.cos(angle);
    const ly = cy + (r + 20) * Math.sin(angle);
    ctx.fillStyle = METRICS[i].color;
    ctx.font = '600 11px Inter, sans-serif';
    ctx.textAlign = Math.abs(Math.cos(angle)) < 0.1 ? 'center' : Math.cos(angle) > 0 ? 'left' : 'right';
    ctx.textBaseline = Math.abs(Math.sin(angle)) < 0.1 ? 'middle' : Math.sin(angle) > 0 ? 'top' : 'bottom';
    ctx.fillText(METRICS[i].key, lx, ly);
  }

  // Draw filled polygon
  ctx.beginPath();
  for (let i = 0; i < n; i++) {
    const angle = -Math.PI / 2 + i * angleStep;
    const v = normVals[i];
    const x = cx + r * v * Math.cos(angle);
    const y = cy + r * v * Math.sin(angle);
    if (i === 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  }
  ctx.closePath();
  ctx.fillStyle = 'rgba(99, 91, 255, 0.12)';
  ctx.fill();
  ctx.strokeStyle = 'rgba(99, 91, 255, 0.6)';
  ctx.lineWidth = 1.5;
  ctx.stroke();

  // Dots at vertices
  for (let i = 0; i < n; i++) {
    const angle = -Math.PI / 2 + i * angleStep;
    const v = normVals[i];
    const x = cx + r * v * Math.cos(angle);
    const y = cy + r * v * Math.sin(angle);
    ctx.beginPath();
    ctx.arc(x, y, 4, 0, Math.PI * 2);
    ctx.fillStyle = METRICS[i].color;
    ctx.fill();
    ctx.strokeStyle = '#111318';
    ctx.lineWidth = 1.5;
    ctx.stroke();
  }
}

/* ═══════════════════════════════════════════
   Charts — Bar chart (raw values)
   ═══════════════════════════════════════════ */
function drawBarChart(serviceId, metricData) {
  const canvas = $('#barChart');
  const ctx = canvas.getContext('2d');
  const dpr = window.devicePixelRatio || 1;
  const rect = canvas.parentElement.getBoundingClientRect();
  const w = rect.width;
  const h = 280;
  canvas.width = w * dpr;
  canvas.height = h * dpr;
  canvas.style.width = w + 'px';
  canvas.style.height = h + 'px';
  ctx.scale(dpr, dpr);
  ctx.clearRect(0, 0, w, h);

  const svc = state.overview?.services?.find(s => s.id === serviceId);
  const vals = METRICS.map(m => {
    if (svc && svc.metrics && svc.metrics[m.key] != null) return svc.metrics[m.key];
    const series = metricData[m.key];
    return series && series.length > 0 ? series[series.length - 1].value : 0;
  });

  const pad = { top: 16, right: 16, bottom: 36, left: 48 };
  const cw = w - pad.left - pad.right;
  const ch = h - pad.top - pad.bottom;
  const n = METRICS.length;
  const barW = Math.min(cw / n * 0.6, 40);
  const gap = cw / n;
  const maxVal = Math.max(...vals, 1);

  // Grid
  ctx.strokeStyle = 'rgba(255,255,255,0.06)';
  ctx.lineWidth = 1;
  for (let i = 0; i <= 4; i++) {
    const y = pad.top + (i / 4) * ch;
    ctx.beginPath();
    ctx.moveTo(pad.left, y);
    ctx.lineTo(w - pad.right, y);
    ctx.stroke();

    const label = Math.round(maxVal - (i / 4) * maxVal);
    ctx.fillStyle = '#6b727d';
    ctx.font = '10px Inter, sans-serif';
    ctx.textAlign = 'right';
    ctx.fillText(label, pad.left - 8, y + 3);
  }

  // Bars
  for (let i = 0; i < n; i++) {
    const x = pad.left + gap * i + (gap - barW) / 2;
    const barH = (vals[i] / maxVal) * ch;
    const y = pad.top + ch - barH;

    // Bar with rounded top
    const radius = Math.min(barW / 2, 4);
    ctx.beginPath();
    ctx.moveTo(x, pad.top + ch);
    ctx.lineTo(x, y + radius);
    ctx.quadraticCurveTo(x, y, x + radius, y);
    ctx.lineTo(x + barW - radius, y);
    ctx.quadraticCurveTo(x + barW, y, x + barW, y + radius);
    ctx.lineTo(x + barW, pad.top + ch);
    ctx.closePath();
    ctx.fillStyle = METRICS[i].color + 'cc';
    ctx.fill();

    // Value label on top
    ctx.fillStyle = '#f0f0f3';
    ctx.font = '600 11px Inter, sans-serif';
    ctx.textAlign = 'center';
    if (vals[i] > 0) {
      ctx.fillText(fmtShort(vals[i]), x + barW / 2, y - 6);
    }

    // X-axis label
    ctx.fillStyle = METRICS[i].color;
    ctx.font = '600 11px Inter, sans-serif';
    ctx.fillText(METRICS[i].key, x + barW / 2, h - 10);
  }
}

/* ═══════════════════════════════════════════
   Charts — Sparklines (in metric tiles)
   ═══════════════════════════════════════════ */
function drawSparklines(metricData) {
  for (const m of METRICS) {
    const canvas = document.querySelector(`canvas[data-metric="${m.key}"]`);
    if (!canvas) continue;

    const ctx = canvas.getContext('2d');
    const dpr = window.devicePixelRatio || 1;
    const parent = canvas.parentElement;
    const w = parent.clientWidth;
    const h = 32;
    canvas.width = w * dpr;
    canvas.height = h * dpr;
    canvas.style.width = w + 'px';
    canvas.style.height = h + 'px';
    ctx.scale(dpr, dpr);
    ctx.clearRect(0, 0, w, h);

    const series = metricData[m.key];
    if (!series || series.length < 2) continue;

    const vals = series.map(p => p.value);
    const min = Math.min(...vals);
    const max = Math.max(...vals);
    const range = max - min || 1;

    // Area
    const grad = ctx.createLinearGradient(0, 0, 0, h);
    grad.addColorStop(0, m.color + '30');
    grad.addColorStop(1, m.color + '00');

    ctx.beginPath();
    ctx.moveTo(0, h);
    for (let i = 0; i < vals.length; i++) {
      const x = (i / (vals.length - 1)) * w;
      const y = h - 4 - ((vals[i] - min) / range) * (h - 8);
      ctx.lineTo(x, y);
    }
    ctx.lineTo(w, h);
    ctx.closePath();
    ctx.fillStyle = grad;
    ctx.fill();

    // Line
    ctx.beginPath();
    for (let i = 0; i < vals.length; i++) {
      const x = (i / (vals.length - 1)) * w;
      const y = h - 4 - ((vals[i] - min) / range) * (h - 8);
      if (i === 0) ctx.moveTo(x, y);
      else ctx.lineTo(x, y);
    }
    ctx.strokeStyle = m.color;
    ctx.lineWidth = 1.5;
    ctx.lineJoin = 'round';
    ctx.stroke();
  }
}

/* ═══════════════════════════════════════════
   Evidence modal
   ═══════════════════════════════════════════ */
async function openEvidence(metricType) {
  const serviceId = state.selectedService;
  if (!serviceId) return; // No evidence in repo-wide view

  const svc = state.services.find(s => s.id === serviceId);
  const serviceName = svc?.name ?? String(serviceId);
  const metricDef = METRICS.find(m => m.key === metricType);

  $('#modalTitle').textContent = `${metricType} — ${metricDef?.label || metricType}`;
  $('#modalSub').textContent = serviceName;
  $('#modalMeta').textContent = 'Loading...';

  const modal = $('#modal');
  modal.setAttribute('aria-hidden', 'false');

  const tbody = $('#evidenceTable tbody');
  tbody.innerHTML = '';

  try {
    const items = await api(`/services/${serviceId}/metrics/${metricType}/evidence`);
    $('#modalMeta').textContent = `${items.length} evidence item${items.length !== 1 ? 's' : ''} from the latest analysis run`;

    for (const it of items) {
      const tr = document.createElement('tr');
      const manifest = [it.manifestKind, it.manifestName].filter(Boolean).join(' / ');
      tr.innerHTML = `
        <td>${escapeHTML(it.component || '')}</td>
        <td><code>${escapeHTML(it.key || '')}</code></td>
        <td>${escapeHTML(it.value || '')}</td>
        <td>${escapeHTML(manifest || '--')}</td>
        <td style="max-width:220px;overflow:hidden;text-overflow:ellipsis">${escapeHTML(it.sourcePath || '--')}</td>
      `;
      tbody.appendChild(tr);
    }

    if (items.length === 0) {
      const tr = document.createElement('tr');
      tr.innerHTML = `<td colspan="5" style="color:var(--text-tertiary);text-align:center;padding:24px">No evidence recorded for this metric. The metric value may be zero, or evidence generation is not applicable.</td>`;
      tbody.appendChild(tr);
    }
  } catch (e) {
    $('#modalMeta').textContent = 'Error: ' + e.message;
  }
}

function closeModal() {
  $('#modal').setAttribute('aria-hidden', 'true');
}

/* ═══════════════════════════════════════════
   View switching
   ═══════════════════════════════════════════ */
function switchView(view) {
  state.view = view;
  $$('.sidebar-link').forEach(el => {
    el.classList.toggle('active', el.dataset.view === view);
  });
  $('#overviewView').style.display = view === 'overview' ? '' : 'none';
  $('#servicesView').style.display = view === 'services' ? '' : 'none';

  if (view === 'services') {
    $('#pageTitle').textContent = 'Services';
  } else {
    // Re-render overview to set proper title
    renderOverview();
  }

  // Re-render charts when switching to overview (canvas sizing)
  if (view === 'overview' && state.selectedService) {
    setTimeout(() => loadServiceDetail(state.selectedService), 50);
  } else if (view === 'overview' && isRepoView()) {
    setTimeout(() => clearServiceCharts(), 50);
  }
}

/* ═══════════════════════════════════════════
   Event listeners
   ═══════════════════════════════════════════ */
document.addEventListener('DOMContentLoaded', () => {
  // Navigation
  $$('.sidebar-link').forEach(el => {
    el.addEventListener('click', () => switchView(el.dataset.view));
  });

  // Service select
  $('#serviceSelect').addEventListener('change', (e) => {
    const val = e.target.value;
    selectService(val === '' ? null : Number(val));
  });

  // Refresh
  $('#refreshBtn').addEventListener('click', loadAll);

  // Modal
  $('#closeModalBtn').addEventListener('click', closeModal);
  $('#modalBackdrop').addEventListener('click', closeModal);
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closeModal();
  });

  // Mobile sidebar toggle
  $('#menuToggle').addEventListener('click', () => {
    $('#sidebar').classList.toggle('open');
  });

  // Resize handler for charts
  let resizeTimer;
  window.addEventListener('resize', () => {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(() => {
      if (state.selectedService) loadServiceDetail(state.selectedService);
      else clearServiceCharts();
    }, 200);
  });

  // Initial load
  loadAll();
});
