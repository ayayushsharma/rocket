// Dynamic app launcher with:
// - fetch() -> update apps
// - fuzzy search
// - keyboard navigation (↑/↓/Enter)
// - dark mode toggle persisted in localStorage
// - loading & error handling
// - debounced input for smoother UX

const searchBar  = document.getElementById('searchBar');
const appList    = document.getElementById('appList');
const darkModeToggle = document.getElementById('darkModeToggle');
const statusBar  = document.getElementById('statusBar'); // optional <div id="statusBar"></div>

let apps = [];
let filteredApps = [];
let selectedIndex = 0;
let darkMode = false;
let appPort = location.port || (location.protocol === 'https:' ? '443' : '80');
appPort = `:${appPort}`


// --- Utils ---
function fuzzyMatch(str, pattern) {
    const re = pattern.split('').reduce((a, b) => a + '.*' + b, '');
    return new RegExp(re, 'i').test(str);
}

function renderApps() {
    appList.innerHTML = '';
    filteredApps.forEach((app, idx) => {
        const li = document.createElement('li');
        if (idx === selectedIndex) li.classList.add('selected');
        const a = document.createElement('a');
        a.href = app.url;
        a.textContent = app.name;
        li.appendChild(a);
        appList.appendChild(li);
    });
}

function filterApps() {
    const query = searchBar.value.trim();
    filteredApps = !query ? apps : apps.filter(app => fuzzyMatch(app.name, query));
    selectedIndex = 0;
    renderApps();
}

function setStatus(text, type = 'info') {
    if (!statusBar) return;
    statusBar.textContent = text || '';
    statusBar.className = type; // e.g., .info, .error for styling
}

// --- Debounce for input ---
function debounce(fn, delay = 200) {
    let t;
    return (...args) => {
        clearTimeout(t);
        t = setTimeout(() => fn(...args), delay);
    };
}


// {
//   "excalidraw.localhost": {
//     "ContainerURL": "http://rocket-excalidraw-latest:80",
//     "AppName": "Excalidraw",
//     "Description": ""
//   }
// }

// --- Fetch & update apps ---
async function pullData() {
    setStatus('Loading apps…');
    try {
        const res = await fetch('/static/application.json', { cache: 'no-store' });
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data = await res.json();
        const fetchedApps = Object.entries(data).map(
            ([url, name]) => {
            return {
                name: name.AppName,
                url: "http://" + url + appPort,
                description: name.Description
            }
        });

        console.log(fetchedApps)

        apps = fetchedApps;

        filteredApps = apps;
        renderApps();
        setStatus(`Loaded ${apps.length} app${apps.length === 1 ? '' : 's'}.`);
    } catch (err) {
        console.error('Failed to load apps:', err);
        setStatus('Failed to load apps. Showing any cached/default entries.', 'error');

        // Fallback defaults if desired:
        if (apps.length === 0) {
            apps = [
                { name: "Excalidraw", url: "http://draw.localhost" + appPort },
                { name: "Swagger Editor (Next Gen)", url: "http://swagger.localhost" + appPort },
                { name: "Swagger Editor (Legacy)", url: "http://legacy.swagger.localhost" + appPort},
                { name: "DrawSQL", url: "http://sqldraw.localhost" + appPort },
            ];
            filteredApps = apps;
            renderApps();
        }
    }
}

// Optional helper to dedupe by name (if merging defaults + fetched)
function dedupeByName(list) {
    const seen = new Set();
    const out = [];
    for (const item of list) {
        const key = item.name.toLowerCase();
        if (!seen.has(key)) {
            seen.add(key);
            out.push(item);
        }
    }
    return out;
}

// --- Event wiring ---
const debouncedFilter = debounce(filterApps, 120);
searchBar.addEventListener('input', debouncedFilter);

searchBar.addEventListener('keydown', function(e) {
    if (filteredApps.length === 0) return;
    if (e.key === 'ArrowDown') {
        selectedIndex = (selectedIndex + 1) % filteredApps.length;
        renderApps();
        e.preventDefault();
    } else if (e.key === 'ArrowUp') {
        selectedIndex = (selectedIndex - 1 + filteredApps.length) % filteredApps.length;
        renderApps();
        e.preventDefault();
    } else if (e.key === 'Enter') {
        const target = filteredApps[selectedIndex];
        if (target) window.location.href = target.url;
    }
});

// --- Dark Mode ---
if (localStorage.getItem('darkMode') === 'true') {
    document.body.classList.add('dark-mode');
    darkMode = true;
    darkModeToggle.textContent = '☀️';
}

darkModeToggle.addEventListener('click', function() {
    darkMode = !darkMode;
    if (darkMode) {
        document.body.classList.add('dark-mode');
        darkModeToggle.textContent = '☀️';
    } else {
        document.body.classList.remove('dark-mode');
        darkModeToggle.textContent = '🌙';
    }
    localStorage.setItem('darkMode', darkMode);
});

// --- Initialize ---
window.addEventListener('load', async () => {
    searchBar.focus();
    await pullData();
});
