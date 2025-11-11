(async function () {
  // nav
  const nav = document.getElementById("nav-actions");
  if (nav) {
    addDefaultNav("nav-actions");
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
      // Use searchTickets which properly calls the backend API
      const { items: list } = await API.searchTickets({
        q: q ? q.value : "",
        status: fs ? fs.value : "",
        limit: 50,
        offset: 0,
        sort: "updated_at",
        order: "desc",
      });
      if (!tbody) return;

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

  q && q.addEventListener("input", render);
  fs && fs.addEventListener("change", render);
  render();
})();
