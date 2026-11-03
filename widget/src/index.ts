// SearchCraft 2KB Standalone Command Palette Search Widget
(function () {
  if (typeof window === "undefined" || (window as any).__SEARCHCRAFT_INITIALIZED__) {
    return;
  }
  (window as any).__SEARCHCRAFT_INITIALIZED__ = true;

  // Localization Dictionary
  const i18nDict: Record<string, any> = {
    en: {
      placeholder: "Search documentation or products (Cmd+K)...",
      typeToSearch: "Type a query to search across",
      noResults: "No matching results found for",
      didYouMean: "Did you mean:",
      navigate: "navigate",
      select: "select",
      close: "close",
      poweredBy: "Powered by",
      suggestions: "Suggestions:",
    },
    ro: {
      placeholder: "Căutare rapidă (Cmd+K)...",
      typeToSearch: "Tastați pentru a căuta în",
      noResults: "Niciun rezultat găsit pentru",
      didYouMean: "Ați vrut să căutați:",
      navigate: "navigare",
      select: "selectare",
      close: "închidere",
      poweredBy: "Propulsat de",
      suggestions: "Sugestii:",
    },
    es: {
      placeholder: "Buscar en la documentación (Cmd+K)...",
      typeToSearch: "Escriba una consulta para buscar en",
      noResults: "No se encontraron resultados para",
      didYouMean: "¿Quiso decir:",
      navigate: "navegar",
      select: "seleccionar",
      close: "cerrar",
      poweredBy: "Desarrollado por",
      suggestions: "Sugerencias:",
    },
    fr: {
      placeholder: "Rechercher des documents (Cmd+K)...",
      typeToSearch: "Tapez une requête pour rechercher dans",
      noResults: "Aucun résultat trouvé pour",
      didYouMean: "Vouliez-vous dire :",
      navigate: "naviguer",
      select: "sélectionner",
      close: "fermer",
      poweredBy: "Propulsé par",
      suggestions: "Suggestions :",
    },
    de: {
      placeholder: "Dokumentation durchsuchen (Cmd+K)...",
      typeToSearch: "Tippen Sie eine Suchanfrage für",
      noResults: "Keine Ergebnisse gefunden für",
      didYouMean: "Meinten Sie:",
      navigate: "navigieren",
      select: "auswählen",
      close: "schließen",
      poweredBy: "Bereitgestellt von",
      suggestions: "Vorschläge:",
    },
  };

  // Extract configuration from current script tag
  const currentScript = document.currentScript as HTMLScriptElement | null;
  const baseUrl = (currentScript?.getAttribute("data-url") || window.location.origin).replace(/\/+$/, "");
  const indexName = currentScript?.getAttribute("data-index") || "products";
  const apiKey = currentScript?.getAttribute("data-api-key") || "";
  const langKey = (currentScript?.getAttribute("data-lang") || "en").toLowerCase();
  const theme = (currentScript?.getAttribute("data-theme") || "auto").toLowerCase();
  const accentColor = currentScript?.getAttribute("data-accent") || "#3b82f6";
  const shortcutKey = (currentScript?.getAttribute("data-shortcut") || "k").toLowerCase();
  const limit = parseInt(currentScript?.getAttribute("data-limit") || "8", 10);
  const searchMode = currentScript?.getAttribute("data-mode") || "hybrid";

  const t = i18nDict[langKey] || i18nDict.en;
  const placeholder = currentScript?.getAttribute("data-placeholder") || t.placeholder;

  let isOpen = false;
  let selectedIndex = 0;
  let currentHits: any[] = [];
  let debounceTimer: any = null;
  let activeAbortController: AbortController | null = null;

  // Set accent color variable
  document.documentElement.style.setProperty("--sc-accent", accentColor);

  // Create UI Elements
  const overlay = document.createElement("div");
  overlay.className = `sc-overlay sc-theme-${theme}`;
  overlay.setAttribute("role", "dialog");
  overlay.setAttribute("aria-modal", "true");

  overlay.innerHTML = `
    <div class="sc-modal">
      <div class="sc-search-bar">
        <svg class="sc-search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8"></circle>
          <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
        </svg>
        <input type="text" class="sc-input" placeholder="${escapeHtml(placeholder)}" autocomplete="off" spellcheck="false" />
        <div class="sc-spinner"></div>
        <kbd class="sc-kbd">ESC</kbd>
      </div>
      <div class="sc-suggestions-bar" style="display: none;"></div>
      <div class="sc-did-you-mean" style="display: none;"></div>
      <div class="sc-results-container" role="listbox">
        <div class="sc-empty-state">${escapeHtml(t.typeToSearch)} ${escapeHtml(indexName)}...</div>
      </div>
      <div class="sc-footer">
        <div class="sc-footer-shortcuts">
          <span><kbd class="sc-kbd">↑</kbd> <kbd class="sc-kbd">↓</kbd> ${escapeHtml(t.navigate)}</span>
          <span><kbd class="sc-kbd">↵</kbd> ${escapeHtml(t.select)}</span>
          <span><kbd class="sc-kbd">ESC</kbd> ${escapeHtml(t.close)}</span>
        </div>
        <div class="sc-footer-brand">
          ${escapeHtml(t.poweredBy)} <span>SearchCraft</span>
        </div>
      </div>
    </div>
  `;

  document.body.appendChild(overlay);

  const input = overlay.querySelector(".sc-input") as HTMLInputElement;
  const spinner = overlay.querySelector(".sc-spinner") as HTMLElement;
  const resultsContainer = overlay.querySelector(".sc-results-container") as HTMLElement;
  const suggestionsBar = overlay.querySelector(".sc-suggestions-bar") as HTMLElement;
  const didYouMeanBar = overlay.querySelector(".sc-did-you-mean") as HTMLElement;

  function openModal() {
    if (isOpen) return;
    isOpen = true;
    overlay.classList.add("sc-open");
    input.value = "";
    selectedIndex = 0;
    renderResults([]);
    suggestionsBar.style.display = "none";
    didYouMeanBar.style.display = "none";
    setTimeout(() => input.focus(), 30);
  }

  function closeModal() {
    if (!isOpen) return;
    isOpen = false;
    overlay.classList.remove("sc-open");
    if (activeAbortController) {
      activeAbortController.abort();
    }
  }

  function escapeHtml(str: string): string {
    return str.replace(/[&<>'"]/g, (tag: string) => ({
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      "'": "&#39;",
      '"': "&quot;",
    }[tag] || tag));
  }

  function renderResults(hits: any[]) {
    currentHits = hits;
    selectedIndex = 0;

    if (!input.value.trim()) {
      resultsContainer.innerHTML = `<div class="sc-empty-state">${escapeHtml(t.typeToSearch)} ${escapeHtml(indexName)}...</div>`;
      return;
    }

    if (hits.length === 0) {
      resultsContainer.innerHTML = `<div class="sc-empty-state">${escapeHtml(t.noResults)} "${escapeHtml(input.value)}".</div>`;
      return;
    }

    resultsContainer.innerHTML = hits.map((hit, idx) => {
      const doc = hit.document || {};
      const title = hit.highlights?.title || escapeHtml(doc.title || doc.name || `Document #${hit.id}`);
      const desc = hit.highlights?.description || hit.highlights?.content || escapeHtml(doc.description || doc.content || doc.summary || "");
      const category = doc.category || doc.brand || doc.type || "";
      const isSelected = idx === selectedIndex ? "sc-selected" : "";

      return `
        <div class="sc-result-item ${isSelected}" data-index="${idx}" role="option">
          <div class="sc-result-header">
            <span class="sc-result-title">${title}</span>
            ${category ? `<span class="sc-badge">${escapeHtml(String(category))}</span>` : ""}
          </div>
          ${desc ? `<div class="sc-result-desc">${desc}</div>` : ""}
        </div>
      `;
    }).join("");
  }

  function renderSuggestions(suggestions: string[]) {
    if (!suggestions || suggestions.length === 0) {
      suggestionsBar.style.display = "none";
      return;
    }

    suggestionsBar.style.display = "flex";
    suggestionsBar.innerHTML = `
      <span class="sc-suggestions-label">${escapeHtml(t.suggestions)}</span>
      ${suggestions.map(s => `<span class="sc-suggestion-pill" data-suggestion="${escapeHtml(s)}">${escapeHtml(s)}</span>`).join("")}
    `;
  }

  function renderDidYouMean(candidate: string) {
    if (!candidate) {
      didYouMeanBar.style.display = "none";
      return;
    }

    didYouMeanBar.style.display = "flex";
    didYouMeanBar.innerHTML = `
      <span>${escapeHtml(t.didYouMean)}</span>
      <button class="sc-did-you-mean-btn" data-correction="${escapeHtml(candidate)}">${escapeHtml(candidate)}</button>
    `;
  }

  function updateSelection() {
    const items = resultsContainer.querySelectorAll(".sc-result-item");
    items.forEach((item, idx) => {
      if (idx === selectedIndex) {
        item.classList.add("sc-selected");
        item.scrollIntoView({ block: "nearest", behavior: "smooth" });
      } else {
        item.classList.remove("sc-selected");
      }
    });
  }

  function selectCurrentItem() {
    if (currentHits.length === 0 || selectedIndex < 0 || selectedIndex >= currentHits.length) {
      return;
    }

    const hit = currentHits[selectedIndex];
    const doc = hit.document || {};

    const event = new CustomEvent("searchcraft:select", {
      detail: { hit, document: doc },
      bubbles: true,
    });
    window.dispatchEvent(event);

    const url = doc.url || doc.link || doc.href || doc.path;
    closeModal();
    if (url && typeof url === "string") {
      window.location.href = url;
    }
  }

  async function executeSearch(query: string) {
    if (!query.trim()) {
      renderResults([]);
      suggestionsBar.style.display = "none";
      didYouMeanBar.style.display = "none";
      spinner.classList.remove("sc-loading");
      return;
    }

    if (activeAbortController) {
      activeAbortController.abort();
    }
    activeAbortController = new AbortController();

    spinner.classList.add("sc-loading");

    try {
      const headers: Record<string, string> = { "Content-Type": "application/json" };
      if (apiKey) {
        headers["Authorization"] = `Bearer ${apiKey}`;
      }

      const res = await fetch(`${baseUrl}/api/v1/indexes/${encodeURIComponent(indexName)}/search`, {
        method: "POST",
        headers,
        body: JSON.stringify({
          query: query,
          mode: searchMode,
          limit: limit,
          highlight: true,
        }),
        signal: activeAbortController.signal,
      });

      if (!res.ok) {
        throw new Error(`HTTP error ${res.status}`);
      }

      const data = await res.json();
      renderResults(data.hits || []);
      renderSuggestions(data.suggestions || []);
      renderDidYouMean(data.did_you_mean || "");
    } catch (err: any) {
      if (err.name !== "AbortError") {
        resultsContainer.innerHTML = `<div class="sc-empty-state">Search request failed. Please check network.</div>`;
      }
    } finally {
      spinner.classList.remove("sc-loading");
    }
  }

  // Click suggestions
  suggestionsBar.addEventListener("click", (e: MouseEvent) => {
    const pill = (e.target as HTMLElement).closest("[data-suggestion]") as HTMLElement;
    if (pill) {
      const val = pill.getAttribute("data-suggestion") || "";
      input.value = val;
      executeSearch(val);
    }
  });

  // Click did you mean
  didYouMeanBar.addEventListener("click", (e: MouseEvent) => {
    const btn = (e.target as HTMLElement).closest("[data-correction]") as HTMLElement;
    if (btn) {
      const val = btn.getAttribute("data-correction") || "";
      input.value = val;
      executeSearch(val);
    }
  });

  // Event Listeners
  input.addEventListener("input", (e: any) => {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      executeSearch(e.target.value);
    }, 140);
  });

  input.addEventListener("keydown", (e: KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      if (currentHits.length > 0) {
        selectedIndex = (selectedIndex + 1) % currentHits.length;
        updateSelection();
      }
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      if (currentHits.length > 0) {
        selectedIndex = (selectedIndex - 1 + currentHits.length) % currentHits.length;
        updateSelection();
      }
    } else if (e.key === "Enter") {
      e.preventDefault();
      selectCurrentItem();
    } else if (e.key === "Escape") {
      e.preventDefault();
      closeModal();
    }
  });

  // Click on result item
  resultsContainer.addEventListener("click", (e: MouseEvent) => {
    const item = (e.target as HTMLElement).closest(".sc-result-item") as HTMLElement;
    if (item) {
      const idx = parseInt(item.getAttribute("data-index") || "0", 10);
      selectedIndex = idx;
      selectCurrentItem();
    }
  });

  // Click backdrop to close
  overlay.addEventListener("click", (e: MouseEvent) => {
    if (e.target === overlay) {
      closeModal();
    }
  });

  // Global Keyboard Shortcuts
  window.addEventListener("keydown", (e: KeyboardEvent) => {
    if (shortcutKey !== "none") {
      const isCmdOrCtrl = e.metaKey || e.ctrlKey;
      if (isCmdOrCtrl && e.key.toLowerCase() === shortcutKey) {
        e.preventDefault();
        if (isOpen) closeModal();
        else openModal();
        return;
      }
    }

    if (e.key === "/" && !isOpen) {
      const activeEl = document.activeElement;
      const isInput = activeEl && (activeEl.tagName === "INPUT" || activeEl.tagName === "TEXTAREA" || (activeEl as HTMLElement).isContentEditable);
      if (!isInput) {
        e.preventDefault();
        openModal();
      }
    }
  });

  document.addEventListener("click", (e: MouseEvent) => {
    const trigger = (e.target as HTMLElement).closest("[data-searchcraft-trigger]");
    if (trigger) {
      e.preventDefault();
      openModal();
    }
  });

  // Global SearchCraft API
  (window as any).SearchCraft = {
    open: openModal,
    close: closeModal,
    search: (q: string) => {
      openModal();
      input.value = q;
      executeSearch(q);
    },
  };
})();
