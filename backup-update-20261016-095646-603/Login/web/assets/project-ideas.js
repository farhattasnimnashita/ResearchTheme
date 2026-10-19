
(() => {
  const page = document.getElementById("project-ideas");
  if (!page) return;

  const ideas = [
    ["AI-Based Student Skill Matching System", "Artificial Intelligence"],
    ["Smart Attendance Management System", "Web Development"],
    ["Dengue Prediction Using Machine Learning", "Data Science"],
    ["University Student Helpdesk Chatbot", "Natural Language Processing"],
    ["Thesis Team Formation Platform", "Software Engineering"],
    ["Online Blood Donation Management", "Web Development"],
    ["Smart Classroom Monitoring System", "Internet of Things"],
    ["Fake News Detection Using AI", "Machine Learning"],
    ["Crop Disease Detection from Images", "Computer Vision"],
    ["Library Book Recommendation System", "Recommendation Systems"],
    ["University Bus Tracking System", "Mobile Application"],
    ["Cybersecurity Awareness Learning Platform", "Cybersecurity"],
    ["Smart Waste Collection Monitoring", "Internet of Things"],
    ["Road Accident Analysis Dashboard", "Data Analytics"],
    ["Mental Wellness Resource Finder", "Web Development"],
    ["Automatic Class Routine Generator", "Algorithms"],
    ["Student Result Analysis Dashboard", "Data Science"],
    ["Research Paper Recommendation System", "Information Retrieval"],
    ["Campus Event Management Platform", "Software Engineering"],
    ["Flood Risk Information Dashboard", "Data Visualization"]
  ];

  const search = page.querySelector("#project-idea-search");
  const list = page.querySelector("#project-idea-list");
  const form = page.querySelector("#project-idea-form");
  const message = page.querySelector("#project-idea-status");
  const title = page.querySelector("#new-project-title");
  const area = page.querySelector("#new-project-area");
  const abstract = page.querySelector("#new-project-description");
  const save = page.querySelector("#save-project-idea");

  const escapeHTML = text => String(text).replace(/[&<>"']/g, c => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;",
    '"': "&quot;", "'": "&#39;"
  })[c]);

  function renderIdeas() {
    const q = search.value.trim().toLowerCase();

    const matches = ideas.map((idea, index) => ({ idea, index }))
      .filter(({idea}) =>
        idea.some(s => s.toLowerCase().includes(q))
      );

    list.innerHTML = matches.map(({idea, index}) => `
      <button type="button" class="project-idea-option"
              data-idea="${index}"
              aria-pressed="${title.value.trim().toLowerCase() === idea[0].toLowerCase()}">
        <strong>${escapeHTML(idea[0])}</strong>
        <span>${escapeHTML(idea[1])}</span>
      </button>
    `).join("") ||
      '<p>No matching ideas. Try another keyword or add your own.</p>';
  }

  list.addEventListener("click", event => {
    const option = event.target.closest("button[data-idea]");
    if (!option) return;

    const [name, category] = ideas[Number(option.dataset.idea)];

    title.value = name;
    area.value = category;
    abstract.value =
      `Draft idea: Design and evaluate a practical solution for ${name.toLowerCase()}, using suitable methods and measurable results.`;

    message.textContent =
      "Selected! Edit the title or description if you want.";

    renderIdeas();
    form.scrollIntoView({behavior: "smooth", block: "start"});
  });

  page.querySelector("#custom-project-button")
    .addEventListener("click", () => {
      form.reset();
      message.textContent =
        "Enter your own project title, area and description.";
      renderIdeas();
      title.focus();
    });

  search.addEventListener("input", renderIdeas);
  title.addEventListener("input", renderIdeas);

  form.addEventListener("submit", async event => {
    event.preventDefault();
    if (!form.reportValidity()) return;

    const data = {
      title: title.value.trim(),
      area: area.value.trim(),
      abstract: abstract.value.trim(),
      teacher_id: 0
    };

    if (data.title.length < 5 || data.title.length > 160 ||
        data.area.length < 2 || data.area.length > 80 ||
        data.abstract.length < 20 || data.abstract.length > 6000) {
      message.textContent =
        "Title: 5–160, area: 2–80, description: 20–6000 characters.";
      return;
    }

    save.disabled = true;
    message.textContent = "Saving your project...";

    try {
      const existingResponse = await fetch("/api/projects", {
        credentials: "same-origin"
      });

      const existing = await existingResponse.json();

      if (!existingResponse.ok) {
        throw new Error(existing.error || "Cannot check saved projects.");
      }

      if (existing.some(p =>
        p.title.trim().toLowerCase() === data.title.toLowerCase()
      )) {
        throw new Error(
          "You already saved a project with this title. Choose a different name."
        );
      }

      const response = await fetch("/api/projects", {
        method: "POST",
        credentials: "same-origin",
        headers: {
          "Content-Type": "application/json",
          "X-Requested-With": "ResearchBridge"
        },
        body: JSON.stringify(data)
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || "Could not save project.");
      }

      message.textContent = "Project saved! Opening My Team...";

      window.location.replace(
        `/dashboard/student?project_created=${result.id}#team`
      );

    } catch (error) {
      message.textContent = "Could not save: " + error.message;
    } finally {
      save.disabled = false;
    }
  });

  renderIdeas();
})();
