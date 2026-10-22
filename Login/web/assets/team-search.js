
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
            Search students by name, department, skills or interests
            and invite them to your project.
          </p>
        </div>
        <div class="team-badge">Student Directory</div>
      </div>

      <article class="card fixed-team-panel">
        <p class="eyebrow">CONFIRMED TEAM</p>
        <h3 id="fixed-team-title">No fixed team yet</h3>
        <p id="fixed-team-status" role="status">
          An invited student must accept before the team is fixed.
        </p>
        <div id="fixed-team-members">
          <p>No confirmed team members yet.</p>
        </div>
      </article>

      <div class="team-toolbar">
        <div class="team-field">
          <label for="directory-project">Select My Project</label>
          <select id="directory-project"></select>
        </div>

        <div class="team-field team-search-field">
          <label for="directory-search">Search Students</label>
          <div class="team-search-box">
            <input
              type="search"
              id="directory-search"
              placeholder="Name, university, department, skill..."
            >
            <button
              class="button"
              id="directory-search-button"
              type="button"
            >Search</button>
            <button
              class="button secondary"
              id="directory-clear-button"
              type="button"
            >Show All</button>
          </div>
        </div>
      </div>

      <div class="team-info-strip">
        <p id="directory-members" class="team-members-note"></p>
        <p id="directory-status" class="team-status" role="status">
          Loading students...
        </p>
      </div>

      <div id="directory-results" class="student-grid"></div>
    </article>
  `;

  const select = section.querySelector("#directory-project");
  const searchInput = section.querySelector("#directory-search");
  const searchButton = section.querySelector("#directory-search-button");
  const clearButton = section.querySelector("#directory-clear-button");
  const results = section.querySelector("#directory-results");
  const status = section.querySelector("#directory-status");
  const membersText = section.querySelector("#directory-members");
  const fixedTeamTitle = section.querySelector("#fixed-team-title");
  const fixedTeamStatus = section.querySelector("#fixed-team-status");
  const fixedTeamMembers = section.querySelector("#fixed-team-members");

  let projectList = [];
  let currentProject = null;
  let currentStudents = [];
  let searchNumber = 0;

  const esc = value =>
    String(value ?? "").replace(/[&<>"']/g, ch => ({
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#39;"
    })[ch]);

  function formatYear(value) {
    if (value === null || value === undefined || value === "") {
      return "Not added";
    }

    return /^\d+$/.test(String(value))
      ? "Year " + value
      : String(value);
  }

  function formatSemester(value) {
    if (value === null || value === undefined || value === "") {
      return "Not added";
    }

    return /^\d+$/.test(String(value))
      ? "Semester " + value
      : String(value);
  }

  async function api(path, data) {
    const options = {
      credentials: "same-origin"
    };

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

  function updateDashboardTeam(members, project) {
    const summary = document.getElementById("dashboard-team-summary");
    const dashboardMembers = document.getElementById("dashboard-team-members");
    const dashboardTopic = document.getElementById("dashboard-team-topic");

    const ownerID = Number(project?.student_id || 0);

    const hasConfirmedMembers = members.some(
      member => Number(member.id) !== ownerID
    );

    if (summary) {
      summary.textContent = hasConfirmedMembers
        ? "Fixed Team"
        : "No team yet";
    }

    if (dashboardMembers) {
      dashboardMembers.textContent = members.length
        ? members.map(member => member.name).join(", ")
        : "No members added yet";
    }

    if (dashboardTopic) {
      dashboardTopic.textContent =
        project?.title || "No topic selected yet";
    }
  }

  async function loadMembers() {
    if (!select.value) {
      currentProject = null;
      membersText.textContent =
        "Select or create a project to manage your team.";

      fixedTeamTitle.textContent = "No fixed team yet";
      fixedTeamStatus.textContent =
        "Select a project to check its confirmed members.";

      fixedTeamMembers.innerHTML =
        "<p>No confirmed team members yet.</p>";

      updateDashboardTeam([], null);
      return;
    }

    try {
      currentProject = projectList.find(
        project => String(project.id) === select.value
      ) || null;

      const members = await api(
        `/api/workspace/${Number(select.value)}/members`
      );

      const ownerID = Number(currentProject?.student_id || 0);

      const hasConfirmedMembers = members.some(
        member => Number(member.id) !== ownerID
      );

      membersText.textContent =
        "Current Team: " +
        (members.length
          ? members.map(member => member.name).join(", ")
          : "No members listed yet.");

      if (hasConfirmedMembers) {
        fixedTeamTitle.textContent = "Fixed Team";
        fixedTeamStatus.textContent =
          "An invited student has accepted. Confirmed members are listed below.";

        fixedTeamMembers.innerHTML = members.map(member => `
          <p class="fixed-team-member">
            <strong>${esc(member.name)}</strong>
            <span>
              — ${Number(member.id) === ownerID
                ? "Team owner"
                : "Confirmed member"}
            </span>
          </p>
        `).join("");
      } else {
        fixedTeamTitle.textContent = "Waiting for Team Confirmation";
        fixedTeamStatus.textContent =
          "Your team will be fixed after an invited student accepts.";

        fixedTeamMembers.innerHTML = members.length
          ? members.map(member => `
              <p class="fixed-team-member">
                <strong>${esc(member.name)}</strong>
                <span> — Team owner</span>
              </p>
            `).join("")
          : "<p>No confirmed team members yet.</p>";
      }

      updateDashboardTeam(members, currentProject);
    } catch (error) {
      membersText.textContent = error.message;
      fixedTeamTitle.textContent = "Team unavailable";
      fixedTeamStatus.textContent =
        "Could not load team members. Please refresh the page.";
    }
  }

  async function searchStudents() {
    const requestNumber = ++searchNumber;

    status.textContent = "Searching...";
    results.replaceChildren();

    try {
      const query = searchInput.value.trim();

      const students = await api(
        "/api/students?q=" + encodeURIComponent(query)
      );

      // Ignore an older search response.
      if (requestNumber !== searchNumber) return;

      currentStudents = students;

      status.textContent = students.length
        ? `Showing ${students.length} students (up to 50).`
        : "No students found.";

      results.innerHTML = students.map(student => `
        <article class="student-card">
          <div class="student-card-top">
            <div class="student-avatar">
              ${esc((student.name || "?").trim().charAt(0).toUpperCase())}
            </div>

            <div class="student-main">
              <h3>${esc(student.name)}</h3>
              <p class="student-meta">
                ${esc(student.institution)} · ${esc(student.department)}
              </p>
              <p class="student-meta">
                ${esc(formatYear(student.academic_year))} ·
                ${esc(formatSemester(student.semester))}
              </p>
            </div>
          </div>

          <div class="student-tags">
            <span class="student-tag">Student</span>
            <span class="student-tag">Verified</span>
          </div>

          <div class="student-details">
            <p><strong>Skills:</strong>
              ${esc(student.skills || "Not added yet")}
            </p>
            <p><strong>Research Interests:</strong>
              ${esc(student.interests || "Not added yet")}
            </p>
          </div>

          <div class="student-actions">
            <button
              type="button"
              class="button invite-student-button"
              data-student-id="${Number(student.id)}"
            >Invite to Team</button>
            <p class="invitation-message" role="status"></p>
          </div>
        </article>
      `).join("");
    } catch (error) {
      if (requestNumber !== searchNumber) return;
      status.textContent = error.message;
    }
  }

  searchButton.addEventListener("click", searchStudents);

  searchInput.addEventListener("keydown", event => {
    if (event.key === "Enter") {
      event.preventDefault();
      searchStudents();
    }
  });

  clearButton.addEventListener("click", () => {
    searchInput.value = "";
    searchStudents();
  });

  select.addEventListener("change", async () => {
    await loadMembers();
  });

  results.addEventListener("click", async event => {
    const button = event.target.closest("button[data-student-id]");
    if (!button) return;

    if (!select.value) {
      status.textContent = "Please create or select a project first.";
      return;
    }

    const studentID = Number(button.dataset.studentId);

    if (!studentID) {
      status.textContent = "Invalid student ID.";
      return;
    }

    const card = button.closest(".student-card");
    const message = card.querySelector(".invitation-message");

    button.disabled = true;
    button.textContent = "Sending Invitation...";
    message.textContent = "";

    try {
      const result = await api("/api/team/invite", {
        project_id: Number(select.value),
        student_id: studentID
      });

      status.textContent =
        result.message || "Invitation sent successfully.";

      button.textContent = "Waiting for Response";
      message.textContent =
        "Invitation sent. Waiting for acceptance.";

    } catch (error) {
      status.textContent = error.message;
      button.textContent = "Invite to Team";
      button.disabled = false;
      message.textContent = error.message;
    }
  });

  async function initialize() {
    try {
      await api("/api/me");

      const projects = await api("/api/collaboration-projects");

      projectList = projects;
      select.replaceChildren();

      if (!projects.length) {
        select.add(new Option("No project yet", ""));
        membersText.innerHTML =
          'No project found. <a class="inline-link" ' +
          'href="/workspace/student#projects">' +
          'Create a project in Research Workspace</a>';

        await loadMembers();
      } else {
        projects.forEach(project => {
          select.add(new Option(project.title, String(project.id)));
        });

        const requestedProject = new URLSearchParams(
          window.location.search
        ).get("project_created");


const approvedProject = projects.find(
    project => project.status === "approved"
);

const selectedProject = approvedProject ||
    projects.find(
        project => String(project.id) === requestedProject
    ) ||
    projects[0];

if (selectedProject) {
    select.value = String(selectedProject.id);
    select.disabled = Boolean(approvedProject);
}


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

  // Refresh team membership while the page is visible.
  setInterval(() => {
    if (
      document.visibilityState === "visible" &&
      select.value
    ) {
      loadMembers();
    }
  }, 15000);

  initialize();
})();
