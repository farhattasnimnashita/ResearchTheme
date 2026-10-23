"use strict";

(() => {
    const page = document.getElementById("proposal");

    if (!page) return;

    const form = document.getElementById("proposal-form");
    const projectSelect = document.getElementById("proposal-project");
    const message = document.getElementById("proposal-message");
    const projectSummary = document.getElementById("proposal-project-summary");
    const proposalList = document.getElementById("proposal-list");

    const saveButton = document.getElementById("proposal-save-button");
    const submitButton = document.getElementById("proposal-submit-button");

    const fields = {
        problem_statement: document.getElementById("proposal-problem"),
        objectives: document.getElementById("proposal-objectives"),
        methodology: document.getElementById("proposal-methodology"),
        expected_outcomes: document.getElementById("proposal-outcomes"),
        scope_limitations: document.getElementById("proposal-scope")
    };

    let projects = [];

    function escapeHTML(value) {
        return String(value ?? "").replace(/[&<>"']/g, char => ({
            "&": "&amp;",
            "<": "&lt;",
            ">": "&gt;",
            '"': "&quot;",
            "'": "&#39;"
        })[char]);
    }

    function hasProposal(project) {
        return project.proposal_started === true ||
            project.proposal_started === 1;
    }

    function editable(project) {
        return project.status === "draft" ||
            (project.status === "changes_requested" && hasProposal(project));
    }

    function showMessage(text, isError = false) {
        message.textContent = text;
        message.classList.toggle("proposal-error", isError);
    }

    async function api(path, method = "GET", data) {
        const options = {
            method,
            credentials: "same-origin",
            cache: "no-store"
        };

        if (method !== "GET") {
            options.headers = {
                "Content-Type": "application/json",
                "X-Requested-With": "ResearchBridge"
            };

            options.body = JSON.stringify(data ?? {});
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
            throw new Error(result.error || "The request failed.");
        }

        return result;
    }

    function currentProject() {
        const id = Number(projectSelect.value);

        return projects.find(project => project.id === id);
    }

    function renderProjectOptions(preferredID = 0) {
        const candidates = projects.filter(editable);

        projectSelect.innerHTML = "";

        if (!candidates.length) {
            projectSelect.innerHTML =
                '<option value="">No editable project available</option>';

            projectSummary.textContent =
                "Go to Choose Project and save a research project first.";

            fillSelectedProject();
            return;
        }

        candidates.forEach(project => {
            const option = document.createElement("option");

            option.value = String(project.id);
            option.textContent = project.title;

            projectSelect.appendChild(option);
        });

        const preferred = candidates.find(
            project => project.id === Number(preferredID)
        );

        const existingDraft = candidates.find(hasProposal);
        const selected = preferred || existingDraft || candidates[0];

        projectSelect.value = String(selected.id);

        fillSelectedProject();
    }

    function fillSelectedProject() {
        const project = currentProject();

        if (!project) {
            Object.values(fields).forEach(field => {
                field.value = "";
            });

            saveButton.disabled = true;
            submitButton.disabled = true;
            return;
        }

        projectSummary.textContent =
            "Project: " + project.title +
            " | Research Area: " + (project.area || "Not specified");
        const supervisorSummary = document.getElementById(
    "proposal-supervisor-summary"
);

if (supervisorSummary) {
    supervisorSummary.textContent =
        project.teacher && project.teacher !== "Not assigned"
            ? project.teacher
            : "No supervisor selected. Open the Supervisor section first.";
}

        fields.problem_statement.value =
            project.problem_statement || "";

        fields.objectives.value =
            project.objectives || "";

        fields.methodology.value =
            project.methodology || "";

        fields.expected_outcomes.value =
            project.expected_outcomes || "";

        fields.scope_limitations.value =
            project.scope_limitations || "";

const isLocked =
    project.status === "submitted" ||
    project.status === "approved";

projectSelect.disabled = isLocked;

Object.values(fields).forEach(field => {
    field.readOnly = isLocked;
});

saveButton.disabled = isLocked;
submitButton.disabled = isLocked;
        }
    function renderSavedProposals() {
        const saved = projects.filter(hasProposal);

        if (!saved.length) {
            proposalList.innerHTML = `
                <div class="proposal-empty">
                    <h3>No proposal draft yet</h3>
                    <p>
                        Choose an existing project, write your research plan,
                        and click Save Draft.
                    </p>
                </div>
            `;
            return;
        }

        proposalList.innerHTML = saved.map(project => {
            const status = escapeHTML(
                String(project.status || "draft").replaceAll("_", " ")
            );

            const details = [
                ["Problem Statement", project.problem_statement],
                ["Objectives", project.objectives],
                ["Proposed Methodology", project.methodology],
                ["Expected Outcomes", project.expected_outcomes],
                ["Scope and Limitations", project.scope_limitations]
            ].filter(item => String(item[1] || "").trim());

            const detailHTML = details.map(([label, value]) => `
                <div class="proposal-detail">
                    <dt>${escapeHTML(label)}</dt>
                    <dd>${escapeHTML(value)}</dd>
                </div>
            `).join("");

            const feedback = project.feedback
                ? `<div class="proposal-feedback">
                    <strong>Supervisor Feedback</strong>
                    <p>${escapeHTML(project.feedback)}</p>
                   </div>`
                : "";

            const updated = project.updated
                ? new Date(project.updated * 1000).toLocaleString()
                : "Not available";

            return `
                <article class="proposal-record">
                    <div class="proposal-record-heading">
                        <div>
                            <h3>${escapeHTML(project.title)}</h3>
                            <p>${escapeHTML(project.area || "Research project")}</p>
                        </div>

                        <span class="proposal-status">
                            ${status}
                        </span>
                    </div>

                    <p class="proposal-meta">
                        Supervisor: ${escapeHTML(project.teacher || "Not assigned")}
                    </p>

                    <details class="proposal-record-details">
                        <summary>View research plan</summary>
                        <dl>${detailHTML || "<p>No details saved.</p>"}</dl>
                    </details>

                    ${feedback}

                    <p class="proposal-meta">Last updated: ${escapeHTML(updated)}</p>

                    ${
                        editable(project)
                            ? `<button
                                type="button"
                                class="button secondary"
                                data-edit-proposal="${project.id}"
                            >Edit Draft / Revision</button>`
                            : ""
                    }
                </article>
            `;
        }).join("");
    }

    async function load(preferredID = 0) {
        try {
            const result = await api("/api/projects");
            if (!Array.isArray(result)) throw new Error("Invalid project list. Please reload.");
            projects = result;
            renderProjectOptions(preferredID);
            renderSavedProposals();

            if (!projects.some(editable)) {
                showMessage(
                    "First go to Choose Project and save a project."
                );
            } else {
                showMessage(
                    "Write your proposal below. You can save an incomplete draft."
                );
            }
        } catch (error) {
            showMessage(error.message, true);
            projectSelect.innerHTML = '<option value="">Could not load projects — reload to retry</option>';
            saveButton.disabled = true;
            submitButton.disabled = true;
            proposalList.textContent = "Could not load saved proposals.";
        }
    }

function collectProposal() {
    return {
        problem_statement: fields.problem_statement.value.trim(),
        objectives: fields.objectives.value.trim(),
        methodology: fields.methodology.value.trim(),
        expected_outcomes: fields.expected_outcomes.value.trim(),
        scope_limitations: fields.scope_limitations.value.trim()
    };
}

    async function saveDraft(showSuccess = true) {
        const projectID = Number(projectSelect.value);

        if (!projectID) {
            throw new Error("Choose a saved project first.");
        }

        const data = collectProposal();

        await api(
            `/api/projects/${projectID}/proposal`,
            "POST",
            data
        );

        await load(projectID);

        if (showSuccess) {
            showMessage(
                "Draft saved successfully. You can continue editing it later."
            );
        }

        return projectID;
    }

    saveButton.addEventListener("click", async () => {
        saveButton.disabled = true;
        submitButton.disabled = true;

        try {
            await saveDraft();
        } catch (error) {
            showMessage(error.message, true);
        } finally {
            const locked = !currentProject() || !editable(currentProject());
            saveButton.disabled = locked;
            submitButton.disabled = locked;
        }
    });

    form.addEventListener("submit", async event => {
        event.preventDefault();

        const projectID = Number(projectSelect.value);
        const data = collectProposal();

        if (!projectID) {
            showMessage("Choose a project before submitting.", true);
            return;
        }

        if (!Number(currentProject()?.teacher_id)) {
            showMessage(
                "Choose a supervisor before submitting the proposal.",
                true
            );
            return;
        }

        const requirements = [
            ["Problem Statement", data.problem_statement, 20, 6000],
            ["Objectives", data.objectives, 10, 4000],
            ["Proposed Methodology", data.methodology, 20, 6000],
            ["Expected Outcomes", data.expected_outcomes, 10, 4000],
            ["Scope and Limitations", data.scope_limitations, 10, 3000]
        ];

        for (const [label, value, min, max] of requirements) {
            if (value.length < min || value.length > max) {
                showMessage(
                    `${label} must contain ${min}–${max} characters.`,
                    true
                );
                return;
            }
        }

        saveButton.disabled = true;
        submitButton.disabled = true;

        try {
            // First persist the final form contents.
            const id = await saveDraft(false);

            // Then submit the saved proposal for review.
            await api(
                `/api/projects/${id}/proposal/submit`,
                "POST",
                {}
            );

            await load();

            showMessage(
                "Proposal submitted successfully. You can track the teacher's decision below."
            );
        } catch (error) {
            showMessage(error.message, true);
        } finally {
            const locked = !currentProject() || !editable(currentProject());
            saveButton.disabled = locked;
            submitButton.disabled = locked;
        }
    });

    projectSelect.addEventListener("change", fillSelectedProject);

    proposalList.addEventListener("click", event => {
        const button = event.target.closest("[data-edit-proposal]");

        if (!button) return;

        const id = Number(button.dataset.editProposal);

        projectSelect.value = String(id);
        fillSelectedProject();

        form.scrollIntoView({
            behavior: "smooth",
            block: "start"
        });

        showMessage("Edit your draft or make the requested revisions.");
    });

    load();
})();