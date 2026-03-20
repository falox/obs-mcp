// MCP Apps trace waterfall application.
// Handles MCP Apps lifecycle, span tree construction, and waterfall rendering.
(function() {
  "use strict";

  // ===== State =====
  var hostContext = null;
  var requestId = 1;
  var collapsedSpans = {};

  // ===== Color Palettes (PatternFly v6 multi-color ordered chart colors) =====
  var LIGHT_COLORS = [
    "#0066cc", "#63993d", "#37a3a3", "#ca6c0f", "#9e4a06",
    "#004d99", "#3d7317", "#147878", "#b98412", "#96640f"
  ];
  var DARK_COLORS = [
    "#4394e5", "#87bb62", "#63bdbd", "#ffcc17", "#f5921b",
    "#92c5f9", "#afdc8f", "#9ad8d8", "#ffe072", "#f8ae54"
  ];

  // ===== Theme =====
  function isDark() {
    return hostContext && hostContext.theme === "dark";
  }

  function getColors() {
    return isDark() ? DARK_COLORS : LIGHT_COLORS;
  }

  function applyTheme() {
    document.documentElement.setAttribute("data-theme", isDark() ? "dark" : "light");
  }

  // ===== JSON-RPC Messaging =====
  function send(msg) {
    window.parent.postMessage(msg, "*");
  }

  function sendRequest(method, params) {
    var id = requestId++;
    send({ jsonrpc: "2.0", method: method, id: id, params: params || {} });
    return id;
  }

  function sendNotification(method, params) {
    send({ jsonrpc: "2.0", method: method, params: params || {} });
  }

  // ===== Duration Formatting =====
  function formatDuration(ms) {
    if (ms >= 1000) return (ms / 1000).toFixed(2) + "s";
    if (ms >= 1) return ms.toFixed(2) + "ms";
    return (ms * 1000).toFixed(0) + "\u00b5s";
  }

  // ===== Span Tree =====
  function buildSpanTree(spans) {
    var byId = {};
    var roots = [];

    for (var i = 0; i < spans.length; i++) {
      byId[spans[i].spanId] = { span: spans[i], children: [] };
    }

    for (var j = 0; j < spans.length; j++) {
      var s = spans[j];
      var parent = byId[s.parentSpanId];
      if (parent) {
        parent.children.push(byId[s.spanId]);
      } else {
        roots.push(byId[s.spanId]);
      }
    }

    // Sort children by start time
    function sortChildren(node) {
      node.children.sort(function(a, b) {
        return a.span.startTimeMs - b.span.startTimeMs;
      });
      for (var k = 0; k < node.children.length; k++) {
        sortChildren(node.children[k]);
      }
    }

    roots.sort(function(a, b) {
      return a.span.startTimeMs - b.span.startTimeMs;
    });

    for (var r = 0; r < roots.length; r++) {
      sortChildren(roots[r]);
    }

    return roots;
  }

  // Flatten tree into ordered list with depth info
  function flattenTree(roots) {
    var result = [];
    function walk(node, depth) {
      result.push({ span: node.span, depth: depth, hasChildren: node.children.length > 0 });
      if (!collapsedSpans[node.span.spanId]) {
        for (var i = 0; i < node.children.length; i++) {
          walk(node.children[i], depth + 1);
        }
      }
    }
    for (var i = 0; i < roots.length; i++) {
      walk(roots[i], 0);
    }
    return result;
  }

  // ===== Service Color Map =====
  function buildServiceColorMap(spans) {
    var services = [];
    var seen = {};
    for (var i = 0; i < spans.length; i++) {
      var svc = spans[i].serviceName || "";
      if (svc && !seen[svc]) {
        seen[svc] = true;
        services.push(svc);
      }
    }
    var palette = getColors();
    var map = {};
    for (var j = 0; j < services.length; j++) {
      map[services[j]] = palette[j % palette.length];
    }
    return map;
  }

  // ===== Rendering =====
  var lastSpans = null;

  function renderWaterfall(spans) {
    lastSpans = spans;
    var container = document.getElementById("trace-waterfall");
    var header = document.getElementById("trace-header");

    if (!spans || spans.length === 0) {
      container.innerHTML = '<div class="no-data">No spans to display</div>';
      header.innerHTML = "";
      return;
    }

    // Compute trace bounds
    var traceStart = Infinity, traceEnd = -Infinity;
    for (var i = 0; i < spans.length; i++) {
      var s = spans[i];
      if (s.startTimeMs < traceStart) traceStart = s.startTimeMs;
      var end = s.startTimeMs + s.durationMs;
      if (end > traceEnd) traceEnd = end;
    }
    var traceDuration = traceEnd - traceStart;

    header.innerHTML = "Duration: <strong>" + formatDuration(traceDuration) + "</strong> &middot; " + spans.length + " spans";

    var tree = buildSpanTree(spans);
    var flat = flattenTree(tree);
    var serviceColors = buildServiceColorMap(spans);

    container.innerHTML = "";

    for (var f = 0; f < flat.length; f++) {
      var item = flat[f];
      var span = item.span;
      var depth = item.depth;

      var row = document.createElement("div");
      row.className = "span-row";
      if (span.statusCode === 2) row.classList.add("error");

      // Left: operation info
      var info = document.createElement("div");
      info.className = "span-info";

      // Indentation
      if (depth > 0) {
        var indent = document.createElement("span");
        indent.className = "span-indent";
        indent.style.width = (depth * 16) + "px";
        info.appendChild(indent);
      }

      // Collapse toggle
      var toggle = document.createElement("span");
      toggle.className = "span-collapse" + (item.hasChildren ? "" : " leaf");
      toggle.textContent = collapsedSpans[span.spanId] ? "\u25b6" : "\u25bc";
      (function(spanId) {
        toggle.addEventListener("click", function() {
          collapsedSpans[spanId] = !collapsedSpans[spanId];
          renderWaterfall(lastSpans);
        });
      })(span.spanId);
      info.appendChild(toggle);

      var op = document.createElement("span");
      op.className = "span-operation";
      op.textContent = span.operationName;
      op.title = span.operationName;
      info.appendChild(op);

      if (span.serviceName) {
        var badge = document.createElement("span");
        badge.className = "span-service";
        badge.textContent = span.serviceName;
        info.appendChild(badge);
      }

      row.appendChild(info);

      // Right: duration bar
      var barArea = document.createElement("div");
      barArea.className = "span-bar-area";

      var track = document.createElement("div");
      track.className = "span-bar-track";

      var bar = document.createElement("div");
      bar.className = "span-bar";

      var color = serviceColors[span.serviceName] || getColors()[0];
      bar.style.backgroundColor = color;

      var offsetPct = traceDuration > 0 ? ((span.startTimeMs - traceStart) / traceDuration) * 100 : 0;
      var widthPct = traceDuration > 0 ? (span.durationMs / traceDuration) * 100 : 100;
      bar.style.left = offsetPct + "%";
      bar.style.width = Math.max(widthPct, 0.3) + "%";

      var durationText = formatDuration(span.durationMs);
      var pctText = traceDuration > 0 ? " (" + (span.durationMs / traceDuration * 100).toFixed(2) + "%)" : "";

      var label = document.createElement("span");
      label.className = "span-bar-label";

      // Place label inside bar if bar is wide enough, otherwise outside
      if (widthPct > 15) {
        label.classList.add("inside");
        label.textContent = durationText + pctText;
        bar.appendChild(label);
      } else {
        label.classList.add("outside");
        label.textContent = durationText + pctText;
        bar.appendChild(label);
      }

      track.appendChild(bar);
      barArea.appendChild(track);
      row.appendChild(barArea);

      container.appendChild(row);
    }

    // Report size after rendering so the host can resize the iframe
    requestAnimationFrame(function() {
      var card = document.getElementById("card");
      var style = getComputedStyle(card);
      var marginW = parseFloat(style.marginLeft) + parseFloat(style.marginRight);
      var marginH = parseFloat(style.marginTop) + parseFloat(style.marginBottom);
      sendNotification("ui/notifications/size-changed", {
        width: Math.round(card.scrollWidth + marginW),
        height: Math.round(card.scrollHeight + marginH)
      });
    });
  }

  // ===== Message Handler =====
  window.addEventListener("message", function(e) {
    var msg = e.data;
    if (!msg || !msg.jsonrpc) return;

    // Response to ui/initialize
    if (msg.id === 1 && msg.result) {
      hostContext = msg.result.hostContext || {};
      applyTheme();
      sendNotification("ui/notifications/initialized");
      return;
    }

    // Tool input: capture title and description
    if (msg.method === "ui/notifications/tool-input") {
      var input = msg.params || {};
      var args = input.arguments || input.input || input;
      var titleEl = document.getElementById("trace-title");
      if (args.title) {
        titleEl.textContent = args.title;
        titleEl.classList.add("visible");
      } else {
        titleEl.textContent = "";
        titleEl.classList.remove("visible");
      }
      var descEl = document.getElementById("trace-description");
      if (args.description) {
        descEl.textContent = args.description;
        descEl.classList.add("visible");
      } else {
        descEl.textContent = "";
        descEl.classList.remove("visible");
      }
      return;
    }

    // Tool result: render waterfall
    if (msg.method === "ui/notifications/tool-result") {
      var params = msg.params || {};
      var sc = params.structuredContent;
      if (sc && sc.spans) {
        document.getElementById("card").classList.add("active");
        collapsedSpans = {};
        renderWaterfall(sc.spans);
      }
      return;
    }

    // Theme change
    if (msg.method === "ui/notifications/host-context-changed") {
      hostContext = msg.params || hostContext;
      applyTheme();
      if (lastSpans) renderWaterfall(lastSpans);
      return;
    }

    // Cleanup
    if (msg.method === "ui/resource-teardown") {
      document.getElementById("card").classList.remove("active");
      document.getElementById("trace-waterfall").innerHTML = "";
      document.getElementById("trace-header").innerHTML = "";
      document.getElementById("trace-title").textContent = "";
      document.getElementById("trace-title").classList.remove("visible");
      document.getElementById("trace-description").textContent = "";
      document.getElementById("trace-description").classList.remove("visible");
      sendNotification("ui/notifications/size-changed", { width: 0, height: 0 });
      lastSpans = null;
      collapsedSpans = {};
      return;
    }
  });

  // ===== Resize Observer =====
  var resizeObserver = new ResizeObserver(function(entries) {
    var card = entries[0].target;
    var style = getComputedStyle(card);
    var marginW = parseFloat(style.marginLeft) + parseFloat(style.marginRight);
    var marginH = parseFloat(style.marginTop) + parseFloat(style.marginBottom);
    sendNotification("ui/notifications/size-changed", {
      width: Math.round(card.offsetWidth + marginW),
      height: Math.round(card.offsetHeight + marginH)
    });
  });
  resizeObserver.observe(document.getElementById("card"));

  // ===== Start MCP Apps Lifecycle =====
  sendNotification("ui/notifications/size-changed", { width: 0, height: 0 });
  sendRequest("ui/initialize", {
    appCapabilities: {
      availableDisplayModes: ["inline"]
    }
  });
})();
