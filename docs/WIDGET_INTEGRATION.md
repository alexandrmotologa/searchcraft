# Drop-In Search Widget Integration

SearchCraft includes a standalone command palette search widget compiled into a single script under 3KB. It requires zero external dependencies, libraries, or CSS stylesheets.

---

## 1. Quick Integration

Add the script tag to your HTML document right before the closing `</body>` tag:

```html
<script
  src="https://search.yourdomain.com/widget.js"
  data-url="https://search.yourdomain.com"
  data-index="documentation"
  data-api-key="sc_live_search_..."
  data-theme="auto"
  data-lang="en"
  data-placeholder="Search documentation (Cmd+K)..."
  data-accent="#3b82f6"
></script>
```

When included, the script listens for the standard shortcut (`Cmd+K` on macOS, `Ctrl+K` on Windows and Linux) and renders an accessible modal command palette.

---

## 2. Configuration Attributes

All configuration options are specified via HTML data attributes on the `<script>` tag:

| Attribute | Type | Default | Description |
|---|---|---|---|
| `data-url` | string | `window.location.origin` | The SearchCraft server endpoint. |
| `data-index` | string | `"products"` | Target index name to query against. |
| `data-api-key` | string | `""` | Scoped public search API key. |
| `data-theme` | string | `"auto"` | Visual theme: `"auto"` (matches system), `"dark"`, or `"light"`. |
| `data-lang` | string | `"en"` | Language localization: `"en"`, `"ro"`, `"es"`, `"fr"`, `"de"`. |
| `data-placeholder` | string | localized | Placeholder text displayed in the search input. |
| `data-accent` | string | `"#3b82f6"` | Primary accent color for badges, borders, and selection. |
| `data-shortcut` | string | `"k"` | Character key to pair with Cmd / Ctrl. Set to `"none"` to disable. |
| `data-limit` | number | `8` | Maximum number of search results displayed in the modal list. |
| `data-mode` | string | `"hybrid"` | Search mode: `"hybrid"`, `"bm25"`, or `"vector"`. |

---

## 3. Interactive Features

### Autocomplete Suggestions
As you type, real-time query suggestions appear below the search bar. Clicking a suggestion pill updates the input and executes the search immediately.

### "Did You Mean?" Typo Correction
If an interogation produces zero hits, SearchCraft computes the closest vocabulary candidate via Damerau-Levenshtein distance and displays a clickable correction pill.

---

## 4. Triggering the Widget Manually

You can open the search palette from any custom button, link, or menu item by adding the `data-searchcraft-trigger` attribute:

```html
<button data-searchcraft-trigger class="search-btn">
  <span>Quick Search</span>
  <kbd>Cmd+K</kbd>
</button>
```

Alternatively, you can interact with the widget programmatically through the global JavaScript window object:

```javascript
// Open search modal
window.SearchCraft.open();

// Close search modal
window.SearchCraft.close();

// Programmatically execute a search
window.SearchCraft.search("keyboard shortcuts");
```

---

## 5. Custom Result Selection Handling

By default, selecting an item navigates to its `url` property if present in the indexed document. If you want custom behavior (such as internal SPA router navigation), define a custom event listener:

```html
<script>
  window.addEventListener("searchcraft:select", function(event) {
    const doc = event.detail.document;
    console.log("Selected document:", doc);
    // Custom router navigation example:
    // router.push(doc.path);
  });
</script>
```

---

## 6. Keyboard Navigation Support

The widget includes built-in keyboard accessibility:
- `Cmd+K` / `Ctrl+K`: Toggle search modal.
- `Escape`: Close search modal.
- `Arrow Down`: Move focus to the next search result.
- `Arrow Up`: Move focus to the previous search result.
- `Enter`: Activate selected item.
- `/`: Open palette when no input is focused.
