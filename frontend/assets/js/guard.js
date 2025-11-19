// Simple Auth Guard for IT Ticketing
// - Redirects unauthenticated users to /pages/login.html
// - Supports ?next= redirect back after login
// - Exposes window.AuthGuard with tiny helpers

(function () {
  const LOGIN_PATH = "/pages/login.html";
  const REGISTER_PATH = "/pages/register.html";
  const HOME_PATH = "/index.html";

  // Public pages that do NOT require auth (adjust as you like)
  const PUBLIC_PATHS = new Set([
    LOGIN_PATH,
    REGISTER_PATH,
  ]);

  // Cache the current user for the page lifetime
  let _currentUser = null;
  let _checked = false;
  let _loginInProgress = false; // Flag to prevent guard from interfering during login

  // Read current path (no origin)
  const PATH = location.pathname;
  
  // Expose flag for login page to set
  window._loginInProgress = () => {
    _loginInProgress = true;
  };

  // ---------- Core ----------
  async function fetchMe() {
    try {
      // Ensure credentials are included and add cache-busting for debugging
      const res = await fetch("/api/auth/me", { 
        credentials: "include",
        cache: "no-store",
        headers: {
          "Cache-Control": "no-cache"
        }
      });
      if (!res.ok) {
        // Log for debugging (remove in production if needed)
        if (res.status === 401) {
          console.debug("[guard.js] Not authenticated (401)");
        } else {
          console.debug("[guard.js] Auth check failed:", res.status, res.statusText);
        }
        return null;
      }
      const user = await res.json();
      console.debug("[guard.js] Auth check successful:", user?.email || "no email");
      return user;
    } catch (err) {
      console.debug("[guard.js] Auth check error:", err);
      return null;
    }
  }

  function isPublicPath(pathname) {
    // Normalize the path
    const normalized = pathname.endsWith("/") ? pathname.slice(0, -1) : pathname;
    // Check if it matches any public path (exact match or ends with the path)
    for (const pubPath of PUBLIC_PATHS) {
      if (normalized === pubPath || normalized.endsWith(pubPath)) {
        return true;
      }
    }
    return false;
  }

  function withNext(url) {
    const next = encodeURIComponent(location.pathname + location.search);
    const hasQuery = url.includes("?");
    return url + (hasQuery ? "&" : "?") + "next=" + next;
  }

  function getNextOr(defaultPath) {
    const p = new URLSearchParams(location.search).get("next");
    return p || defaultPath;
  }

  async function ensureAuth() {
    if (_checked) return _currentUser;
    _currentUser = await fetchMe();
    _checked = true;
    return _currentUser;
  }

  async function guardPage() {
    // Don't run guard logic if login is in progress (let login form handle redirect)
    if (_loginInProgress) {
      console.debug("[guard.js] Login in progress, skipping guard");
      return;
    }
    
    const pub = isPublicPath(PATH);
    console.debug("[guard.js] Checking auth for path:", PATH, "isPublic:", pub);
    
    const me = await ensureAuth();
    console.debug("[guard.js] Auth result:", me ? `authenticated (${me.email})` : "not authenticated");

    // Case 1: user not logged in on a protected page -> go to login
    if (!pub && !me) {
      // Double-check: maybe cookie wasn't sent, try one more time after a short delay
      // This handles cases where the page loads before cookies are fully available
      console.debug("[guard.js] Not authenticated on protected page, retrying after delay...");
      await new Promise(resolve => setTimeout(resolve, 200));
      
      // Reset the check flag to force a new fetch
      _checked = false;
      const retryMe = await ensureAuth();
      
      if (retryMe) {
        console.debug("[guard.js] Retry successful, user authenticated:", retryMe.email);
        // User is authenticated, continue normally
        try {
          document.dispatchEvent(new CustomEvent("auth:user", { detail: retryMe }));
        } catch {}
        return;
      }
      
      console.debug("[guard.js] Still not authenticated after retry, redirecting to login");
      location.replace(withNext(LOGIN_PATH));
      return;
    }

    // Case 2: user already logged in but on login/register -> send to next/home
    // Only redirect if login is not in progress
    const isLoginPage = PATH.includes('login.html') || PATH === LOGIN_PATH;
    const isRegisterPage = PATH.includes('register.html') || PATH === REGISTER_PATH;
    
    if (pub && me && (isLoginPage || isRegisterPage) && !_loginInProgress) {
      const target = getNextOr(HOME_PATH);
      // Small delay to avoid race condition with login redirect
      setTimeout(() => {
        if (!_loginInProgress) {
          location.replace(target);
        }
      }, 100);
      return;
    }

    // Otherwise continue; optionally fire an event for other scripts
    if (me) {
      try {
        document.dispatchEvent(new CustomEvent("auth:user", { detail: me }));
      } catch {}
    }
  }

  // ---------- Extras ----------
  // Minimal role checker for pages that need it
  async function requireRole(...roles) {
    const me = await ensureAuth();
    if (!me) {
      location.replace(withNext(LOGIN_PATH));
      return false;
    }
    if (!roles.includes(me.role)) {
      // Forbidden: send home (or a dedicated 403 page if you add one)
      location.replace(HOME_PATH);
      return false;
    }
    return true;
  }

  // API for other scripts
  window.AuthGuard = {
    ensureAuth, // returns user or null
    requireRole, // await AuthGuard.requireRole("admin")
    get user() {
      return _currentUser;
    },
  };

  // Kick it off automatically, but wait for DOM to be ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
      // Small delay to ensure cookies are available
      setTimeout(guardPage, 50);
    });
  } else {
    // DOM already ready, but still wait a bit for cookies
    setTimeout(guardPage, 50);
  }
})();
