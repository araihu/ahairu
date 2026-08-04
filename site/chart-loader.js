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
    installPajeLifecycleMotion();
    installHeartMotion();
    document.dispatchEvent(new CustomEvent("araihu:charts-loaded"));
    fragment.remove();
  }

  function installPajeLifecycleMotion() {
    const art = document.querySelector("[data-paje-actual-chart]");
    const trigger = art?.closest(".project-card");
    const host = art?.querySelector("[_echarts_instance_]");
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)");
    const touchInput = window.matchMedia("(hover: none), (pointer: coarse)");
    if (!trigger || !host || !window.echarts) return;

    const chart = window.echarts.getInstanceByDom(host);
    const series = chart?.getOption().series?.[0];
    const baseNodes = series?.data;
    const baseLinks = series?.links;
    if (!chart || !Array.isArray(baseNodes) || !Array.isArray(baseLinks)) return;

    const waves = [[0], [1], [2, 3], [4], [5, 6, 7], [8], [9], [10], [11]];
    const indexByName = new Map(baseNodes.map((node, index) => [node.name, index]));
    const baseLine = series.lineStyle || {};
    const baseLineOpacity = Number(baseLine.opacity ?? 0.72);
    const baseLineWidth = Number(baseLine.width ?? 2);
    const waveDelay = 60;
    let timers = [];
    let run = 0;

    const finiteSize = (value) => {
      const numeric = Number(value);
      return Number.isFinite(numeric) && numeric > 0 ? numeric : 10;
    };
    const scaleSize = (size, scale) => Array.isArray(size)
      ? size.map((value) => finiteSize(value) * scale)
      : finiteSize(size) * scale;

    const frame = (active, current, duration = 180) => {
      const currentSet = new Set(current);
      chart.setOption({
        animationDurationUpdate: duration,
        animationEasingUpdate: "cubicOut",
        series: [{
          data: baseNodes.map((node, index) => {
            const isActive = active.has(index);
            const isCurrent = currentSet.has(index);
            const color = node.itemStyle?.color || "#b8ff39";
            return {
              ...node,
              symbolSize: scaleSize(node.symbolSize, isCurrent ? 1.24 : (isActive ? 1 : 0.72)),
              itemStyle: {
                ...node.itemStyle,
                opacity: isActive ? 1 : 0.22,
                borderWidth: isCurrent ? 4 : (node.itemStyle?.borderWidth ?? 2),
                shadowBlur: isCurrent ? 22 : 0,
                shadowColor: isCurrent ? color : "transparent",
              },
            };
          }),
          links: baseLinks.map((link) => {
            const targetIndex = indexByName.get(link.target);
            const isActive = active.has(targetIndex);
            const isCurrent = currentSet.has(targetIndex);
            return {
              ...link,
              lineStyle: {
                ...link.lineStyle,
                opacity: isCurrent ? 1 : (isActive ? baseLineOpacity : 0.1),
                width: isCurrent ? baseLineWidth + 1 : baseLineWidth,
              },
            };
          }),
        }],
      }, { notMerge: false, lazyUpdate: false, silent: true });
    };

    const cancel = () => {
      run += 1;
      timers.forEach(window.clearTimeout);
      timers = [];
    };

    const settle = (duration = 0) => {
      frame(new Set(baseNodes.map((_, index) => index)), [], duration);
      art.dataset.pajeMotion = "idle";
      delete art.dataset.pajeMotionStep;
    };

    const animate = () => {
      if (reducedMotion.matches) return;
      cancel();
      const currentRun = run;
      const active = new Set();
      art.dataset.pajeMotion = "running";
      art.dataset.pajeMotionStep = "0";
      frame(active, [], 100);

      waves.forEach((wave, waveIndex) => {
        timers.push(window.setTimeout(() => {
          if (currentRun !== run) return;
          wave.forEach((index) => active.add(index));
          art.dataset.pajeMotionStep = String(waveIndex + 1);
          frame(active, wave);
        }, waveDelay * (waveIndex + 1)));
      });

      timers.push(window.setTimeout(() => {
        if (currentRun !== run) return;
        settle(120);
      }, waveDelay * (waves.length + 1) + 80));
    };

    art.dataset.pajeMotion = "idle";
    trigger.addEventListener("pointerenter", animate);
    trigger.addEventListener("focusin", animate);
    reducedMotion.addEventListener("change", () => {
      if (!reducedMotion.matches) return;
      cancel();
      settle();
    });
    if ("IntersectionObserver" in window) {
      let visible = false;
      new IntersectionObserver((entries) => {
        const nextVisible = entries.some((entry) => entry.isIntersecting);
        if (touchInput.matches && nextVisible && !visible) animate();
        visible = nextVisible;
      }, { threshold: 0.2 }).observe(trigger);
    }
    window.addEventListener("resize", () => chart.resize(), { passive: true });
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
