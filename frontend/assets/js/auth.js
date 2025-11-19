// Authentication guard - checks real API authentication state
// Note: This file works with the real backend API, not mock data

(async function(){
  // Pages that don't require authentication
  const publicPages = [
    '/login.html',
    '/pages/login.html',
    '../pages/login.html',
    '/register.html',
    '/pages/register.html',
    '../pages/register.html'
  ];
  
  const currentPath = location.pathname;
  const isPublicPage = publicPages.some(page => 
    currentPath.endsWith(page) || currentPath.includes(page)
  );
  
  // If on a public page, don't check auth
  if(isPublicPage) {
    // But if user is already logged in, redirect them away from login/register
    try {
      const user = await Auth.me();
      if(user && user.email) {
        // User is logged in, redirect to home
        const params = new URLSearchParams(location.search);
        const next = params.get('next');
        location.replace(next || '/index.html');
      }
    } catch(e) {
      // Not logged in, stay on login page
    }
    return;
  }
  
  // Check real authentication via API
  try {
    const user = await Auth.me();
    if(!user || !user.email) {
      // Not authenticated, redirect to login with next parameter
      const next = encodeURIComponent(location.pathname + location.search);
      location.href = './pages/login.html?next=' + next;
    }
  } catch(e) {
    // API call failed or not authenticated, redirect to login
    console.warn('[auth.js] Authentication check failed:', e);
    const next = encodeURIComponent(location.pathname + location.search);
    location.href = './pages/login.html?next=' + next;
  }
})();
