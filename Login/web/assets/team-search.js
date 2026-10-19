
(() => {
  const section = document.getElementById("team");
  if (!section) return;

section.innerHTML = `
  <article class="card team-directory-card">
    <div class="team-header">
      <div>
        <p class="eyebrow">BUILD YOUR RESEARCH TEAM</p>
        <h2>Find Students & Build Your Team</h2>
        <p class="description">
          Search students by name, department, skills or interests and invite them to your project.
        </p>
      </div>
      <div class="team-badge">Student Directory</div>
    </div>

    <div class="team-toolbar">
      <div class="team-field">
        <label for="directory-project">My Project</label>
        <select id="directory-project"></select>
      </div>

      <div class="team-field team-search-field">
        <label for="directory-search">Search Students</label>
        <div class="team-search-box">
          <input
            type="search"
            id="directory-search"
            placeholder="Name, CSE, Python, AI..."
          >
          <button class="button" type="submit" form="directory-search-form">
            Search
          </button>
        </div>
      </div>
    </div>

    <div class="team-info-strip">
      <p id="directory-members" class="team-members-note"></p>
      <p id="directory-status" class="team-status" role="status"></p>
    </div>

    <form id="directory-search-form" hidden></form>

    <div id="directory-results" class="student-grid"></div>
  </article>
`;

  const select = section.querySelector("#directory-project");
  const results = section.querySelector("#directory-results");
  const status = section.querySelector("#directory-status");
  const membersText = section.querySelector("#directory-members");

  const esc = value => String(value ?? "").replace(
    /[&<>"']/g,
    ch => ({
      "&": "&amp;", "<": "&lt;", ">": "&gt;",
      '"': "&quot;", "'": "&#39;"
    })[ch]
  );

  async function api(path, data) {
    const options = { credentials: "same-origin" };

    if (data !== undefined) {
      options.method = "POST";
      options.headers = {
        "Content-Type": "application/json",
        "X-Requested-With": "ResearchBridge"
      };
      options.body = JSON.stringify(data);
    }

    const response = await fetch(path, options);
    const result = await response.json();

    if (!response.ok) {
      throw new Error(result.error || "Request failed.");
    }
    return result;
  }

  async function loadMembers() {
    if (!select.value) {
      membersText.textContent = "Create a project first.";
      return;
    }

    try {
      const members = await api(
        `/api/workspace/${Number(select.value)}/members`
      );
      membersText.textContent = "Current Team: " +
        members.map(m => m.name).join(", ");
    } catch (error) {
      membersText.textContent = error.message;
    }
  }

  async function searchStudents() {
    status.textContent = "Searching...";
    results.replaceChildren();

    try {
      const q = section.querySelector("#directory-search").value.trim();
      const students = await api(
        "/api/students?q=" + encodeURIComponent(q)
      );

      status.textContent = students.length
        ? `Showing ${students.length} students (up to 50).`
        : "No students found.";

results.innerHTML = students.map(s => `
  <article class="student-card">
    <div class="student-card-top">
      <div class="student-avatar">
        ${esc((s.name || "?").trim().charAt(0).toUpperCase())}
      </div>

      <div class="student-main">
        <h3>${esc(s.name)}</h3>
        <p class="student-meta">
          ${esc(s.institution)} · ${esc(s.department)}
        </p>
        <p class="student-meta">
          Year ${Number(s.academic_year)} · Semester ${Number(s.semester)}
        </p>
      </div>
    </div>

    <div class="student-tags">
      <span class="student-tag">Student</span>
      <span class="student-tag">Verified</span>
    </div>

    <div class="student-details">
      <p><strong>Skills:</strong> ${esc(s.skills || "Not added yet")}</p>
      <p><strong>Research Interests:</strong> ${esc(s.interests || "Not added yet")}</p>
    </div>

    <div class="student-actions">
      <button
        type="button"
        class="button"
        data-student-id="${Number(s.id)}"
      >
        Invite to Team
      </button>
    </div>
  </article>
`).join("");
    } catch (error) {
      status.textContent = error.message;
    }
  }

  section.querySelector("#directory-search-form")
    .addEventListener("submit", event => {
      event.preventDefault();
      searchStudents();
    });

  select.addEventListener("change", loadMembers);

 
results.addEventListener("click", async event => {
  const button = event.target.closest("button[data-student-id]");
  if (!button) return;

  if (!select.value) {
    status.textContent = "❌ Please create or select a project first.";
    return;
  }

  button.disabled = true;
  button.textContent = "Sending Invitation...";
  status.textContent = "Sending invitation...";

  try {
    const result = await api("/api/team/invite", {
      project_id: Number(select.value),
      student_id: Number(button.dataset.studentId)
    });

    // Success message
    status.textContent = "✅ " + result.message;

    // Change button
    button.textContent = "⏳ Waiting for Response...";
    button.disabled = true;

    // Find selected student card
    const card = button.closest(".student-card");
    const actions = button.closest(".student-actions");

    // Add pending status message
    let pendingMessage = card.querySelector(".invitation-pending");

    if (!pendingMessage) {
      pendingMessage = document.createElement("div");
      pendingMessage.className = "invitation-pending";
      pendingMessage.setAttribute("role", "status");
      actions.appendChild(pendingMessage);
    }

    pendingMessage.innerHTML = `
      <strong>✅ Invitation Sent Successfully!</strong>
      <p>Your invitation has been sent.</p>
      <span>⏳ Pending — Waiting for Acceptance</span>
    `;

  } catch (error) {
    status.textContent = "❌ " + error.message;
    button.textContent = "Invite to Team";
    button.disabled = false;
  }
});


  async function initialize() {
    try {
      const me = await api("/api/me");
      const projects = await api("/api/collaboration-projects");
      const mine = projects;

      select.replaceChildren();

     if (!mine.length) {
  select.add(new Option("No project yet", ""));
  membersText.innerHTML =
    'No project found. <a class="inline-link" href="/workspace/student#projects">Create a project in Research Workspace →</a>';
} else {
        mine.forEach(p => {
          select.add(new Option(p.title, String(p.id)));
        });
        await loadMembers();
      }

      await searchStudents();
    } catch (error) {
      status.textContent = error.message;
    }
  }

window.addEventListener("team-members-changed", initialize);

window.addEventListener("hashchange", () => {
    if (location.hash === "#team") {
        loadMembers();
    }
});
initialize();
})();
