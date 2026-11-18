"use strict";
(() => {
  const supervisorLabel = document.getElementById("dashboard-team-supervisor");
  if (!supervisorLabel) return;

  let busy = false;

  async function updateSupervisor() {
    if (busy || document.hidden) return;
    busy = true;
    try {
      const response = await fetch("/api/student-placement", {
        credentials: "same-origin",
        cache: "no-store",
        signal: AbortSignal.timeout(12000)
      });
      if (!response.ok) throw new Error("Status unavailable");

      const placement = await response.json();
      supervisorLabel.textContent =
        placement.supervision === "Supervisor assigned" && placement.teacher
          ? placement.teacher
          : placement.supervision;
    } catch (_) {
      // Keep the existing dashboard label if the status endpoint is temporarily unavailable.
    } finally {
      busy = false;
    }
  }

  updateSupervisor();
  window.addEventListener("focus", updateSupervisor);
  window.addEventListener("placement-updated", updateSupervisor);
  window.addEventListener("team-members-changed", updateSupervisor);
  document.addEventListener("visibilitychange", updateSupervisor);
  window.setInterval(updateSupervisor, 15000);
})();
