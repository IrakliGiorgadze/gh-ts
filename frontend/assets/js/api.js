/* frontend/assets/js/api.js
   API client for the IT Ticketing System.
   Assumes Nginx proxies /api → backend (so same-origin, cookies ok).
*/
(() => {
  const BASE = "/api"; // Nginx proxy in frontend/nginx.conf

  // ---------- Utilities ----------
  const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

  class ApiError extends Error {
    constructor(message, status, details) {
      super(message);
      this.name = "ApiError";
      this.status = status;
      this.details = details;
    }
  }

  // Optional bearer token support (kept for future); cookies are primary.
  const Token = {
    get: () => localStorage.getItem("auth_token") || null,
    set: (t) => localStorage.setItem("auth_token", t),
    clear: () => localStorage.removeItem("auth_token"),
  };

  // Core fetch with retries + JSON/text handling
  async function fetchJSON(
    path,
    {
      method = "GET",
      body,
      headers = {},
      retries = 2,
      retryBackoffBase = 250,
      credentials = "include",
    } = {}
  ) {
    const url = path.startsWith("http") ? path : `${BASE}${path}`;
    const h = { "Content-Type": "application/json", ...headers };
    const token = Token.get();
    if (token) h["Authorization"] = `Bearer ${token}`;

    const payload =
      body && typeof body !== "string" ? JSON.stringify(body) : body;

    for (let attempt = 0; attempt <= retries; attempt++) {
      try {
        const res = await fetch(url, {
          method,
          headers: h,
          body: payload,
          credentials,
        });

        // Success
        if (res.ok) {
          if (res.status === 204) return null;
          const ct = res.headers.get("content-type") || "";
          if (ct.includes("application/json")) return await res.json();
          return await res.text();
        }

        // Error: maybe retry
        const isRetryable = [409, 429, 500, 502, 503, 504].includes(res.status);
        let errPayload = null;
        try {
          errPayload = await res.json();
        } catch {
          errPayload = { error: await res.text() };
        }

        if (isRetryable && attempt < retries) {
          await sleep(2 ** attempt * retryBackoffBase);
          continue;
        }
        throw new ApiError(
          errPayload?.error || errPayload?.message || `HTTP ${res.status}`,
          res.status,
          errPayload
        );
      } catch (e) {
        const isNetwork =
          e instanceof TypeError || (e.name === "ApiError" && e.status === 0);
        if (isNetwork && attempt < retries) {
          await sleep(2 ** attempt * retryBackoffBase);
          continue;
        }
        throw e;
      }
    }
  }

  // ---------- Public API ----------
  const API = {
    // New: server-side filters + pagination; returns { items, total }
    async searchTickets({
      q = "",
      status = "",
      priority = "",
      category = "",
      assignee = "",
      limit = 10,
      offset = 0,
      sort = "updated_at",
      order = "desc",
    } = {}) {
        const params = new URLSearchParams();
        if (q) params.set("q", q);
        if (status) params.set("status", status);
        if (priority) params.set("priority", priority);
        if (category) params.set("category", category);
        if (assignee) params.set("assignee", assignee);
        params.set("limit", String(limit));
        params.set("offset", String(offset));
        params.set("sort", sort);
        params.set("order", order);

        // Backend returns { items, total } (advanced) OR [] (legacy)
        const data = await fetchJSON(`/tickets?${params.toString()}`, {
          method: "GET",
        });
        if (Array.isArray(data)) {
          // legacy fallback
          return { items: data, total: data.length };
        }
        // expected modern shape
        if (data && Array.isArray(data.items)) {
          return {
            items: data.items,
            total:
              typeof data.total === "number" ? data.total : data.items.length,
          };
        }
        // ultimate fallback
        return { items: [], total: 0 };
    },

    // Backward-compat — returns array only (older pages). Internally uses searchTickets.
    async listTickets({ q = "", status = "" } = {}) {
      const { items } = await this.searchTickets({
        q,
        status,
        limit: 100,
        offset: 0,
      });
      return items;
    },

    async getTicket(id) {
      return await fetchJSON(`/tickets/${encodeURIComponent(id)}`, {
        method: "GET",
      });
    },

    async createTicket(t) {
      return await fetchJSON(`/tickets`, { method: "POST", body: t });
    },

    async updateTicket(id, patch) {
      return await fetchJSON(`/tickets/${encodeURIComponent(id)}`, {
        method: "PATCH",
        body: patch,
      });
    },

    async addComment(id, text) {
      // Server returns full updated ticket (your backend does this)
      return await fetchJSON(`/tickets/${encodeURIComponent(id)}/comments`, {
        method: "POST",
        body: { text },
      });
    },

    // NEW: backend reports summary (with client-side fallback)
    async getReportSummary() {
      try {
        // backend returns: { open, resolved7d, highCriticalOpen }
        return await fetchJSON(`/reports/summary`, { method: "GET" });
      } catch (e) {
        console.warn(
          "[API.getReportSummary] falling back to client-side stats:",
          e
        );
        const { open, resolved7d, risk } = await this.getStats();
        return { open, resolved7d, highCriticalOpen: risk };
      }
    },

    // Simple KPIs; uses server list when available
    async getStats() {
      const { items } = await this.searchTickets({ limit: 200, offset: 0 });
      const now = Date.now();
      const open = items.filter(
        (t) => !["Resolved", "Closed"].includes(t.status)
      ).length;
      const resolved7d = items.filter((t) => {
        if (!["Resolved", "Closed"].includes(t.status)) return false;
        const ts = new Date(t.updatedAt || t.createdAt || Date.now()).getTime();
        return now - ts < 7 * 86400_000;
      }).length;
      const risk = items.filter(
        (t) =>
          ["High", "Critical"].includes(t.priority) &&
          !["Resolved", "Closed"].includes(t.status)
      ).length;
      return { open, resolved7d, risk };
    },
  };

  // ---------- Auth (cookie-based; works with your Go backend) ----------
  const Auth = {
    async login(email, password) {
      const res = await fetch(`${BASE}/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ email, password }),
      });
      if (!res.ok) return false;
      // backend returns the user json; cookie set via Set-Cookie
      const user = await res.json().catch(() => null);
      if (user?.email) localStorage.setItem("demo_email", user.email); // keep minimal UI state
      return true;
    },
    async logout() {
      await fetch(`${BASE}/auth/logout`, {
        method: "POST",
        credentials: "include",
      }).catch(() => {});
      localStorage.removeItem("demo_email");
      Token.clear();
    },
    async me() {
      const res = await fetch(`${BASE}/auth/me`, { credentials: "include" });
      if (!res.ok) return null;
      return await res.json();
    },
  };

  // ---------- Users (admin only) ----------
  const Users = {
    // List users (admin only) - can filter by role
    async list({ role = "", active = true, limit = 100, offset = 0 } = {}) {
      try {
        const params = new URLSearchParams();
        if (role) params.set("role", role);
        if (active !== null) params.set("active", String(active));
        params.set("limit", String(limit));
        params.set("offset", String(offset));

        const data = await fetchJSON(`/users?${params.toString()}`, {
          method: "GET",
        });
        if (data && Array.isArray(data.items)) {
          return {
            items: data.items,
            total: typeof data.total === "number" ? data.total : data.items.length,
          };
        }
        return { items: [], total: 0 };
      } catch (e) {
        console.error("[Users.list] Failed:", e);
        throw e;
      }
    },
    // Get agents and admins (for assignment dropdown)
    async getAgents() {
      const [agents, admins] = await Promise.all([
        this.list({ role: "agent", active: true, limit: 200 }),
        this.list({ role: "admin", active: true, limit: 200 })
      ]);
      return [...agents.items, ...admins.items];
    },
  };

  // expose globally
  window.API = API;
  window.Auth = Auth;
  window.Users = Users;
})();
