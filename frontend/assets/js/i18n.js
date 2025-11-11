/* Minimal i18n runtime for vanilla JS.
 * - Loads /assets/i18n/{lang}.json
 * - Persists lang in localStorage ('lang')
 * - Applies translations to elements with:
 *   - data-i18n="key" (textContent)
 *   - data-i18n-attr="placeholder|title|value|aria-label" (attribute)
 * - Exposes: window.I18n.t, setLang, getLang, formatDate
 */
(function () {
  const STORAGE_KEY = "lang";
  const DEFAULT_LANG = "en";
  const SUPPORTED = ["en", "ka"];

  let currentLang = null;
  let dict = {};
  const listeners = new Set();

  function getLang() {
    if (currentLang) return currentLang;
    const saved =
      localStorage.getItem(STORAGE_KEY) ||
      (navigator.language || "en").slice(0, 2).toLowerCase();
    currentLang = SUPPORTED.includes(saved) ? saved : DEFAULT_LANG;
    return currentLang;
  }

  async function loadDict(lang) {
    const res = await fetch(`/assets/i18n/${lang}.json`, {
      credentials: "omit",
      cache: "no-cache",
    });
    if (!res.ok) throw new Error(`Failed to load i18n for ${lang}`);
    return await res.json();
  }

  function t(key, params) {
    const val =
      key.split(".").reduce((o, k) => (o && o[k] !== undefined ? o[k] : null), dict) ??
      key;
    if (!params) return String(val);
    return String(val).replace(/\{(\w+)\}/g, (_, k) => (k in params ? params[k] : `{${k}}`));
  }

  function applyTranslations(root = document) {
    // Elements with data-i18n (text)
    root.querySelectorAll("[data-i18n]").forEach((el) => {
      const key = el.getAttribute("data-i18n");
      if (!key) return;
      el.textContent = t(key);
    });
    // Elements with data-i18n-attr="placeholder|title|value|aria-label"
    root.querySelectorAll("[data-i18n-attr]").forEach((el) => {
      const attrList = (el.getAttribute("data-i18n-attr") || "")
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean);
      for (const attr of attrList) {
        const key = el.getAttribute(`data-i18n-${attr}`);
        if (key) el.setAttribute(attr, t(key));
      }
    });
  }

  async function setLang(lang) {
    const next = SUPPORTED.includes(lang) ? lang : DEFAULT_LANG;
    if (next === currentLang && Object.keys(dict).length) return;
    const loaded = await loadDict(next);
    currentLang = next;
    dict = loaded || {};
    localStorage.setItem(STORAGE_KEY, currentLang);
    applyTranslations(document);
    listeners.forEach((fn) => {
      try {
        fn(currentLang);
      } catch {}
    });
  }

  function onChange(cb) {
    listeners.add(cb);
    return () => listeners.delete(cb);
  }

  function formatDate(input) {
    const lang = getLang();
    try {
      const dt = input instanceof Date ? input : new Date(input);
      return new Intl.DateTimeFormat(lang, {
        year: "numeric",
        month: "short",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
      }).format(dt);
    } catch {
      return String(input);
    }
  }

  // Wire language toggle: any <select data-i18n-toggle>
  function initToggles() {
    document.querySelectorAll('select[data-i18n-toggle]').forEach((sel) => {
      const lang = getLang();
      if (sel.value !== lang) sel.value = lang;
      sel.addEventListener("change", (e) => {
        setLang(e.target.value);
      });
    });
  }

  window.I18n = {
    t,
    setLang,
    getLang,
    onChange,
    formatDate,
    // simple domain value mappers
    mapStatus(value) {
      const m = {
        New: { en: "New", ka: "ახალი" },
        Open: { en: "Open", ka: "გახსნილი" },
        "In Progress": { en: "In Progress", ka: "მიმდინარე" },
        Pending: { en: "Pending", ka: "მოლოდინში" },
        Resolved: { en: "Resolved", ka: "გადაწყვეტილი" },
        Closed: { en: "Closed", ka: "დახურული" }
      };
      const lang = getLang();
      return (m[value] && m[value][lang]) || value;
    },
    mapPriority(value) {
      const m = {
        Low: { en: "Low", ka: "დაბალი" },
        Medium: { en: "Medium", ka: "საშუალო" },
        High: { en: "High", ka: "მაღალი" },
        Critical: { en: "Critical", ka: "კრიტიკული" }
      };
      const lang = getLang();
      return (m[value] && m[value][lang]) || value;
    }
  };

  // Initial load
  document.addEventListener("DOMContentLoaded", async () => {
    await setLang(getLang());
    initToggles();
  });
})();


