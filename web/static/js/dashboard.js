/* FinançasLua — Dashboard JS com CRUD completo */
const API = window.location.origin + '/api';

const CAT_COLOR = {
  'alimentação':'#f87171','transporte':'#fb923c','saúde':'#34d399',
  'lazer':'#60a5fa','educação':'#a78bfa','moradia':'#f5a623','outros':'#94a3b8'
};
const CAT_ICON = {
  'alimentação':'🍽️','transporte':'🚗','saúde':'💊',
  'lazer':'🎮','educação':'📚','moradia':'🏠','outros':'📦'
};
const BUDGET_MAP = {
  'alimentação':1200,'transporte':400,'saúde':300,
  'lazer':500,'educação':600,'moradia':2500,'outros':300
};

let charts = {};
let state = { expenses:[], goals:[], budget:{income:8500,budgets:{}}, calEvents:[], calYear:new Date().getFullYear(), calMonth:new Date().getMonth() };

// ── INIT ──────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
  document.getElementById('page-date').textContent =
    new Date().toLocaleDateString('pt-BR',{weekday:'long',year:'numeric',month:'long',day:'numeric'});
  loadDashboard();
});

// ── NAVIGATION ────────────────────────────────
function switchTab(tab, el) {
  document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
  document.getElementById('tab-'+tab).classList.add('active');
  if (el) el.classList.add('active');
  const titles = {overview:'Visão Geral',expenses:'Gastos',goals:'Metas',calendar:'Calendário',forecast:'Previsão de Gastos',lua:'Scripts Lua'};
  document.getElementById('page-title').textContent = titles[tab]||tab;
  if (tab==='expenses') renderAllExpenses();
  if (tab==='goals')    renderGoals();
  if (tab==='calendar') loadCalendar();
  if (tab==='forecast' && state.budget) renderForecast();
  if (tab==='lua')      loadLuaScripts();
}

// ── LOAD DATA ─────────────────────────────────
async function loadDashboard(btn) {
  if (btn) { btn.classList.add('spinning'); setTimeout(()=>btn.classList.remove('spinning'),650); }
  try {
    const [dash, bgt] = await Promise.all([
      fetch(`${API}/dashboard`).then(r=>r.json()),
      fetch(`${API}/budget`).then(r=>r.json()),
    ]);
    state.budget = bgt;
    state.goals  = dash.goals || [];
    state.expenses = [];

    renderKPIs(dash);
    renderDonut(dash.expenses_by_category);
    renderBar(dash.monthly_trend);
    renderBudget(dash.expenses_by_category);
    renderRecentTx(dash.recent_expenses);
    document.getElementById('insight-text').textContent = dash.lua_insights || '—';

    // also refresh active tab
    const active = document.querySelector('.tab.active')?.id?.replace('tab-','');
    if (active==='goals') renderGoals();
  } catch(e) { toast('Erro ao carregar dados','error'); console.error(e); }
}

// ── FORMATTERS ────────────────────────────────
const fmtBRL = n => new Intl.NumberFormat('pt-BR',{style:'currency',currency:'BRL'}).format(n||0);
const fmtDate = iso => new Date(iso).toLocaleDateString('pt-BR',{day:'2-digit',month:'2-digit'});
const fmtDateInput = iso => { const d=new Date(iso); return d.toISOString().slice(0,10); };

// ── RENDER: KPIs ──────────────────────────────
function renderKPIs(d) {
  document.getElementById('kpi-income').textContent  = fmtBRL(d.total_income);
  document.getElementById('kpi-expense').textContent = fmtBRL(d.total_expenses);
  document.getElementById('kpi-balance').textContent = fmtBRL(d.balance);
  document.getElementById('kpi-goals-count').textContent = (d.goals||[]).length;
  const pct = d.total_income>0 ? (d.total_expenses/d.total_income*100).toFixed(1):0;
  document.getElementById('kpi-expense-pct').textContent = `${pct}% da renda`;
  const sp = d.total_income>0 ? (d.balance/d.total_income*100).toFixed(1):0;
  document.getElementById('kpi-savings-pct').textContent = `${sp}% poupança`;
  const done = (d.goals||[]).filter(g=>g.current_amount>=g.target_amount).length;
  document.getElementById('kpi-goals-sub').textContent = done>0?`${done} concluída(s)`:'ativas';
}

// ── RENDER: DONUT ─────────────────────────────
function renderDonut(cats) {
  const labels=Object.keys(cats||{}), vals=labels.map(c=>cats[c]), colors=labels.map(c=>CAT_COLOR[c]||'#94a3b8');
  const total=vals.reduce((a,b)=>a+b,0);
  document.getElementById('donut-total').textContent=fmtBRL(total);
  if(charts.donut) charts.donut.destroy();
  charts.donut=new Chart(document.getElementById('chartDonut').getContext('2d'),{
    type:'doughnut',
    data:{labels,datasets:[{data:vals,backgroundColor:colors,borderColor:'#18181d',borderWidth:3,hoverOffset:5}]},
    options:{cutout:'66%',plugins:{legend:{display:false},tooltip:{callbacks:{label:c=>` ${fmtBRL(c.raw)} (${(c.raw/total*100).toFixed(1)}%)`}}}}
  });
  document.getElementById('donut-legend').innerHTML=labels.map((l,i)=>`
    <div class="legend-item"><div class="legend-dot" style="background:${colors[i]}"></div>${l} ${fmtBRL(vals[i])}</div>`).join('');
}

// ── RENDER: BAR ───────────────────────────────
function renderBar(trend) {
  if(!trend?.length) return;
  if(charts.bar) charts.bar.destroy();
  charts.bar=new Chart(document.getElementById('chartBar').getContext('2d'),{
    type:'bar',
    data:{labels:trend.map(t=>t.month),datasets:[
      {label:'Gastos',data:trend.map(t=>t.actual||0),backgroundColor:'rgba(248,113,113,.72)',borderColor:'#f87171',borderWidth:1,borderRadius:4},
      {label:'Renda',data:trend.map(t=>t.income||0),type:'line',borderColor:'#34d399',backgroundColor:'rgba(52,211,153,.07)',borderWidth:2,pointRadius:3,tension:.4,fill:true},
      {label:'Previsão ▲',data:trend.map(t=>t.predicted||0),backgroundColor:'rgba(245,166,35,.65)',borderColor:'#f5a623',borderWidth:1,borderRadius:4},
    ]},
    options:{responsive:true,plugins:{legend:{labels:{color:'#a09590',font:{family:'DM Mono',size:11},boxWidth:11}},tooltip:{callbacks:{label:c=>` ${fmtBRL(c.raw)}`}}},
      scales:{x:{grid:{color:'rgba(255,255,255,.04)'},ticks:{color:'#5a5450',font:{family:'DM Mono',size:10}}},
              y:{grid:{color:'rgba(255,255,255,.04)'},ticks:{color:'#5a5450',font:{family:'DM Mono',size:10},callback:v=>fmtBRL(v)}}}}
  });
}

// ── RENDER: BUDGET ────────────────────────────
function renderBudget(expenses) {
  const bdg = state.budget?.budgets || BUDGET_MAP;
  document.getElementById('budget-grid').innerHTML=Object.entries(bdg).map(([cat,lim])=>{
    const spent=expenses?.[cat]||0, pct=Math.min(spent/lim*100,100);
    const cls=pct>=90?'danger':pct>=70?'warn':'ok';
    return `<div class="budget-item">
      <div class="budget-meta"><span class="budget-cat">${CAT_ICON[cat]||''} ${cat}</span><span class="budget-pct ${cls}">${pct.toFixed(0)}%</span></div>
      <div class="budget-bar-bg"><div class="budget-bar-fill ${cls}" style="width:${pct}%"></div></div>
      <div class="budget-amounts"><span>${fmtBRL(spent)}</span><span>de ${fmtBRL(lim)}</span></div>
    </div>`;
  }).join('');
}

// ── RENDER: RECENT TX ─────────────────────────
function renderRecentTx(expenses) {
  document.getElementById('tx-list-recent').innerHTML=(expenses||[]).slice(0,10).map(e=>txHTML(e,false)).join('')||emptyMsg();
}

// ── RENDER: ALL EXPENSES ──────────────────────
async function renderAllExpenses() {
  try {
    const data = await fetch(`${API}/expenses`).then(r=>r.json());
    state.expenses = data;
  } catch { state.expenses = []; }
  filterExpenses();
}

function filterExpenses() {
  const q   = document.getElementById('filter-search').value.toLowerCase();
  const cat = document.getElementById('filter-cat').value;
  let list  = state.expenses;
  if(q)   list=list.filter(e=>e.description.toLowerCase().includes(q)||e.category.includes(q));
  if(cat) list=list.filter(e=>e.category===cat);
  document.getElementById('tx-list-all').innerHTML=list.map(e=>txHTML(e,true)).join('')||emptyMsg('Nenhum gasto encontrado.');
}

function txHTML(e, withActions) {
  const color=CAT_COLOR[e.category]||'#94a3b8', icon=CAT_ICON[e.category]||'📦';
  const actions = withActions ? `
    <div class="tx-actions">
      <button class="act-btn edit" onclick="openExpenseModal(${JSON.stringify(e).replace(/"/g,'&quot;')})" title="Editar">✎</button>
      <button class="act-btn del"  onclick="deleteExpense(${e.id})" title="Excluir">✕</button>
    </div>` : '';
  return `<div class="tx-item" id="tx-${e.id}">
    <div class="tx-icon" style="background:${color}22">${icon}</div>
    <div class="tx-info">
      <div class="tx-desc">${e.description}</div>
      <div class="tx-sub">${e.category}${e.is_recurring?' · recorrente':''}</div>
    </div>
    <div class="tx-right">
      <div class="tx-amount">${fmtBRL(e.amount)}</div>
      <div class="tx-date">${fmtDate(e.date)}</div>
      ${actions}
    </div>
  </div>`;
}

function emptyMsg(msg='Sem transações.') {
  return `<p style="color:var(--tm);font-size:.85rem;padding:.5rem 0">${msg}</p>`;
}

// ── RENDER: GOALS ─────────────────────────────
function renderGoals() {
  document.getElementById('goals-grid').innerHTML=(state.goals||[]).map(g=>{
    const pct=Math.min(g.current_amount/g.target_amount*100,100);
    const dl=new Date(g.deadline).toLocaleDateString('pt-BR',{month:'short',year:'numeric'});
    return `<div class="goal-card">
      <div class="goal-actions">
        <button class="act-btn edit" onclick='openGoalModal(${JSON.stringify(g)})' title="Editar">✎</button>
        <button class="act-btn del"  onclick="deleteGoal(${g.id})" title="Excluir">✕</button>
        <button class="act-btn edit" onclick="depositGoal(${g.id}, '${g.name}')" title="Depositar">+</button>
      </div>
      <div class="goal-top">
        <div class="goal-icon" style="border:1px solid ${g.color}33">${g.icon}</div>
        <div><div class="goal-name">${g.name}</div><div class="goal-dl">📅 ${dl}</div></div>
      </div>
      <div class="goal-amounts">
        <div class="goal-curr" style="color:${g.color}">${fmtBRL(g.current_amount)}</div>
        <div class="goal-tgt">meta: ${fmtBRL(g.target_amount)}</div>
      </div>
      <div class="goal-bar-bg"><div class="goal-bar-fill" style="width:${pct}%;background:${g.color}"></div></div>
      <div class="goal-pct">${pct.toFixed(1)}%</div>
    </div>`;
  }).join('') || '<p style="color:var(--tm);padding:1rem 0">Nenhuma meta cadastrada. Crie uma!</p>';
}

// ── RENDER: FORECAST ──────────────────────────
async function renderForecast() {
  const dash = await fetch(`${API}/dashboard`).then(r=>r.json()).catch(()=>({}));
  const trend = dash.monthly_trend || [];
  if(!trend.length) return;
  if(charts.fc) charts.fc.destroy();
  charts.fc=new Chart(document.getElementById('chartForecast').getContext('2d'),{
    type:'line',
    data:{labels:trend.map(t=>t.month),datasets:[
      {label:'Realizado',data:trend.map(t=>t.actual||0),borderColor:'#f87171',backgroundColor:'rgba(248,113,113,.1)',borderWidth:2.5,pointRadius:5,tension:.4,fill:true},
      {label:'Previsão Lua',data:trend.map(t=>t.predicted||0),borderColor:'#f5a623',backgroundColor:'rgba(245,166,35,.06)',borderWidth:2,borderDash:[6,4],pointRadius:6,pointStyle:'star',tension:.3},
      {label:'Renda',data:trend.map(t=>t.income||0),borderColor:'#34d399',borderWidth:1.5,pointRadius:0,tension:.4,borderDash:[3,3]},
    ]},
    options:{responsive:true,plugins:{legend:{labels:{color:'#a09590',font:{family:'DM Mono',size:11},boxWidth:13}},tooltip:{callbacks:{label:c=>` ${fmtBRL(c.raw)}`}}},
      scales:{x:{grid:{color:'rgba(255,255,255,.04)'},ticks:{color:'#5a5450',font:{family:'DM Mono',size:10}}},
              y:{grid:{color:'rgba(255,255,255,.04)'},ticks:{color:'#5a5450',font:{family:'DM Mono',size:10},callback:v=>fmtBRL(v)}}}}
  });
  document.getElementById('forecast-cards').innerHTML=trend.map(t=>{
    const isPred=t.predicted>0&&t.actual===0;
    return `<div class="fc ${isPred?'predicted':''}">
      <div class="fc-month">${t.month}</div>
      <div class="fc-amount">${fmtBRL(isPred?t.predicted:t.actual)}</div>
      <div class="fc-label">${isPred?'⬡ previsão':'realizado'}</div>
    </div>`;
  }).join('');
}

async function depositGoal(id, name) {
  const val = prompt(`Depositar na meta "${name}"\nValor (R$):`);
  if (!val) return;
  const amount = parseFloat(val.replace(',','.'));
  if (isNaN(amount) || amount <= 0) { toast('Valor inválido', 'error'); return; }
  try {
    await fetch(`${API}/goals/deposit/${id}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ amount })
    });
    toast(`+${fmtBRL(amount)} depositado ✓`, 'success');
    const r = await fetch(`${API}/goals`).then(r => r.json());
    state.goals = r;
    renderGoals();
    loadDashboard();
  } catch { toast('Erro ao depositar', 'error'); }
}

// ── CALENDAR ──────────────────────────────────
async function loadCalendar() {
  try { state.calEvents = await fetch(`${API}/calendar`).then(r=>r.json()); }
  catch { state.calEvents = []; }
  renderCalendar();
}

function renderCalendar() {
  const {calYear:y, calMonth:m, calEvents:evts} = state;
  document.getElementById('cal-month-label').textContent =
    new Date(y,m,1).toLocaleDateString('pt-BR',{month:'long',year:'numeric'});
  const firstDay=new Date(y,m,1).getDay(), daysInMonth=new Date(y,m+1,0).getDate();
  const now=new Date();
  const byDay={};
  (evts||[]).forEach(e=>{ const d=new Date(e.date).getDate(); (byDay[d]=byDay[d]||[]).push(e); });
  let html='';
  for(let i=0;i<firstDay;i++) html+='<div class="cal-day empty"></div>';
  for(let d=1;d<=daysInMonth;d++){
    const isToday=d===now.getDate()&&m===now.getMonth()&&y===now.getFullYear();
    const dayEvts=byDay[d]||[], total=dayEvts.reduce((a,e)=>a+(e.amount||0),0);
    const cls=['cal-day',isToday?'today':'',dayEvts.some(e=>e.type==='expense')?'has-exp':''].filter(Boolean).join(' ');
    html+=`<div class="${cls}" onclick="showDayEvents(${d})">
      <div class="cal-day-num">${d}</div>
      ${total>0?`<div class="cal-day-amt">${fmtBRL(total)}</div>`:''}
    </div>`;
  }
  document.getElementById('cal-grid').innerHTML=html;
  document.getElementById('cal-events').innerHTML=
    (evts||[]).sort((a,b)=>new Date(a.date)-new Date(b.date)).slice(0,10).map(e=>`
    <div class="cal-evt ${e.type==='goal'?'goal':''}">
      <div class="cal-evt-title">${CAT_ICON[e.category]||'💳'} ${e.title}</div>
      <div class="cal-evt-date">${fmtDate(e.date)}</div>
      ${e.amount?`<div class="cal-evt-amt">${fmtBRL(e.amount)}</div>`:''}
    </div>`).join('');
}

function showDayEvents(day) {
  const evts=(state.calEvents||[]).filter(e=>new Date(e.date).getDate()===day);
  if(!evts.length) return;
  document.getElementById('cal-events').innerHTML=
    `<div style="font-family:var(--ff-mono);font-size:.72rem;color:var(--gold);margin-bottom:.6rem">Dia ${day}</div>`+
    evts.map(e=>`<div class="cal-evt"><div class="cal-evt-title">${e.title}</div>${e.amount?`<div class="cal-evt-amt">${fmtBRL(e.amount)}</div>`:''}</div>`).join('');
}

function changeMonth(dir) {
  state.calMonth+=dir;
  if(state.calMonth>11){state.calMonth=0;state.calYear++;}
  if(state.calMonth<0) {state.calMonth=11;state.calYear--;}
  renderCalendar();
}

// ── LUA SCRIPTS ───────────────────────────────
async function loadLuaScripts() {
  try {
    const s=await fetch(`${API}/lua-scripts`).then(r=>r.json());
    document.getElementById('lua-scripts').innerHTML=Object.entries(s).map(([n,src])=>`
      <div><div class="lua-block-name">${n}.lua</div><div class="lua-code">${esc(src.trim())}</div></div>`).join('');
  } catch { document.getElementById('lua-scripts').innerHTML='<p style="color:var(--tm)">Erro ao carregar.</p>'; }
}
const esc = s => s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');

// ══ EXPENSE MODAL ════════════════════════════
function openExpenseModal(exp) {
  document.getElementById('modal-expense-title').textContent = exp ? 'Editar Gasto' : 'Novo Gasto';
  document.getElementById('exp-id').value      = exp?.id || '';
  document.getElementById('exp-desc').value    = exp?.description || '';
  document.getElementById('exp-amount').value  = exp?.amount || '';
  document.getElementById('exp-cat').value     = exp?.category || 'alimentação';
  document.getElementById('exp-date').value    = exp?.date ? fmtDateInput(exp.date) : new Date().toISOString().slice(0,10);
  document.getElementById('exp-recurring').checked = exp?.is_recurring || false;
  openModal('modal-expense');
}

async function saveExpense() {
  const id     = document.getElementById('exp-id').value;
  const desc   = document.getElementById('exp-desc').value.trim();
  const amount = parseFloat(document.getElementById('exp-amount').value);
  const cat    = document.getElementById('exp-cat').value;
  const date   = document.getElementById('exp-date').value;
  const recur  = document.getElementById('exp-recurring').checked;
  if(!desc||isNaN(amount)||amount<=0) { toast('Preencha descrição e valor','error'); return; }

  const body = {description:desc,amount,category:cat,date:new Date(date+'T12:00:00').toISOString(),is_recurring:recur};
  try {
    if(id) {
      body.id=parseInt(id);
      await fetch(`${API}/expenses/${id}`,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
      toast('Gasto atualizado ✓','success');
    } else {
      await fetch(`${API}/expenses`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
      toast('Gasto adicionado ✓','success');
    }
    closeModal('modal-expense');
    await loadDashboard();
    if(document.getElementById('tab-expenses').classList.contains('active')) renderAllExpenses();
  } catch(e) { toast('Erro ao salvar','error'); }
}

async function deleteExpense(id) {
  if(!confirm('Excluir este gasto?')) return;
  try {
    await fetch(`${API}/expenses/${id}`,{method:'DELETE'});
    toast('Gasto excluído','success');
    await loadDashboard();
    renderAllExpenses();
  } catch { toast('Erro ao excluir','error'); }
}

// ══ GOAL MODAL ═══════════════════════════════
function openGoalModal(goal) {
  document.getElementById('modal-goal-title').textContent = goal ? 'Editar Meta' : 'Nova Meta';
  document.getElementById('goal-id').value      = goal?.id || '';
  document.getElementById('goal-name').value    = goal?.name || '';
  document.getElementById('goal-target').value  = goal?.target_amount || '';
  document.getElementById('goal-current').value = goal?.current_amount || '';
  document.getElementById('goal-deadline').value= goal?.deadline ? fmtDateInput(goal.deadline) : '';
  // icon & color
  const icon  = goal?.icon  || '💰';
  const color = goal?.color || '#4ade80';
  document.getElementById('goal-icon').value  = icon;
  document.getElementById('goal-color').value = color;
  document.querySelectorAll('.emoji-picker span').forEach(s=>s.classList.toggle('selected',s.textContent===icon));
  document.querySelectorAll('.color-swatch').forEach(s=>s.classList.toggle('selected',s.style.background===color||s.style.backgroundColor===color));
  openModal('modal-goal');
}

async function saveGoal() {
  const id      = document.getElementById('goal-id').value;
  const name    = document.getElementById('goal-name').value.trim();
  const target  = parseFloat(document.getElementById('goal-target').value);
  const current = parseFloat(document.getElementById('goal-current').value)||0;
  const dl      = document.getElementById('goal-deadline').value;
  const icon    = document.getElementById('goal-icon').value||'💰';
  const color   = document.getElementById('goal-color').value||'#4ade80';
  if(!name||isNaN(target)||target<=0) { toast('Preencha nome e valor alvo','error'); return; }
  const body={name,target_amount:target,current_amount:current,icon,color,deadline:dl?new Date(dl+'T12:00:00').toISOString():new Date(Date.now()+365*86400000).toISOString()};
  try {
    if(id) {
      body.id=parseInt(id);
      await fetch(`${API}/goals/${id}`,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
      toast('Meta atualizada ✓','success');
    } else {
      await fetch(`${API}/goals`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
      toast('Meta criada ✓','success');
    }
    closeModal('modal-goal');
    const r=await fetch(`${API}/goals`).then(r=>r.json());
    state.goals=r;
    renderGoals();
    loadDashboard();
  } catch { toast('Erro ao salvar meta','error'); }
}

async function deleteGoal(id) {
  if(!confirm('Excluir esta meta?')) return;
  try {
    await fetch(`${API}/goals/${id}`,{method:'DELETE'});
    toast('Meta excluída','success');
    const r=await fetch(`${API}/goals`).then(r=>r.json());
    state.goals=r;
    renderGoals();
    loadDashboard();
  } catch { toast('Erro ao excluir','error'); }
}

// ══ BUDGET MODAL ═════════════════════════════
async function openBudgetModal() {
  try { state.budget = await fetch(`${API}/budget`).then(r=>r.json()); } catch{}
  const b=state.budget;
  document.getElementById('budget-income').value=b.income||'';
  const cats=['alimentação','transporte','saúde','lazer','educação','moradia','outros'];
  cats.forEach(c=>{ const el=document.getElementById('b-'+c); if(el) el.value=b.budgets?.[c]||''; });
  openModal('modal-budget');
}

async function saveBudget() {
  const income=parseFloat(document.getElementById('budget-income').value);
  if(isNaN(income)||income<0){ toast('Renda inválida','error'); return; }
  const cats=['alimentação','transporte','saúde','lazer','educação','moradia','outros'];
  const budgets={};
  cats.forEach(c=>{ const v=parseFloat(document.getElementById('b-'+c)?.value)||0; budgets[c]=v; });
  const now=new Date();
  const body={month:now.getMonth()+1,year:now.getFullYear(),income,budgets};
  try {
    state.budget=await fetch(`${API}/budget`,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)}).then(r=>r.json());
    toast('Orçamento salvo ✓','success');
    closeModal('modal-budget');
    loadDashboard();
  } catch { toast('Erro ao salvar orçamento','error'); }
}

// ══ EMOJI / COLOR PICKERS ════════════════════
function pickEmoji(e) {
  document.getElementById('goal-icon').value=e;
  document.querySelectorAll('.emoji-picker span').forEach(s=>s.classList.toggle('selected',s.textContent===e));
}
function pickColor(c,el) {
  document.getElementById('goal-color').value=c;
  document.querySelectorAll('.color-swatch').forEach(s=>s.classList.remove('selected'));
  el.classList.add('selected');
}

// ══ MODAL UTILS ══════════════════════════════
function openModal(id)  { document.getElementById(id).classList.add('open'); }
function closeModal(id) { document.getElementById(id).classList.remove('open'); }

// ── TOAST ─────────────────────────────────────
let toastTimer;
function toast(msg, type='success') {
  const el=document.getElementById('toast');
  el.textContent=msg; el.className='toast show '+(type||'');
  clearTimeout(toastTimer);
  toastTimer=setTimeout(()=>el.classList.remove('show'),2800);
}
