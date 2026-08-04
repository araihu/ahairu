(() => {
  const video = document.querySelector("[data-storm-backdrop]");
  const montage = document.querySelector("[data-featured-montage]");
  if (!video && !montage) return;

  const hero = video?.closest(".storm-hero");
  const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)");
  const darkMode = window.matchMedia("(prefers-color-scheme: dark)");
  const touchInput = window.matchMedia("(hover: none), (pointer: coarse)");
  const saveData = navigator.connection?.saveData === true;
  let animationFrame = 0;
  let montageActive = false;
  let montageVisible = false;

  const updateParallax = () => {
    animationFrame = 0;
    if (!hero) return;
    if (reducedMotion.matches) {
      hero.style.removeProperty("--storm-parallax");
      return;
    }
    const offset = Math.max(-48, Math.min(48, -hero.getBoundingClientRect().top * 0.12));
    hero.style.setProperty("--storm-parallax", `${offset.toFixed(1)}px`);
  };

  const scheduleParallax = () => {
    if (!animationFrame) animationFrame = window.requestAnimationFrame(updateParallax);
  };

  const syncPlayback = () => {
    if (video) {
      if (reducedMotion.matches || saveData) {
        video.pause();
        hero.style.removeProperty("--storm-parallax");
      } else {
        if (video.networkState === HTMLMediaElement.NETWORK_EMPTY) video.load();
        video.play().catch(() => {});
        scheduleParallax();
      }
    }
    if (montage) {
      const montageEngaged = touchInput.matches ? montageVisible : montageActive;
      if (reducedMotion.matches || saveData || !montageVisible || !montageEngaged) {
        montage.pause();
      } else {
        if (montage.networkState === HTMLMediaElement.NETWORK_EMPTY) montage.load();
        montage.play().catch(() => {});
      }
    }
  };

  reducedMotion.addEventListener("change", syncPlayback);
  touchInput.addEventListener("change", syncPlayback);
  darkMode.addEventListener("change", () => {
    video?.load();
    syncPlayback();
  });
  window.addEventListener("scroll", scheduleParallax, { passive: true });
  window.addEventListener("resize", scheduleParallax, { passive: true });
  const montageTrigger = montage?.closest(".featured-visual");
  if (montageTrigger) {
    const setMontageActive = (active) => {
      montageActive = active;
      syncPlayback();
    };
    montageTrigger.addEventListener("pointerenter", () => setMontageActive(true));
    montageTrigger.addEventListener("pointerleave", () => setMontageActive(false));
    montageTrigger.addEventListener("focusin", () => setMontageActive(true));
    montageTrigger.addEventListener("focusout", (event) => {
      if (!montageTrigger.contains(event.relatedTarget)) setMontageActive(false);
    });
  }
  if (montage && "IntersectionObserver" in window) {
    new IntersectionObserver((entries) => {
      montageVisible = entries.some((entry) => entry.isIntersecting);
      syncPlayback();
    }, { threshold: 0.2 }).observe(montage);
  } else {
    montageVisible = Boolean(montage);
  }
  syncPlayback();

  const cards = document.querySelectorAll(".project-card, .more-row");
  const cardVisibility = new Map();
  const syncTouchCard = (card, active) => {
    card.toggleAttribute("data-card-viewport-active", active && touchInput.matches && !reducedMotion.matches);
  };
  if ("IntersectionObserver" in window) {
    const cardObserver = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        cardVisibility.set(entry.target, entry.isIntersecting);
        syncTouchCard(entry.target, entry.isIntersecting);
      });
    }, { threshold: 0.2 });
    cards.forEach((card) => cardObserver.observe(card));
    const refreshCards = () => cards.forEach((card) => syncTouchCard(card, cardVisibility.get(card) === true));
    reducedMotion.addEventListener("change", refreshCards);
    touchInput.addEventListener("change", refreshCards);
  }
})();
