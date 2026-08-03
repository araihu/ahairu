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

    fragment.querySelectorAll("template[data-chart-target]").forEach((template) => {
      const target = document.querySelector(template.dataset.chartTarget);
      if (!target) return;
      window.htmx.swap(target, template.innerHTML.trim(), {
        swapStyle: "outerHTML",
        settleDelay: 0,
      });
    });

    window.htmx.process(document.body);
    document.dispatchEvent(new CustomEvent("araihu:charts-loaded"));
    fragment.remove();
  }

  hydrateCharts().catch(() => {
    fragment.dataset.chartLoadFailed = "true";
  });
})();
