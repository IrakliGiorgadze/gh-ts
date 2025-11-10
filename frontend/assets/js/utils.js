
const byId = (id)=>document.getElementById(id);
const val = (id)=>byId(id).value;
const qs = (sel)=>document.querySelector(sel);
const qsa = (sel)=>Array.from(document.querySelectorAll(sel));
const escapeHtml = (str)=>str.replace(/[&<>"']/g, m => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[m]));

async function addDefaultNav(targetId){
  const wrap = byId(targetId);
  if(!wrap) return;
  
  // Get real user from API - only show nav if authenticated
  try {
    const user = await Auth.me();
    if(user && user.email) {
      // User is authenticated - show email and logout button
      wrap.innerHTML = `<span class="badge">${escapeHtml(user.email)}</span> <button class="btn ghost" id="btn-logout">Logout</button>`;
      const btn = byId('btn-logout'); 
      if(btn){ 
        btn.onclick = async ()=>{ 
          await Auth.logout(); 
          location.href='../pages/login.html'; 
        }; 
      }
    } else {
      // Not authenticated - show login link
      wrap.innerHTML = `<a href="./pages/login.html" class="btn ghost">Login</a>`;
    }
  } catch(e) {
    // Not authenticated or API error - show login link
    console.warn('[addDefaultNav] Failed to get user:', e);
    wrap.innerHTML = `<a href="./pages/login.html" class="btn ghost">Login</a>`;
  }
}
