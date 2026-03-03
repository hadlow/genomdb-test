const nodes = [
  { id: "node1", name: "Node 1", url: "http://localhost:8001" },
  { id: "node2", name: "Node 2", url: "http://localhost:8002" },
  { id: "node3", name: "Node 3", url: "http://localhost:8003" },
];

const refreshBtn = document.getElementById("refreshBtn");
const lastUpdated = document.getElementById("lastUpdated");
const nodesTable = document.getElementById("nodesTable");
const keysContainer = document.getElementById("keysContainer");
const summary = document.getElementById("summary");

async function fetchNodeState(node) {
  const pingURL = `${node.url}/ping`;
  const statusURL = `${node.url}/status`;

  let pingOk = false;
  let pingLatencyMs = null;

  const pingStart = performance.now();
  try {
    const pingRes = await fetch(pingURL, { method: "GET" });
    pingOk = pingRes.ok;
    pingLatencyMs = Math.round(performance.now() - pingStart);
  } catch {
    pingOk = false;
  }

  try {
    const statusRes = await fetch(statusURL, { method: "GET" });
    if (!statusRes.ok) {
      return { node, healthy: false, pingOk, pingLatencyMs, error: `status HTTP ${statusRes.status}` };
    }
    const status = await statusRes.json();
    return {
      node,
      healthy: pingOk,
      pingOk,
      pingLatencyMs,
      status,
    };
  } catch (error) {
    return {
      node,
      healthy: false,
      pingOk,
      pingLatencyMs,
      error: error instanceof Error ? error.message : "failed to fetch status",
    };
  }
}

function renderSummary(states) {
  const healthyCount = states.filter((s) => s.healthy).length;
  const leaders = states.filter((s) => s.status?.raft_state === "Leader");
  const totalKeys = states.reduce((sum, s) => sum + (s.status?.key_count || 0), 0);

  summary.innerHTML = [
    card("Total Nodes", `${states.length}`),
    card("Healthy Nodes", `${healthyCount}`),
    card("Leaders Seen", `${leaders.length}`),
    card("Total Keys (sum)", `${totalKeys}`),
  ].join("");
}

function card(label, value) {
  return `
    <div class="summary-card">
      <div class="label">${label}</div>
      <div class="value">${value}</div>
    </div>
  `;
}

function renderNodesTable(states) {
  nodesTable.innerHTML = states
    .map((s) => {
      const role = s.status?.raft_state || "Unknown";
      const roleClass = role === "Leader" ? "leader" : "follower";
      const peersCount = s.status?.peers?.length ?? 0;
      const leader = s.status?.raft_leader || "-";
      const raftAddr = s.status?.raft_advertise_address || s.status?.raft_bind_address || "-";
      const httpAddr = s.status?.http_address || s.node.url.replace("http://", "");
      const keyCount = s.status?.key_count ?? 0;
      const healthBadge = s.healthy
        ? `<span class="badge healthy">healthy (${s.pingLatencyMs ?? "-"}ms)</span>`
        : `<span class="badge unhealthy">unhealthy</span>`;

      return `
        <tr>
          <td>${s.status?.node_id || s.node.id}</td>
          <td>${healthBadge}</td>
          <td><span class="badge ${roleClass}">${role}</span></td>
          <td>${httpAddr}</td>
          <td>${raftAddr}</td>
          <td>${leader}</td>
          <td>${peersCount}</td>
          <td>${keyCount}</td>
        </tr>
      `;
    })
    .join("");
}

function renderKeys(states) {
  keysContainer.innerHTML = states
    .map((s) => {
      const nodeName = s.status?.node_id || s.node.id;
      if (!s.status?.store) {
        return `
          <div class="keys-card">
            <h3>${nodeName}</h3>
            <div class="muted">No data available (${s.error || "node unavailable"})</div>
          </div>
        `;
      }

      const entries = Object.entries(s.status.store);
      if (entries.length === 0) {
        return `
          <div class="keys-card">
            <h3>${nodeName}</h3>
            <div class="muted">No keys stored</div>
          </div>
        `;
      }

      const keysHTML = entries
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([key, value]) => `<li><strong>${key}</strong>: <span class="kv-value">${escapeHTML(String(value))}</span></li>`)
        .join("");

      return `
        <div class="keys-card">
          <h3>${nodeName}</h3>
          <ul class="keys-list">${keysHTML}</ul>
        </div>
      `;
    })
    .join("");
}

function escapeHTML(value) {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

async function refresh() {
  refreshBtn.disabled = true;
  const states = await Promise.all(nodes.map(fetchNodeState));

  renderSummary(states);
  renderNodesTable(states);
  renderKeys(states);

  lastUpdated.textContent = `Last updated: ${new Date().toLocaleTimeString()}`;
  refreshBtn.disabled = false;
}

refreshBtn.addEventListener("click", refresh);

refresh();
setInterval(refresh, 5000);
