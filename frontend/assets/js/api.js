/* frontend/assets/js/api.js
   Advanced API client for the IT Ticketing System.
   Assumes Nginx proxies /api → backend (so same-origin, no CORS headaches).
*/
(() => {
    const BASE = "/api"; // Nginx proxy in frontend/nginx.conf
    const DEMO_NS = "it_demo_v1"; // for offline fallback

    // -------- Utilities --------
    const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

    class ApiError extends Error {
        constructor(message, status, details) {
            super(message);
            this.name = "ApiError";
            this.status = status;
            this.details = details;
        }
    }

    // Optional bearer token support (when you add real auth)
    const Token = {
        get: () => localStorage.getItem("auth_token") || null,
        set: (t) => localStorage.setItem("auth_token", t),
        clear: () => localStorage.removeItem("auth_token"),
    };

    // Core fetch with retries + JSON parsing
    async function fetchJSON(path, { method = "GET", body, headers = {}, retries = 3, retryBackoffBase = 250, credentials = "include" } = {}) {
        const url = path.startsWith("http") ? path : `${BASE}${path}`;
        const h = { "Content-Type": "application/json", ...headers };
        const token = Token.get();
        if (token) h["Authorization"] = `Bearer ${token}`;

        for (let attempt = 0; attempt <= retries; attempt++) {
            try {
                const res = await fetch(url, {
                    method,
                    headers: h,
                    credentials, // include cookies if you switch to cookie sessions
                    body: body ? JSON.stringify(body) : undefined,
                });

                // Success path
                if (res.ok) {
                    // no content
                    if (res.status === 204) return null;
                    // try json; if not json, return text
                    const ct = res.headers.get("content-type") || "";
                    if (ct.includes("application/json")) return await res.json();
                    return await res.text();
                }

                // Non-2xx: decide whether to retry
                const isRetryable = [409, 429, 500, 502, 503, 504].includes(res.status);
                let payload = null;
                try { payload = await res.json(); } catch { payload = { error: await res.text() }; }
                if (isRetryable && attempt < retries) {
                    await sleep((2 ** attempt) * retryBackoffBase);
                    continue;
                }
                throw new ApiError(payload?.error || payload?.message || `HTTP ${res.status}`, res.status, payload);
            } catch (e) {
                // Network error: retry
                const isNetwork = e instanceof TypeError || (e.name === "ApiError" && e.status === 0);
                if (isNetwork && attempt < retries) {
                    await sleep((2 ** attempt) * retryBackoffBase);
                    continue;
                }
                throw e;
            }
        }
    }

    // -------- Offline fallback (localStorage) --------
    function _dbLoad() {
        const raw = localStorage.getItem(DEMO_NS);
        if (raw) return JSON.parse(raw);
        const now = Date.now();
        const data = {
            users: [{ id: "u1", email: "admin@example.com", name: "Admin", role: "admin", password: "admin" }],
            tickets: [
                { id: "T-1001", title: "VPN not connecting", description: "Fails on step 2", category: "Network", priority: "High", department: "Finance", status: "Open", assignee: "Admin", updatedAt: now - 3600_000, createdAt: now - 7200_000, comments: [{ text: "Checking logs", createdAt: now - 3500_000 }] },
                { id: "T-1002", title: "Email quota exceeded", description: "Cannot send", category: "Software", priority: "Medium", department: "Sales", status: "Pending", assignee: "", updatedAt: now - 1800_000, createdAt: now - 5400_000, comments: [] },
                { id: "T-1003", title: "Laptop fan noise", description: "Very loud", category: "Hardware", priority: "Low", department: "HR", status: "New", assignee: "", updatedAt: now - 600_000, createdAt: now - 500_000, comments: [] },
            ],
            nextId: 1004,
        };
        localStorage.setItem(DEMO_NS, JSON.stringify(data));
        return data;
    }
    const _dbSave = (d) => localStorage.setItem(DEMO_NS, JSON.stringify(d));

    // Filter helper for offline list
    function _filterList(list, { q = "", status = "" } = {}) {
        let arr = list.slice().sort((a, b) => (b.updatedAt || 0) - (a.updatedAt || 0));
        if (q) {
            const needle = q.toLowerCase();
            arr = arr.filter((t) => (t.title + t.description).toLowerCase().includes(needle));
        }
        if (status) arr = arr.filter((t) => t.status === status);
        return arr;
    }

    // -------- Public API (matches your mock) --------
    const API = {
        async listTickets({ q = "", status = "" } = {}) {
            try {
                const query = new URLSearchParams();
                if (q) query.set("q", q);
                if (status) query.set("status", status);
                const path = `/tickets${query.toString() ? `?${query.toString()}` : ""}`;
                return await fetchJSON(path, { method: "GET" });
            } catch (e) {
                // offline fallback
                console.warn("[API.listTickets] Falling back to localStorage:", e);
                const db = _dbLoad();
                return _filterList(db.tickets, { q, status });
            }
        },

        async getTicket(id) {
            try {
                return await fetchJSON(`/tickets/${encodeURIComponent(id)}`, { method: "GET" });
            } catch (e) {
                console.warn("[API.getTicket] Falling back to localStorage:", e);
                const db = _dbLoad();
                return db.tickets.find((t) => t.id === id) || null;
            }
        },

        async createTicket(t) {
            try {
                return await fetchJSON(`/tickets`, { method: "POST", body: t });
            } catch (e) {
                console.warn("[API.createTicket] Falling back to localStorage:", e);
                const db = _dbLoad();
                const id = `T-${db.nextId++}`;
                const now = Date.now();
                const rec = {
                    id,
                    title: t.title,
                    description: t.description || "",
                    category: t.category || "Software",
                    priority: t.priority || "Low",
                    department: t.department || "",
                    status: "New",
                    assignee: "",
                    updatedAt: now,
                    createdAt: now,
                    comments: [],
                };
                db.tickets.push(rec);
                _dbSave(db);
                return rec;
            }
        },

        async updateTicket(id, patch) {
            try {
                return await fetchJSON(`/tickets/${encodeURIComponent(id)}`, { method: "PATCH", body: patch });
            } catch (e) {
                console.warn("[API.updateTicket] Falling back to localStorage:", e);
                const db = _dbLoad();
                const t = db.tickets.find((x) => x.id === id);
                if (!t) throw new ApiError("not found", 404);
                Object.assign(t, patch);
                t.updatedAt = Date.now();
                _dbSave(db);
                return t;
            }
        },

        async addComment(id, text) {
            try {
                return await fetchJSON(`/tickets/${encodeURIComponent(id)}/comments`, { method: "POST", body: { text } });
            } catch (e) {
                console.warn("[API.addComment] Falling back to localStorage:", e);
                const db = _dbLoad();
                const t = db.tickets.find((x) => x.id === id);
                if (!t) throw new ApiError("not found", 404);
                t.comments = t.comments || [];
                t.comments.push({ text, createdAt: Date.now() });
                t.updatedAt = Date.now();
                _dbSave(db);
                return t;
            }
        },

        // Derive KPIs on the client from list (works online or offline)
        async getStats() {
            const list = await this.listTickets({});
            const now = Date.now();
            const open = list.filter((t) => !["Resolved", "Closed"].includes(t.status)).length;
            const resolved7d = list.filter((t) => t.status === "Resolved" && now - new Date(t.updatedAt || Date.now()).getTime() < 7 * 86400_000).length;
            const risk = list.filter((t) => ["High", "Critical"].includes(t.priority) && !["Resolved", "Closed"].includes(t.status)).length;
            return { open, resolved7d, risk };
        },
    };

    // Simple “session” facade for now (keeps parity with your mock)
    const Auth = {
        async login(email, password) {
            // When you add real backend auth, call it here and store token/cookie.
            // For now, keep the demo login behavior.
            const db = _dbLoad();
            const user = db.users.find((u) => u.email === email && u.password === password);
            if (user) {
                localStorage.setItem("demo_email", user.email);
                return true;
            }
            return false;
        },
        logout() {
            localStorage.removeItem("demo_email");
            Token.clear();
        },
    };

    // expose globally
    window.API = API;
    window.Auth = Auth;
})();