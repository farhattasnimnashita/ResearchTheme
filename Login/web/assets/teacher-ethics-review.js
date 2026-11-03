/* Adds a separate supervisor-review tool to the existing teacher workspace. */
"use strict";

(() => {
    if (typeof modules === "undefined" || typeof teacher !== "function") return;

    modules.ethics_review = [
        "Ethics Review",
        "Review ethics records submitted by your assigned research teams."
    ];
    toolIcons.ethics_review = "🛡️";
    if (!globalKinds.includes("ethics_review")) globalKinds.push("ethics_review");

    // Keep this supervisor-only item out of the student view at /research.
    const originalFeatureNavigation = featureNavigation;
    featureNavigation = function () {
        originalFeatureNavigation();
        if (!teacher()) {
            document.querySelectorAll('#tools-nav a[data-tool="ethics_review"]').forEach(link => link.remove());
        }
    };

    const originalRenderModule = renderModule;
    renderModule = function () {
        if (S.module !== "ethics_review") {
            originalRenderModule();
            return;
        }
        const target = document.getElementById("module-body");
        if (!target) return;
        renderEthicsReview(target).catch(error => notice(error.message || "Could not load ethics reviews.", true));
    };

    async function renderEthicsReview(target) {
        if (!teacher()) {
            target.innerHTML = '<section class="panel"><h2>Supervisor Ethics Review</h2><p>This tool is available to assigned supervisors only.</p></section>';
            return;
        }

        target.innerHTML = '<p class="loading">Loading ethics review requests…</p>';
        const records = await api("/api/research-ethics");
        const requested = records.filter(record =>
            ["Pending", "Changes requested", "Reviewed"].includes(record.supervisor_review_status)
        );

        let html = '<section class="panel"><p class="eyebrow">SUPERVISOR WORKFLOW</p><h2>Ethics review requests</h2><p class="role-note">Review the ethics plan and provide academic feedback. This supervisor review is not official ethics committee approval and cannot change the committee status.</p></section>';

        const waiting = requested.filter(record => record.supervisor_review_status === "Pending").length;
        html += `<div class="section-title"><h2>Review queue</h2><span>${waiting} pending</span></div>`;

        if (!requested.length) {
            target.innerHTML = html + blank(
                "No supervisor review requests yet.",
                "Students can save an ethics record and request review after a supervisor is assigned to their project."
            );
            return;
        }

        for (const record of requested) {
            const id = Number(record.project_id);
            const status = String(record.supervisor_review_status || "Pending");
            const documentLink = safeEthicsDocumentLink(record.document_url);
            html += `<article class="project-card">
                <div class="card-top"><div><p class="eyebrow">ETHICS RECORD · ${h(record.student_name || "Student")}</p><h3>${h(record.project_title || "Research project")}</h3></div><span class="badge ${status === "Reviewed" ? "approved" : status === "Changes requested" ? "rejected" : ""}">${h(status)}</span></div>
                <p><strong>Committee status:</strong> ${h(record.review_status || "Not started")} · <strong>Review required:</strong> ${h(record.ethics_required || "Not sure")}</p>
                <p><strong>Review body:</strong> ${h(record.committee || "Not specified")}</p>
                <p><strong>Application submitted:</strong> ${h(record.submission_date || "Not submitted")}</p>
                <p><strong>Participants:</strong> ${h(record.participants_involved || "Not sure")} · <strong>Consent method:</strong> ${h(record.consent_method || "Not specified")}</p>
                ${record.consent_plan ? `<div class="feedback-box"><strong>Consent plan</strong><br>${h(record.consent_plan)}</div>` : ""}
                ${record.committee_conditions ? `<div class="feedback-box"><strong>Committee conditions</strong><br>${h(record.committee_conditions)}</div>` : ""}
                ${record.supervisor_comment ? `<div class="feedback-box"><strong>Your previous feedback</strong><br>${h(record.supervisor_comment)}</div>` : ""}
                ${documentLink ? `<p><a class="record-link" href="${h(documentLink)}" target="_blank" rel="noopener noreferrer">Open official document ↗</a></p>` : ""}
                ${status !== "Reviewed" ? `<form class="teacher-ethics-review-form" data-project-id="${id}">
                    <label class="field">Supervisor review decision<select name="supervisor_review_status" required><option value="Reviewed">Reviewed — feedback complete</option><option value="Changes requested" ${status === "Changes requested" ? "selected" : ""}>Changes requested</option></select></label>
                    <label class="field">Feedback to student<textarea name="supervisor_comment" rows="3" maxlength="5000" placeholder="Share suggestions or explain the changes needed">${h(record.supervisor_comment || "")}</textarea></label>
                    <p class="mini">This action updates only the supervisor-review status, not the official ethics committee status.</p>
                    <p class="mini" data-ethics-form-message role="status"></p>
                    <button class="button">Save supervisor feedback</button>
                </form>` : `<p class="role-note">Supervisor review completed. If the student updates the ethics record and requests a new review, it will return to the queue.</p>`}
            </article>`;
        }

        target.innerHTML = html;
        target.querySelectorAll(".teacher-ethics-review-form").forEach(form => {
            form.addEventListener("submit", async event => {
                event.preventDefault();
                const button = event.submitter;
                const message = form.querySelector("[data-ethics-form-message]");
                const values = Object.fromEntries(new FormData(form));
                if (values.supervisor_review_status === "Changes requested" && !values.supervisor_comment.trim()) {
                    message.textContent = "Please explain the changes needed.";
                    return;
                }

                button.disabled = true;
                message.textContent = "Saving review…";
                try {
                    const result = await api("/api/research-ethics", {
                        action: "review",
                        project_id: Number(form.dataset.projectId),
                        supervisor_review_status: values.supervisor_review_status,
                        supervisor_comment: values.supervisor_comment.trim()
                    });
                    notice(result.message || "Supervisor feedback saved.");
                    await renderEthicsReview(target);
                } catch (error) {
                    message.textContent = error.message || "Could not save review.";
                    button.disabled = false;
                }
            });
        });
    }

    function safeEthicsDocumentLink(value) {
        if (!value) return "";
        try {
            const parsed = new URL(String(value));
            return parsed.protocol === "https:" || parsed.protocol === "http:" ? parsed.href : "";
        } catch {
            return "";
        }
    }
})();
