/* ─── MOTORIA — APP.JS ──────────────────────────────────────────────────── */

// ── State ──────────────────────────────────────────────────────────────────
const state = {
  cars: [],
  manufacturers: [],
  categories: [],
  compareIds: new Set(),
  currentView: 'gallery',
};

// ── Helpers ────────────────────────────────────────────────────────────────
const $ = id => document.getElementById(id);
const $$ = sel => document.querySelectorAll(sel);

async function apiFetch(path) {
  const res = await fetch(path);
  if (!res.ok) throw new Error(`API error: ${res.status}`);
  return res.json();
}

function imgSrc(filename) {
  return filename ? `/static/img/${filename}` : '';
}

// ── View Navigation ────────────────────────────────────────────────────────
function showView(name) {
  state.currentView = name;
  $$('.view').forEach(v => v.classList.remove('active'));
  $$('.nav-btn').forEach(b => b.classList.remove('active'));
  const view = $(`view-${name}`);
  const btn = document.querySelector(`[data-view="${name}"]`);
  if (view) view.classList.add('active');
  if (btn) btn.classList.add('active');
}

$$('.nav-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    showView(btn.dataset.view);
    if (btn.dataset.view === 'manufacturers') renderManufacturers();
    if (btn.dataset.view === 'recommendations') populatePrefCategory();
  });
});

// ── Bootstrap ──────────────────────────────────────────────────────────────
async function init() {
  try {
    const [cars, manufacturers, categories] = await Promise.all([
      apiFetch('/api/search'),
      apiFetch('/api/manufacturers'),
      apiFetch('/api/categories'),
    ]);
    state.cars = cars;
    state.manufacturers = manufacturers;
    state.categories = categories;

    populateFilters();
    renderGallery(cars);
  } catch (err) {
    $('car-grid').innerHTML = `<p class="loading">Failed to load data</p>`;
    console.error(err);
  }
}

// ── Populate Filters ───────────────────────────────────────────────────────
function populateFilters() {
  const catSel = $('filter-category');
  const mfrSel = $('filter-manufacturer');

  state.categories.forEach(c => {
    catSel.innerHTML += `<option value="${c.name}">${c.name}</option>`;
  });
  state.manufacturers.forEach(m => {
    mfrSel.innerHTML += `<option value="${m.name}">${m.name}</option>`;
  });

  // Compare slots
  $$('.slot-select').forEach(sel => {
    state.cars.forEach(c => {
      sel.innerHTML += `<option value="${c.id}">${c.name} (${c.year})</option>`;
    });
  });
}

function populatePrefCategory() {
  const sel = $('pref-category');
  if (sel.children.length > 1) return;
  state.categories.forEach(c => {
    sel.innerHTML += `<option value="${c.name}">${c.name}</option>`;
  });
}

// ── Gallery Rendering ──────────────────────────────────────────────────────
function renderGallery(cars) {
  const grid = $('car-grid');
  const noRes = $('no-results');

  if (!cars || cars.length === 0) {
    grid.innerHTML = '';
    noRes.classList.remove('hidden');
    return;
  }
  noRes.classList.add('hidden');

  grid.innerHTML = cars.map((car, i) => {
    const catName = car.categoryName || getCategoryName(car.categoryId);
    const mfrName = car.manufacturerName || getManufacturerName(car.manufacturerId);
    const isSelected = state.compareIds.has(car.id) ? 'selected' : '';

    return `
    <div class="car-card" data-id="${car.id}" style="animation-delay:${i * 0.04}s">
      <div class="card-img-wrap">
        ${car.image
          ? `<img class="card-img" src="${imgSrc(car.image)}" alt="${car.name}" onerror="this.style.display='none'">`
          : `<div class="card-img-placeholder"><svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#333" stroke-width="1.5"><path d="M5 17H3a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2h11l5 5v5h-2m-7 0a2 2 0 1 0 4 0 2 2 0 0 0-4 0M5 17a2 2 0 1 0 4 0 2 2 0 0 0-4 0"/></svg></div>`
        }
        <div class="card-overlay"><span class="card-overlay-text">VIEW DETAILS →</span></div>
        <span class="card-badge">${catName}</span>
        <button class="card-compare-btn ${isSelected}" data-id="${car.id}" title="Add to compare">
          ${state.compareIds.has(car.id) ? '✓ COMPARING' : '+ COMPARE'}
        </button>
      </div>
      <div class="card-body">
        <div class="card-category">${catName}</div>
        <div class="card-name">${car.name}</div>
        <div class="card-specs">
          <div class="spec-item">
            <span class="spec-label">Engine</span>
            <span class="spec-value">${car.specifications?.engine || '—'}</span>
          </div>
          <div class="spec-item">
            <span class="spec-label">Power</span>
            <span class="spec-value hp">${car.specifications?.horsepower || '—'} HP</span>
          </div>
          <div class="spec-item">
            <span class="spec-label">Transmission</span>
            <span class="spec-value">${car.specifications?.transmission || '—'}</span>
          </div>
          <div class="spec-item">
            <span class="spec-label">Drivetrain</span>
            <span class="spec-value">${car.specifications?.drivetrain || '—'}</span>
          </div>
        </div>
      </div>
      <div class="card-footer">
        <span class="card-year">${car.year}</span>
        <span class="card-maker">${mfrName}</span>
      </div>
    </div>`;
  }).join('');

  // Click handlers
  grid.querySelectorAll('.car-card').forEach(card => {
    card.addEventListener('click', e => {
      if (e.target.closest('.card-compare-btn')) return;
      openModal(parseInt(card.dataset.id));
    });
  });

  grid.querySelectorAll('.card-compare-btn').forEach(btn => {
    btn.addEventListener('click', e => {
      e.stopPropagation();
      toggleCompare(parseInt(btn.dataset.id));
      // re-render to update button state
      runSearch();
    });
  });
}

function getCategoryName(id) {
  const c = state.categories.find(c => c.id === id);
  return c ? c.name : '';
}

function getManufacturerName(id) {
  const m = state.manufacturers.find(m => m.id === id);
  return m ? m.name : '';
}

// ── Compare Toggle ─────────────────────────────────────────────────────────
function toggleCompare(id) {
  if (state.compareIds.has(id)) {
    state.compareIds.delete(id);
  } else {
    if (state.compareIds.size >= 3) {
      // Remove oldest
      const first = state.compareIds.values().next().value;
      state.compareIds.delete(first);
    }
    state.compareIds.add(id);
  }

  // Sync compare dropdowns
  const ids = [...state.compareIds];
  $$('.slot-select').forEach((sel, i) => {
    sel.value = ids[i] || '';
  });
}

// ── Search / Filter ────────────────────────────────────────────────────────
let debounceTimer;

function runSearch() {
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(async () => {
    const q = $('search-input').value;
    const category = $('filter-category').value;
    const manufacturer = $('filter-manufacturer').value;
    const sort = $('filter-sort').value;
    const minHP = $('hp-min').value;
    const maxHP = $('hp-max').value;

    const params = new URLSearchParams();
    if (q) params.set('q', q);
    if (category) params.set('category', category);
    if (manufacturer) params.set('manufacturer', manufacturer);
    if (sort) params.set('sort', sort);
    if (minHP > 0) params.set('minHP', minHP);
    if (maxHP < 500) params.set('maxHP', maxHP);

    try {
      const results = await apiFetch(`/api/search?${params}`);
      renderGallery(results);
    } catch(e) {
      console.error(e);
    }
  }, 300);
}

$('search-input').addEventListener('input', runSearch);
$('filter-category').addEventListener('change', runSearch);
$('filter-manufacturer').addEventListener('change', runSearch);
$('filter-sort').addEventListener('change', runSearch);

// HP range
function updateHPDisplay() {
  const min = parseInt($('hp-min').value);
  const max = parseInt($('hp-max').value);
  $('hp-display').textContent = `${min}–${max}`;
  runSearch();
}

$('hp-min').addEventListener('input', () => {
  const min = parseInt($('hp-min').value);
  const max = parseInt($('hp-max').value);
  if (min > max) $('hp-min').value = max;
  updateHPDisplay();
});

$('hp-max').addEventListener('input', () => {
  const min = parseInt($('hp-min').value);
  const max = parseInt($('hp-max').value);
  if (max < min) $('hp-max').value = min;
  updateHPDisplay();
});

// ── Modal ──────────────────────────────────────────────────────────────────
async function openModal(id) {
  const overlay = $('modal-overlay');
  const inner = $('modal-inner');
  overlay.classList.remove('hidden');
  inner.innerHTML = `<div class="loading">LOADING</div>`;

  try {
    const car = await apiFetch(`/api/models/${id}`);
    const mfr = car.manufacturer;
    const cat = car.category;

    inner.innerHTML = `
      <div class="modal-hero">
        ${car.image
          ? `<img class="modal-hero-img" src="${imgSrc(car.image)}" alt="${car.name}" onerror="this.style.opacity=0">`
          : ''
        }
        <div class="modal-hero-gradient"></div>
        <div class="modal-hero-content">
          <div class="modal-category">${cat?.name || ''}</div>
          <div class="modal-name">${car.name}</div>
        </div>
      </div>
      <div class="modal-body">
        <div class="modal-grid">
          <div class="modal-stat">
            <div class="modal-stat-label">Horsepower</div>
            <div class="modal-stat-value big">${car.specifications?.horsepower}</div>
          </div>
          <div class="modal-stat">
            <div class="modal-stat-label">Year</div>
            <div class="modal-stat-value big">${car.year}</div>
          </div>
          <div class="modal-stat">
            <div class="modal-stat-label">Engine</div>
            <div class="modal-stat-value">${car.specifications?.engine}</div>
          </div>
          <div class="modal-stat">
            <div class="modal-stat-label">Transmission</div>
            <div class="modal-stat-value">${car.specifications?.transmission}</div>
          </div>
          <div class="modal-stat">
            <div class="modal-stat-label">Drivetrain</div>
            <div class="modal-stat-value">${car.specifications?.drivetrain}</div>
          </div>
          <div class="modal-stat">
            <div class="modal-stat-label">Category</div>
            <div class="modal-stat-value">${cat?.name || '—'}</div>
          </div>
        </div>
        ${mfr ? `
        <div class="modal-maker-section">
          <div class="modal-maker-title">MANUFACTURER</div>
          <div class="modal-maker-name">${mfr.name}</div>
          <div class="modal-maker-meta">
            ${mfr.country} · Est. ${mfr.foundingYear}
          </div>
        </div>` : ''}
        <button class="modal-add-compare" data-id="${car.id}">
          ${state.compareIds.has(car.id) ? '✓ ALREADY IN COMPARISON' : '+ ADD TO COMPARISON'}
        </button>
      </div>
    `;

    inner.querySelector('.modal-add-compare').addEventListener('click', () => {
      toggleCompare(car.id);
      runSearch();
      closeModal();
      showView('compare');
    });

  } catch(e) {
    inner.innerHTML = `<p style="padding:2rem;color:var(--accent)">Failed to load details</p>`;
  }
}

function closeModal() {
  $('modal-overlay').classList.add('hidden');
}

$('modal-close').addEventListener('click', closeModal);
$('modal-overlay').addEventListener('click', e => {
  if (e.target === $('modal-overlay')) closeModal();
});

document.addEventListener('keydown', e => {
  if (e.key === 'Escape') closeModal();
});

// ── Compare ────────────────────────────────────────────────────────────────
$('compare-btn').addEventListener('click', async () => {
  const ids = $$('.slot-select')
    .values()
    .filter(s => s.value)
    .map(s => s.value)
    .filter(Boolean);

  if (ids.length < 2) {
    $('compare-results').innerHTML = `<p style="font-family:var(--font-mono);color:var(--accent);font-size:0.8rem;padding:1rem;letter-spacing:.1em;">SELECT AT LEAST 2 VEHICLES</p>`;
    return;
  }

  try {
    const cars = await apiFetch(`/api/compare?ids=${ids.join(',')}`);
    renderCompareTable(cars);
  } catch(e) {
    console.error(e);
  }
});

function renderCompareTable(cars) {
  const wrap = $('compare-results');
  if (!cars || cars.length === 0) {
    wrap.innerHTML = '<p style="color:var(--text-dim);font-family:var(--font-mono)">No data</p>';
    return;
  }

  const maxHP = Math.max(...cars.map(c => c.specifications.horsepower));
  const minHP = Math.min(...cars.map(c => c.specifications.horsepower));
  const maxYear = Math.max(...cars.map(c => c.year));

  const rows = [
    { label: 'Photo', key: 'image', render: (car) =>
      car.image ? `<img class="thumb" src="${imgSrc(car.image)}" alt="${car.name}" onerror="this.style.opacity=0.1">` : '—'
    },
    { label: 'Category', key: 'cat', render: (car) => car.category?.name || '—' },
    { label: 'Manufacturer', key: 'mfr', render: (car) => car.manufacturer?.name || '—' },
    { label: 'Year', key: 'year', render: (car, _) => {
      const cls = car.year === maxYear ? 'best-value' : '';
      return `<span class="${cls}">${car.year}</span>`;
    }},
    { label: 'Engine', key: 'engine', render: (car) => car.specifications?.engine || '—' },
    { label: 'Horsepower', key: 'hp', render: (car) => {
      const cls = car.specifications.horsepower === maxHP ? 'best-value'
                : car.specifications.horsepower === minHP ? 'worst-value' : '';
      return `<span class="${cls}">${car.specifications.horsepower} HP</span>`;
    }},
    { label: 'Transmission', key: 'trans', render: (car) => car.specifications?.transmission || '—' },
    { label: 'Drivetrain', key: 'drive', render: (car) => car.specifications?.drivetrain || '—' },
  ];

  wrap.innerHTML = `
    <table class="compare-table">
      <thead>
        <tr>
          <th>Spec</th>
          ${cars.map(c => `<th><div class="car-col-header">${c.name}<small>${c.year}</small></div></th>`).join('')}
        </tr>
      </thead>
      <tbody>
        ${rows.map(row => `
          <tr>
            <td style="font-family:var(--font-mono);font-size:.65rem;letter-spacing:.1em;color:var(--text-dim);text-transform:uppercase">${row.label}</td>
            ${cars.map(car => `<td>${row.render(car)}</td>`).join('')}
          </tr>
        `).join('')}
      </tbody>
    </table>
  `;
}

// ── Manufacturers ──────────────────────────────────────────────────────────
function renderManufacturers() {
  const grid = $('makers-grid');
  if (grid.children.length > 0) return; // already rendered

  grid.innerHTML = state.manufacturers.map(m => `
    <div class="maker-card" data-id="${m.id}">
      <div class="maker-name">${m.name}</div>
      <div class="maker-meta">
        <span class="maker-country">${m.country}</span> · Est. ${m.foundingYear}
      </div>
    </div>
  `).join('');

  grid.querySelectorAll('.maker-card').forEach(card => {
    card.addEventListener('click', () => loadMakerDetail(parseInt(card.dataset.id), card));
  });
}

async function loadMakerDetail(id, cardEl) {
  $$('.maker-card').forEach(c => c.classList.remove('active'));
  cardEl.classList.add('active');

  const detail = $('maker-detail');
  detail.classList.remove('hidden');
  detail.innerHTML = `<div class="loading">LOADING</div>`;

  try {
    const data = await apiFetch(`/api/manufacturers/${id}`);
    const m = data.manufacturer;
    const models = data.models || [];

    detail.innerHTML = `
      <div class="modal-maker-title">MANUFACTURER PROFILE</div>
      <h3>${m.name}</h3>
      <div class="maker-detail-meta">${m.country} · Founded ${m.foundingYear} · ${models.length} model${models.length !== 1 ? 's' : ''} in fleet</div>
      <div style="font-family:var(--font-mono);font-size:.7rem;color:var(--text-dim);letter-spacing:.1em;margin-bottom:.75rem">MODELS</div>
      <div class="maker-models-grid">
        ${models.map(car => `
          <div class="maker-model-card" data-id="${car.id}">
            <div class="maker-model-name">${car.name}</div>
            <div class="maker-model-sub">${car.year} · ${car.specifications?.horsepower} HP</div>
          </div>
        `).join('')}
      </div>
    `;

    detail.querySelectorAll('.maker-model-card').forEach(card => {
      card.addEventListener('click', () => openModal(parseInt(card.dataset.id)));
    });

    detail.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
  } catch(e) {
    detail.innerHTML = `<p style="color:var(--accent);font-family:var(--font-mono)">Failed to load</p>`;
  }
}

// ── Recommendations ────────────────────────────────────────────────────────
$('get-recs-btn').addEventListener('click', async () => {
  const category = $('pref-category').value;
  const minHP = $('pref-minhp').value;
  const maxHP = $('pref-maxhp').value;

  const params = new URLSearchParams();
  if (category) params.set('category', category);
  if (minHP) params.set('minHP', minHP);
  if (maxHP) params.set('maxHP', maxHP);

  const results = $('rec-results');
  results.innerHTML = `<div class="loading" style="grid-column:1/-1">CALCULATING</div>`;

  try {
    const recs = await apiFetch(`/api/recommendations?${params}`);
    if (!recs || recs.length === 0) {
      results.innerHTML = `<p style="font-family:var(--font-mono);color:var(--text-dim);padding:2rem">No recommendations found</p>`;
      return;
    }

    results.innerHTML = recs.map((r, i) => `
      <div class="rec-card" data-id="${r.car.id}" style="animation-delay:${i * 0.1}s">
        <div class="rec-rank">0${i+1}</div>
        <div class="rec-name">${r.car.name}</div>
        <div class="rec-reason">${r.reason}</div>
        <div class="rec-hp">${r.car.specifications.horsepower} <span>HP · ${r.car.year}</span></div>
      </div>
    `).join('');

    results.querySelectorAll('.rec-card').forEach(card => {
      card.addEventListener('click', () => openModal(parseInt(card.dataset.id)));
    });
  } catch(e) {
    results.innerHTML = `<p style="color:var(--accent);font-family:var(--font-mono)">Error loading recommendations</p>`;
  }
});

// ── Init ───────────────────────────────────────────────────────────────────
init();
