const $ = (sel) => document.querySelector(sel);

async function api(path) {
  const res = await fetch(`/api${path}`, { headers: { 'Accept': 'application/json' } });
  if (!res.ok) {
    const txt = await res.text();
    throw new Error(`${res.status} ${txt}`);
  }
  return res.json();
}

function fmt(n) {
  if (n === null || n === undefined) return '-';
  if (Number.isInteger(n)) return String(n);
  return (Math.round(n * 1000) / 1000).toFixed(3);
}

function setPill(el, ok, text) {
  el.textContent = text;
  el.style.borderColor = ok ? 'rgba(61,214,198,0.35)' : 'rgba(246,193,119,0.35)';
  el.style.background = ok ? 'rgba(61,214,198,0.10)' : 'rgba(246,193,119,0.10)';
}

function drawTrend(canvas, points) {
  const ctx = canvas.getContext('2d');
  const w = canvas.width;
  const h = canvas.height;
  ctx.clearRect(0, 0, w, h);

  // frame
  ctx.strokeStyle = 'rgba(255,255,255,0.12)';
  ctx.lineWidth = 1;
  ctx.strokeRect(0.5, 0.5, w - 1, h - 1);

  if (!points || points.length === 0) {
    ctx.fillStyle = 'rgba(255,255,255,0.60)';
    ctx.font = '12px ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace';
    ctx.fillText('No data yet. Run the analyzer again to add timepoints.', 16, 28);
    return;
  }

  // OCM score is normalized; keep axis stable for readability.
  const min = 0.0;
  const max = 1.0;
  const pad = 18;
  const x0 = pad, y0 = pad, x1 = w - pad, y1 = h - pad;

  // grid
  ctx.strokeStyle = 'rgba(255,255,255,0.06)';
  for (let i = 1; i <= 4; i++) {
    const y = y0 + (i * (y1 - y0) / 5);
    ctx.beginPath(); ctx.moveTo(x0, y); ctx.lineTo(x1, y); ctx.stroke();
  }

  const norm = (v) => Math.max(0, Math.min(1, (v - min) / (max - min)));

  const step = (points.length === 1) ? 0 : (x1 - x0) / (points.length - 1);
  const xy = (i) => {
    const x = x0 + i * step;
    const y = y1 - norm(points[i].score) * (y1 - y0);
    return { x, y };
  };

  // line gradient
  const grad = ctx.createLinearGradient(x0, y0, x1, y1);
  grad.addColorStop(0, 'rgba(61,214,198,0.9)');
  grad.addColorStop(1, 'rgba(246,193,119,0.85)');

  ctx.lineWidth = 2;
  ctx.strokeStyle = grad;
  ctx.beginPath();
  const p0 = xy(0);
  ctx.moveTo(p0.x, p0.y);
  for (let i = 1; i < points.length; i++) {
    const p = xy(i);
    ctx.lineTo(p.x, p.y);
  }
  ctx.stroke();

  // points
  for (let i = 0; i < points.length; i++) {
    const p = xy(i);
    ctx.fillStyle = 'rgba(10,15,20,0.95)';
    ctx.beginPath(); ctx.arc(p.x, p.y, 4.5, 0, Math.PI * 2); ctx.fill();
    ctx.strokeStyle = 'rgba(255,255,255,0.18)';
    ctx.lineWidth = 1;
    ctx.stroke();

    ctx.fillStyle = i === points.length - 1 ? 'rgba(255,255,255,0.92)' : 'rgba(255,255,255,0.55)';
    ctx.font = '11px ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace';
    ctx.fillText(fmt(points[i].score), p.x + 8, p.y - 8);
  }

  // axis labels
  ctx.fillStyle = 'rgba(255,255,255,0.55)';
  ctx.font = '11px ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace';
  ctx.fillText('1.0', 6, y0 + 4);
  ctx.fillText('0.0', 6, y1);
}

async function load() {
  const apiPill = $('#apiPill');
  try {
    await api('/healthz');
    setPill(apiPill, true, 'API: ok');
  } catch (e) {
    setPill(apiPill, false, 'API: down');
    $('#raw').textContent = String(e);
    return;
  }

  const services = await api('/services');
  const sel = $('#serviceSelect');
  sel.innerHTML = '';
  for (const s of services) {
    const opt = document.createElement('option');
    opt.value = String(s.id);
    opt.textContent = s.name;
    sel.appendChild(opt);
  }

  if (services.length === 0) {
    $('#raw').textContent = 'No services found in DB.';
    drawTrend($('#trend'), []);
    return;
  }

  await loadService(Number(sel.value));
}

async function loadService(id) {
  const [scores, csa, dd] = await Promise.all([
    api(`/services/${id}/scores`),
    api(`/services/${id}/metrics/CSA`),
    api(`/services/${id}/metrics/DD`),
  ]);

  const latestScore = (scores.length ? scores[scores.length - 1].score : null);
  const latestCSA = (csa.length ? csa[csa.length - 1].value : null);
  const latestDD = (dd.length ? dd[dd.length - 1].value : null);

  const latestAt = (scores.length ? scores[scores.length - 1].timestamp : null);
  $('#latestAt').textContent = latestAt ? `latest: ${latestAt}` : '';

  $('#ocmValue').textContent = fmt(latestScore);
  $('#csaValue').textContent = fmt(latestCSA);
  $('#ddValue').textContent = fmt(latestDD);

  drawTrend($('#trend'), scores);
  $('#raw').textContent = JSON.stringify({ scores, metrics: { CSA: csa, DD: dd } }, null, 2);
}

document.addEventListener('DOMContentLoaded', () => {
  $('#refreshBtn').addEventListener('click', load);
  $('#serviceSelect').addEventListener('change', (e) => loadService(Number(e.target.value)));
  load();
});
