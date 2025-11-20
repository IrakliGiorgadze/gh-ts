(async function () {
  // nav
  const nav = document.getElementById("nav-actions");
  if (nav) {
    addDefaultNav("nav-actions");
  }

  // Show admin quick access if user is admin
  async function checkAdminAccess() {
    try {
      // Don't redirect here - let index.html handle auth redirects
      // Just check if user is admin and show/hide admin panel
      const me = await Auth.me();
      
      if (me && me.email) {
        console.log("[Dashboard] Current user:", me.email, "role:", me.role);
        if (me.role === "admin") {
          const adminAccess = document.getElementById("admin-quick-access");
          if (adminAccess) {
            adminAccess.style.display = "block";
            console.log("[Dashboard] Admin access card shown");
          } else {
            console.warn("[Dashboard] Admin access element not found");
          }
        }
      } else {
        console.log("[Dashboard] Not authenticated (auth check will handle redirect)");
      }
    } catch (e) {
      // Don't redirect here - let index.html handle it
      console.warn("[Dashboard] Could not check admin status:", e);
    }
  }
  
  // Wait for DOM and auth to be ready
  function initAdminAccess() {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", () => {
        setTimeout(checkAdminAccess, 200);
      });
    } else {
      setTimeout(checkAdminAccess, 200);
    }
  }
  
  initAdminAccess();

  // Helper to escape HTML (if not available globally)
  function escapeHtml(str) {
    if (typeof window.escapeHtml === 'function') {
      return window.escapeHtml(str);
    }
    return String(str).replace(/[&<>"']/g, (m) => ({
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#39;'
    }[m]));
  }

  // Helper to show error message
  function showError(message) {
    const tbody = document.getElementById("tickets-body");
    if (tbody) {
      tbody.innerHTML = `<tr><td colspan="6" style="text-align:center;padding:20px;color:#999">${escapeHtml(
        message
      )}</td></tr>`;
    }
  }

  // Load KPIs with error handling
  try {
    const stats = await API.getStats();
    const safe = (id, val) => {
      const el = document.getElementById(id);
      if (el) el.textContent = String(val);
    };
    safe("kpi-open", stats.open || 0);
    safe("kpi-risk", stats.risk || 0);
    safe("kpi-resolved", stats.resolved7d || 0);
  } catch (e) {
    console.error("[Dashboard] Failed to load stats:", e);
    // Stats will show 0 if API fails
  }

  const q = document.getElementById("q");
  const fs = document.getElementById("filter-status");
  const tbody = document.getElementById("tickets-body");

  async function render() {
    try {
      // Get search query and trim whitespace
      const searchQuery = q ? q.value.trim() : "";
      const statusFilter = fs ? fs.value.trim() : "";
      
      // Use searchTickets which properly calls the backend API
      const { items: list } = await API.searchTickets({
        q: searchQuery,
        status: statusFilter,
        limit: 50,
        offset: 0,
        sort: "updated_at",
        order: "desc",
      });
      if (!tbody) return;
      
      // Debug logging (remove in production if needed)
      console.log("[Dashboard] Search query:", searchQuery, "Results:", list.length);

      if (!list || list.length === 0) {
        tbody.innerHTML =
          '<tr><td colspan="6" style="text-align:center;padding:20px;color:#999">No tickets found</td></tr>';
        return;
      }

      tbody.innerHTML = "";
      list.forEach((t) => {
        const tr = document.createElement("tr");
        // Use assigneeName or assigneeEmail if available, otherwise fallback to assignee ID
        const assigneeDisplay = (
          t.assigneeName ||
          t.assigneeEmail ||
          t.assignee ||
          ""
        ).trim();
        const ticketLabel = t.alias || t.id;
        tr.innerHTML = `
          <td>${escapeHtml(ticketLabel)}</td>
          <td><a href="./pages/ticket-detail.html?id=${t.id}">${escapeHtml(
          t.title
        )}</a></td>
          <td>${t.priority}</td>
          <td><span class="status ${t.status
            .toUpperCase()
            .replace(" ", "_")}">${t.status}</span></td>
          <td>${escapeHtml(assigneeDisplay)}</td>
          <td>${t.updatedAt ? new Date(t.updatedAt).toLocaleString() : ""}</td>
        `;
        tbody.appendChild(tr);
      });
    } catch (e) {
      console.error("[Dashboard] Failed to load tickets:", e);
      const errorMsg =
        e.message || "Failed to load tickets. Please check your connection.";
      showError(errorMsg);
    }
  }

  // Debounce search input to avoid too many API calls
  let searchTimeout = null;
  if (q) {
    q.addEventListener("input", () => {
      // Clear previous timeout
      if (searchTimeout) {
        clearTimeout(searchTimeout);
      }
      // Set new timeout - search after 300ms of no typing
      searchTimeout = setTimeout(() => {
        render();
      }, 300);
    });
    // Also trigger on Enter key for immediate search
    q.addEventListener("keydown", (e) => {
      if (e.key === "Enter") {
        if (searchTimeout) {
          clearTimeout(searchTimeout);
        }
        render();
      }
    });
  }
  fs && fs.addEventListener("change", render);
  render();
})();
