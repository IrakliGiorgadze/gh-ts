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

  // Read current path (no origin)
  const PATH = location.pathname;

  // ---------- Core ----------
  async function fetchMe() {
    try {
      const res = await fetch("/api/auth/me", { credentials: "include" });
      if (!res.ok) return null;
      return await res.json();
    } catch {
      return null;
    }
  }

  function isPublicPath(pathname) {
    // Normalize trailing slashes
    if (pathname.endsWith("/") && PUBLIC_PATHS.has(pathname.slice(0, -1)))
      return true;
    return PUBLIC_PATHS.has(pathname);
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
    const pub = isPublicPath(PATH);
    const me = await ensureAuth();

    // Case 1: user not logged in on a protected page -> go to login
    if (!pub && !me) {
      location.replace(withNext(LOGIN_PATH));
      return;
    }

    // Case 2: user already logged in but on login/register -> send to next/home
    if (pub && me && (PATH === LOGIN_PATH || PATH === REGISTER_PATH)) {
      const target = getNextOr(HOME_PATH);
      location.replace(target);
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

  // Kick it off automatically
  guardPage();
})();
