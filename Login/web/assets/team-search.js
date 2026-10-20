
(() => {
  const section = document.getElementById("team");
  if (!section) return;

  section.innerHTML = `
    <article class="card team-directory-card">
      <div class="team-header">
        <div>
          <p class="eyebrow">BUILD YOUR RESEARCH TEAM</p>
          <h2>My Team — Student Directory</h2>
          <p class="description">
            Browse students and invite them to your research project.
          </p>
        </div>
        <div class="team-badge">All Students</div>
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
              placeholder="Search name, university, department, skill..."
            >
            <button class="button" id="directory-search-button" type="button">
              Search
            </button>
            <button class="button secondary" id="directory-clear-button" type="button">
              Show All
            </button>
          </div>
        </div>
      </div>

      <div class="team-info-strip">
        <p id="directory-members" class="team-members-note"></p>
        <p id="directory-status" class="team-status" role="status">
          Loading students...
        </p>
      </div>

      <div id="directory-results" class="student-grid" aria-live="polite"></div>
    </article>
  `;

  const select = section.querySelector("#directory-project");
  const searchInput = section.querySelector("#directory-search");
  const searchButton = section.querySelector("#directory-search-button");
  const clearButton = section.querySelector("#directory-clear-button");
  const results = section.querySelector("#directory-results");
  const status = section.querySelector("#directory-status");
  const membersText = section.querySelector("#directory-members");
  const fixedTeamTitle =
    section.querySelector("#fixed-team-title");

  const fixedTeamStatus =
    section.querySelector("#fixed-team-status");

  const fixedTeamMembers =
    section.querySelector("#fixed-team-members");

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

    const text = String(value);
    if (/^\d+$/.test(text)) {
      return "Year " + text;
    }

    return text;
  }

  function formatSemester(value) {
    if (value === null || value === undefined || value === "") {
      return "Not added";
    }

    const text = String(value);
    if (/^\d+$/.test(text)) {
      return "Semester " + text;
    }

    return text;
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

    let result;
    try {
      result = await response.json();
    } catch {
      throw new Error("Server returned an invalid response.");
    }

    if (!response.ok) {
      throw new Error(result.error || "Request failed.");
    }

    return result;
  }

  function updateDashboardTeam(members, project) {
    const summary = document.getElementById(
      "dashboard-team-summary"
    );

    const dashboardMembers = document.getElementById(
      "dashboard-team-members"
    );

    const dashboardTopic = document.getElementById(
      "dashboard-team-topic"
    );

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

      const ownerID = Number(
        currentProject?.student_id || 0
      );

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
          "The invited student accepted. Team members are saved.";

        fixedTeamMembers.innerHTML = members.map(member => `
          <p class="fixed-team-member">
            <strong>${esc(member.name)}</strong>
            <span> — Confirmed member</span>
          </p>
        `).join("");
      } else {
        fixedTeamTitle.textContent =
          "Waiting for Team Confirmation";

        fixedTeamStatus.textContent =
          "Your team will be marked Fixed Team after an invited student accepts.";

        fixedTeamMembers.innerHTML = members.map(member => `
          <p class="fixed-team-member">
            <strong>${esc(member.name)}</strong>
            <span> — Team owner</span>
          </p>
        `).join("");
      }

      updateDashboardTeam(members, currentProject);

    } catch (error) {
      membersText.textContent = error.message;

      fixedTeamTitle.textContent = "Team unavailable";

      fixedTeamStatus.textContent =
        "Could not load team members. Please refresh the page.";
    }
  }

  function renderStudents(students) {
    currentStudents = students;

    if (!students.length) {
      results.innerHTML = `
        <article class="card">
          <h3>No students found</h3>
          <p>Try another name, university, department or skill.</p>
          <button type="button" class="button secondary" id="empty-show-all">
            Show All Students
          </button>
        </article>
      `;

      results.querySelector("#empty-show-all").addEventListener("click", () => {
        searchInput.value = "";
        searchStudents();
      });

      return;
    }

    results.innerHTML = students.map(student => {
      const id = Number(student.id);
      const alreadyInvited = student.invitation_pending === true;

      return `
        <article class="student-card">
          <div class="student-card-top">
            <div class="student-avatar">
              ${esc((student.name || "?").trim().charAt(0).toUpperCase())}
            </div>

            <div class="student-main">
              <h3>${esc(student.name || "Unnamed student")}</h3>
              <p class="student-meta">
                <strong>Student ID:</strong> ${id}
              </p>
              <p class="student-meta">
                <strong>University:</strong>
                ${esc(student.institution || "Not added")}
              </p>
              <p class="student-meta">
                <strong>Department:</strong>
                ${esc(student.department || "Not added")}
              </p>
              <p class="student-meta">
                ${esc(formatYear(student.academic_year))}
                · ${esc(formatSemester(student.semester))}
              </p>
            </div>
          </div>

          <div class="student-tags">
            <span class="student-tag">Student</span>
            <span class="student-tag">Verified</span>
          </div>

          <div class="student-details">
            <p>
              <strong>Skills:</strong>
              ${esc(student.skills || "Not added yet")}
            </p>
            <p>
              <strong>Research Interests:</strong>
              ${esc(student.interests || "Not added yet")}
            </p>
          </div>

          <div class="student-actions">
            <button
              type="button"
              class="button invite-student-button"
              data-student-id="${id}"
              ${!select.value || alreadyInvited ? "disabled" : ""}
            >
              ${alreadyInvited ? "Invitation Pending" : "Invite to Team"}
            </button>
            <p class="invitation-message" role="status"></p>
          </div>
        </article>
      `;
    }).join("");

    results.querySelectorAll(".invite-student-button").forEach(button => {
      button.title = select.value
        ? "Send a team invitation"
        : "Select a project first";
    });
  }

  async function searchStudents() {
    const thisSearch = ++searchNumber;
    const query = searchInput.value.trim();

    status.textContent = query
      ? "Searching students..."
      : "Loading all students...";

    results.replaceChildren();

    try {
      const students = await api(
        "/api/students?q=" + encodeURIComponent(query)
      );

      // Ignore an older response if a newer search has started.
      if (thisSearch !== searchNumber) return;

      renderStudents(students);

      status.textContent = students.length
        ? `Showing ${students.length} student(s).`
        : "No students found.";
    } catch (error) {
      if (thisSearch !== searchNumber) return;

      status.textContent = error.message;
      results.innerHTML = `
        <article class="card">
          <h3>Could not load students</h3>
          <p>${esc(error.message)}</p>
          <button type="button" class="button" id="directory-retry">
            Try Again
          </button>
        </article>
      `;

      results.querySelector("#directory-retry").addEventListener(
        "click",
        searchStudents
      );
    }
  }

  async function initialize() {
    status.textContent = "Loading students...";

    try {
      await api("/api/me");

      const projects = await api("/api/collaboration-projects");
             projectList = projects;
      select.replaceChildren();

      if (!Array.isArray(projects) || projects.length === 0) {
        select.add(new Option("No project yet — create one first", ""));
        membersText.textContent =
          "You can browse students now. Create a project before sending invitations.";
      } else {
        select.add(new Option("Select a project", ""));

        projects.forEach(project => {
          select.add(new Option(project.title, String(project.id)));
        });

        membersText.textContent =
          "Select your project before inviting a student.";
      }
           if (projects.length > 0) {
        const requestedProject = new URLSearchParams(
          window.location.search
        ).get("project_created");

        const selectedProject = projects.find(
          project => String(project.id) === requestedProject
        ) || projects[0];

        select.value = String(selectedProject.id);
      }

      await loadMembers();
      // Show students immediately, without requiring a search.
      await searchStudents();
    } catch (error) {
      status.textContent = error.message;
      results.innerHTML = `
        <article class="card">
          <h3>Could not initialize Student Directory</h3>
          <p>${esc(error.message)}</p>
          <button type="button" class="button" id="directory-retry">
            Try Again
          </button>
        </article>
      `;

      results.querySelector("#directory-retry").addEventListener(
        "click",
        initialize
      );
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
    renderStudents(currentStudents);
  });

  results.addEventListener("click", async event => {
    const button = event.target.closest(".invite-student-button");
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

    const message = button.closest(".student-actions")
      .querySelector(".invitation-message");

    button.disabled = true;
    button.textContent = "Sending Invitation...";
    message.textContent = "";

    try {
      const result = await api("/api/team/invite", {
        project_id: Number(select.value),
        student_id: studentID
      });

      button.textContent = "Invitation Pending";
      message.textContent =
        result.message || "Invitation sent. Waiting for response.";

      status.textContent = "Invitation sent successfully.";

    } catch (error) {
      button.disabled = false;
      button.textContent = "Invite to Team";
      message.textContent = error.message;
      status.textContent = "Could not send invitation.";
    }
  });

  window.addEventListener("team-members-changed", initialize);
  // Refresh the team list periodically so the owner's
  // dashboard can show newly accepted invitations.
  setInterval(() => {
    if (
      document.visibilityState === "visible" &&
      select.value
    ) {
      loadMembers();
    }
  }, 15000);
  window.addEventListener("hashchange", () => {
    if (location.hash === "#team") {
      loadMembers();
      searchStudents();
    }
  });

  initialize();
})();
