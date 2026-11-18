"use strict";

(() => {
    const definitions = {
        "citation-manager": {
            intro: "Manage citation details for papers already saved in Literature Review Matrix. Paper title, authors, year, DOI and URL are reused there, so you do not need to enter them twice.",
            fields: [
                { key: "paper_id", label: "Source paper from Literature Review Matrix", type: "paper", required: true },
                { key: "source_type", label: "Source Type", type: "select", required: true, options: ["Journal article", "Conference paper", "Book", "Book chapter", "Thesis", "Dataset", "Website", "Preprint", "Other"] },
                { key: "volume_issue_pages", label: "Volume / Issue / Pages", type: "text", placeholder: "e.g. Vol. 12, Issue 3, pp. 20–35" },
                { key: "accessed_date", label: "Accessed Date", type: "date", hint: "Use this for online sources where an access date is relevant." },
                { key: "citation_style", label: "Citation Style", type: "select", required: true, options: ["IEEE", "APA 7", "ACM", "Chicago", "Harvard", "Other"] },
                { key: "bibtex", label: "Verified BibTeX Entry / Citation", type: "textarea", rows: 4, hint: "Paste a verified entry. ResearchBridge does not invent or verify bibliographic details." },
                { key: "notes", label: "Notes — which research claim does this source support?", type: "textarea", rows: 3 }
            ]
        },
        "risk-decision-log": {
            intro: "Keep research risks and team decisions in one log. Record evidence and rationale honestly; do not present unverified assumptions as confirmed outcomes.",
            fields: [
                { key: "title", label: "Title", type: "text", required: true, placeholder: "Short risk or decision name" },
                { key: "entry_type", label: "Entry Type", type: "select", required: true, options: ["Risk", "Decision"] },
                { key: "description", label: "Description", type: "textarea", required: true, rows: 3, placeholder: "What could happen, or what decision was made?" },
                { key: "likelihood", label: "Likelihood", type: "select", options: ["Low", "Medium", "High", "Not applicable"] },
                { key: "impact", label: "Impact", type: "select", options: ["Low", "Medium", "High", "Critical", "Not applicable"] },
                { key: "mitigation", label: "Mitigation / Alternative Plan", type: "textarea", rows: 3 },
                { key: "decision_rationale", label: "Decision Rationale", type: "textarea", rows: 3 },
                { key: "owner", label: "Owner", type: "text", placeholder: "Responsible team member" },
                { key: "status", label: "Status", type: "select", required: true, options: ["Open", "Monitoring", "Mitigated", "Decided", "Closed"] },
                { key: "review_date", label: "Review Date", type: "date" }
            ]
        },
        "research-gap-tracker": {
            intro: "Turn literature observations into evidence-backed gap records. A gap should be supported by reviewed papers, not simply by something the team has not studied yet.",
            fields: [
                { key: "title", label: "Gap Title", type: "text", required: true, placeholder: "What is not yet sufficiently known?" },
                { key: "source_papers", label: "Source Papers", type: "textarea", rows: 3, hint: "List paper titles or their Literature Review record IDs." },
                { key: "existing_evidence", label: "Existing Evidence", type: "textarea", rows: 3 },
                { key: "missing_evidence", label: "Missing Evidence", type: "textarea", required: true, rows: 3 },
                { key: "importance", label: "Importance", type: "select", required: true, options: ["Low", "Medium", "High"] },
                { key: "feasibility", label: "Feasibility", type: "select", options: ["Low", "Medium", "High", "Needs assessment"] },
                { key: "proposed_approach", label: "Proposed Approach", type: "textarea", rows: 3 },
                { key: "status", label: "Status", type: "select", required: true, options: ["New", "Under discussion", "Selected", "In progress", "Addressed", "Not proceeding"] },
                { key: "owner", label: "Owner", type: "text" }
            ]
        },
        "research-question-map": {
            intro: "Connect the main question, sub-questions, hypothesis and method. Use the same wording consistently in your proposal and research notes.",
            fields: [
                { key: "title", label: "Question Label", type: "text", required: true, placeholder: "Short label for this question" },
                { key: "question_type", label: "Question Type", type: "select", required: true, options: ["Main question", "Sub-question", "Exploratory question"] },
                { key: "question", label: "Research Question", type: "textarea", required: true, rows: 3 },
                { key: "parent_question", label: "Parent Question", type: "question", hint: "Select the main question when this is a sub-question. Leave blank for a main question." },
                { key: "rationale", label: "Rationale", type: "textarea", rows: 3 },
                { key: "linked_gap", label: "Linked Research Gap", type: "gap", hint: "Choose a gap already saved in Research Gap Tracker." },
                { key: "hypothesis", label: "Hypothesis", type: "textarea", rows: 2 },
                { key: "variables", label: "Variables / Key Concepts", type: "textarea", rows: 2 },
                { key: "population_dataset", label: "Population / Dataset", type: "text" },
                { key: "method", label: "Method", type: "textarea", required: true, rows: 3 },
                { key: "status", label: "Status", type: "select", required: true, options: ["Draft", "Discussing", "Approved by team", "In progress", "Answered", "Revised"] }
            ]
        },
        "dataset-registry": {
            intro: "Record dataset provenance, version, licence, consent basis and handling restrictions. Do not publish sensitive or confidential dataset links to people who are not authorised.",
            fields: [
                { key: "title", label: "Dataset Name / Registry ID", type: "text", required: true },
                { key: "description", label: "Description", type: "textarea", required: true, rows: 3 },
                { key: "source_url", label: "Official Source URL", type: "url", required: true },
                { key: "creator_owner", label: "Creator / Owner", type: "text" },
                { key: "version", label: "Version / Release Date", type: "text" },
                { key: "format_size", label: "Format / Size", type: "text", placeholder: "CSV · 2 GB, JSON, image folder..." },
                { key: "license", label: "Licence / Restrictions", type: "text" },
                { key: "consent_basis", label: "Consent Basis", type: "textarea", rows: 2 },
                { key: "sensitivity", label: "Sensitivity", type: "select", required: true, options: ["Public", "Internal", "Confidential", "Sensitive"] },
                { key: "access_group", label: "Access Group", type: "text", placeholder: "e.g. project team only" },
                { key: "retention_date", label: "Retention / Review Date", type: "date" },
                { key: "quality_notes", label: "Quality Notes", type: "textarea", rows: 3 }
            ]
        },
        "task-dependencies": {
            intro: "This tool links tasks and milestones already created in the existing Milestones workspace. It does not create a second copy of those tasks.",
            fields: [
                { key: "task_id", label: "Task that depends on another", type: "task", required: true },
                { key: "depends_on_id", label: "Prerequisite Task", type: "task", required: true },
                { key: "dependency_type", label: "Dependency Type", type: "select", required: true, options: ["Finish-to-start", "Start-to-start", "Finish-to-finish", "Informational only"] },
                { key: "blocker", label: "Blocker", type: "textarea", rows: 2, placeholder: "What prevents the dependent task from progressing?" },
                { key: "notes", label: "Coordination Notes", type: "textarea", rows: 2 }
            ]
        },
        "team-workload": {
            intro: "Estimate capacity for a reporting period. Open task counts and completed work are calculated from existing assigned project tasks, not re-entered here.",
            fields: [
                { key: "member_id", label: "Team Member", type: "member", required: true },
                { key: "role", label: "Role / Work Area", type: "text", placeholder: "Writing, data analysis, experiment..." },
                { key: "reporting_period", label: "Reporting Period", type: "text", required: true, placeholder: "e.g. 17–23 November 2026" },
                { key: "estimated_hours", label: "Estimated Hours", type: "number", required: true, min: 0, step: 0.5 },
                { key: "capacity_hours", label: "Capacity Hours", type: "number", required: true, min: 0, step: 0.5 },
                { key: "blockers", label: "Blockers", type: "textarea", rows: 2 },
                { key: "wellbeing_note", label: "Wellbeing Note", type: "textarea", rows: 2, hint: "Keep this respectful and work-related. Do not enter private health information." }
            ]
        },
        "authorship-agreement": {
            intro: "Record each contributor's proposed order and actual contributions. Only the selected member can mark their own record as Agreed.",
            fields: [
                { key: "member_id", label: "Member Name", type: "member", required: true },
                { key: "proposed_order", label: "Proposed Author Order", type: "number", required: true, min: 1, step: 1 },
                { key: "contribution_roles", label: "Contribution Roles", type: "textarea", required: true, rows: 3, placeholder: "Conceptualization, data curation, analysis, writing, visualization..." },
                { key: "equal_contribution", label: "Equal Contribution", type: "checkbox" },
                { key: "acknowledgment", label: "Acknowledgment", type: "textarea", rows: 2 },
                { key: "agreement_status", label: "Agreement Status", type: "select", required: true, options: ["Not discussed", "Proposed", "Agreed", "Needs discussion", "Disputed"] },
                { key: "review_date", label: "Review Date", type: "date" },
                { key: "notes", label: "Notes / Follow-up", type: "textarea", rows: 2 }
            ]
        },
        "research-budget": {
            intro: "Track estimates separately from actual expenses. Leave Actual Cost blank until the amount is known, and use one currency consistently for the project.",
            fields: [
                { key: "title", label: "Item Name", type: "text", required: true },
                { key: "category", label: "Category", type: "select", required: true, options: ["Equipment", "Software / cloud", "Travel", "Printing / publication", "Data collection", "Other"] },
                { key: "description", label: "Description / Research Need", type: "textarea", rows: 2 },
                { key: "quantity", label: "Quantity", type: "number", required: true, min: 0, step: 0.01 },
                { key: "estimated_cost", label: "Estimated Total Cost", type: "number", required: true, min: 0, step: 0.01 },
                { key: "actual_cost", label: "Actual Cost", type: "number", min: 0, step: 0.01 },
                { key: "currency", label: "Currency", type: "select", required: true, options: ["BDT", "USD", "EUR", "GBP", "Other"] },
                { key: "funding_source", label: "Funding Source", type: "text", placeholder: "Grant, department, self-funded..." },
                { key: "requested_by", label: "Requested By", type: "text" },
                { key: "approval_status", label: "Approval Status", type: "select", required: true, options: ["Not requested", "Pending", "Approved", "Rejected", "Paid / purchased"] },
                { key: "expense_date", label: "Expense Date", type: "date" },
                { key: "receipt_link", label: "Receipt / Invoice Link", type: "url" },
                { key: "notes", label: "Notes", type: "textarea", rows: 2 }
            ]
        },
        "project-export": {
            intro: "Create a handover plan and generate a real ZIP metadata bundle. Only members with project access can generate it. Selected project metadata and uploaded source-code files may be included; planned Data Access rules are not automatically enforced.",
            fields: [
                { key: "title", label: "Export / Handover Plan Name", type: "text", required: true },
                { key: "export_format", label: "Planned Export Format", type: "select", required: true, options: ["ZIP metadata bundle"] },
                { key: "include_proposal", label: "Include Proposal", type: "checkbox" },
                { key: "include_literature", label: "Include Literature Review", type: "checkbox" },
                { key: "include_code", label: "Include uploaded source-code files (if available)", type: "checkbox" },
                { key: "include_dataset_metadata", label: "Include Dataset Metadata", type: "checkbox" },
                { key: "include_experiments", label: "Include Experiments", type: "checkbox" },
                { key: "include_decisions", label: "Include Risk / Decision Records", type: "checkbox" },
                { key: "recipient", label: "Recipient", type: "text", placeholder: "Person or institution" },
                { key: "handover_date", label: "Handover Date", type: "date" },
                { key: "handover_status", label: "Handover Status", type: "select", required: true, options: ["Planned", "Preparing", "Ready for review", "Transferred", "Archived"] },
                { key: "rights_notes", label: "Rights / Export Restrictions", type: "textarea", rows: 3 },
                { key: "notes", label: "Notes / Follow-up", type: "textarea", rows: 2 }
            ]
        },
        "data-access-settings": {
            intro: "This records the intended access and embargo policy. It is a planning record only: saving a policy here does not yet enforce permissions on uploaded files or datasets.",
            fields: [
                { key: "title", label: "Resource Name / ID", type: "text", required: true },
                { key: "classification", label: "Classification", type: "select", required: true, options: ["Public", "Team only", "Restricted", "Confidential"] },
                { key: "access_rule", label: "Access Rule", type: "textarea", required: true, rows: 2, placeholder: "Which role or named members may access this resource?" },
                { key: "purpose", label: "Purpose", type: "textarea", rows: 2 },
                { key: "embargo_start", label: "Embargo Start", type: "date" },
                { key: "embargo_end", label: "Embargo End", type: "date" },
                { key: "approval_reference", label: "Approval Reference", type: "text" },
                { key: "retention_rule", label: "Retention / Access Review Rule", type: "text" },
                { key: "policy_status", label: "Policy Status", type: "select", required: true, options: ["Draft", "Pending approval", "Active", "Expired", "Revoked"] },
                { key: "review_owner", label: "Review Owner", type: "text" },
                { key: "notes", label: "Notes / Exceptions", type: "textarea", rows: 2 }
            ]
        },
        "research-impact": {
            intro: "Record outputs and evidence-backed impact. Enter only metrics verified from the named source; do not estimate citation or download counts.",
            fields: [
                { key: "title", label: "Output Title", type: "text", required: true },
                { key: "output_type", label: "Output Type", type: "select", required: true, options: ["Journal paper", "Conference paper", "Dataset", "Software", "Thesis", "Poster", "Other"] },
                { key: "publication_date", label: "Publication Date", type: "date" },
                { key: "doi_url", label: "DOI / Official URL", type: "url" },
                { key: "citations", label: "Verified Citations", type: "number", min: 0, step: 1 },
                { key: "downloads", label: "Platform Downloads", type: "number", min: 0, step: 1 },
                { key: "adoption", label: "Adoption", type: "textarea", rows: 2, placeholder: "Known organisations or researchers using this output" },
                { key: "outcomes", label: "Measurable Outcomes", type: "textarea", rows: 2 },
                { key: "impact_evidence", label: "Impact Evidence Link", type: "url" },
                { key: "metric_source", label: "Metric Source", type: "text", placeholder: "Repository analytics, citation database..." },
                { key: "verified_date", label: "Metrics Verified Date", type: "date" },
                { key: "notes", label: "Notes / Period / Limitations", type: "textarea", rows: 2 }
            ]
        }
    };

    const labels = {
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

    const contexts = new Map();
    const instances = new Map();
    let projects = [];

    function el(tag, className, text) {
        const node = document.createElement(tag);
        if (className) node.className = className;
        if (text !== undefined) node.textContent = text;
        return node;
    }

    function safe(value) {
        return String(value ?? "");
    }

    async function api(path, method = "GET", payload = null) {
        const options = { method, credentials: "same-origin", cache: "no-store" };
        if (payload !== null) {
            options.headers = { "Content-Type": "application/json", "X-Requested-With": "ResearchBridge" };
            options.body = JSON.stringify(payload);
        }
        const response = await fetch(path, options);
        if (response.status === 401) {
            window.location.replace("/login/student");
            throw new Error("Please log in again.");
        }
        let result;
        try { result = await response.json(); }
        catch { throw new Error("The server returned an invalid response."); }
        if (!response.ok) throw new Error(result.error || "Request failed.");
        return result;
    }

    async function loadProjects() {
        projects = await api("/api/projects");
        if (!Array.isArray(projects)) projects = [];
    }

    async function loadContext(projectID, force = false) {
        if (!force && contexts.has(projectID)) return contexts.get(projectID);
        const ctx = { members: [], tasks: [], papers: [], gaps: [], questions: [] };
        await Promise.all([
            api(`/api/workspace/${projectID}/members`).then(value => { ctx.members = Array.isArray(value) ? value : []; }).catch(() => {}),
            api(`/api/workspace/${projectID}/records`).then(value => {
                ctx.tasks = (Array.isArray(value) ? value : []).filter(item => item.kind === "task" || item.kind === "milestone");
            }).catch(() => {}),
            api(`/api/research-tools/citation-sources?project_id=${projectID}`).then(value => { ctx.papers = Array.isArray(value) ? value : []; }).catch(() => {}),
            api(`/api/research-tools/research-gap-tracker?project_id=${projectID}`).then(value => { ctx.gaps = Array.isArray(value) ? value : []; }).catch(() => {}),
            api(`/api/research-tools/research-question-map?project_id=${projectID}`).then(value => { ctx.questions = Array.isArray(value) ? value : []; }).catch(() => {})
        ]);
        contexts.set(projectID, ctx);
        return ctx;
    }

    function makeSelectOptions(select, field, ctx) {
        const add = (value, label) => {
            const option = document.createElement("option");
            option.value = safe(value);
            option.textContent = label;
            select.appendChild(option);
        };
        const placeholder = {
            paper: "Choose a saved literature paper",
            member: "Choose a confirmed team member",
            task: "Choose a task or milestone",
            gap: "Choose a saved research gap",
            question: "Choose a parent question (optional)"
        }[field.type] || "Choose an option";
        add("", placeholder);
        let items = [];
        if (field.type === "paper") items = ctx.papers.map(p => ({ id: p.id, label: `${p.title} — ${p.authors || "Author not recorded"}` }));
        if (field.type === "member") items = ctx.members.map(m => ({ id: m.id, label: `${m.name}${m.owner ? " (Project owner)" : ""}` }));
        if (field.type === "task") items = ctx.tasks.map(t => ({ id: t.id, label: `${t.title} · ${t.kind} · ${t.status}` }));
        if (field.type === "gap") items = ctx.gaps.map(g => ({ id: g.id, label: g.title }));
        if (field.type === "question") items = ctx.questions.map(q => ({ id: q.id, label: q.title }));
        items.forEach(item => add(item.id, item.label));
    }

    function makeField(field, ctx, value = "") {
        const wrapper = el("div", "rt-field" + (field.type === "checkbox" ? " rt-checkbox-field" : ""));
        const id = `rt-${field.key}-${Math.random().toString(36).slice(2, 8)}`;
        const control = field.type === "textarea" ? document.createElement("textarea") : field.type === "select" || ["paper", "member", "task", "gap", "question"].includes(field.type) ? document.createElement("select") : document.createElement("input");
        control.id = id;
        control.name = field.key;
        if (field.type === "textarea") control.rows = field.rows || 3;
        if (field.type === "select") {
            const blank = document.createElement("option"); blank.value = ""; blank.textContent = "Choose an option"; control.appendChild(blank);
            field.options.forEach(option => { const node = document.createElement("option"); node.value = option; node.textContent = option; control.appendChild(node); });
        } else if (["paper", "member", "task", "gap", "question"].includes(field.type)) {
            makeSelectOptions(control, field, ctx);
        } else if (field.type === "checkbox") {
            control.type = "checkbox";
            control.checked = value === "true" || value === true;

} else {
    if (field.type !== "textarea") {
        control.type = field.type === "number" ? "number" : field.type;

        if (field.min !== undefined) control.min = String(field.min);
        if (field.step !== undefined) control.step = String(field.step);

        control.placeholder = field.placeholder || "";
    }

    control.value = safe(value);
}

        if (field.placeholder && field.type === "textarea") control.placeholder = field.placeholder;
        if (field.required) control.required = true;
        const label = el("label", "rt-label", field.label + (field.required ? " *" : ""));
        label.htmlFor = id;
        if (field.type === "checkbox") {
            wrapper.append(control, label);
        } else {
            wrapper.append(label, control);
        }
        if (field.hint) wrapper.appendChild(el("small", "rt-hint", field.hint));
        return { wrapper, control };
    }

    function displayValue(feature, field, value, ctx) {
        if (!value) return "";
        if (field.type === "checkbox") return value === "true" ? "Yes" : "No";
        if (field.type === "paper") {
            const paper = ctx.papers.find(p => String(p.id) === String(value));
            return paper ? `${paper.title} — ${paper.authors}` : `Literature record #${value}`;
        }
        if (field.type === "member") {
            const member = ctx.members.find(m => String(m.id) === String(value));
            return member ? member.name : `Member #${value}`;
        }
        if (field.type === "task") {
            const task = ctx.tasks.find(t => String(t.id) === String(value));
            return task ? task.title : `Task #${value}`;
        }
        if (field.type === "gap") {
            const gap = ctx.gaps.find(item => String(item.id) === String(value));
            return gap ? gap.title : `Gap #${value}`;
        }
        if (field.type === "question") {
            const question = ctx.questions.find(item => String(item.id) === String(value));
            return question ? question.title : `Question #${value}`;
        }
        return safe(value);
    }

    function createRecordCard(feature, record, ctx, refresh) {
        const card = el("article", "rt-record-card");
        const header = el("div", "rt-record-heading");
        header.append(el("h3", "", record.title || labels[feature]), el("span", "rt-record-badge", labels[feature]));
        card.appendChild(header);
        if (feature === "citation-manager") {
            const paper = ctx.papers.find(item => String(item.id) === String(record.fields?.paper_id));
            if (paper) {
                const sourceMeta = el("p", "rt-source-meta", [paper.authors, paper.year || "Year not recorded", paper.venue, paper.doi ? `DOI: ${paper.doi}` : ""].filter(Boolean).join(" · "));
                card.appendChild(sourceMeta);
                if (paper.paper_url) {
                    try {
                        const sourceURL = new URL(paper.paper_url);
                        if (sourceURL.protocol === "http:" || sourceURL.protocol === "https:") {
                            const link = el("a", "rt-source-link", "Open source paper");
                            link.href = sourceURL.href;
                            link.target = "_blank";
                            link.rel = "noopener noreferrer";
                            card.appendChild(link);
                        }
                    } catch { /* Invalid legacy links are shown as text only in Literature Review. */ }
                }
            }
        }
        const fields = definitions[feature].fields;
        const details = el("dl", "rt-record-details");
        fields.forEach(field => {
            const value = record.fields?.[field.key];
            if (value === undefined || value === "") return;
            const dt = el("dt", "", field.label);
            const dd = el("dd", "", displayValue(feature, field, value, ctx));
            details.append(dt, dd);
        });
        if (details.children.length) card.appendChild(details);
        if (feature === "team-workload") {
            const memberID = safe(record.fields.member_id);
            const assigned = ctx.tasks.filter(task => String(task.assignee_id) === memberID);
            const done = assigned.filter(task => ["completed", "closed", "approved"].includes(String(task.status).toLowerCase()));
            const open = assigned.filter(task => !["completed", "closed", "approved"].includes(String(task.status).toLowerCase()));
            const summary = el("p", "rt-derived-summary", `From existing assigned tasks: ${open.length} open · ${done.length} completed`);
            card.appendChild(summary);
            const estimated = Number(record.fields?.estimated_hours || 0);
            const capacity = Number(record.fields?.capacity_hours || 0);
            if (capacity > 0 && Number.isFinite(estimated)) {
                const percent = Math.max(0, Math.min(100, estimated / capacity * 100));
                const meter = el("div", "rt-workload-meter");
                meter.appendChild(el("div", "rt-workload-meter-label", `${estimated} / ${capacity} hours planned (${Math.round(estimated / capacity * 100)}% of capacity)`));
                const track = el("div", "rt-workload-track");
                const fill = el("span", "rt-workload-fill" + (estimated > capacity ? " is-over" : ""));
                fill.style.width = `${percent}%`;
                track.appendChild(fill);
                meter.appendChild(track);
                if (estimated > capacity) meter.appendChild(el("small", "rt-workload-warning", "Planned hours exceed the stated capacity."));
                card.appendChild(meter);
            }
            if (open.length) card.appendChild(el("p", "rt-derived-list", `Open: ${open.map(t => t.title).join(" · ")}`));
            if (done.length) card.appendChild(el("p", "rt-derived-list", `Completed: ${done.map(t => t.title).join(" · ")}`));
        }
        if (feature === "data-access-settings") {
            card.appendChild(el("p", "rt-caution", "Planning record only — file permissions are not enforced by this feature."));
        }
        const actions = el("div", "rt-actions");
        const edit = el("button", "rt-secondary", "Edit");
        edit.type = "button";
        edit.addEventListener("click", () => loadIntoForm(record));
        const remove = el("button", "rt-danger", "Delete");
        remove.type = "button";
        remove.addEventListener("click", async () => {
            if (!window.confirm("Delete this saved record? This cannot be undone.")) return;
            try {
                await api(`/api/research-tools/${feature}?id=${record.id}`, "DELETE", {});
                await refresh();
            } catch (error) { setStatus(feature, error.message, true); }
        });
        actions.append(edit, remove);
        if (feature === "project-export") {
            const download = el("button", "rt-primary", "Download selected metadata as ZIP");
            download.type = "button";
            download.addEventListener("click", () => downloadExport(record.id, feature));
            actions.appendChild(download);
        }
        card.appendChild(actions);
        return card;
    }

    function setStatus(feature, message, isError = false) {
        const status = document.getElementById(`${feature}-status`);
        if (status) {
            status.textContent = message;
            status.classList.toggle("is-error", isError);
            status.classList.toggle("is-success", !isError && Boolean(message));
        }
    }

    function loadIntoForm(record) {
        const instance = instances.get(record.feature);
        if (!instance) return;
        instance.currentID = record.id;
        instance.form.querySelectorAll("[name]").forEach(control => {
            const value = record.fields?.[control.name] ?? "";
            if (control.type === "checkbox") control.checked = value === "true";
            else control.value = value;
        });
        instance.saveButton.textContent = "Update Record";
        instance.cancelButton.hidden = false;
        instance.form.scrollIntoView({ behavior: "smooth", block: "start" });
        setStatus(record.feature, "Editing saved record.");
    }

    async function downloadExport(planID, feature) {
        try {
            const response = await fetch(`/api/research-tools/export?plan_id=${encodeURIComponent(planID)}`, { credentials: "same-origin", cache: "no-store" });
            if (response.status === 401) { window.location.replace("/login/student"); return; }
            if (!response.ok) {
                let result = {};
                try { result = await response.json(); } catch {}
                throw new Error(result.error || "Could not create the project export.");
            }
            const blob = await response.blob();
            const url = URL.createObjectURL(blob);
            const link = document.createElement("a");
            link.href = url;
            link.download = `researchbridge-project-export-${planID}.zip`;
            document.body.appendChild(link);
            link.click();
            link.remove();
            URL.revokeObjectURL(url);
            setStatus(feature, "ZIP metadata bundle created. Review its manifest for items not included.");
        } catch (error) { setStatus(feature, error.message, true); }
    }

    async function mountFeature(feature, root) {
        if (instances.has(feature)) return;
        const definition = definitions[feature];
        const header = el("div", "rt-intro");
        header.append(el("p", "rt-eyebrow", "PROJECT RESEARCH TOOL"), el("p", "rt-description", definition.intro));
        root.appendChild(header);
        const projectRow = el("div", "rt-project-row");
        const projectLabel = el("label", "", "Research Project");
        const projectSelect = document.createElement("select");
        projectSelect.className = "rt-project-select";
        projectSelect.setAttribute("aria-label", "Select research project");
        const projectPlaceholder = document.createElement("option");
        projectPlaceholder.value = ""; projectPlaceholder.textContent = "Choose a project";
        projectSelect.appendChild(projectPlaceholder);
        projects.forEach(project => {
            const option = document.createElement("option");
            option.value = String(project.id);
            option.textContent = project.title;
            projectSelect.appendChild(option);
        });
        projectLabel.appendChild(projectSelect);
        projectRow.append(projectLabel, el("span", "rt-project-hint", "Records are saved to the selected project and shared with people who can access that project."));
        root.appendChild(projectRow);

        const formCard = el("section", "rt-form-card");
        formCard.append(el("h3", "", "Add a new record"));
        const form = document.createElement("form");
        form.className = "rt-form";
        const grid = el("div", "rt-fields-grid");
        const controls = new Map();
        const ctx = { members: [], tasks: [], papers: [], gaps: [], questions: [] };
        definition.fields.forEach(field => {
            const control = makeField(field, ctx);
            controls.set(field.key, control.control);
            grid.appendChild(control.wrapper);
        });
        form.appendChild(grid);
        const actions = el("div", "rt-form-actions");
        const saveButton = el("button", "rt-primary", "Save Record");
        saveButton.type = "submit";
        const cancelButton = el("button", "rt-secondary", "Cancel Edit");
        cancelButton.type = "button";
        cancelButton.hidden = true;
        actions.append(saveButton, cancelButton);
        form.appendChild(actions);
        formCard.appendChild(form);
        root.appendChild(formCard);
        const listHeader = el("div", "rt-list-header");
        listHeader.append(el("h3", "", "Saved Records"));
        const search = document.createElement("input");
        search.type = "search"; search.placeholder = "Search this project's records..."; search.className = "rt-search";
        search.setAttribute("aria-label", "Search saved records");
        listHeader.appendChild(search);
        root.appendChild(listHeader);
        const empty = el("p", "rt-empty", "Choose a project to view or add records.");
        const list = el("div", "rt-record-grid");
        root.append(empty, list);

        const instance = { feature, root, form, projectSelect, controls, saveButton, cancelButton, search, empty, list, currentID: 0, records: [], context: ctx };
        instances.set(feature, instance);

        async function loadRecords() {
            const projectID = projectSelect.value;
            list.replaceChildren();
            if (!projectID) {
                empty.hidden = false; empty.textContent = "Choose a project to view or add records.";
                return;
            }
            try {
                const data = await api(`/api/research-tools/${feature}?project_id=${encodeURIComponent(projectID)}`);
                instance.records = Array.isArray(data) ? data : [];
                renderRecords();
            } catch (error) {
                empty.hidden = false; empty.textContent = error.message;
            }
        }

        function renderRecords() {
            const needle = search.value.trim().toLowerCase();
            const matches = instance.records.filter(record => JSON.stringify(record).toLowerCase().includes(needle));
            list.replaceChildren();
            empty.hidden = matches.length > 0;
            empty.textContent = instance.records.length ? "No records match your search." : "No records saved for this project yet. Add the first one above.";
            if (feature === "research-budget" && matches.length) {
                const estimatedTotals = new Map();
                const actualTotals = new Map();
                matches.forEach(record => {
                    const currency = record.fields?.currency || "Other";
                    const estimated = Number(record.fields?.estimated_cost || 0);
                    if (Number.isFinite(estimated)) estimatedTotals.set(currency, (estimatedTotals.get(currency) || 0) + estimated);
                    if (record.fields?.actual_cost !== undefined && record.fields.actual_cost !== "") {
                        const actual = Number(record.fields.actual_cost);
                        if (Number.isFinite(actual)) actualTotals.set(currency, (actualTotals.get(currency) || 0) + actual);
                    }
                });
                const formatTotals = totals => Array.from(totals.entries()).map(([currency, total]) => `${currency} ${total.toLocaleString(undefined, { maximumFractionDigits: 2 })}`).join(" · ") || "Not recorded";
                const summary = el("p", "rt-budget-summary", `Estimated totals: ${formatTotals(estimatedTotals)} · Actual totals: ${formatTotals(actualTotals)}`);
                list.appendChild(summary);
            }
            matches.forEach(record => list.appendChild(createRecordCard(feature, record, instance.context, loadRecords)));
        }

        function resetForm() {
            form.reset();
            instance.currentID = 0;
            saveButton.textContent = "Save Record";
            cancelButton.hidden = true;
        }

        async function loadProjectContext(force = false) {
            const projectID = projectSelect.value;
            if (!projectID) return;
            try {
                const fresh = await loadContext(projectID, force);
                instance.context = fresh;
                for (const field of definition.fields) {
                    if (!["paper", "member", "task", "gap", "question"].includes(field.type)) continue;
                    const old = controls.get(field.key);
                    const previousValue = old.value;
                    const replacement = document.createElement("select");
                    replacement.id = old.id;
                    replacement.name = old.name;
                    replacement.required = old.required;
                    replacement.className = old.className;
                    makeSelectOptions(replacement, field, fresh);
                    replacement.value = previousValue;
                    old.replaceWith(replacement);
                    controls.set(field.key, replacement);
                }
            } catch (error) { setStatus(feature, error.message, true); }
        }

        form.addEventListener("submit", async event => {
            event.preventDefault();
            if (!projectSelect.value) { setStatus(feature, "Choose a research project first.", true); return; }
            const fields = {};
            controls.forEach((control, key) => {
                fields[key] = control.type === "checkbox" ? String(control.checked) : control.value.trim();
            });
            if (feature === "task-dependencies" && fields.task_id === fields.depends_on_id) {
                setStatus(feature, "A task cannot depend on itself. Choose two different tasks.", true); return;
            }
            const payload = { id: instance.currentID, project_id: Number(projectSelect.value), fields };
            saveButton.disabled = true;
            setStatus(feature, instance.currentID ? "Updating record..." : "Saving record...");
            try {
                await api(`/api/research-tools/${feature}`, instance.currentID ? "PUT" : "POST", payload);
                resetForm();
                await loadRecords();
                setStatus(feature, "Record saved successfully.");
            } catch (error) { setStatus(feature, error.message, true); }
            finally { saveButton.disabled = false; }
        });
        cancelButton.addEventListener("click", () => { resetForm(); setStatus(feature, "Edit cancelled."); });
        projectSelect.addEventListener("change", async () => {
            resetForm();
            if (projectSelect.value) await loadProjectContext();
            await loadRecords();
            setStatus(feature, "");
        });
        search.addEventListener("input", renderRecords);
        instance.refresh = loadRecords;
        instance.refreshContext = loadProjectContext;

        if (!projects.length) {
            projectSelect.disabled = true;
            empty.textContent = "No accessible research project found. Create or join a project first.";
        }
    }

    async function loadActiveFeature() {
        const feature = location.hash.slice(1);
        const root = document.querySelector(`[data-feature-root="${feature}"]`);
        const instance = instances.get(feature);
        if (!root || !instance || !instance.projectSelect.value) return;
        await instance.refreshContext(true);
        await instance.refresh();
    }

    document.addEventListener("DOMContentLoaded", async () => {
        try {
            await loadProjects();
            for (const [feature, definition] of Object.entries(definitions)) {
                const root = document.querySelector(`[data-feature-root="${feature}"]`);
                if (root) await mountFeature(feature, root);
            }
            window.addEventListener("hashchange", loadActiveFeature);
            await loadActiveFeature();
        } catch (error) {
            console.error("Research tools could not be initialized:", error);
            document.querySelectorAll("[data-feature-root]").forEach(root => {
                root.textContent = "Research tools could not load. Refresh the page and try again.";
            });
        }
    });
})();
