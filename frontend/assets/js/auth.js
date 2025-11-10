// Authentication guard - checks real API authentication state
// Note: This file works with the real backend API, not mock data

(async function(){
  // Pages that don't require authentication
  const publicPages = [
    '/login.html',
    '/pages/login.html',
    '/register.html',
    '/pages/register.html'
  ];
  
  const isPublicPage = publicPages.some(page => 
    location.pathname.endsWith(page)
  );
  
  // If on a public page, don't check auth
  if(isPublicPage) return;
  
  // Check real authentication via API
  try {
    const user = await Auth.me();
    if(!user || !user.email) {
      // Not authenticated, redirect to login
      location.href = './pages/login.html';
    }
  } catch(e) {
    // API call failed or not authenticated, redirect to login
    console.warn('[auth.js] Authentication check failed:', e);
    location.href = './pages/login.html';
  }
})();
