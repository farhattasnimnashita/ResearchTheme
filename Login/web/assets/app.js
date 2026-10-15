"use strict";
const root = document.querySelector("#app");
let state = {
  role: location.pathname.includes("teacher") ? "teacher" : "student",
  mode: location.pathname === "/signup" ? "signup" : "login",
  user: null,
  projects: [],
  teachers: [],
  view: "overview",
};
const esc = (s) =>
  String(s ?? "").replace(
    /[&<>"']/g,
    (c) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        c
      ],
  );
const label = (s) => s.replaceAll("_", " ");
async function api(path, data) {
  const options =
    data === undefined
      ? {}
      : {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-Requested-With": "ResearchBridge",
          },
          body: JSON.stringify(data),
        };
  const response = await fetch(path, options);
  let result;
  try {
    result = await response.json();
  } catch {
    throw Error(
      "The server could not complete your request. Please try again.",
    );
  }
  if (!response.ok) {
    const error = new Error(result.error || "Request failed.");
    error.code = result.code;
    error.suggestSignup = !!result.suggest_signup;
    throw error;
  }
  return result;
}
function message(id, text, error = false) {
  const node = document.getElementById(id);
  node.textContent = text;
  node.className = "notice" + (error ? " error" : "");
}
function busy(button, on) {
  if (button) {
    button.disabled = on;
    button.setAttribute("aria-busy", String(on));
  }
}
const brand =
  '<a class="brand" href="/" aria-label="ResearchBridge home"><span class="mark">r</span>ResearchBridge<span class="muted">.</span></a>';
const academicOptions = (count, value) =>
  Array.from(
    { length: count },
    (_, i) =>
      `<option value="${i + 1}" ${Number(value) === i + 1 ? "selected" : ""}>${i + 1}</option>`,
  ).join("");
function rememberAuth() {
  const form = document.querySelector("#auth-form");
  if (form)
    state.authDraft = {
      ...state.authDraft,
      ...Object.fromEntries(new FormData(form)),
    };
}
function changeAuth(mode, role = state.role) {
  rememberAuth();
  state.mode = mode;
  state.role = role;
  // Passwords stay only in the current form and are cleared when changing account mode.
  if (state.authDraft) delete state.authDraft.password;
  authPage();
}
// Uses the same 1672×941 coordinate system and cover alignment as the background image.
// Decorative SVG layers animate around the globe without moving the login form.
function orbitArtwork() {
  return `<div class="orbit-layer" aria-hidden="true"><svg class="orbit-artwork" viewBox="0 0 1672 941" xmlns="http://www.w3.org/2000/svg" focusable="false">
 <g class="orbit-band" transform="rotate(-19 445 535)">
  <path id="research-orbit-a" class="orbit-track" d="M 135 535 a 310 118 0 1 0 620 0 a 310 118 0 1 0 -620 0"/>
  <circle class="orbit-spark" r="4"><animateMotion dur="24s" repeatCount="indefinite" calcMode="paced"><mpath href="#research-orbit-a"/></animateMotion></circle>
  <circle class="orbit-spark orbit-spark-small" r="2.5"><animateMotion dur="24s" begin="-12s" repeatCount="indefinite" calcMode="paced"><mpath href="#research-orbit-a"/></animateMotion></circle>
 </g>
 <g class="orbit-band" transform="rotate(29 445 535)">
  <path id="research-orbit-b" class="orbit-track orbit-track-fine" d="M 153 535 a 292 101 0 1 0 584 0 a 292 101 0 1 0 -584 0"/>
  <circle class="orbit-spark" r="3.3"><animateMotion dur="31s" begin="-8s" repeatCount="indefinite" calcMode="paced"><mpath href="#research-orbit-b"/></animateMotion></circle>
 </g>
 <g class="orbit-band" transform="rotate(72 445 535)">
  <path id="research-orbit-c" class="orbit-track orbit-track-dotted" d="M 204 535 a 241 137 0 1 0 482 0 a 241 137 0 1 0 -482 0"/>
  <circle class="orbit-spark" r="3"><animateMotion dur="38s" begin="-19s" repeatCount="indefinite" calcMode="paced"><mpath href="#research-orbit-c"/></animateMotion></circle>
 </g>
 </svg></div>`;
}
function authPage() {
  const signup = state.mode === "signup";
  const student = state.role === "student";
  const draft = state.authDraft || {};
  root.innerHTML = `<div class="auth-screen">${orbitArtwork()}<header class="auth-header">${brand}<span class="auth-header-note">A shared space for curious minds</span></header>
 <main class="auth-layout"><section class="auth-story" aria-label="About ResearchBridge"><p class="eyebrow">Connect · Collaborate · Discover</p><h1>Make room for<br>your next<br><em>discovery.</em></h1><span class="gold-rule"></span><p class="story-description">Connect with your team and<br>collaborate with supervisors.</p><div class="story-caption"><span class="caption-dot"></span>Ideas grow stronger together.</div></section>
 <section class="auth-panel" aria-labelledby="auth-title"><div class="auth-panel-heading"><span class="step-label">${signup ? "Create your account" : "Your research workspace"}</span>${signup ? '<button type="button" id="back-login" class="auth-back">← Log in</button>' : '<span class="secure-label">Verified community</span>'}</div>
 <h2 id="auth-title">${signup ? "A new chapter starts here." : "Welcome back."}</h2><p class="auth-sub">${signup ? "Tell us a little about yourself and your institution." : "Log in to continue your research journey."}</p>
 <div class="roles" aria-label="Choose account role">${["student", "teacher"].map((r) => `<button type="button" data-role="${r}" class="role ${r === state.role ? "active" : ""}" aria-pressed="${r === state.role}"><span class="role-icon" aria-hidden="true">${r === "student" ? "♧" : "♙"}</span><span><strong>${r === "student" ? "Student" : "Teacher"}</strong><small>${r === "student" ? "Explore & build" : "Guide & inspire"}</small></span><span class="role-check" aria-hidden="true">${r === state.role ? "✓" : ""}</span></button>`).join("")}</div>
 <div id="auth-message" class="notice" role="status" aria-live="polite"></div><div id="auth-next-action"></div>
 <form id="auth-form">${
   signup
     ? `<label class="field">Full name<input name="name" autocomplete="name" placeholder="Your full name" required minlength="2" maxlength="100" value="${esc(draft.name || "")}"></label>
 <label class="field">Institution <span class="required-mark">*</span><input name="institution" list="institution-options" autocomplete="organization" placeholder="Select or type your institution" required minlength="2" maxlength="180" value="${esc(draft.institution || "")}"></label><datalist id="institution-options"><option value="Shahjalal University of Science and Technology (SUST)"><option value="University of Dhaka"><option value="Bangladesh University of Engineering and Technology (BUET)"><option value="Jahangirnagar University"><option value="University of Rajshahi"><option value="University of Chittagong"><option value="Khulna University"><option value="North South University"><option value="BRAC University"></datalist>
 ${student ? `<fieldset class="academic-fields"><legend>Student details</legend><label class="field">Department <span class="required-mark">*</span><input name="department" list="department-options" placeholder="Select or type your department" required minlength="2" maxlength="120" value="${esc(draft.department || "")}"></label><datalist id="department-options"><option value="Computer Science and Engineering (CSE)"><option value="Software Engineering (SWE)"><option value="Electrical and Electronic Engineering (EEE)"><option value="Civil Engineering"><option value="Mathematics"><option value="Statistics"><option value="Physics"><option value="Business Administration"></datalist><div class="field-row"><label class="field">Academic year <span class="required-mark">*</span><select name="academic_year" required><option value="">Choose year</option>${academicOptions(6, draft.academic_year)}</select></label><label class="field">Semester <span class="required-mark">*</span><select name="semester" required><option value="">Choose semester</option>${academicOptions(12, draft.semester)}</select></label></div></fieldset>` : ""}`
     : ""
 }
 <label class="field">Email address<input name="email" type="email" autocomplete="email" placeholder="you@university.edu" required maxlength="254" value="${esc(draft.email || "")}"></label>
 <label class="field">Password<span class="password-wrap"><input id="password" name="password" type="password" autocomplete="${signup ? "new-password" : "current-password"}" placeholder="${signup ? "At least 8 characters" : "Your ResearchBridge password"}" required minlength="${signup ? "8" : "1"}" maxlength="72"><button type="button" id="show-password" aria-label="Show password">Show</button></span></label>
 <button class="btn wide auth-submit" type="submit">${signup ? "Create " + state.role + " account" : "Log in as " + state.role}<span aria-hidden="true">→</span></button></form>
 ${signup ? "" : '<p class="register-prompt">New here? <a id="register-link" href="/signup">Register an account</a></p>'}
 <p class="auth-note">${signup ? "Email verification is required before login." : "Use the password you created on ResearchBridge."}</p><div class="auth-verification"><span>Need to verify your email?</span><button type="button" class="btn link" id="resend">Resend verification</button></div>
 <div class="auth-panel-footer"><span>Thoughtful research starts with connection.</span><span aria-hidden="true">✧</span></div></section></main>
 <footer class="auth-footer"><span>© ${new Date().getFullYear()} ResearchBridge</span><span>Student ideas. Teacher guidance. Shared progress.</span></footer></div>`;
  document
    .querySelector("#register-link")
    ?.addEventListener("click", (event) => {
      event.preventDefault();
      changeAuth("signup");
      document.querySelector('[name="name"]')?.focus();
    });
  document
    .querySelector("#back-login")
    ?.addEventListener("click", () => changeAuth("login"));
  document
    .querySelectorAll("[data-role]")
    .forEach(
      (button) =>
        (button.onclick = () => changeAuth(state.mode, button.dataset.role)),
    );
  document.querySelector("#show-password").onclick = (event) => {
    const input = document.querySelector("#password");
    const show = input.type === "password";
    input.type = show ? "text" : "password";
    event.currentTarget.textContent = show ? "Hide" : "Show";
    event.currentTarget.setAttribute(
      "aria-label",
      show ? "Hide password" : "Show password",
    );
  };
  document.querySelector("#resend").onclick = resendDialog;
  document.querySelector("#auth-form").onsubmit = async (event) => {
    event.preventDefault();
    const button = event.submitter;
    busy(button, true);
    document.querySelector("#auth-next-action").replaceChildren();
    const data = Object.fromEntries(new FormData(event.target));
    data.role = state.role;
    if (signup && student) {
      data.academic_year = Number(data.academic_year);
      data.semester = Number(data.semester);
    }
    try {
      const result = await api("/api/" + (signup ? "register" : "login"), data);
      if (result.redirect) {
        location.assign(result.redirect);
        return;
      }
      message("auth-message", result.message, !!result.delivery_failed);
      state.authDraft = { email: data.email };
      event.target.querySelector("[name=password]").value = "";
      if (signup) {
        event.target.hidden = true;
        const next = document.createElement("button");
        next.type = "button";
        next.className = "btn wide auth-submit";
        next.textContent = "Continue to login →";
        next.onclick = () => changeAuth("login");
        document.querySelector("#auth-next-action").append(next);
      }
    } catch (err) {
      message("auth-message", err.message, true);
      if (err.code === "account_not_found") {
        state.authDraft = { email: data.email };
        const suggestion = document.createElement("div");
        suggestion.className = "signup-suggestion";
        const text = document.createElement("p");
        text.textContent =
          "New here? Create your " + state.role + " account to get started.";
        const next = document.createElement("button");
        next.type = "button";
        next.className = "btn wide";
        next.textContent = "Sign up as " + state.role + " →";
        next.onclick = () => changeAuth("signup");
        suggestion.append(text, next);
        document.querySelector("#auth-next-action").append(suggestion);
        next.focus();
      }
    } finally {
      busy(button, false);
    }
  };
}

function makeDialog(title, content) {
  document.querySelector("dialog")?.remove();
  const d = document.createElement("dialog");
  d.innerHTML = `<div class="dialog-top"><h2>${esc(title)}</h2><button type="button" class="close" aria-label="Close dialog">×</button></div>${content}`;
  document.body.append(d);
  d.querySelector(".close").onclick = () => d.close();
  d.addEventListener("close", () => d.remove());
  d.showModal();
  return d;
}
function resendDialog() {
  const d = makeDialog(
    "Verify your email.",
    `<p class="muted small">Enter the email you used to register. Links expire after 30 minutes.</p><div class="notice" id="resend-message" role="status"></div><form id="resend-form"><label class="field">Email address<input type="email" name="email" autocomplete="email" required></label><button class="btn wide">Send verification link →</button></form>`,
  );
  d.querySelector("form").onsubmit = async (e) => {
    e.preventDefault();
    busy(e.submitter, true);
    try {
      const result = await api(
        "/api/resend",
        Object.fromEntries(new FormData(e.target)),
      );
      message("resend-message", result.message);
    } catch (err) {
      message("resend-message", err.message, true);
    } finally {
      busy(e.submitter, false);
    }
  };
}
function verifyPage() {
  const t = new URLSearchParams(location.search).get("token");
  history.replaceState(null, "", "/verify");
  root.innerHTML = `<header class="topbar">${brand}</header><main class="verify-card"><p class="eyebrow">One last step</p><h1>A verified start<br>to your research.</h1><p class="muted">Confirm your email to activate your ResearchBridge account.</p><div class="notice" id="verify-message" role="status"></div><button class="btn wide" id="verify-button">Verify email →</button><p class="auth-note"><a href="/">Back to login</a></p></main>`;
  const b = document.querySelector("#verify-button");
  if (!t) {
    b.hidden = true;
    message(
      "verify-message",
      "Open the full verification link from your email or server terminal. If it expired, return to login and choose Resend verification.",
      true,
    );
    return;
  }
  b.onclick = async () => {
    busy(b, true);
    try {
      const result = await api("/api/verify", { token: t });
      message("verify-message", result.message);
      b.outerHTML = `<a class="btn wide" href="/login/${result.role}">Continue to ${esc(result.role)} login →</a>`;
    } catch (err) {
      message("verify-message", err.message, true);
      busy(b, false);
    }
  };
}
function initials(name) {
  return name
    .split(/\s+/)
    .map((s) => s[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}
async function loadDashboard() {
  try {
    state.user = await api("/api/me");
    if (location.pathname !== "/dashboard/" + state.user.role) {
      location.replace("/dashboard/" + state.user.role);
      return;
    }
    state.projects = await api("/api/projects");
    if (state.user.role === "student")
      state.teachers = await api("/api/teachers");
    dashboard();
  } catch (err) {
    root.innerHTML = `<main class="verify-card"><h2>Unable to open your workspace.</h2><p>${esc(err.message)}</p><a class="btn" href="/">Return to login</a><button class="btn light" id="retry">Try again</button></main>`;
    document.querySelector("#retry").onclick = loadDashboard;
  }
}
function projectCard(p) {
  const student = state.user.role === "student";
  return `<article class="project"><div class="project-top"><div><p class="eyebrow muted small">${esc(p.area)}</p><h3>${esc(p.title)}</h3></div><span class="badge ${esc(p.status)}">${esc(label(p.status))}</span></div><p class="project-meta">${student ? "Supervisor: " + esc(p.teacher) : "Student: " + esc(p.student)} · Updated ${new Date(p.updated * 1000).toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" })}</p><p class="abstract">${esc(p.abstract)}</p>${p.feedback ? `<div class="feedback"><strong>Supervisor feedback</strong><br>${esc(p.feedback)}</div>` : ""}<div class="project-actions">${student && ["draft", "changes_requested"].includes(p.status) ? `<button class="btn soft" data-submit="${p.id}">${p.status === "draft" ? "Submit for review" : "Revise & resubmit"} →</button>` : ""}${!student && p.status === "submitted" ? `<button class="btn" data-review="${p.id}">Review proposal →</button>` : ""}</div></article>`;
}
function dashboard() {
  document.querySelector("dialog")?.close();
  const u = state.user;
  const student = u.role === "student";
  const pending = state.projects.filter((p) => p.status === "submitted").length;
  const approved = state.projects.filter((p) => p.status === "approved").length;
  const v = state.view;
  const visible =
    v === "feedback"
      ? state.projects.filter((p) => p.feedback)
      : v === "review"
        ? state.projects.filter((p) => p.status === "submitted")
        : state.projects;
  const titles = {
    overview: "Your research, moving forward.",
    projects: student ? "Your research proposals." : "Supervised proposals.",
    review: "A fresh perspective starts here.",
    feedback: "Feedback that moves ideas forward.",
    account: "Your profile.",
  };
  root.innerHTML = `<div class="shell"><aside class="sidebar"><div class="sidebar-header"><div>${brand}</div><div class="portal-tag">${student ? "Student workspace" : "Teacher workspace"}</div></div><nav class="nav" aria-label="Dashboard navigation">${[
    ["overview", "◫ Overview"],
    ["projects", student ? "▤ My proposals" : "▤ Supervised proposals"],
    [
      student ? "feedback" : "review",
      student ? "☷ Feedback" : "☷ Review queue",
    ],
    ["account", "♙ My profile"],
  ]
    .map(
      ([id, text]) =>
        `<button type="button" data-view="${id}" class="${id === v ? "active" : ""}" ${id === v ? 'aria-current="page"' : ""}>${text}</button>`,
    )
    .join(
      "",
    )}</nav><div class="sidebar-bottom"><div class="user-chip"><span class="avatar">${esc(initials(u.name))}</span><span><strong>${esc(u.name)}</strong>${student ? "Student researcher" : "Research supervisor"}</span></div><p>Small questions.<br>Meaningful discoveries.</p><button class="btn light wide" data-logout>Log out ↗</button></div></aside><main class="main"><header class="dash-top"><span>Workspace / ${esc(label(v))}</span><div><span class="verified">✓ Email verified</span> <button class="btn link" data-logout>Log out</button></div></header><div class="page-title"><div><p class="eyebrow muted">${esc(u.name.split(" ")[0])}'s workspace</p><h1>${titles[v]}</h1><p>${student ? "A little structure for your next big idea." : "Guide promising ideas toward meaningful research."}</p></div>${student && v !== "account" ? '<button class="btn" id="new-project">＋ New proposal</button>' : ""}</div><div class="notice" id="dash-message" role="status"></div>${v === "account" ? `<section class="account"><div class="avatar">${esc(initials(u.name))}</div><h2>${esc(u.name)}</h2><dl><dt>Email</dt><dd>${esc(u.email)}</dd><dt>Account role</dt><dd>${student ? "Student researcher" : "Teacher / Supervisor"}</dd><dt>Institution</dt><dd>${esc(u.institution || "Not provided (registered before this update)")}</dd>${student ? `<dt>Department</dt><dd>${esc(u.department || "Not provided")}</dd><dt>Academic year</dt><dd>${u.academic_year ? esc(u.academic_year) : "Not provided"}</dd><dt>Semester</dt><dd>${u.semester ? esc(u.semester) : "Not provided"}</dd>` : ""}<dt>Verification</dt><dd>✓ Verified</dd></dl><p class="small muted">Your account role controls which workspace and actions you can access.</p></section>` : `${v === "overview" ? `<section class="dash-banner"><p class="eyebrow">A shared space for discovery</p><h2>${student ? "Big ideas begin<br>with a first draft." : "Good guidance.<br>Greater possibilities."}</h2><p>${student ? "Turn a question into a proposal. Connect with a supervisor, get feedback, and take your next step." : "Review student proposals, share your perspective, and help the next generation of researchers find their way."}</p></section><section class="stats"><div class="stat"><label>${student ? "YOUR PROPOSALS" : "ASSIGNED PROPOSALS"}</label><strong>${state.projects.length.toString().padStart(2, "0")}</strong></div><div class="stat"><label>${student ? "AWAITING REVIEW" : "TO REVIEW"}</label><strong>${String(pending).padStart(2, "0")}</strong></div><div class="stat"><label>APPROVED</label><strong>${String(approved).padStart(2, "0")}</strong></div></section>` : ""}<div class="section-title"><h3>${v === "feedback" ? "Supervisor feedback" : v === "review" ? "Ready for your review" : student ? "Your proposals" : "Student proposals"}</h3><span>${visible.length} ${visible.length === 1 ? "proposal" : "proposals"}</span></div>${visible.length ? visible.map(projectCard).join("") : `<section class="empty"><span class="feature-icon">${student ? "↗" : "♧"}</span><h3>${v === "feedback" ? "Your feedback will appear here." : v === "review" ? "Your review queue is clear." : student ? "Make space for your first idea." : "Ready for the next promising idea."}</h3><p>${student ? "Create a proposal and submit it to a verified supervisor." : "Student proposals assigned to you will appear here after submission."}</p>${student && v !== "feedback" ? '<button class="btn light" id="empty-create">Create your first proposal →</button>' : ""}</section>`}<section class="hint-panel"><h3>${student ? "Your next step" : "A note on supervision"}</h3><p>${student ? "Draft your idea → choose a verified teacher → submit for review → read feedback. Your drafts are visible only to you." : "You can review proposals assigned to you. Your decision and feedback are shared with the student immediately."}</p></section>`}</main></div>`;
  document.querySelectorAll("[data-view]").forEach(
    (b) =>
      (b.onclick = () => {
        state.view = b.dataset.view;
        dashboard();
      }),
  );
  document.querySelectorAll("[data-logout]").forEach(
    (b) =>
      (b.onclick = async () => {
        busy(b, true);
        try {
          await api("/api/logout", {});
          location.assign("/");
        } catch (err) {
          message("dash-message", err.message, true);
          busy(b, false);
        }
      }),
  );
  document.querySelector("#new-project")?.addEventListener("click", newProject);
  document
    .querySelector("#empty-create")
    ?.addEventListener("click", newProject);
  document
    .querySelectorAll("[data-submit]")
    .forEach(
      (b) =>
        (b.onclick = () =>
          submitProject(
            state.projects.find((p) => p.id === Number(b.dataset.submit)),
          )),
    );
  document
    .querySelectorAll("[data-review]")
    .forEach(
      (b) =>
        (b.onclick = () =>
          reviewProject(
            state.projects.find((p) => p.id === Number(b.dataset.review)),
          )),
    );
}
function teacherOptions() {
  return (
    '<option value="0">Choose a supervisor</option>' +
    state.teachers
      .map((t) => `<option value="${t.id}">${esc(t.name)}</option>`)
      .join("")
  );
}
function bindDialog(d, endpoint, transform) {
  d.querySelector("form").onsubmit = async (e) => {
    e.preventDefault();
    const b = e.submitter;
    busy(b, true);
    try {
      const result = await api(
        endpoint,
        transform(Object.fromEntries(new FormData(e.target))),
      );
      d.close();
      await loadDashboard();
      if (document.querySelector("#dash-message"))
        message("dash-message", result.message);
    } catch (err) {
      message("dialog-message", err.message, true);
    } finally {
      busy(b, false);
    }
  };
}
function newProject() {
  const d = makeDialog(
    "Make room for an idea.",
    `<p class="muted small">Start with a draft. Only you can see it until you submit.</p><div id="dialog-message" class="notice" role="status"></div><form><label class="field">Research title<input name="title" placeholder="e.g. Detecting misinformation in Bengali text" required minlength="5" maxlength="160"></label><label class="field">Research area<input name="area" placeholder="e.g. Natural Language Processing" required minlength="2" maxlength="80"></label><label class="field">Abstract<textarea name="abstract" placeholder="Describe your research question, approach, and expected outcome…" required minlength="20" maxlength="6000"></textarea></label><label class="field">Preferred supervisor (optional)<select name="teacher_id">${teacherOptions()}</select></label>${state.teachers.length ? "" : '<p class="small muted">No verified teachers yet. A teacher must register and verify their email before you can submit to them.</p>'}<div class="form-actions"><button class="btn" type="submit">Save draft →</button></div></form>`,
  );
  bindDialog(d, "/api/projects", (data) => ({
    ...data,
    teacher_id: Number(data.teacher_id),
  }));
}
function submitProject(p) {
  const d = makeDialog(
    p.status === "draft"
      ? "Ready for a fresh perspective?"
      : "Refine your proposal.",
    `<p class="small muted">${esc(p.title)}</p><div id="dialog-message" class="notice" role="status"></div><form><label class="field">Choose a verified supervisor<select name="teacher_id" required>${teacherOptions()}</select></label><label class="field">${p.status === "changes_requested" ? "Revised abstract" : "Abstract"}<textarea name="abstract" required minlength="20" maxlength="6000">${esc(p.abstract)}</textarea></label>${p.feedback ? `<p class="feedback">${esc(p.feedback)}</p>` : ""}<p class="small muted">Submitting shares your proposal with the selected supervisor. It can be revised again if they request changes.</p><button class="btn wide" ${state.teachers.length ? "" : "disabled"}>Submit for review →</button></form>`,
  );
  const match = state.teachers.find((t) => t.id === p.teacher_id);
  if (match) d.querySelector("select").value = String(match.id);
  bindDialog(d, `/api/projects/${p.id}/submit`, (data) => ({
    ...data,
    teacher_id: Number(data.teacher_id),
  }));
}
function reviewProject(p) {
  const d = makeDialog(
    "Help an idea move forward.",
    `<p class="small muted">${esc(p.title)} · ${esc(p.student)}</p><p class="abstract">${esc(p.abstract)}</p><div id="dialog-message" class="notice" role="status"></div><form><label class="field">Your decision<select name="status"><option value="approved">Approve proposal</option><option value="changes_requested">Request changes</option><option value="rejected">Reject proposal</option></select></label><label class="field">Feedback<textarea name="feedback" placeholder="Share clear, constructive guidance…" required minlength="5" maxlength="3000"></textarea></label><button class="btn wide">Save review & share feedback →</button></form>`,
  );
  bindDialog(d, `/api/projects/${p.id}/review`, (data) => data);
}
if (location.pathname === "/verify") verifyPage();
else if (location.pathname.startsWith("/dashboard/")) loadDashboard();
else authPage();
