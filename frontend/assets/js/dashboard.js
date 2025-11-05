
(async function(){
  // nav
  const nav = document.getElementById('nav-actions');
  if(nav){ addDefaultNav('nav-actions'); }
  // kpis
  const stats = await API.getStats();
  const safe = (id,val)=>{ const el=document.getElementById(id); if(el) el.textContent = String(val); };
  safe('kpi-open', stats.open);
  safe('kpi-risk', stats.risk);
  safe('kpi-resolved', stats.resolved7d);

  const q = document.getElementById('q');
  const fs = document.getElementById('filter-status');
  async function render(){
    const list = await API.listTickets({q:q.value, status:fs.value});
    const tbody = document.getElementById('tickets-body'); tbody.innerHTML='';
    list.forEach(t => {
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td>${t.id}</td>
        <td><a href="./pages/ticket-detail.html?id=${t.id}">${escapeHtml(t.title)}</a></td>
        <td>${t.priority}</td>
        <td><span class="status ${t.status.toUpperCase().replace(' ','_')}">${t.status}</span></td>
        <td>${t.assignee||''}</td>
        <td>${new Date(t.updatedAt).toLocaleString()}</td>
      `;
      tbody.appendChild(tr);
    });
  }
  q && q.addEventListener('input', render);
  fs && fs.addEventListener('change', render);
  render();
})();
