(() => {
  const loader = document.currentScript;
  const fragment = loader && loader.closest("[data-chart-fragment]");
  if (!loader || !fragment || !window.htmx) return;

  function runtimeScript(src) {
    const absolute = new URL(src, window.location.href).href;
    return Array.from(document.scripts).find((script) => script.src === absolute);
  }

  function loadRuntime(src, ready) {
    if (ready()) return Promise.resolve();

    const existing = runtimeScript(src);
    if (existing) {
      if (existing.dataset.chartRuntimeReady === "true") return Promise.resolve();
      return new Promise((resolve, reject) => {
        existing.addEventListener("load", resolve, { once: true });
        existing.addEventListener("error", reject, { once: true });
      });
    }

    return new Promise((resolve, reject) => {
      const script = document.createElement("script");
      script.src = src;
      script.async = false;
      script.addEventListener("load", () => {
        script.dataset.chartRuntimeReady = "true";
        resolve();
      }, { once: true });
      script.addEventListener("error", reject, { once: true });
      document.head.appendChild(script);
    });
  }

  async function hydrateCharts() {
    await loadRuntime(loader.dataset.echartsSrc, () => Boolean(window.echarts));
    await loadRuntime(loader.dataset.threeDSrc, () => false);

    for (const template of fragment.querySelectorAll("template[data-chart-target]")) {
      const target = document.querySelector(template.dataset.chartTarget);
      if (!target) continue;
      await window.htmx.swap({
        sourceElement: target,
        target: template.dataset.chartTarget,
        text: template.innerHTML.trim(),
        swap: "outerHTML settle:0",
      });
    }

    window.htmx.process(document.body);
    installHeartMotion();
    document.dispatchEvent(new CustomEvent("araihu:charts-loaded"));
    fragment.remove();
  }

  function installHeartMotion() {
    const art = document.querySelector("[data-goshtoso-heart-chart]");
    const trigger = art?.closest(".featured-family-card, .more-row");
    const host = art?.querySelector("[_echarts_instance_]");
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)");
    if (!trigger || !host || !window.echarts) return;

    const chart = () => window.echarts.getInstanceByDom(host);
    let visible = false;
    let rotating = false;

    // Surface3D owns rendering and topology. These presentation-only overrides
    // keep its 3D grid while removing the legend and axis labels from the card.
    chart()?.setOption({
      legend: [{ show: false }],
      xAxis3D: [{ show: false }],
      yAxis3D: [{ show: false }],
      zAxis3D: [{ show: false }],
      grid3D: [{ show: true, viewControl: { autoRotateSpeed: 12 } }],
    }, { notMerge: false, lazyUpdate: false, silent: true });
    art.dataset.heartFramed = "true";
    art.dataset.heartMotion = "paused";

    const sync = () => {
      const next = visible && !reducedMotion.matches;
      if (next === rotating) return;
      rotating = next;
      chart()?.setOption({
        grid3D: [{ viewControl: { autoRotate: next } }],
      }, { notMerge: false, lazyUpdate: false, silent: true });
      art.dataset.heartMotion = next ? "running" : "paused";
    };
    reducedMotion.addEventListener("change", sync);
    window.addEventListener("resize", () => chart()?.resize(), { passive: true });

    if ("IntersectionObserver" in window) {
      new IntersectionObserver((entries) => {
        visible = entries.some((entry) => entry.isIntersecting);
        sync();
      }, { threshold: 0.2 }).observe(trigger);
    } else {
      visible = true;
    }
    sync();
  }

  hydrateCharts().catch(() => {
    fragment.dataset.chartLoadFailed = "true";
  });
})();
