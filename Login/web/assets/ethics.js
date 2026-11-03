"use strict";

document.addEventListener("DOMContentLoaded", async function () {
    const page = document.getElementById("research-ethics");
    if (!page) return;

    const form = document.getElementById("ethics-form");
    const projectSelect = document.getElementById("ethics-project");
    const requiredSelect = document.getElementById("ethics-required");
    const reviewStatus = document.getElementById("ethics-review-status");
    const statusMessage = document.getElementById("research-ethics-status");
    const statusHelp = document.getElementById("ethics-status-help");
    const projectHint = document.getElementById("ethics-project-hint");
    const noProject = document.getElementById("ethics-no-project");
    const saveButton = document.getElementById("ethics-save-button");
    const requestButton = document.getElementById("ethics-request-review-button");
    const discardButton = document.getElementById("ethics-discard-button");
    const recordsContainer = document.getElementById("ethics-records");
    const recordsEmpty = document.getElementById("ethics-records-empty");

    if (!form || !projectSelect || !requiredSelect || !reviewStatus || !statusMessage || !saveButton || !requestButton || !recordsContainer) {
        console.error("Research Ethics page elements are missing.");
        return;
    }

    let projects = [];
    let records = [];
    let currentProjectID = 0;
    let fillingForm = false;

    const fields = [
        "ethics_required", "committee", "application_reference", "review_status",
        "submission_date", "decision_date", "approval_number", "expiry_date",
        "participants_involved", "consent_method", "consent_plan",
        "committee_conditions", "document_url"
    ];

    const labels = {
        ethics_required: "Ethics review required",
        committee: "Review body / committee",
        application_reference: "Application reference",
        review_status: "Ethics committee status",
        submission_date: "Submission date",
        decision_date: "Decision date",
        approval_number: "Approval / exemption number",
        expiry_date: "Approval expiry date",
        participants_involved: "Human participants involved",
        consent_method: "Consent method",
        consent_plan: "Consent plan",
        committee_conditions: "Committee conditions / requested changes",
        document_url: "Official document link"
    };

    async function api(path, method = "GET", data) {
        const options = { method, credentials: "same-origin", cache: "no-store" };
        if (method !== "GET") {
            options.headers = {
                "Content-Type": "application/json",
                "X-Requested-With": "ResearchBridge"
            };
            options.body = JSON.stringify(data || {});
        }
        const response = await fetch(path, options);
        if (response.status === 401) {
            window.location.replace("/login/student");
            throw new Error("Please log in again.");
        }
        let result;
        try { result = await response.json(); }
        catch { throw new Error("The server returned an invalid response."); }
        if (!response.ok) throw new Error(result.error || "The request failed.");
        return result;
    }

    function recordForProject(projectID) {
        return records.find(record => Number(record.project_id) === Number(projectID));
    }

    function valueForRecord(record, key) {
        if (!record) return "";
        return record[key] ?? "";
    }

    function updateGuidance() {
        if (reviewStatus.value === "Approved") {
            statusHelp.textContent = "Official approval only. Add the submission date, approval number, decision date and official approval document link.";
        } else if (reviewStatus.value === "Exempt") {
            statusHelp.textContent = "Use Exempt only when the appropriate review body has confirmed the exemption. Add the decision date and official evidence link.";
        } else if (requiredSelect.value === "No") {
            statusHelp.textContent = "Confirm your institution's rules first. The ethics status must be Not required.";
        } else if (reviewStatus.value === "Changes requested") {
            statusHelp.textContent = "Record the requested changes in Committee Conditions and update the status after responding.";
        } else {
            statusHelp.textContent = "Check your university's rules to decide whether review or exemption is required.";
        }
    }

    function updateRequestButton(record, project) {
        const status = record?.supervisor_review_status || "Not requested";
        const hasSupervisor = Number(project?.teacher_id || 0) > 0;
        requestButton.disabled = !record || !hasSupervisor || status === "Pending";
        if (!record) requestButton.textContent = "Save record before requesting review";
        else if (!hasSupervisor) requestButton.textContent = "Supervisor not assigned";
        else if (status === "Pending") requestButton.textContent = "Supervisor review pending";
        else if (status === "Reviewed") requestButton.textContent = "Request another review";
        else requestButton.textContent = "Request Supervisor Review";
        requestButton.title = !hasSupervisor ? "Assign a supervisor to the project first." : "Send the saved ethics record to the assigned supervisor for review.";
    }

    function populateForm(record, projectID) {
        fillingForm = true;
        form.reset();
        projectSelect.value = projectID ? String(projectID) : "";

        if (record) {
            fields.forEach(key => {
                const control = form.elements.namedItem(key);
                if (control) control.value = valueForRecord(record, key);
            });
            projectHint.textContent = "Saved record loaded. Saving changes resets the previous supervisor review so the updated information can be reviewed again.";
        } else {
            requiredSelect.value = "Not sure";
            reviewStatus.value = "Not started";
            const participants = form.elements.namedItem("participants_involved");
            if (participants) participants.value = "Not sure";
            projectHint.textContent = "No ethics record has been saved for this project yet.";
        }

        currentProjectID = Number(projectID) || 0;
        form.dataset.dirty = "false";
        const project = projects.find(item => Number(item.id) === currentProjectID);
        updateRequestButton(record, project);
        fillingForm = false;
        updateGuidance();
    }

    function updateStats() {
        document.getElementById("ethics-total").textContent = String(records.length);
        document.getElementById("ethics-awaiting").textContent = String(records.filter(record => ["Preparing", "Submitted", "Changes requested"].includes(record.review_status)).length);
        document.getElementById("ethics-approved").textContent = String(records.filter(record => record.review_status === "Approved").length);
    }

    function addRecordLine(parent, label, value) {
        if (value === undefined || value === null || String(value).trim() === "") return;
        const line = document.createElement("div");
        line.className = "ethics-record-line";
        const strong = document.createElement("strong");
        strong.textContent = label;
        const valueNode = document.createElement("p");
        valueNode.textContent = String(value);
        line.append(strong, valueNode);
        parent.appendChild(line);
    }

    function safeLink(value) {
        if (!value) return null;
        try {
            const parsed = new URL(String(value));
            if (parsed.protocol !== "https:" && parsed.protocol !== "http:") return null;
            return parsed.href;
        } catch { return null; }
    }

    function renderRecords() {
        recordsContainer.replaceChildren();
        if (recordsEmpty) recordsEmpty.hidden = records.length > 0;

        records.forEach(record => {
            const card = document.createElement("article");
            card.className = "ethics-record-card";

            const top = document.createElement("div");
            top.className = "ethics-record-top";
            const heading = document.createElement("h4");
            heading.textContent = record.project_title || "Research project";
            const pill = document.createElement("span");
            const officialSlug = String(record.review_status || "Not started").toLowerCase().replace(/[^a-z]+/g, "-");
            pill.className = "ethics-status-pill ethics-status-" + officialSlug;
            pill.textContent = "Committee: " + (record.review_status || "Not started");
            top.append(heading, pill);

            const supervisor = document.createElement("p");
            supervisor.className = "ethics-supervisor-line";
            supervisor.textContent = "Supervisor review: " + (record.supervisor_review_status || "Not requested");

            const details = document.createElement("details");
            const summary = document.createElement("summary");
            summary.textContent = "View saved information";
            const body = document.createElement("div");
            body.className = "ethics-record-details";
            addRecordLine(body, "Committee", record.committee);
            addRecordLine(body, "Submission date", record.submission_date);
            addRecordLine(body, "Decision date", record.decision_date);
            addRecordLine(body, "Approval / exemption number", record.approval_number);
            addRecordLine(body, "Expiry date", record.expiry_date);
            addRecordLine(body, "Consent plan", record.consent_plan);
            addRecordLine(body, "Committee conditions", record.committee_conditions);
            if (record.supervisor_comment) addRecordLine(body, "Supervisor feedback", record.supervisor_comment);
            const link = safeLink(record.document_url);
            if (link) {
                const anchor = document.createElement("a");
                anchor.href = link;
                anchor.target = "_blank";
                anchor.rel = "noopener noreferrer";
                anchor.textContent = "Open official document ↗";
                body.appendChild(anchor);
            }
            details.append(summary, body);

            const open = document.createElement("button");
            open.type = "button";
            open.className = "button secondary";
            open.textContent = "Open / Edit Record";
            open.addEventListener("click", () => {
                if (form.dataset.dirty === "true" && !window.confirm("Discard your unsaved changes?")) return;
                projectSelect.value = String(record.project_id);
                populateForm(recordForProject(record.project_id), record.project_id);
                form.scrollIntoView({ behavior: "smooth", block: "start" });
            });

            card.append(top, supervisor, details, open);
            recordsContainer.appendChild(card);
        });
        updateStats();
    }

    async function loadData(preferredProjectID = 0) {
        const [projectList, ethicsList] = await Promise.all([
            api("/api/projects"),
            api("/api/research-ethics")
        ]);
        if (!Array.isArray(projectList) || !Array.isArray(ethicsList)) throw new Error("The server returned unexpected ethics data.");
        projects = projectList;
        records = ethicsList;

        projectSelect.replaceChildren();
        projects.forEach(project => {
            const option = document.createElement("option");
            option.value = String(project.id);
            option.textContent = project.title;
            projectSelect.appendChild(option);
        });
        noProject.hidden = projects.length > 0;
        saveButton.disabled = projects.length === 0;
        if (!projects.length) {
            const option = document.createElement("option");
            option.value = "";
            option.textContent = "No project available";
            projectSelect.appendChild(option);
            populateForm(null, 0);
        } else {
            let projectID = Number(preferredProjectID);
            if (!projects.some(project => Number(project.id) === projectID)) {
                const existing = records.find(record => projects.some(project => Number(project.id) === Number(record.project_id)));
                projectID = existing ? Number(existing.project_id) : Number(projects[0].id);
            }
            populateForm(recordForProject(projectID), projectID);
        }
        renderRecords();
    }

    projectSelect.addEventListener("change", function () {
        const nextID = Number(projectSelect.value);
        if (form.dataset.dirty === "true" && !window.confirm("Discard unsaved changes and switch projects?")) {
            projectSelect.value = String(currentProjectID || "");
            return;
        }
        populateForm(recordForProject(nextID), nextID);
    });

    form.addEventListener("input", event => {
        if (!fillingForm && event.target !== projectSelect) form.dataset.dirty = "true";
    });
    form.addEventListener("change", event => {
        if (!fillingForm && event.target !== projectSelect) form.dataset.dirty = "true";
    });

    requiredSelect.addEventListener("change", function () {
        if (requiredSelect.value === "No") reviewStatus.value = "Not required";
        else if (reviewStatus.value === "Not required") reviewStatus.value = "Not started";
        updateGuidance();
    });
    reviewStatus.addEventListener("change", updateGuidance);

    discardButton.addEventListener("click", function () {
        populateForm(recordForProject(currentProjectID), currentProjectID);
        statusMessage.textContent = "Unsaved changes discarded.";
    });

    form.addEventListener("submit", async function (event) {
        event.preventDefault();
        if (!form.reportValidity()) return;
        if (!currentProjectID) {
            statusMessage.textContent = "Select a project first.";
            return;
        }

        const payload = Object.fromEntries(new FormData(form).entries());
        payload.action = "save";
        payload.project_id = Number(payload.project_id);
        saveButton.disabled = true;
        requestButton.disabled = true;
        statusMessage.textContent = "Saving ethics record...";

        try {
            const result = await api("/api/research-ethics", "POST", payload);
            await loadData(payload.project_id);
            statusMessage.textContent = result.message || "Ethics record saved.";
        } catch (error) {
            statusMessage.textContent = error.message || "Could not save the ethics record.";
            console.error(error);
        } finally {
            saveButton.disabled = projects.length === 0;
            updateRequestButton(recordForProject(currentProjectID), projects.find(item => Number(item.id) === currentProjectID));
        }
    });

    requestButton.addEventListener("click", async function () {
        const record = recordForProject(currentProjectID);
        if (!record) {
            statusMessage.textContent = "Save the ethics record before requesting supervisor review.";
            return;
        }
        if (!window.confirm("Send this saved ethics record to the assigned supervisor for review? This is not a request for official committee approval.")) return;

        requestButton.disabled = true;
        statusMessage.textContent = "Sending supervisor review request...";
        try {
            const result = await api("/api/research-ethics", "POST", {
                action: "request_review",
                project_id: Number(currentProjectID)
            });
            await loadData(currentProjectID);
            statusMessage.textContent = result.message || "Supervisor review requested.";
        } catch (error) {
            statusMessage.textContent = error.message || "Could not request supervisor review.";
        } finally {
            updateRequestButton(recordForProject(currentProjectID), projects.find(item => Number(item.id) === currentProjectID));
        }
    });

    try {
        await loadData();
    } catch (error) {
        statusMessage.textContent = error.message || "Could not load ethics records.";
        console.error(error);
    }
});
