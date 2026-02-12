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

  try {
    const ov = await api('/overview');
    $('#repoCsaValue').textContent = fmt(ov.csaSum);
    const count = ov.serviceCount ?? 0;
    const csaAvg = ov.csaAvg;
    $('#repoCsaHint').textContent = `Sum of latest CSA across ${count} services${csaAvg !== null && csaAvg !== undefined ? ` (avg ${fmt(csaAvg)})` : ''}`;

    const ocmAvg = ov.ocmAvg;
    $('#repoOcmValue').textContent = ocmAvg !== null && ocmAvg !== undefined ? fmt(ocmAvg) : '-';
    const withOCM = ov.servicesWithOCM ?? 0;
    const ocmSum = ov.ocmSum;
    $('#repoOcmHint').textContent = `Average of latest OCM across ${withOCM}/${count} services${ocmSum !== null && ocmSum !== undefined ? ` (sum ${fmt(ocmSum)})` : ''}`;
  } catch (e) {
    $('#repoCsaValue').textContent = '-';
    $('#repoCsaHint').textContent = 'Repo aggregate unavailable';

    $('#repoOcmValue').textContent = '-';
    $('#repoOcmHint').textContent = 'Repo aggregate unavailable';
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
    return;
  }

  await loadService(Number(sel.value));
}

async function loadService(id) {
  const [csa, scores] = await Promise.all([
    api(`/services/${id}/metrics/CSA`),
    api(`/services/${id}/scores`),
  ]);

  const latestCSA = (csa.length ? csa[csa.length - 1].value : null);
  const latestOCM = (scores.length ? scores[scores.length - 1].score : null);
  $('#csaValue').textContent = fmt(latestCSA);
  $('#ocmValue').textContent = fmt(latestOCM);

  $('#raw').textContent = JSON.stringify({ scores, metrics: { CSA: csa } }, null, 2);
}

async function openEvidence(metricType) {
  const sel = $('#serviceSelect');
  const serviceID = Number(sel.value);
  const serviceName = sel.options[sel.selectedIndex]?.textContent ?? String(serviceID);

  $('#modalTitle').textContent = `${metricType} evidence`;
  $('#modalSub').textContent = `${serviceName}`;
  $('#modalMeta').textContent = 'Loading...';

  const modal = $('#modal');
  modal.setAttribute('aria-hidden', 'false');

  const tbody = $('#evidenceTable tbody');
  tbody.innerHTML = '';

  try {
    const items = await api(`/services/${serviceID}/metrics/${metricType}/evidence`);
    $('#modalMeta').textContent = `${items.length} evidence item(s) (latest run)`;
    for (const it of items) {
      const tr = document.createElement('tr');
      const manifest = [it.manifestKind, it.manifestName].filter(Boolean).join(' / ');
      tr.innerHTML = `
        <td>${escapeHTML(it.component || '')}</td>
        <td>${escapeHTML(it.key || '')}</td>
        <td>${escapeHTML(it.value || '')}</td>
        <td>${escapeHTML(manifest)}</td>
        <td>${escapeHTML(it.sourcePath || '')}</td>
      `;
      tbody.appendChild(tr);
    }
    if (items.length === 0) {
      const tr = document.createElement('tr');
      tr.innerHTML = `<td colspan="5" style="color: rgba(255,255,255,0.62)">No evidence recorded yet. Re-run analysis to populate.</td>`;
      tbody.appendChild(tr);
    }
  } catch (e) {
    $('#modalMeta').textContent = String(e);
  }
}

function closeModal() {
  $('#modal').setAttribute('aria-hidden', 'true');
}

function escapeHTML(s) {
  return String(s)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;');
}

document.addEventListener('DOMContentLoaded', () => {
  $('#refreshBtn').addEventListener('click', load);
  $('#serviceSelect').addEventListener('change', (e) => loadService(Number(e.target.value)));

  $('#csaTile').addEventListener('click', () => openEvidence('CSA'));
  $('#closeModalBtn').addEventListener('click', closeModal);
  $('#modalScrim').addEventListener('click', closeModal);
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closeModal();
  });

  load();
});
