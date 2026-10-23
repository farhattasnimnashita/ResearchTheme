 const pages = document.querySelectorAll(".page");
        const menuLinks = document.querySelectorAll("nav a");
        const title = document.getElementById("page-title");

        const titles = {
            dashboard: "Student Dashboard",
            profile: "My Profile",
            "project-ideas": "Choose Project",
            team: "My Team",
            supervisor: "Supervisor",
            proposal: "My Proposal",
            milestones: "Milestones",
            files: "Files & Feedback",
            meetings: "Meetings",
            "literature-review": "Literature Review Matrix",
            "research-ethics": "Research Ethics & Approval",
            "experiment-reproducibility": "Experiment Reproducibility",
            "feedback-resolution": "Reviewer Feedback Resolution Tracker",
            "citation-manager": "Citation & Reference Manager",
            "risk-decision-log": "Risk Register & Decision Log",
"research-gap-tracker": "Research Gap Tracker",
"research-question-map": "Research Question Map",
"dataset-registry": "Dataset Registry",
"task-dependencies": "Task Dependency Map",
"team-workload": "Team Workload Insights",
"authorship-agreement": "Authorship & Contribution Agreement",
"research-budget": "Research Budget Tracker",
"project-export": "Project Export & Handover",
"data-access-settings": "Data Access & Embargo Settings",
"research-impact": "Research Impact Tracker"
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

        async function loadMyProfile() {
    const message = document.getElementById("profile-message");

    if (!message) return;

    const fields = {
        "profile-name": "name",
        "profile-email": "email",
        "profile-institution": "institution",
        "profile-department": "department",
        "profile-year": "academic_year",
        "profile-semester": "semester"
    };

    try {
        const response = await fetch("/api/me", {
            credentials: "same-origin",
            cache: "no-store"
        });

        if (response.status === 401) {
            window.location.replace("/login/student");
            return;
        }

        if (!response.ok) {
            throw new Error("Could not load your profile.");
        }

        const user = await response.json();
        const welcome = document.getElementById("student-welcome");

if (welcome) {
    welcome.textContent = "Welcome, " + user.name + "! ";
}

        for (const [elementId, fieldName] of Object.entries(fields)) {
            const element = document.getElementById(elementId);

            if (element) {
                element.textContent = user[fieldName] || "Not added";
            }
        }

        message.textContent = "";
    } catch (error) {
        for (const elementId of Object.keys(fields)) {
            const element = document.getElementById(elementId);

            if (element) {
                element.textContent = "Unavailable";
            }
        }

        message.textContent =
            "Profile load হয়নি। Page refresh করে আবার চেষ্টা করো.";
    }
}

loadMyProfile();

(() => {
    function setupProfileEditor() {
        const section = document.getElementById("profile");
        if (!section || section.querySelector("#student-edit-form")) return;

       const oldButton = section.querySelector("#edit-profile-button");

        if (!oldButton) {
            console.error("Edit Profile button was not found.");
            return;
        }

        const editButton = document.createElement("button");
        editButton.type = "button";
        editButton.className = "button";
        editButton.textContent = "Edit Profile";

        // Replacing the button removes its old preview click handler.
        oldButton.replaceWith(editButton);

        const form = document.createElement("form");
        form.id = "student-edit-form";
        form.hidden = true;

        form.innerHTML = `
            <p>
                <label>Name<br>
                    <input name="name" required
                           minlength="2" maxlength="100">
                </label>
            </p>
            <p>
                <label>Institution<br>
                    <input name="institution" required
                           minlength="2" maxlength="180">
                </label>
            </p>
            <p>
                <label>Department<br>
                    <input name="department" required
                           minlength="2" maxlength="120">
                </label>
            </p>
            <p>
                <label>Academic Year<br>
                    <input name="academic_year" type="number"
                           min="1" max="6" step="1" required>
                </label>
            </p>
            <p>
                <label>Semester<br>
                    <input name="semester" type="number"
                           min="1" max="12" step="1" required>
                </label>
            </p>
            <p>
    <label>Skills<br>
        <textarea name="expertise"
                  rows="3"
                  maxlength="1500"
                  placeholder="C++, Python, Java, Web Development"></textarea>
    </label>
</p>

<p>
    <label>Research Interests<br>
        <textarea name="interests"
                  rows="3"
                  maxlength="1500"
                  placeholder="Artificial Intelligence, Machine Learning"></textarea>
    </label>
</p>

<p>
    <label>About Me<br>
        <textarea name="bio"
                  rows="4"
                  maxlength="3000"
                  placeholder="Write something about yourself"></textarea>
    </label>
</p>
            <div class="actions">
                <button type="submit" class="button">Save Profile</button>
                <button type="button" class="button secondary"
                        data-cancel>Cancel</button>
            </div>
        `;

        const status = document.createElement("p");
        status.setAttribute("role", "status");

        editButton.after(form, status);

        const saveButton = form.querySelector('[type="submit"]');
        const cancelButton = form.querySelector("[data-cancel]");

        async function readProfile() {
            const response = await fetch("/api/me", {
                credentials: "same-origin",
                cache: "no-store"
            });

            if (response.status === 401) {
                window.location.replace("/login/student");
                throw new Error("Please log in.");
            }

            if (!response.ok) {
                throw new Error("Could not load your profile.");
            }

            return response.json();
        }

        function showProfile(user) {
            const details = section.querySelector("dl.details");
            if (!details) return;

            const rows = [
                ["Name", user.name],
                ["Email", user.email],
                ["Institution", user.institution],
                ["Department", user.department],
                ["Academic Year", user.academic_year],
                ["Semester", user.semester],
["Skills", user.extra?.expertise || "Not added"],
["Research Interests", user.extra?.interests || "Not added"],
["About Me", user.extra?.bio || "Not added"]
            ];

            details.replaceChildren();

            for (const [label, value] of rows) {
                const row = document.createElement("div");
                const dt = document.createElement("dt");
                const dd = document.createElement("dd");

                dt.textContent = label;
                dd.textContent = value || "Not added";

                row.append(dt, dd);
                details.append(row);
            }
        }

        editButton.addEventListener("click", async () => {
            editButton.disabled = true;
            status.textContent = "Loading profile...";

            try {
                const user = await readProfile();
                const extraResponse = await fetch("/api/researcher-profile");

if (!extraResponse.ok) {
    throw new Error("Could not load research profile.");
}

const extra = await extraResponse.json();

form.elements.namedItem("expertise").value = extra.expertise || "";
form.elements.namedItem("interests").value = extra.interests || "";
form.elements.namedItem("bio").value = extra.bio || "";
                for (const key of [
                    "name", "institution", "department",
                    "academic_year", "semester"
                ]) {
                    form.elements.namedItem(key).value = user[key] || "";
                }

                form.hidden = false;
                editButton.hidden = true;
                status.textContent = "";
                form.elements.namedItem("name").focus();
            } catch (error) {
                status.textContent = error.message;
            } finally {
                editButton.disabled = false;
            }
        });

        cancelButton.addEventListener("click", () => {
            form.hidden = true;
            editButton.hidden = false;
            status.textContent = "";
            editButton.focus();
        });

        form.addEventListener("submit", async (event) => {
            event.preventDefault();
            if (!form.reportValidity()) return;

            const values = new FormData(form);
            const payload = {
                name: values.get("name").trim(),
                institution: values.get("institution").trim(),
                department: values.get("department").trim(),
                academic_year: Number(values.get("academic_year")),
                semester: Number(values.get("semester"))
            };

            saveButton.disabled = true;
            cancelButton.disabled = true;
            status.textContent = "Saving...";

            try {
                const response = await fetch("/api/student-profile", {
                    method: "POST",
                    credentials: "same-origin",
                    headers: {
                        "Content-Type": "application/json",
                        "X-Requested-With": "ResearchBridge"
                    },
                    body: JSON.stringify(payload)
                });

                const data = await response.json();

                if (!response.ok) {
                    throw new Error(data.error || "Could not save profile.");
                }
                // Load existing profile settings
const extraResponse = await fetch("/api/researcher-profile");

if (!extraResponse.ok) {
    throw new Error("Could not load research profile.");
}

const existingProfile = await extraResponse.json();

// Skills, Research Interests, About Me
const extraPayload = {
    ...existingProfile,
    expertise: String(values.get("expertise") || "").trim(),
    interests: String(values.get("interests") || "").trim(),
    bio: String(values.get("bio") || "").trim()
};

// Save additional profile information
const extraSave = await fetch("/api/researcher-profile", {
    method: "POST",
    credentials: "same-origin",
    headers: {
        "Content-Type": "application/json",
        "X-Requested-With": "ResearchBridge"
    },
    body: JSON.stringify(extraPayload)
});

if (!extraSave.ok) {
    const error = await extraSave.json();
    throw new Error(error.error || "Could not save skills.");
}

                form.hidden = true;
                editButton.hidden = false;
                status.textContent = "Profile saved successfully.";

                try {
                    const updatedUser = await readProfile();

const updatedResponse = await fetch("/api/researcher-profile");
if (!updatedResponse.ok) {
    throw new Error("Could not reload profile.");
}

updatedUser.extra = await updatedResponse.json();

showProfile(updatedUser);
                } catch {
                    status.textContent =
                        "Profile saved. Refresh to load the updated details.";
                }
            } catch (error) {
                status.textContent = error.message;
            } finally {
                saveButton.disabled = false;
                cancelButton.disabled = false;
            }
        });

   readProfile()
    .then(async (user) => {
        // Load Skills, Research Interests and About Me
        const extraResponse = await fetch("/api/researcher-profile", {
            credentials: "same-origin",
            cache: "no-store"
        });

        if (!extraResponse.ok) {
            throw new Error("Could not load research profile.");
        }

        // Attach saved research profile to the user
        user.extra = await extraResponse.json();

        // Display all profile details
        showProfile(user);
    })
    .catch((error) => {
        status.textContent = error.message;
    });
    }

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", setupProfileEditor);
    } else {
        setupProfileEditor();
    }
})();
// Student Logout



const studentLogoutButton = document.getElementById("student-logout");

if (studentLogoutButton) {
    studentLogoutButton.addEventListener("click", async () => {

        studentLogoutButton.disabled = true;
        studentLogoutButton.textContent = "Logging out...";

        try {
            const response = await fetch("/api/logout", {
                method: "POST",
                credentials: "same-origin",
                headers: {
                    "Content-Type": "application/json",
                    "X-Requested-With": "ResearchBridge"
                },
                body: JSON.stringify({})
            });

            const result = await response.json();

            if (!response.ok) {
                throw new Error(result.error || "Logout failed.");
            }

            window.location.replace("/login/student");

        } catch (error) {
            alert(error.message);

            studentLogoutButton.disabled = false;
            studentLogoutButton.textContent = "↪ Logout";
        }
    });
}


/* Team Invitation Notifications */

/* Notification Bell Dropdown */

(() => {
    const bell = document.getElementById("notification-bell");
    const panel = document.getElementById("notification-panel");
    const badge = document.getElementById("notification-count");
    const list = document.getElementById("notification-list");
    const message = document.getElementById("notification-message");

    if (!bell || !panel || !badge || !list || !message) return;

    const escapeHTML = value => String(value ?? "").replace(
        /[&<>"']/g,
        c => ({
            "&": "&amp;",
            "<": "&lt;",
            ">": "&gt;",
            '"': "&quot;",
            "'": "&#39;"
        })[c]
    );

    async function api(url, data) {
        const options = {
            credentials: "same-origin",
            cache: "no-store"
        };

        if (data !== undefined) {
            options.method = "POST";
            options.headers = {
                "Content-Type": "application/json",
                "X-Requested-With": "ResearchBridge"
            };
            options.body = JSON.stringify(data);
        }

        const response = await fetch(url, options);
        const result = await response.json();

        if (!response.ok) {
            throw new Error(result.error || "Request failed.");
        }

        return result;
    }

    let loading = false;

    async function loadNotifications() {
        if (loading) return;
        loading = true;

        try {
            const [invites, updates] = await Promise.all([
                api("/api/invitations"),
                api("/api/notifications")
            ]);

            const unread = updates.filter(n => !n.read_at).length;
            const total = invites.length + unread;

            badge.textContent = total > 99 ? "99+" : String(total);
            badge.hidden = total === 0;

            list.innerHTML = `
                <h4>Team Invitations</h4>

                ${invites.map(inv => `
                    <div class="bell-item">
                        <strong>New Team Invitation</strong>
                        <p>
                            ${escapeHTML(inv.name)}
                            invited you to join
                            ${escapeHTML(inv.title)}.
                        </p>

                        <div class="bell-actions">
                            <button type="button" class="button"
                                data-invite-id="${Number(inv.id)}"
                                data-accept="true">
                                Accept
                            </button>

                            <button type="button" class="button secondary"
                                data-invite-id="${Number(inv.id)}"
                                data-accept="false">
                                Reject
                            </button>
                        </div>
                    </div>
                `).join("") || "<p>No pending invitations.</p>"}

                <h4>Recent Activity</h4>

                ${updates.map(n => `
                    <div class="bell-item">
                        <p>${escapeHTML(n.message)}</p>
                        <small>
                            ${new Date(n.created_at * 1000).toLocaleString()}
                        </small>
                    </div>
                `).join("") || "<p>No recent activity.</p>"}
            `;

            message.textContent = "";

        } catch (error) {
            message.textContent = "❌ " + error.message;
        } finally {
            loading = false;
        }
    }

    bell.addEventListener("click", async () => {
        panel.hidden = !panel.hidden;
        bell.setAttribute("aria-expanded", String(!panel.hidden));

        if (!panel.hidden) {
            await loadNotifications();

            try {
                await api("/api/notifications", {});
                await loadNotifications();
            } catch (error) {
                message.textContent = error.message;
            }
        }
    });

    list.addEventListener("click", async event => {
        const button = event.target.closest("button[data-invite-id]");
        if (!button) return;

        button.disabled = true;

        try {
            const result = await api("/api/invitations", {
                id: Number(button.dataset.inviteId),
                accept: button.dataset.accept === "true"
            });

            await loadNotifications();
            message.textContent = "✅ " + result.message;

            window.dispatchEvent(
                new Event("team-members-changed")
            );

        } catch (error) {
            message.textContent = "❌ " + error.message;
            button.disabled = false;
        }
    });

    document.addEventListener("click", event => {
        if (!event.target.closest(".bell-wrapper")) {
            panel.hidden = true;
            bell.setAttribute("aria-expanded", "false");
        }
    });

    setInterval(() => {
        if (document.visibilityState === "visible") {
            loadNotifications();
        }
    }, 15000);

    loadNotifications();
})();

/* Student Profile Picture */
(() => {
    const image = document.getElementById("student-profile-picture");
    const input = document.getElementById("profile-picture-input");
    const chooseButton = document.getElementById("choose-profile-picture");
    const uploadButton = document.getElementById("upload-profile-picture");
    const message = document.getElementById("profile-picture-message");

    if (!image || !input || !chooseButton || !uploadButton || !message) {
        return;
    }

    let selectedFile = null;
    let previewURL = null;

    image.addEventListener("error", () => {
        image.style.visibility = "hidden";
    });

    image.addEventListener("load", () => {
        image.style.visibility = "visible";
    });

    chooseButton.addEventListener("click", () => {
        input.click();
    });

    input.addEventListener("change", () => {
        const file = input.files[0];

        if (!file) {
            return;
        }

        const allowedTypes = ["image/jpeg", "image/png", "image/webp"];

        if (!allowedTypes.includes(file.type)) {
            message.textContent = "Please choose a JPG, PNG or WEBP image.";
            input.value = "";
            selectedFile = null;
            uploadButton.disabled = true;
            return;
        }

        if (file.size > 5 * 1024 * 1024) {
            message.textContent = "Image must be 5 MB or smaller.";
            input.value = "";
            selectedFile = null;
            uploadButton.disabled = true;
            return;
        }

        selectedFile = file;

        if (previewURL) {
            URL.revokeObjectURL(previewURL);
        }

        previewURL = URL.createObjectURL(file);
        image.src = previewURL;
        image.style.visibility = "visible";

        uploadButton.disabled = false;
        message.textContent = "Preview ready. Click Upload Picture to save.";
    });

    uploadButton.addEventListener("click", async () => {
        if (!selectedFile) {
            message.textContent = "Please choose a picture first.";
            return;
        }

        const formData = new FormData();
        formData.append("picture", selectedFile);

        uploadButton.disabled = true;
        chooseButton.disabled = true;
        message.textContent = "Uploading picture...";

        try {
const response = await fetch("/api/profile-picture", {
    method: "POST",
    body: formData,
    credentials: "same-origin",
    headers: {
        "X-Requested-With": "ResearchBridge"
    }
});

            if (!response.ok) {
                let errorMessage = "Could not upload picture.";

                try {
                    const result = await response.json();
                    errorMessage = result.message || result.error || errorMessage;
                } catch (_) {
                    // Keep the default error message.
                }

                throw new Error(errorMessage);
            }

            message.textContent = "Profile picture saved successfully.";
            selectedFile = null;
            input.value = "";

            image.src = "/api/profile-picture?t=" + Date.now();

            if (previewURL) {
                URL.revokeObjectURL(previewURL);
                previewURL = null;
            }
        } catch (error) {
            message.textContent = error.message || "Upload failed. Please try again.";
        } finally {
            chooseButton.disabled = false;
            uploadButton.disabled = true;
        }
    });
})();


/* ============================================
   SUPERVISOR DIRECTORY + SUPERVISION REQUESTS
   Uses GET /api/teacher-directory
============================================ */

(() => {
    const form = document.getElementById("supervisor-request-form");
    if (!form) return;

    const teacherSelect = document.getElementById("supervisor-teacher");
    const cardsBox = document.getElementById("supervisor-teacher-cards");
    const searchBox = document.getElementById("supervisor-search");
    const countBox = document.getElementById("supervisor-teacher-count");
    const directoryMessage = document.getElementById("supervisor-directory-message");
    const requestMessage = document.getElementById("supervisor-request-message");
    const requestList = document.getElementById("supervisor-request-list");
    const sendButton = document.getElementById("supervisor-send-button");

    let teachers = [];

    function escapeHTML(value) {
        return String(value ?? "").replace(/[&<>"']/g, char => ({
            "&": "&amp;",
            "<": "&lt;",
            ">": "&gt;",
            '"': "&quot;",
            "'": "&#39;"
        })[char]);
    }

    function shown(value) {
        const text = String(value ?? "").trim();
        return text ? escapeHTML(text) : "Not added yet";
    }

    function initials(name) {
        return String(name || "Teacher")
            .trim()
            .split(/\s+/)
            .slice(0, 2)
            .map(part => Array.from(part)[0] || "")
            .join("")
            .toUpperCase();
    }

    function listValues(value) {
        if (Array.isArray(value)) {
            return value.map(item => String(item).trim()).filter(Boolean);
        }

        return String(value || "")
            .split(/[,;\n|]+/)
            .map(item => item.trim())
            .filter(Boolean);
    }

    function tags(value) {
        const items = listValues(value);

        if (!items.length) {
            return '<span class="rb-tag rb-tag-muted">Not added yet</span>';
        }

        return items.map(item =>
            `<span class="rb-tag">${escapeHTML(item)}</span>`
        ).join("");
    }

    function safeHTTPURL(value) {
        try {
            const url = new URL(String(value || ""));
            if (url.protocol === "https:" || url.protocol === "http:") {
                return url.href;
            }
        } catch {
            // Empty, malformed, or unsafe link.
        }

        return "";
    }

    function externalLink(url, label) {
        if (!url) return "";
        return `<a href="${escapeHTML(url)}"
                   target="_blank"
                   rel="noopener noreferrer">${label} ↗</a>`;
    }

    function acceptingStatus(value) {
        if (value === true) return "Accepting students";
        if (value === false) return "Not accepting students";
        return "Not specified";
    }

    function teacherCard(teacher) {
        const appID = Number(teacher.account_id || 0);
        const websiteURL = safeHTTPURL(teacher.website_url);
        const photoURL = safeHTTPURL(teacher.photo_url);
        const email = String(teacher.email || "").trim();
        const phone = String(teacher.phone || "").trim();

        const validEmail =
            /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);

        return `
            <article class="rb-teacher-card"
                     data-directory-id="${escapeHTML(teacher.id)}"
                     data-account-id="${appID || ""}">

                <div class="rb-teacher-top">
                    <div class="rb-teacher-avatar" aria-hidden="true">
                        ${escapeHTML(initials(teacher.name))}
                    </div>

                    <div class="rb-teacher-identity">
                        <span class="rb-teacher-role">
                            ${teacher.verified
                                ? "Verified directory entry"
                                : "Directory entry"}
                        </span>
                        <h3>${shown(teacher.name)}</h3>
                        <p class="rb-teacher-email">
                            ${validEmail
                                ? escapeHTML(email)
                                : "Email not added"}
                        </p>
                    </div>
                </div>

                <dl class="rb-teacher-info">
                    <div>
                        <dt>Designation</dt>
                        <dd>${shown(teacher.designation)}</dd>
                    </div>
                    <div>
                        <dt>Institution</dt>
                        <dd>${shown(teacher.institution)}</dd>
                    </div>
                    <div>
                        <dt>Department</dt>
                        <dd>${shown(teacher.department)}</dd>
                    </div>
                    <div>
                        <dt>Phone</dt>
                        <dd>${phone ? escapeHTML(phone) : "Not added yet"}</dd>
                    </div>
                    <div>
                        <dt>Student Capacity</dt>
                        <dd>${acceptingStatus(teacher.accepting_students)}</dd>
                    </div>
                    <div>
                        <dt>Directory Source</dt>
                        <dd>${shown(teacher.source)}</dd>
                    </div>
                </dl>

                <div class="rb-tag-section">
                    <span class="rb-tag-title">Research Interests</span>
                    <div class="rb-tag-list">
                        ${tags(teacher.research_interests)}
                    </div>
                </div>

                <div class="rb-tag-section">
                    <span class="rb-tag-title">Expertise</span>
                    <div class="rb-tag-list">
                        ${tags(teacher.expertise)}
                    </div>
                </div>

                <div class="rb-teacher-actions">
                    ${validEmail
                        ? `<a href="mailto:${escapeHTML(email)}">Email Teacher ↗</a>`
                        : ""}

                    ${externalLink(websiteURL, "Website")}
                    ${externalLink(photoURL, "View Photo")}
                </div>

                <div class="rb-teacher-select-row">
                    ${appID
                        ? `<button
                               type="button"
                               class="rb-select-teacher"
                               data-select-account="${appID}">
                               Select for Request →
                           </button>`
                        : `<button type="button"
                                   class="rb-select-teacher"
                                   disabled
                                   title="This teacher has no matching verified ResearchBridge account">
                               Account Not Linked
                           </button>`
                    }
                </div>
            </article>
        `;
    }

    function markSelectedTeacher() {
        const selectedID = String(teacherSelect.value || "");

        cardsBox.querySelectorAll(".rb-teacher-card").forEach(card => {
            const cardID = String(card.dataset.accountId || "");
            card.classList.toggle(
                "is-selected",
                selectedID !== "" && cardID === selectedID
            );
        });
    }

    function renderTeachers(list) {
        countBox.textContent =
            list.length + (list.length === 1 ? " teacher" : " teachers");

        if (!list.length) {
            cardsBox.innerHTML = `
                <div class="rb-empty-state">
                    <strong>No teachers found</strong>
                    <p>Try a different name, institution, department or research interest.</p>
                </div>
            `;
            return;
        }

        cardsBox.innerHTML = list.map(teacherCard).join("");
        markSelectedTeacher();
    }

    function filterTeachers() {
        const query = searchBox.value.trim().toLowerCase();

        const filtered = teachers.filter(teacher => {
            const allFields = [
                teacher.name,
                teacher.designation,
                teacher.institution,
                teacher.department,
                teacher.phone,
                teacher.email,
                teacher.website_url,
                teacher.photo_url,
                teacher.research_interests,
                teacher.expertise,
                teacher.source,
                acceptingStatus(teacher.accepting_students),
                teacher.verified ? "verified" : "not verified"
            ].flat().join(" ").toLowerCase();

            return allFields.includes(query);
        });

        renderTeachers(filtered);
    }

    async function api(path, data) {
        const options = {
            method: data === undefined ? "GET" : "POST",
            credentials: "same-origin",
            cache: "no-store",
            headers: {}
        };

        if (data !== undefined) {
            options.headers["Content-Type"] = "application/json";
            options.headers["X-Requested-With"] = "ResearchBridge";
            options.body = JSON.stringify(data);
        }

        const response = await fetch(path, options);

        if (response.status === 401) {
            window.location.replace("/login/student");
            throw new Error("Please log in again.");
        }

        let result;
        try {
            result = await response.json();
        } catch {
            throw new Error("The server returned an invalid response.");
        }

        if (!response.ok) {
            throw new Error(result.error || "Request failed.");
        }

        return result;
    }

    async function loadTeachers() {
        cardsBox.innerHTML = "";
        countBox.textContent = "Loading...";
        directoryMessage.textContent = "Loading teacher profiles...";

        try {
            const result = await api("/api/teacher-directory");
            teachers = Array.isArray(result) ? result : [];

            teacherSelect.replaceChildren();

            const firstOption = document.createElement("option");
            firstOption.value = "";
            firstOption.textContent = "Choose a registered teacher";
            teacherSelect.appendChild(firstOption);

            // Only verified ResearchBridge accounts can receive requests.
            const linkedIDs = new Set();

            teachers.forEach(teacher => {
                const accountID = Number(teacher.account_id || 0);

                if (!accountID || linkedIDs.has(String(accountID))) return;

                linkedIDs.add(String(accountID));

                const option = document.createElement("option");
                option.value = String(accountID);
                option.textContent =
                    (teacher.name || "Unnamed teacher") +
                    (teacher.designation
                        ? " — " + teacher.designation
                        : "");

                teacherSelect.appendChild(option);
            });

            const linkedCount = linkedIDs.size;

            directoryMessage.textContent = teachers.length
                ? `${teachers.length} directory profile(s) loaded. ` +
                  `${linkedCount} profile(s) match a verified ResearchBridge teacher account.`
                : "No teacher directory records are available.";

            filterTeachers();

            if (teachers.length > 0 && linkedCount === 0) {
                requestMessage.textContent =
                    "Profiles are visible, but no matching verified ResearchBridge teacher accounts were found. Requests cannot be sent until accounts are linked by email.";
            }
        } catch (error) {
            teachers = [];
            countBox.textContent = "Unavailable";
            directoryMessage.textContent = error.message;
            cardsBox.innerHTML = `
                <div class="rb-empty-state">
                    <strong>Teacher directory unavailable</strong>
                    <p>${escapeHTML(error.message)}</p>
                    <p>Check the PostgreSQL connection and teacher_directory table.</p>
                </div>
            `;
        }
    }

    async function loadMyRequests() {
        requestList.textContent = "Loading your requests...";

        try {
            const projects = await api("/api/projects");
            if (Array.isArray(projects)) {
    populateSupervisorProjects(projects, teachers);
}

            if (!Array.isArray(projects) || projects.length === 0) {
                requestList.innerHTML = `
                    <div class="rb-empty-state">
                        <strong>No requests yet</strong>
                        <p>Your submitted proposals and teacher decisions will appear here.</p>
                    </div>
                `;
                return;
            }

            const statusLabels = {
                draft: "Draft — not sent",
                submitted: "Pending teacher decision",
                approved: "Accepted — supervision approved",
                changes_requested: "Changes requested",
                rejected: "Rejected"
            };

            requestList.innerHTML = projects.map(project => {
                const status = project.status || "draft";
                const statusClass =
                    status === "approved" ? "rb-status-approved" :
                    status === "rejected" ? "rb-status-rejected" : "";

                return `
                    <article class="rb-request-card">
                        <h3>${shown(project.title)}</h3>
                        <p><strong>Research Area:</strong> ${shown(project.area)}</p>
                        <p><strong>Teacher:</strong> ${shown(project.teacher)}</p>
                        <span class="rb-request-status ${statusClass}">
                            ${escapeHTML(statusLabels[status] || status)}
                        </span>
                        ${project.feedback
                            ? `<p><strong>Teacher feedback:</strong>
                               ${escapeHTML(project.feedback)}</p>`
                            : ""}
                    </article>
                `;
            }).join("");
        } catch (error) {
            requestList.textContent = error.message;
        }
    }
    function populateSupervisorProjects(projects, teachers) {
        const projectSelect = document.getElementById("supervisor-project");
        if (!projectSelect) return;
        const previousProject = projectSelect.value;
        projectSelect.replaceChildren();
        for (const project of projects) {
            projectSelect.add(new Option(project.title, String(project.id)));
        }
        if (!projects.length) projectSelect.add(new Option("Create a project first", ""));
        const selected = projects.find(p => String(p.id) === previousProject) || projects[0];
        if (selected) projectSelect.value = String(selected.id);
        projectSelect.disabled = !projects.length;
        const updateSelection = () => {
            const project = projects.find(p => String(p.id) === projectSelect.value);
            teacherSelect.replaceChildren(new Option("Choose a registered teacher", ""));
            const added = new Set();
            teachers.forEach(t => {
                const accountID = Number(t.account_id || 0);
                if (!accountID || added.has(accountID)) return;
                added.add(accountID);
                teacherSelect.add(new Option(t.name || "Teacher", String(accountID)));
            });
            if (project?.teacher_id) {
                if (!added.has(Number(project.teacher_id))) {
                    teacherSelect.add(new Option(project.teacher || "Selected supervisor", String(project.teacher_id)));
                }
                teacherSelect.value = String(project.teacher_id);
            }
            teacherSelect.disabled = Boolean(project?.teacher_id);
            sendButton.disabled = !project || !["draft", "changes_requested"].includes(project.status);
            markSelectedTeacher();
        };
        projectSelect.onchange = updateSelection;
        updateSelection();
    }

    searchBox.addEventListener("input", filterTeachers);
    teacherSelect.addEventListener("change", markSelectedTeacher);

    cardsBox.addEventListener("click", event => {
        const button = event.target.closest("[data-select-account]");
        if (!button) return;

        const accountID = String(button.dataset.selectAccount || "");
        const teacher = teachers.find(
            item => String(item.account_id || "") === accountID
        );

        if (!teacher) return;

        teacherSelect.value = accountID;
        markSelectedTeacher();

        requestMessage.textContent =
            "Selected " + (teacher.name || "teacher") +
            ". Complete the form below to send your request.";

        document.getElementById("supervisor-request-panel")
            .scrollIntoView({ behavior: "smooth", block: "start" });
    });

    form.addEventListener("submit", async event => {
        event.preventDefault();
        const projectID = Number(document.getElementById("supervisor-project").value);
        const teacherID = Number(teacherSelect.value);
        if (!projectID || !teacherID) {
            requestMessage.textContent = "Choose a project and a registered supervisor first.";
            return;
        }
        sendButton.disabled = true;
        try {
            await api(`/api/projects/${projectID}/supervisor`, { teacher_id: teacherID });
            requestMessage.textContent = "Supervisor saved. Complete your Proposal and click Submit for review to send the request and notify your teacher.";
            await loadMyRequests();
        } catch (error) {
            requestMessage.textContent = error.message;
        } finally {
            sendButton.disabled = false;
        }
    });

    loadTeachers().then(loadMyRequests);
})();
