// Simple component loader + navbar wiring
(async function () {
  async function loadNavbar(into = "#navbar") {
    const host = document.querySelector(into);
    if (!host) return;
    const res = await fetch("/components/navbar.html", {
      credentials: "include",
    });
    host.innerHTML = await res.text();

    // Populate username if logged in
    try {
      const meRes = await fetch("/api/auth/me", { credentials: "include" });
      if (meRes.ok) {
        const me = await meRes.json();
        const u = document.getElementById("nav-username");

        if (u) {
          if (me.name && me.name.trim() !== "") {
            u.textContent = me.name;
          } else if (me.email) {
            u.textContent = me.email;
          } else {
            u.textContent = me.id;
          }
        }
      } else {
        // Not authenticated — optionally hide logout and show Sign in
        const btn = document.getElementById("btn-logout");
        if (btn) btn.style.display = "none";
      }
    } catch (_) {}

    // Wire logout
    document.addEventListener("click", (e) => {
      const el = e.target.closest("#btn-logout");
      if (!el) return;
      e.preventDefault();
      fetch("/api/auth/logout", { method: "POST", credentials: "include" })
        .catch(() => {})
        .finally(() => {
          try {
            localStorage.removeItem("demo_email");
          } catch (_) {}
          window.location.href = "/pages/login.html";
        });
    });
  }

  // Expose and auto-load
  window.Components = window.Components || {};
  window.Components.loadNavbar = loadNavbar;

  // Auto-mount if a container exists
  if (document.querySelector("#navbar")) {
    loadNavbar("#navbar");
  }
})();
