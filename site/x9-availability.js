(() => {
  const bucketCount = 36;
  const tickMilliseconds = 2000;
  const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)");

  function stateAt(value) {
    const phase = value % 24;
    if (phase >= 8 && phase <= 10) return 1;
    if (phase >= 17 && phase <= 19) return 2;
    return 0;
  }

  function clockLabel(timestamp) {
    return new Date(timestamp).toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    });
  }

  function snapshot(step) {
    const end = Date.now();
    const categories = [];
    const series = [[], [], []];
    for (let index = 0; index < bucketCount; index += 1) {
      categories.push(clockLabel(end - (bucketCount - index - 1) * tickMilliseconds));
      const state = stateAt(step + index);
      for (let seriesIndex = 0; seriesIndex < series.length; seriesIndex += 1) {
        series[seriesIndex].push(state === seriesIndex ? 1 : 0);
      }
    }
    return { categories, series };
  }

  function chartFor(root) {
    const host = root.querySelector("[_echarts_instance_]");
    if (!host || !window.echarts) return null;
    return window.echarts.getInstanceByDom(host);
  }

  function start(root) {
    if (root.dataset.x9AvailabilityReady === "true") return;
    root.dataset.x9AvailabilityReady = "true";
    let step = 0;
    let attempts = 0;
    const trigger = root.closest(".project-card") || root;

    function connect() {
      const chart = chartFor(root);
      if (!chart) {
        attempts += 1;
        if (attempts < 100) window.setTimeout(connect, 50);
        return;
      }

      function tick() {
        const data = snapshot(step);
        chart.setOption({
          xAxis: [{ data: data.categories }],
          series: [
            { name: "Healthy", data: data.series[0] },
            { name: "Degraded", data: data.series[1] },
            { name: "Down", data: data.series[2] },
          ],
        }, { notMerge: false, lazyUpdate: false, silent: true });
        root.dataset.x9Categories = String(bucketCount);
        root.dataset.x9Tick = String(step);
        step += 1;
      }

      let active = trigger.matches(":hover") || trigger.contains(document.activeElement);
      let interval = 0;
      const stopTicks = () => {
        if (!interval) return;
        window.clearInterval(interval);
        interval = 0;
      };
      const syncTicks = () => {
        if (!active || reducedMotion.matches) {
          stopTicks();
          return;
        }
        if (interval) return;
        tick();
        interval = window.setInterval(tick, tickMilliseconds);
      };
      const setActive = (nextActive) => {
        active = nextActive;
        syncTicks();
      };

      tick();
      trigger.addEventListener("pointerenter", () => setActive(true));
      trigger.addEventListener("pointerleave", () => setActive(false));
      trigger.addEventListener("focusin", () => setActive(true));
      trigger.addEventListener("focusout", (event) => {
        if (!trigger.contains(event.relatedTarget)) setActive(false);
      });
      reducedMotion.addEventListener("change", syncTicks);
      syncTicks();
    }

    connect();
  }

  function initialize() {
    document.querySelectorAll("[data-x9-live-availability]").forEach(start);
  }

  initialize();
  document.addEventListener("htmx:afterSettle", initialize);
  document.addEventListener("araihu:charts-loaded", initialize);
})();
