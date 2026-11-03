document.addEventListener("DOMContentLoaded", async function () {
    const form = document.querySelector(
        'form[data-feature-form="literature-review"]'
    );

    const list = document.getElementById("literature-paper-list");
    const empty = document.getElementById("literature-paper-empty");
    const noResults = document.getElementById("literature-paper-no-results");
    const search = document.getElementById("literature-paper-search");
    const count = document.getElementById("literature-paper-count");
    const status = form?.querySelector(".research-feature-status");
    const submitButton = form?.querySelector('button[type="submit"]');

    if (!form || !list || !empty || !noResults || !search || !count || !status) {
        console.error("Literature Review elements not found.");
        return;
    }

    let savedPapers = [];

    const fields = [
        ["title", "Paper Title"],
        ["authors", "Authors"],
        ["year", "Publication Year"],
        ["venue", "Journal / Conference"],
        ["doi", "DOI"],
        ["paper_url", "Paper URL / PDF Link"],
        ["objective", "Research Objective"],
        ["methodology", "Methodology / Algorithm"],
        ["dataset", "Dataset / Sample"],
        ["key_findings", "Key Findings"],
        ["limitations", "Limitations"],
        ["research_gap", "Research Gap"],
        ["relevance", "Relevance to Our Project"]
    ];

    function paperValue(paper, key) {
        if (key === "year") {
            return paper.year ?? paper.publication_year ?? "";
        }

        return paper[key] ?? "";
    }

    function addInfo(parent, label, value) {
        if (value === undefined || value === null || String(value).trim() === "") {
            return;
        }

        const section = document.createElement("div");
        section.className = "literature-paper-info";

        const heading = document.createElement("strong");
        heading.textContent = label;

        const text = document.createElement("p");
        text.textContent = String(value);

        section.append(heading, text);
        parent.appendChild(section);
    }

    function createPaperCard(paper) {
        const card = document.createElement("article");
        card.className = "literature-paper-card";

        const title = document.createElement("h4");
        title.textContent = paperValue(paper, "title") || "Untitled paper";

        const authors = document.createElement("p");
        authors.className = "literature-paper-authors";
        authors.textContent = paperValue(paper, "authors") || "Authors not provided";

        const meta = document.createElement("div");
        meta.className = "literature-paper-meta";

        const year = paperValue(paper, "year");

        if (year && Number(year) !== 0) {
            const yearTag = document.createElement("span");
            yearTag.className = "literature-paper-tag";
            yearTag.textContent = String(year);
            meta.appendChild(yearTag);
        }

        const venue = paperValue(paper, "venue");

        if (venue) {
            const venueTag = document.createElement("span");
            venueTag.className = "literature-paper-tag";
            venueTag.textContent = String(venue);
            meta.appendChild(venueTag);
        }

        card.append(title, authors, meta);

        // Important comparison fields.
        addInfo(card, "Methodology", paperValue(paper, "methodology"));
        addInfo(card, "Key Findings", paperValue(paper, "key_findings"));
        addInfo(card, "Research Gap", paperValue(paper, "research_gap"));

        // Other information stays inside the expandable details.
        const details = document.createElement("details");
        const summary = document.createElement("summary");
        summary.textContent = "View full details";

        const dl = document.createElement("dl");
        dl.className = "literature-paper-details-list";

        fields.forEach(([key, label]) => {
            const value = paperValue(paper, key);

            if (
                value === undefined ||
                value === null ||
                String(value).trim() === "" ||
                (key === "year" && Number(value) === 0)
            ) {
                return;
            }

            const dt = document.createElement("dt");
            dt.textContent = label;

            const dd = document.createElement("dd");

            if (key === "paper_url") {
                try {
                    const url = new URL(String(value));

                    if (url.protocol === "https:" || url.protocol === "http:") {
                        const link = document.createElement("a");
                        link.href = url.href;
                        link.target = "_blank";
                        link.rel = "noopener noreferrer";
                        link.textContent = "Open original paper";
                        dd.appendChild(link);
                    } else {
                        dd.textContent = String(value);
                    }
                } catch {
                    dd.textContent = String(value);
                }
            } else {
                dd.textContent = String(value);
            }

            dl.append(dt, dd);
        });

        details.append(summary, dl);
        card.appendChild(details);

        return card;
    }

    function renderPapers() {
        const term = search.value.trim().toLowerCase();

        const filtered = savedPapers.filter((paper) =>
            fields.some(([key]) =>
                String(paperValue(paper, key) ?? "")
                    .toLowerCase()
                    .includes(term)
            )
        );

        list.replaceChildren();

        filtered.forEach((paper) => {
            list.appendChild(createPaperCard(paper));
        });

        empty.hidden = savedPapers.length > 0;
        noResults.hidden = savedPapers.length === 0 || filtered.length > 0;

        count.textContent = term
            ? `${filtered.length} of ${savedPapers.length} papers`
            : `${savedPapers.length} ${savedPapers.length === 1 ? "paper" : "papers"}`;
    }

    async function loadSavedPapers() {
        const response = await fetch("/api/literature-papers", {
            method: "GET",
            credentials: "same-origin"
        });

        const result = await response.json();

        if (!response.ok) {
            throw new Error(result.error || "Could not load saved papers.");
        }

        if (!Array.isArray(result)) {
            throw new Error("Unexpected response while loading research papers.");
        }

        savedPapers = result;
        renderPapers();
    }

    search.addEventListener("input", renderPapers);

    // Load papers already saved in the database.
    try {
        await loadSavedPapers();
    } catch (error) {
        status.textContent = error.message || "Could not load saved papers.";
        console.error(error);
    }

    // Save new paper.
    form.addEventListener("submit", async function (event) {
        event.preventDefault();

        if (!form.reportValidity()) {
            return;
        }

        const payload = Object.fromEntries(new FormData(form).entries());
        payload.year = payload.year ? Number(payload.year) : 0;

        if (submitButton) {
            submitButton.disabled = true;
        }

        status.textContent = "Saving paper...";

        try {
            const response = await fetch("/api/literature-papers", {
                method: "POST",
                credentials: "same-origin",
                headers: {
                    "Content-Type": "application/json",
                    "X-Requested-With": "ResearchBridge"
                },
                body: JSON.stringify(payload)
            });

            const result = await response.json();

            if (!response.ok) {
                throw new Error(result.error || "Could not save the paper.");
            }

            form.reset();
            await loadSavedPapers();

            status.textContent = "Paper saved successfully to your account.";
        } catch (error) {
            status.textContent = error.message || "Could not save the paper.";
            console.error(error);
        } finally {
            if (submitButton) {
                submitButton.disabled = false;
            }
        }
    });
});