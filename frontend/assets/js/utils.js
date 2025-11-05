
const byId = (id)=>document.getElementById(id);
const val = (id)=>byId(id).value;
const qs = (sel)=>document.querySelector(sel);
const qsa = (sel)=>Array.from(document.querySelectorAll(sel));
const escapeHtml = (str)=>str.replace(/[&<>"']/g, m => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[m]));

function addDefaultNav(targetId){
  const wrap = byId(targetId);
  if(!wrap) return;
  const email = localStorage.getItem('demo_email') || 'admin@example.com';
  wrap.innerHTML = `<span class="badge">${email}</span> <button class="btn ghost" id="btn-logout">Logout</button>`;
  const btn = byId('btn-logout'); if(btn){ btn.onclick = ()=>{ Auth.logout(); location.href='../pages/login.html'; }; }
}
