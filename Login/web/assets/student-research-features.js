
/* Load the 16 research feature sections and handle their forms. */

document.addEventListener("DOMContentLoaded", async function () {
    const container = document.getElementById("research-feature-sections");

    if (!container) {
        console.error("Research feature container not found.");
        return;
    }

    try {
        const response = await fetch("/student-research-feature-sections.html");

        if (!response.ok) {
            throw new Error("Could not load feature HTML: " + response.status);
        }

        const html = await response.text();
        container.innerHTML = html;

        setupResearchFeatureForms(container);

        console.log("Research feature sections loaded.");
    } catch (error) {
        console.error(error);
        container.textContent =
            "Research features could not be loaded. Check the browser console.";
    }
});

function setupResearchFeatureForms(container) {
    container.querySelectorAll("form[data-feature-form]").forEach((form) => {
        form.addEventListener("submit", (event) => {
            event.preventDefault();

            if (!form.reportValidity()) return;

            const card = form.closest("[data-feature-card]");
            const table = card?.querySelector("table[data-columns]");
            const tbody = table?.querySelector("tbody");
            const status = form.querySelector(".research-feature-status");
            const empty = card?.querySelector(".research-empty");

            if (!table || !tbody || !status) {
                if (status) {
                    status.textContent = "Could not find this feature's table.";
                }
                return;
            }

            const data = {};

            new FormData(form).forEach((value, key) => {
                if (typeof value === "string" && value.trim()) {
                    data[key] = value.trim();
                }
            });

            const columns = table.dataset.columns.split(",").filter(Boolean);
            const row = document.createElement("tr");

            columns.forEach((key) => {
                const cell = document.createElement("td");
                cell.textContent = data[key] || "—";
                row.appendChild(cell);
            });

            const detailCell = document.createElement("td");
            const details = document.createElement("details");
            const summary = document.createElement("summary");

            summary.textContent = "View full record";
            details.appendChild(summary);

            const list = document.createElement("dl");

            Array.from(form.elements).forEach((control) => {
                if (!control.name || !(control.name in data)) return;

                const dt = document.createElement("dt");
                const dd = document.createElement("dd");
                const label = form.querySelector(
                    `label[for="${CSS.escape(control.id)}"]`
                );

                dt.textContent = label
                    ? label.textContent.replace("*", "").trim()
                    : control.name;

                dd.textContent = data[control.name];
                list.append(dt, dd);
            });

            details.appendChild(list);

            const remove = document.createElement("button");
            remove.type = "button";
            remove.className = "button secondary research-record-delete";
            remove.textContent = "Remove from preview";

            remove.addEventListener("click", () => {
                row.remove();

                if (!tbody.children.length && empty) {
                    empty.hidden = false;
                }

                status.textContent = "Record removed from the preview only.";
            });

            detailCell.append(details, remove);
            row.appendChild(detailCell);
            tbody.appendChild(row);

            if (empty) empty.hidden = true;

            form.reset();

            status.textContent =
                "Added to this page preview only. It is not saved to the server or database.";
        });
    });
}
