/* ============================================================
   FinançasLua Dashboard — Frontend JS
   Consumes Go REST API, renders Chart.js visualizations
   ============================================================ */

const API = window.location.origin + '/api';

const CAT_COLORS = {
  'alimentação': '#f87171',
  'transporte':  '#fb923c',
  'saúde':       '#34d399',
  'lazer':       '#60a5fa',
  'educação':    '#a78bfa',
  'moradia':     '#f5a623',
  'outros':      '#94a3b8',
};

const CAT_ICONS = {
  'alimentação': '🍽️',
  'transporte':  '🚗',
  'saúde':       '💊',
  'lazer':       '🎮',
  'educação':    '📚',
  'moradia':     '🏠',
  'outros':      '📦',
};

let charts = {};
let calendarData = { events: [], year: new Date().getFullYear(), month: new Date().getMonth() };
let dashboardData = null;

// ── INIT ─────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
  setPageDate();
  loadDashboard();
});

function setPageDate() {
  const now = new Date();
  const opts = { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' };
  document.getElementById('page-date').textContent =
    now.toLocaleDateString('pt-BR', opts);
}

// ── TAB NAVIGATION ───────────────────────────
function switchTab(tab, el) {
  document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
  document.getElementById('tab-' + tab).classList.add('active');
  if (el) el.classList.add('active');

  const titles = {
    overview: 'Visão Geral', expenses: 'Gastos',
    goals: 'Metas', calendar: 'Calendário',
    forecast: 'Previsão de Gastos', lua: 'Scripts Lua'
  };
  document.getElementById('page-title').textContent = titles[tab] || tab;

  if (tab === 'expenses' && dashboardData) renderAllExpenses(dashboardData.recent_expenses);
  if (tab === 'goals'    && dashboardData) renderGoals(dashboardData.goals);
  if (tab === 'calendar') loadCalendar();
  if (tab === 'forecast' && dashboardData) renderForecast(dashboardData.monthly_trend);
  if (tab === 'lua') loadLuaScripts();
}

// ── DATA LOADING ─────────────────────────────
async function loadDashboard() {
  try {
    const res = await fetch(`${API}/dashboard`);
    const data = await res.json();
    dashboardData = data;

    renderKPIs(data);
    renderDonut(data.expenses_by_category);
    renderBarChart(data.monthly_trend);
    renderBudget(data.expenses_by_category, data.budget_usage);
    renderRecentTx(data.recent_expenses);
    renderInsight(data.lua_insights);
    renderGoals(data.goals);
  } catch (e) {
    console.error('Dashboard error:', e);
    document.getElementById('insight-text').textContent = 'Erro ao carregar dados da API Go.';
  }
}

async function loadCalendar() {
  try {
    const res = await fetch(`${API}/calendar`);
    calendarData.events = await res.json();
  } catch { calendarData.events = []; }
  renderCalendar();
}

async function loadLuaScripts() {
  try {
    const res = await fetch(`${API}/lua-scripts`);
    const scripts = await res.json();
    renderLuaScripts(scripts);
  } catch (e) {
    document.getElementById('lua-scripts').innerHTML =
      '<p style="color:var(--text-muted)">Erro ao carregar scripts Lua.</p>';
  }
}

// ── FORMATTERS ───────────────────────────────
function fmtBRL(n) {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(n || 0);
}

function fmtDate(iso) {
  return new Date(iso).toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit' });
}

// ── RENDER: KPIs ─────────────────────────────
function renderKPIs(data) {
  document.getElementById('kpi-income').textContent  = fmtBRL(data.total_income);
  document.getElementById('kpi-expense').textContent = fmtBRL(data.total_expenses);
  document.getElementById('kpi-balance').textContent = fmtBRL(data.balance);
  document.getElementById('kpi-goals').textContent   = (data.goals || []).length;

  const pct = data.total_income > 0 ? (data.total_expenses / data.total_income * 100).toFixed(1) : 0;
  document.getElementById('kpi-expense-pct').textContent = `${pct}% da renda`;

  const savingsPct = data.total_income > 0 ? ((data.balance / data.total_income) * 100).toFixed(1) : 0;
  document.getElementById('kpi-savings').textContent = `${savingsPct}% de poupança`;

  const goalsComplete = (data.goals || []).filter(g => g.current_amount >= g.target_amount).length;
  document.getElementById('kpi-goals-sub').textContent =
    goalsComplete > 0 ? `${goalsComplete} concluída(s)` : 'em andamento';
}

// ── RENDER: INSIGHT BADGE ────────────────────
function renderInsight(text) {
  document.getElementById('insight-text').textContent = text || '—';
}

// ── RENDER: DONUT CHART ──────────────────────
function renderDonut(expByCategory) {
  const cats = Object.keys(expByCategory || {});
  const vals = cats.map(c => expByCategory[c]);
  const colors = cats.map(c => CAT_COLORS[c] || '#94a3b8');
  const total = vals.reduce((a, b) => a + b, 0);

  document.getElementById('donut-total').textContent = fmtBRL(total);

  if (charts.donut) charts.donut.destroy();

  const ctx = document.getElementById('chartDonut').getContext('2d');
  charts.donut = new Chart(ctx, {
    type: 'doughnut',
    data: {
      labels: cats,
      datasets: [{ data: vals, backgroundColor: colors, borderColor: '#1a1a1f', borderWidth: 3, hoverOffset: 6 }]
    },
    options: {
      cutout: '68%',
      plugins: {
        legend: { display: false },
        tooltip: {
          callbacks: {
            label: ctx => ` ${fmtBRL(ctx.raw)} (${(ctx.raw/total*100).toFixed(1)}%)`
          }
        }
      }
    }
  });

  // Legend
  const leg = document.getElementById('donut-legend');
  leg.innerHTML = cats.map((c, i) => `
    <div class="legend-item">
      <div class="legend-dot" style="background:${colors[i]}"></div>
      <span>${c} — ${fmtBRL(vals[i])}</span>
    </div>`).join('');
}

// ── RENDER: BAR CHART ────────────────────────
function renderBarChart(trend) {
  if (!trend || !trend.length) return;

  const labels = trend.map(t => t.month);
  const actual    = trend.map(t => t.actual || 0);
  const income    = trend.map(t => t.income || 0);
  const predicted = trend.map(t => t.predicted || 0);

  if (charts.bar) charts.bar.destroy();

  const ctx = document.getElementById('chartBar').getContext('2d');
  charts.bar = new Chart(ctx, {
    type: 'bar',
    data: {
      labels,
      datasets: [
        {
          label: 'Gastos', data: actual,
          backgroundColor: 'rgba(248,113,113,0.7)',
          borderColor: '#f87171', borderWidth: 1, borderRadius: 4,
        },
        {
          label: 'Renda', data: income,
          type: 'line',
          borderColor: '#34d399', backgroundColor: 'rgba(52,211,153,0.08)',
          borderWidth: 2, pointRadius: 3, tension: 0.4, fill: true,
        },
        {
          label: 'Previsão (Lua)', data: predicted,
          backgroundColor: 'rgba(245,166,35,0.6)',
          borderColor: '#f5a623', borderWidth: 1, borderRadius: 4,
          borderDash: [4,4],
        },
      ]
    },
    options: {
      responsive: true,
      plugins: {
        legend: {
          labels: { color: '#b0a898', font: { family: 'DM Mono', size: 11 }, boxWidth: 12 }
        },
        tooltip: { callbacks: { label: ctx => ` ${fmtBRL(ctx.raw)}` } }
      },
      scales: {
        x: { grid: { color: 'rgba(255,255,255,0.04)' }, ticks: { color: '#6b6460', font: { family: 'DM Mono', size: 10 } } },
        y: { grid: { color: 'rgba(255,255,255,0.04)' }, ticks: { color: '#6b6460', font: { family: 'DM Mono', size: 10 }, callback: v => fmtBRL(v) } }
      }
    }
  });
}

// ── RENDER: BUDGET ───────────────────────────
function renderBudget(expenses, usage) {
  const grid = document.getElementById('budget-grid');
  const budgetMap = {
    'alimentação': 1200, 'transporte': 400, 'saúde': 300,
    'lazer': 500, 'educação': 600, 'moradia': 2500, 'outros': 300
  };

  grid.innerHTML = Object.entries(budgetMap).map(([cat, budget]) => {
    const spent = expenses[cat] || 0;
    const pct   = Math.min((spent / budget) * 100, 100);
    const cls   = pct >= 90 ? 'danger' : pct >= 70 ? 'warning' : 'ok';

    return `
    <div class="budget-item">
      <div class="budget-meta">
        <span class="budget-cat">${CAT_ICONS[cat] || ''} ${cat}</span>
        <span class="budget-pct ${cls}">${pct.toFixed(0)}%</span>
      </div>
      <div class="budget-bar-bg">
        <div class="budget-bar-fill ${cls}" style="width:${pct}%"></div>
      </div>
      <div class="budget-amounts">
        <span>${fmtBRL(spent)}</span>
        <span>de ${fmtBRL(budget)}</span>
      </div>
    </div>`;
  }).join('');
}

// ── RENDER: RECENT TRANSACTIONS ──────────────
function renderRecentTx(expenses) {
  const el = document.getElementById('tx-list');
  if (!expenses || !expenses.length) { el.innerHTML = '<p style="color:var(--text-muted)">Sem transações.</p>'; return; }
  el.innerHTML = expenses.slice(0, 10).map(e => txItem(e)).join('');
}

function renderAllExpenses(expenses) {
  const el = document.getElementById('all-expenses');
  if (!expenses || !expenses.length) { el.innerHTML = '<p style="color:var(--text-muted)">Sem transações.</p>'; return; }

  // Load from API
  fetch(`${API}/expenses`)
    .then(r => r.json())
    .then(data => { el.innerHTML = data.map(e => txItem(e)).join(''); })
    .catch(() => { el.innerHTML = expenses.map(e => txItem(e)).join(''); });
}

function txItem(e) {
  const color = CAT_COLORS[e.category] || '#94a3b8';
  const icon  = CAT_ICONS[e.category]  || '📦';
  return `
  <div class="tx-item">
    <div class="tx-icon" style="background:${color}22">${icon}</div>
    <div class="tx-desc">
      <div class="tx-desc-main">${e.description}</div>
      <div class="tx-desc-sub">${e.category}${e.is_recurring ? ' · recorrente' : ''}</div>
    </div>
    <div class="tx-amount">${fmtBRL(e.amount)}</div>
    <div class="tx-date">${fmtDate(e.date)}</div>
  </div>`;
}

// ── RENDER: GOALS ─────────────────────────────
function renderGoals(goals) {
  const grid = document.getElementById('goals-grid');
  if (!goals || !goals.length) { grid.innerHTML = '<p>Sem metas.</p>'; return; }
  grid.innerHTML = goals.map(g => {
    const pct = Math.min((g.current_amount / g.target_amount) * 100, 100);
    const deadline = new Date(g.deadline).toLocaleDateString('pt-BR', { month: 'short', year: 'numeric' });
    return `
    <div class="goal-card">
      <div class="goal-top">
        <div class="goal-icon-wrap" style="border: 1px solid ${g.color}33">${g.icon}</div>
        <div class="goal-info">
          <div class="goal-name">${g.name}</div>
          <div class="goal-deadline">📅 Prazo: ${deadline}</div>
        </div>
      </div>
      <div class="goal-amounts">
        <div class="goal-current" style="color:${g.color}">${fmtBRL(g.current_amount)}</div>
        <div class="goal-target">meta: ${fmtBRL(g.target_amount)}</div>
      </div>
      <div class="goal-bar-bg">
        <div class="goal-bar-fill" style="width:${pct}%;background:${g.color}"></div>
      </div>
      <div class="goal-pct">${pct.toFixed(1)}% concluído</div>
    </div>`;
  }).join('');
}

// ── RENDER: CALENDAR ─────────────────────────
function renderCalendar() {
  const { year, month, events } = calendarData;
  const now = new Date();

  const monthLabel = new Date(year, month, 1).toLocaleDateString('pt-BR', { month: 'long', year: 'numeric' });
  document.getElementById('cal-month-label').textContent = monthLabel;

  const firstDay = new Date(year, month, 1).getDay();
  const daysInMonth = new Date(year, month + 1, 0).getDate();

  // Group events by day
  const byDay = {};
  (events || []).forEach(e => {
    const d = new Date(e.date).getDate();
    if (!byDay[d]) byDay[d] = [];
    byDay[d].push(e);
  });

  let html = '';
  for (let i = 0; i < firstDay; i++) html += `<div class="cal-day empty"></div>`;
  for (let d = 1; d <= daysInMonth; d++) {
    const isToday = d === now.getDate() && month === now.getMonth() && year === now.getFullYear();
    const dayEvents = byDay[d] || [];
    const hasExpense = dayEvents.some(e => e.type === 'expense');
    const hasGoal    = dayEvents.some(e => e.type === 'goal');
    const dayTotal   = dayEvents.reduce((a, e) => a + (e.amount || 0), 0);

    const cls = [
      'cal-day',
      isToday ? 'today' : '',
      hasExpense ? 'has-expense' : '',
      hasGoal ? 'has-goal' : '',
    ].filter(Boolean).join(' ');

    html += `
    <div class="${cls}" onclick="showDayEvents(${d})">
      <div class="cal-day-num">${d}</div>
      ${dayTotal > 0 ? `<div class="cal-day-amt">${fmtBRL(dayTotal)}</div>` : ''}
    </div>`;
  }
  document.getElementById('cal-grid').innerHTML = html;

  // Show all events sorted
  const sorted = [...(events || [])].sort((a, b) => new Date(a.date) - new Date(b.date));
  document.getElementById('cal-events').innerHTML = sorted.slice(0, 8).map(e => `
    <div class="cal-event ${e.type === 'goal' ? 'goal' : ''}">
      <div class="cal-event-title">${CAT_ICONS[e.category] || (e.type === 'goal' ? '🎯' : '💳')} ${e.title}</div>
      <div class="cal-event-date">${fmtDate(e.date)}</div>
      ${e.amount ? `<div class="cal-event-amount">${fmtBRL(e.amount)}</div>` : ''}
    </div>`).join('');
}

function showDayEvents(day) {
  const { year, month, events } = calendarData;
  const dayEvts = (events || []).filter(e => new Date(e.date).getDate() === day);
  if (!dayEvts.length) return;
  const el = document.getElementById('cal-events');
  el.innerHTML = `<div style="font-family:var(--font-mono);font-size:.75rem;color:var(--gold);margin-bottom:.75rem">Dia ${day}</div>` +
    dayEvts.map(e => `
    <div class="cal-event ${e.type === 'goal' ? 'goal' : ''}">
      <div class="cal-event-title">${CAT_ICONS[e.category] || '💳'} ${e.title}</div>
      ${e.amount ? `<div class="cal-event-amount">${fmtBRL(e.amount)}</div>` : ''}
    </div>`).join('');
}

function changeMonth(dir) {
  calendarData.month += dir;
  if (calendarData.month > 11) { calendarData.month = 0; calendarData.year++; }
  if (calendarData.month < 0)  { calendarData.month = 11; calendarData.year--; }
  renderCalendar();
}

// ── RENDER: FORECAST ─────────────────────────
function renderForecast(trend) {
  if (!trend || !trend.length) return;

  const labels = trend.map(t => t.month);
  const actual  = trend.map(t => t.actual || 0);
  const predicted = trend.map(t => t.predicted || 0);
  const income  = trend.map(t => t.income || 0);

  if (charts.forecast) charts.forecast.destroy();

  const ctx = document.getElementById('chartForecast').getContext('2d');
  charts.forecast = new Chart(ctx, {
    type: 'line',
    data: {
      labels,
      datasets: [
        {
          label: 'Gastos Reais', data: actual,
          borderColor: '#f87171', backgroundColor: 'rgba(248,113,113,0.1)',
          borderWidth: 2.5, pointRadius: 5, tension: 0.4, fill: true,
        },
        {
          label: 'Previsão Lua', data: predicted,
          borderColor: '#f5a623', backgroundColor: 'rgba(245,166,35,0.08)',
          borderWidth: 2, borderDash: [6,4], pointRadius: 6, pointStyle: 'star',
          tension: 0.3, fill: false,
        },
        {
          label: 'Renda', data: income,
          borderColor: '#34d399', backgroundColor: 'rgba(52,211,153,0.05)',
          borderWidth: 1.5, pointRadius: 0, tension: 0.4, fill: true,
          borderDash: [3,3],
        },
      ]
    },
    options: {
      responsive: true,
      plugins: {
        legend: { labels: { color: '#b0a898', font: { family: 'DM Mono', size: 11 }, boxWidth: 14 } },
        tooltip: { callbacks: { label: ctx => ` ${fmtBRL(ctx.raw)}` } }
      },
      scales: {
        x: { grid: { color: 'rgba(255,255,255,0.04)' }, ticks: { color: '#6b6460', font: { family: 'DM Mono', size: 10 } } },
        y: { grid: { color: 'rgba(255,255,255,0.04)' }, ticks: { color: '#6b6460', font: { family: 'DM Mono', size: 10 }, callback: v => fmtBRL(v) } }
      }
    }
  });

  // Forecast cards
  const cards = document.getElementById('forecast-cards');
  cards.innerHTML = trend.map(t => {
    const isPred = t.predicted > 0 && t.actual === 0;
    const val = isPred ? t.predicted : t.actual;
    return `
    <div class="forecast-card ${isPred ? 'fc-predicted' : ''}">
      <div class="fc-month">${t.month}</div>
      <div class="fc-amount">${fmtBRL(val)}</div>
      <div class="fc-label">${isPred ? '⬡ previsão lua' : 'realizado'}</div>
    </div>`;
  }).join('');
}

// ── RENDER: LUA SCRIPTS ──────────────────────
function renderLuaScripts(scripts) {
  const el = document.getElementById('lua-scripts');
  el.innerHTML = Object.entries(scripts).map(([name, src]) => `
    <div class="lua-script-block">
      <div class="lua-script-name">${name}.lua</div>
      <div class="lua-code">${escapeHtml(src.trim())}</div>
    </div>`).join('');
}

function escapeHtml(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

// ── ADD EXPENSE ───────────────────────────────
async function addExpense() {
  const desc   = document.getElementById('exp-desc').value.trim();
  const amount = parseFloat(document.getElementById('exp-amount').value);
  const cat    = document.getElementById('exp-cat').value;

  if (!desc || isNaN(amount) || amount <= 0) {
    alert('Preencha descrição e valor corretamente.');
    return;
  }

  try {
    const res = await fetch(`${API}/expenses`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ description: desc, amount, category: cat }),
    });
    if (res.ok) {
      document.getElementById('exp-desc').value = '';
      document.getElementById('exp-amount').value = '';
      await loadDashboard();
      renderAllExpenses([]);
    }
  } catch (e) {
    console.error('Add expense error:', e);
  }
}
