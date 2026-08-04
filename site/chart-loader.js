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
    installPajeRestart();
    installHeartMotion();
    document.dispatchEvent(new CustomEvent("araihu:charts-loaded"));
    fragment.remove();
  }

  function installPajeRestart() {
    const art = document.querySelector("[data-paje-actual-chart]");
    const trigger = art?.closest(".project-card");
    const host = art?.querySelector("[_echarts_instance_]");
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)");
    const touchInput = window.matchMedia("(hover: none), (pointer: coarse)");
    if (!trigger || !host || !window.echarts) return;

    const currentChart = () => window.echarts.getInstanceByDom(host);
    const restart = () => {
      if (reducedMotion.matches) return;
      const current = currentChart();
      if (!current) return;
      const option = current.getOption();
      window.echarts.dispose(host);
      const next = window.echarts.init(host);
      next.setOption(option, true);
      next.resize();
    };

    trigger.addEventListener("pointerenter", restart);
    trigger.addEventListener("focusin", restart);
    if ("IntersectionObserver" in window) {
      let visible = false;
      new IntersectionObserver((entries) => {
        const nextVisible = entries.some((entry) => entry.isIntersecting);
        if (touchInput.matches && nextVisible && !visible) restart();
        visible = nextVisible;
      }, { threshold: 0.2 }).observe(trigger);
    }
    window.addEventListener("resize", () => currentChart()?.resize(), { passive: true });
  }

  function installHeartMotion() {
    const art = document.querySelector("[data-goshtoso-heart-chart]");
    const trigger = art?.closest(".more-row");
    const host = art?.querySelector("[_echarts_instance_]");
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)");
    const touchInput = window.matchMedia("(hover: none), (pointer: coarse)");
    if (!trigger || !host || !window.echarts) return;

    const chart = () => window.echarts.getInstanceByDom(host);
    let visible = false;
    let engaged = trigger.matches(":hover") || trigger.contains(document.activeElement);
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
      const next = visible && !reducedMotion.matches && (touchInput.matches || engaged);
      if (next === rotating) return;
      rotating = next;
      chart()?.setOption({
        grid3D: [{ viewControl: { autoRotate: next } }],
      }, { notMerge: false, lazyUpdate: false, silent: true });
      art.dataset.heartMotion = next ? "running" : "paused";
    };
    const setEngaged = (next) => {
      engaged = next;
      sync();
    };

    trigger.addEventListener("pointerenter", () => setEngaged(true));
    trigger.addEventListener("pointerleave", () => setEngaged(false));
    trigger.addEventListener("focusin", () => setEngaged(true));
    trigger.addEventListener("focusout", (event) => {
      if (!trigger.contains(event.relatedTarget)) setEngaged(false);
    });
    reducedMotion.addEventListener("change", sync);
    touchInput.addEventListener("change", sync);
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
