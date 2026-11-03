"use strict";

document.addEventListener("DOMContentLoaded", async () => {
    const experimentPage = document.getElementById("experiment-reproducibility");
    const feedbackPage = document.getElementById("feedback-resolution");

    if (!experimentPage || !feedbackPage) return;

    let projects = [];
    let experiments = [];
    let feedbackRecords = [];

    async function api(path, method = "GET", data = null) {
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
            if (data) options.body = JSON.stringify(data);
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

    function escapeHTML(value) {
        return String(value ?? "").replace(/[&<>"']/g, char => ({
            "&": "&amp;",
            "<": "&lt;",
            ">": "&gt;",
            '"': "&quot;",
            "'": "&#39;"
        })[char]);
    }

    function safeURL(value) {
        if (!value) return "";
        try {
            const url = new URL(value);
            if (url.protocol === "https:" || url.protocol === "http:") {
                return url.href;
            }
        } catch {
            return "";
        }
        return "";
    }

    function fillProjectSelect(select) {
        const previous = select.value;

        select.replaceChildren();

        const placeholder = document.createElement("option");
        placeholder.value = "";
        placeholder.textContent = projects.length
            ? "Select a research project"
            : "No accessible projects";
        select.appendChild(placeholder);

        projects.forEach(project => {
            const option = document.createElement("option");
            option.value = String(project.id);
            option.textContent = project.title;
            select.appendChild(option);
        });

        if (projects.some(project => String(project.id) === previous)) {
            select.value = previous;
        }
    }

    function formData(form, fields) {
        const data = {};

        fields.forEach(field => {
            const control = form.elements.namedItem(field);
            data[field] = control ? control.value.trim() : "";
        });

        return data;
    }

    function setFormValues(form, record, fields) {
        fields.forEach(field => {
            const control = form.elements.namedItem(field);
            if (control) control.value = record[field] ?? "";
        });
    }

    function resetForm(form, cancelButton, message) {
        form.reset();
        form.elements.namedItem("id").value = "";
        cancelButton.hidden = true;
        message.textContent = "";
    }

    function addText(parent, label, value) {
        if (value === undefined || value === null || String(value).trim() === "") {
            return;
        }

        const item = document.createElement("div");
        item.className = "tracking-detail";

        const heading = document.createElement("strong");
        heading.textContent = label;

        const content = document.createElement("p");
        content.textContent = String(value);

        item.append(heading, content);
        parent.appendChild(item);
    }

    function addLink(parent, label, value) {
        const url = safeURL(value);
        if (!url) return;

        const item = document.createElement("div");
        item.className = "tracking-detail";

        const heading = document.createElement("strong");
        heading.textContent = label;

        const link = document.createElement("a");
        link.href = url;
        link.textContent = "Open link";
        link.target = "_blank";
        link.rel = "noopener noreferrer";

        item.append(heading, link);
        parent.appendChild(item);
    }

    function createActionButton(label, action, id) {
        const button = document.createElement("button");
        button.type = "button";
        button.textContent = label;
        button.dataset.action = action;
        button.dataset.id = String(id);
        return button;
    }

    function createCard(title, projectTitle, id, type, details) {
        const card = document.createElement("article");
        card.className = "tracking-record-card";

        const header = document.createElement("div");
        header.className = "tracking-card-heading";

        const heading = document.createElement("h3");
        heading.textContent = title;

        const project = document.createElement("p");
        project.className = "tracking-project-name";
        project.textContent = projectTitle;

        header.append(heading, project);
        card.appendChild(header);

        const body = document.createElement("div");
        body.className = "tracking-details";

        details(body);
        card.appendChild(body);

        const actions = document.createElement("div");
        actions.className = "tracking-actions";

        actions.append(
            createActionButton("Edit", "edit-" + type, id),
            createActionButton("Delete", "delete-" + type, id)
        );

        card.appendChild(actions);
        return card;
    }

    function renderExperiments() {
        const list = document.getElementById("experiment-record-list");
        const empty = document.getElementById("experiment-empty");
        const search = document.getElementById("experiment-search").value.trim().toLowerCase();
        const projectID = document.getElementById("experiment-project").value;

        list.replaceChildren();

        const matches = experiments.filter(record =>
            (!projectID || String(record.project_id) === projectID) &&
            JSON.stringify(record).toLowerCase().includes(search)
        );

        empty.hidden = matches.length > 0;
        empty.textContent = experiments.length
            ? "No experiments match your search."
            : "No experiment records yet. Add your first experiment above.";

        matches.forEach(record => {
            const card = createCard(
                record.experiment_name,
                record.project_title,
                record.id,
                "experiment",
                body => {
                    addText(body, "Hypothesis", record.hypothesis);
                    addText(body, "Run Date", record.run_date);
                    addText(body, "Protocol", record.protocol);
                    addText(body, "Dataset Version", record.dataset_version);
                    addLink(body, "Code Link", record.code_link);
                    addText(body, "Environment", record.environment);
                    addText(body, "Parameters", record.parameters);
                    addText(body, "Random Seed", record.random_seed);
                    addText(body, "Result Summary", record.result_summary);
                    addLink(body, "Artifact Link", record.artifact_link);
                    addText(body, "Notes", record.notes);
                }
            );
            list.appendChild(card);
        });
    }

    function renderFeedback() {
        const list = document.getElementById("feedback-record-list");
        const empty = document.getElementById("feedback-empty");
        const search = document.getElementById("feedback-search").value.trim().toLowerCase();
        const projectID = document.getElementById("feedback-project").value;

        list.replaceChildren();

        const matches = feedbackRecords.filter(record =>
            (!projectID || String(record.project_id) === projectID) &&
            JSON.stringify(record).toLowerCase().includes(search)
        );

        empty.hidden = matches.length > 0;
        empty.textContent = feedbackRecords.length
            ? "No feedback records match your search."
            : "No feedback records yet. Add your first feedback above.";

        matches.forEach(record => {
            const card = createCard(
                record.comment_summary,
                record.project_title,
                record.id,
                "feedback",
                body => {
                    addText(body, "Source", record.source);
                    addText(body, "Received Date", record.received_date);
                    addText(body, "Priority", record.priority);
                    addText(body, "Owner", record.owner);
                    addText(body, "Response Plan", record.response_plan);
                    addText(body, "Status", record.status);
                    addText(body, "Due Date", record.due_date);
                    addLink(body, "Evidence Link", record.evidence_link);
                    addText(body, "Resolution Note", record.resolution_note);
                }
            );
            list.appendChild(card);
        });
    }

    const experimentForm = document.getElementById("experiment-form");
    const feedbackForm = document.getElementById("feedback-form");

    const experimentCancel = document.getElementById("experiment-cancel");
    const feedbackCancel = document.getElementById("feedback-cancel");

    const experimentMessage = document.getElementById("experiment-message");
    const feedbackMessage = document.getElementById("feedback-message");

    const experimentFields = [
        "id", "project_id", "experiment_name", "hypothesis",
        "run_date", "protocol", "dataset_version", "code_link",
        "environment", "parameters", "random_seed",
        "result_summary", "artifact_link", "notes"
    ];

    const feedbackFields = [
        "id", "project_id", "source", "received_date",
        "comment_summary", "priority", "owner", "response_plan",
        "status", "due_date", "evidence_link", "resolution_note"
    ];

    async function loadRecords() {
        const results = await Promise.all([
            api("/api/experiment-reproducibility"),
            api("/api/feedback-resolution")
        ]);

        experiments = results[0];
        feedbackRecords = results[1];

        renderExperiments();
        renderFeedback();
    }

    experimentForm.addEventListener("submit", async event => {
        event.preventDefault();
        if (!experimentForm.reportValidity()) return;

        const saveButton = document.getElementById("experiment-save");
        saveButton.disabled = true;

        try {
            const data = formData(experimentForm, experimentFields);
            data.id = Number(data.id) || 0;
            data.project_id = Number(data.project_id);

            const result = await api(
                "/api/experiment-reproducibility",
                "POST",
                data
            );

            experimentMessage.textContent = result.message;
            resetForm(experimentForm, experimentCancel, experimentMessage);
            await loadRecords();
        } catch (error) {
            experimentMessage.textContent = error.message;
        } finally {
            saveButton.disabled = false;
        }
    });

    feedbackForm.addEventListener("submit", async event => {
        event.preventDefault();
        if (!feedbackForm.reportValidity()) return;

        const saveButton = document.getElementById("feedback-save");
        saveButton.disabled = true;

        try {
            const data = formData(feedbackForm, feedbackFields);
            data.id = Number(data.id) || 0;
            data.project_id = Number(data.project_id);

            const result = await api(
                "/api/feedback-resolution",
                "POST",
                data
            );

            feedbackMessage.textContent = result.message;
            resetForm(feedbackForm, feedbackCancel, feedbackMessage);
            await loadRecords();
        } catch (error) {
            feedbackMessage.textContent = error.message;
        } finally {
            saveButton.disabled = false;
        }
    });

    experimentCancel.addEventListener("click", () => {
        resetForm(experimentForm, experimentCancel, experimentMessage);
    });

    feedbackCancel.addEventListener("click", () => {
        resetForm(feedbackForm, feedbackCancel, feedbackMessage);
    });

    document.getElementById("experiment-record-list").addEventListener("click", async event => {
        const button = event.target.closest("button[data-action]");
        if (!button) return;

        const id = Number(button.dataset.id);
        const record = experiments.find(item => Number(item.id) === id);
        if (!record) return;

        if (button.dataset.action === "edit-experiment") {
            setFormValues(experimentForm, record, experimentFields);
            experimentCancel.hidden = false;
            experimentMessage.textContent = "Editing existing experiment.";
            experimentForm.scrollIntoView({ behavior: "smooth", block: "start" });
            return;
        }

        if (button.dataset.action === "delete-experiment" &&
            window.confirm("Delete this experiment record?")) {
            try {
                await api("/api/experiment-reproducibility?id=" + id, "DELETE");
                await loadRecords();
                experimentMessage.textContent = "Experiment deleted.";
            } catch (error) {
                experimentMessage.textContent = error.message;
            }
        }
    });

    document.getElementById("feedback-record-list").addEventListener("click", async event => {
        const button = event.target.closest("button[data-action]");
        if (!button) return;

        const id = Number(button.dataset.id);
        const record = feedbackRecords.find(item => Number(item.id) === id);
        if (!record) return;

        if (button.dataset.action === "edit-feedback") {
            setFormValues(feedbackForm, record, feedbackFields);
            feedbackCancel.hidden = false;
            feedbackMessage.textContent = "Editing existing feedback.";
            feedbackForm.scrollIntoView({ behavior: "smooth", block: "start" });
            return;
        }

        if (button.dataset.action === "delete-feedback" &&
            window.confirm("Delete this feedback record?")) {
            try {
                await api("/api/feedback-resolution?id=" + id, "DELETE");
                await loadRecords();
                feedbackMessage.textContent = "Feedback deleted.";
            } catch (error) {
                feedbackMessage.textContent = error.message;
            }
        }
    });

    document.getElementById("experiment-search").addEventListener("input", renderExperiments);
    document.getElementById("feedback-search").addEventListener("input", renderFeedback);
    document.getElementById("experiment-project").addEventListener("change", renderExperiments);
    document.getElementById("feedback-project").addEventListener("change", renderFeedback);

    try {
        projects = await api("/api/projects");

        fillProjectSelect(document.getElementById("experiment-project"));
        fillProjectSelect(document.getElementById("feedback-project"));

        await loadRecords();
    } catch (error) {
        experimentMessage.textContent = error.message;
        feedbackMessage.textContent = error.message;
    }
});