
// Small helpers on top of Auth in mock_api.js

(function(){
  // If we are not on login page and not "logged in", bounce to login.
  const onLogin = location.pathname.endsWith('/login.html') || location.pathname.endsWith('/pages/login.html');
  const logged = !!localStorage.getItem('demo_email');
  if(!onLogin && !logged){
    // allow index.html to be public for demo
    if(!location.pathname.endsWith('/index.html') && !location.pathname.endswith('/')){
      location.href = './pages/login.html';
    }
  }
})();
