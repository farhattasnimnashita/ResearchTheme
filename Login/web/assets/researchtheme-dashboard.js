
        const pages = document.querySelectorAll(".page");
        const menuLinks = document.querySelectorAll("nav a");
        const title = document.getElementById("page-title");

        const titles = {
            dashboard: "Student Dashboard",
            profile: "My Profile",
            team: "My Team",
            supervisor: "Supervisor",
            proposal: "My Proposal",
            milestones: "Milestones",
            files: "Files & Feedback",
            meetings: "Meetings"
        };

        function showPage() {
            let selected = location.hash.slice(1) || "dashboard";

            if (!Object.prototype.hasOwnProperty.call(titles, selected)) {
                selected = "dashboard";
            }

            pages.forEach(function (page) {
                page.hidden = page.id !== selected;
            });

            menuLinks.forEach(function (link) {
                const active = link.getAttribute("href") === "#" + selected;

                link.classList.toggle("active", active);

                if (active) {
                    link.setAttribute("aria-current", "page");
                } else {
                    link.removeAttribute("aria-current");
                }
            });

            title.textContent = titles[selected];
            document.title = titles[selected] + " | ResearchBridge";
        }

        window.addEventListener("hashchange", showPage);
        showPage();

        const notice = document.getElementById("notice");
        const noticeMessage = document.getElementById("notice-message");

        document.querySelectorAll("[data-feature]").forEach(
            function (button) {
                button.addEventListener("click", function () {
                    noticeMessage.textContent =
                        button.dataset.feature +
                        " is not connected yet. This is a design preview.";

                    notice.showModal();
                });
            }
        );
    
async function request(path, options = {}) {
 const response = await fetch(path, {credentials: "same-origin", ...options});
 if (response.status === 401) { location.replace("/login/student"); throw new Error("Please log in again."); }
 const data = await response.json();
 if (!response.ok) throw new Error(data.error || "Could not load your data.");
 return data;
}
document.getElementById("logout").addEventListener("click", async function () {
 this.disabled = true;
 try {
  await request("/api/logout", {method: "POST", headers: {"Content-Type": "application/json", "X-Requested-With": "ResearchBridge"}, body: "{}"});
  location.replace("/login/student");
 } catch (error) { document.getElementById("data-status").textContent = error.message; this.disabled = false; }
});
async function loadStudent() {
 try {
  const user = await request("/api/me");
  if (user.role !== "student") { location.replace("/dashboard/" + user.role); return; }
  document.getElementById("session-name").textContent = user.name + " · Email verified";
  const profile = document.getElementById("student-profile");
  profile.replaceChildren();
  for (const [label, value] of [["Name",user.name],["Email",user.email],["Institution",user.institution],["Department",user.department],["Academic year",user.academic_year],["Semester",user.semester]]) {
   const row = document.createElement("div"), term = document.createElement("dt"), detail = document.createElement("dd");
   term.textContent = label; detail.textContent = value || "Not provided"; row.append(term,detail); profile.append(row);
  }
  const projects = await request("/api/projects");
  document.getElementById("proposal-status").textContent = projects.length ? projects.length + " proposal" + (projects.length === 1 ? "" : "s") : "Not submitted";
  const latest = projects[0];
  const currentProject = document.getElementById("current-project");
  currentProject.textContent = latest ? latest.title + " · " + latest.status.replaceAll("_", " ") : "No proposal yet.";
  if (latest) {
    currentProject.parentElement.querySelector(".empty").hidden = true;
    document.querySelector("#proposal h3").textContent = latest.title;
    document.querySelector("#proposal .empty p").textContent = latest.status.replaceAll("_", " ");
  }
  const feedback = projects.find(project => project.feedback);
  const feedbackNode = document.getElementById("latest-feedback");
  if (feedbackNode && feedback) feedbackNode.textContent = feedback.feedback;
 } catch(error) { document.getElementById("data-status").textContent = error.message + " Refresh to try again."; }
}
loadStudent();

