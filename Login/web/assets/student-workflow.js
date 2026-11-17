(() => {
  'use strict';
  const $ = (id) => document.getElementById(id);
  const state = { projects: [], project: 0, records: [], files: [] };
  const escapeHTML = (value) => String(value ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  const pretty = (value) => String(value || 'pending').replaceAll('_',' ');
  const date = (seconds) => seconds ? new Date(seconds * 1000).toLocaleString() : 'Not scheduled';
  async function api(url, options = {}) {
    const response = await fetch(url, { credentials:'same-origin', ...options });
    let data = {}; try { data = await response.json(); } catch (_) {}
    if (!response.ok) throw new Error(data.error || 'Request failed. Please try again.');
    return data;
  }
  function selectedProject() { return Number(state.project || $('workflow-files-project')?.value || $('workflow-milestone-project')?.value || $('workflow-meeting-project')?.value || 0); }
  function optionsHTML() { return state.projects.map(p => `<option value="${p.id}" ${Number(p.id)===state.project?'selected':''}>${escapeHTML(p.title)}</option>`).join(''); }
  function syncPickers() {
    ['workflow-milestone-project','workflow-files-project','workflow-meeting-project'].forEach(id => { const el=$(id); if(el) el.innerHTML=optionsHTML() || '<option value="">No project available</option>'; });
  }
  function statusBadge(status) { return `<span class="workflow-status ${escapeHTML(status)}">${escapeHTML(pretty(status))}</span>`; }
  function empty(text) { return `<div class="workflow-empty">${escapeHTML(text)}</div>`; }
  function renderMilestones() {
    const items=state.records.filter(r=>r.kind==='milestone'||r.kind==='task').sort((a,b)=>(a.due_at||0)-(b.due_at||0));
    $('workflow-milestones').innerHTML=items.length?items.map(r=>`<article class="workflow-item"><div class="workflow-meta">${statusBadge(r.status)} <span>Due: ${escapeHTML(date(r.due_at))}</span></div><h4>${escapeHTML(r.title)}</h4><p>${escapeHTML(r.body)}</p><div class="workflow-progress"><span style="width:${Math.max(0,Math.min(100,Number(r.progress)||0))}%"></span></div><div class="workflow-meta"><strong>${Math.max(0,Math.min(100,Number(r.progress)||0))}% complete</strong>${r.author?`<span>· Assigned by ${escapeHTML(r.author)}</span>`:''}</div>${r.feedback?`<div class="workflow-feedback"><strong>Supervisor note:</strong> ${escapeHTML(r.feedback)}</div>`:''}</article>`).join(''):empty('Your supervisor has not added any milestones yet. They will appear here when assigned.');
  }
  function renderFiles() {
    const select=$('workflow-submission-file');
    select.innerHTML='<option value="">No file selected</option>'+state.files.map(f=>`<option value="${f.id}">${escapeHTML(f.name)}</option>`).join('');
    $('workflow-files-list').innerHTML=state.files.length?state.files.map(f=>`<article class="workflow-item"><h4>${escapeHTML(f.name)}</h4><div class="workflow-meta"><span>Uploaded by ${escapeHTML(f.author||'Project member')}</span><span>${escapeHTML(date(f.created_at))}</span><span>${(Number(f.size||0)/1024).toFixed(1)} KB</span></div><div class="workflow-actions"><a href="${escapeHTML(f.link)}" target="_blank" rel="noopener">Download / view ↗</a><button class="button secondary" type="button" data-attach-file="${f.id}">Use in submission</button></div></article>`).join(''):empty('No files have been uploaded to this project yet.');
    const submissions=state.records.filter(r=>['submission','final_submission','correction'].includes(r.kind));
    $('workflow-feedback-list').innerHTML=submissions.length?submissions.map(r=>`<article class="workflow-item"><div class="workflow-meta">${statusBadge(r.status)} <span>${escapeHTML(date(r.updated_at/1000))}</span></div><h4>${escapeHTML(r.title)}</h4><p>${escapeHTML(r.body)}</p>${r.link?`<div class="workflow-actions"><a href="${escapeHTML(r.link)}" target="_blank" rel="noopener">Open attached work ↗</a></div>`:''}${r.feedback?`<div class="workflow-feedback"><strong>Supervisor feedback</strong><p>${escapeHTML(r.feedback)}</p></div>`:''}${r.status==='changes_requested'?`<button class="button" type="button" data-revise="${r.id}">Submit a revised version →</button>`:''}</article>`).join(''):empty('No submissions yet. Upload your work and submit it for supervisor review.');
  }
  function renderMeetings() {
    const items=state.records.filter(r=>r.kind==='meeting').sort((a,b)=>(a.due_at||0)-(b.due_at||0));
    $('workflow-meetings-list').innerHTML=items.length?items.map(r=>`<article class="workflow-item"><div class="workflow-meta">${statusBadge(r.status)} <span>${escapeHTML(date(r.due_at))}</span></div><h4>${escapeHTML(r.title)}</h4><p>${escapeHTML(r.body||'No agenda provided.')}</p>${r.feedback?`<div class="workflow-feedback"><strong>Supervisor response:</strong> ${escapeHTML(r.feedback)}</div>`:''}${r.link?`<div class="workflow-actions"><a href="${escapeHTML(r.link)}" target="_blank" rel="noopener">Join meeting ↗</a></div>`:''}</article>`).join(''):empty('No meeting requests yet. Use the form above to request a discussion.');
  }
  async function loadProject(projectID) {
    state.project=Number(projectID)||0; if(!state.project) return;
    try {
      const [records,files]=await Promise.all([api(`/api/workspace/${state.project}/records`),api(`/api/workspace/${state.project}/files`)]);
      state.records=Array.isArray(records)?records:[]; state.files=Array.isArray(files)?files:[];
      renderMilestones();renderFiles();renderMeetings();
    } catch(error) { ['workflow-milestones','workflow-files-list','workflow-feedback-list','workflow-meetings-list'].forEach(id=>{if($(id))$(id).innerHTML=empty(error.message)}); }
  }
  function syncProjectFrom(select) { state.project=Number(select.value)||0; ['workflow-milestone-project','workflow-files-project','workflow-meeting-project'].forEach(id=>{if($(id))$(id).value=String(state.project)}); loadProject(state.project); }
  async function init() {
    try {
      state.projects=await api('/api/projects');
      if(!Array.isArray(state.projects))state.projects=[];
      state.project=Number(state.projects[0]?.id)||0;syncPickers();
      if(state.project) await loadProject(state.project);
      else ['workflow-milestones','workflow-files-list','workflow-feedback-list','workflow-meetings-list'].forEach(id=>$(id).innerHTML=empty('Create a project and get a supervisor assigned to start this workflow.'));
    } catch(error) { ['workflow-milestones','workflow-files-list','workflow-feedback-list','workflow-meetings-list'].forEach(id=>$(id).innerHTML=empty(error.message)); }
  }
  ['workflow-milestone-project','workflow-files-project','workflow-meeting-project'].forEach(id=>$(id)?.addEventListener('change',e=>syncProjectFrom(e.target)));
  $('workflow-upload-form')?.addEventListener('submit',async e=>{e.preventDefault();const form=e.currentTarget,button=form.querySelector('button[type=submit]'),msg=$('workflow-upload-status');if(!state.project){msg.textContent='Choose a project first.';return}button.disabled=true;msg.textContent='Uploading…';try{const file=form.elements.file.files[0];if(!file||file.size>8*1024*1024)throw new Error('Choose a file up to 8 MB.');const body=new FormData(form);await api(`/api/workspace/${state.project}/files`,{method:'POST',headers:{'X-Requested-With':'ResearchBridge'},body});form.reset();msg.textContent='File uploaded successfully.';await loadProject(state.project)}catch(error){msg.textContent=error.message}finally{button.disabled=false}});
  $('workflow-submission-form')?.addEventListener('submit',async e=>{e.preventDefault();const form=e.currentTarget,button=form.querySelector('button[type=submit]'),msg=$('workflow-submission-status');if(!state.project){msg.textContent='Choose a project first.';return}button.disabled=true;try{const data=Object.fromEntries(new FormData(form));const file=state.files.find(f=>f.id===Number(data.file_id));await api(`/api/workspace/${state.project}/records`,{method:'POST',headers:{'Content-Type':'application/json','X-Requested-With':'ResearchBridge'},body:JSON.stringify({kind:'submission',title:data.title,body:data.body,link:file?.link||'',status:'pending'})});form.reset();msg.textContent='Submitted. Your supervisor can now review it.';await loadProject(state.project)}catch(error){msg.textContent=error.message}finally{button.disabled=false}});
  $('workflow-meeting-form')?.addEventListener('submit',async e=>{e.preventDefault();const form=e.currentTarget,button=form.querySelector('button[type=submit]'),msg=$('workflow-meeting-status');if(!state.project){msg.textContent='Choose a project first.';return}button.disabled=true;try{const data=Object.fromEntries(new FormData(form));const due=Math.floor(new Date(data.due_at).getTime()/1000);if(!Number.isFinite(due)||due<=Math.floor(Date.now()/1000))throw new Error('Choose a future meeting date and time.');await api(`/api/workspace/${state.project}/records`,{method:'POST',headers:{'Content-Type':'application/json','X-Requested-With':'ResearchBridge'},body:JSON.stringify({kind:'meeting',title:data.title,body:data.body,link:data.link,status:'pending',due_at:due})});form.reset();msg.textContent='Meeting request sent to your supervisor.';await loadProject(state.project)}catch(error){msg.textContent=error.message}finally{button.disabled=false}});
  document.addEventListener('click',async e=>{const attach=e.target.closest('[data-attach-file]');if(attach){$('workflow-submission-file').value=attach.dataset.attachFile;location.hash='#files';$('workflow-submission-form')?.scrollIntoView({behavior:'smooth',block:'center'});return}const revise=e.target.closest('[data-revise]');if(revise){const old=state.records.find(r=>r.id===Number(revise.dataset.revise));if(!old)return;const title=prompt('Revised submission title:',old.title);if(title===null)return;const body=prompt('Describe the corrections you made:',old.body);if(body===null)return;try{await api(`/api/workspace/${state.project}/records`,{method:'POST',headers:{'Content-Type':'application/json','X-Requested-With':'ResearchBridge'},body:JSON.stringify({id:0,kind:old.kind,title,body,link:old.link,status:'pending',parent_id:old.id})});await loadProject(state.project)}catch(error){alert(error.message)}}});
  window.addEventListener('hashchange',()=>{if(['#milestones','#files','#meetings'].includes(location.hash))loadProject(selectedProject())});
  init();
})();
