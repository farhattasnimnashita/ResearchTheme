/* Enhance the existing teacher profile; retain its save and upload handlers. */
(() => {
  function setup() {
    if (!document.body.classList.contains('teacher-design')) return;
    const host = document.getElementById('research-content');
    if (!host) return;
    function enhance() {
      const form = host.querySelector('#profile-form');
      const grid = form?.closest('.grid-two');
      const summary = grid?.querySelector('.profile-summary');
      if (!form || !grid || !summary || grid.dataset.profileView) return;
      grid.dataset.profileView = 'ready';
      grid.classList.add('profile-reference');
      const panel = form.parentElement;
      const view = document.createElement('section');
      view.className = 'profile-reference-view';
      view.innerHTML = '<h2>My Profile</h2><div class="profile-reference-photo"><div class="profile-reference-avatar"></div><div><h3>Profile Picture</h3><p>Choose a JPG, PNG or WebP image. Maximum size: 300 KB.</p><button type="button" class="button" data-profile-picture>Change Picture</button></div></div><p class="profile-reference-intro">Your academic information, expertise and research interests.</p><dl class="profile-reference-details"></dl><button type="button" class="button" data-profile-edit>Edit Profile</button>';
      const avatar = summary.querySelector('.profile-avatar');
      if (avatar) view.querySelector('.profile-reference-avatar').append(avatar.cloneNode(true));
      const paragraphs = summary.querySelectorAll(':scope > p');
      const rows = [
        ['Name', summary.querySelector('h2')?.textContent],
        ['Email', paragraphs[0]?.textContent],
        ['Institution', paragraphs[1]?.textContent],
        ['Role', 'Teacher / Supervisor'],
        ['Educational Qualifications', form.elements.qualification?.value],
        ['Expertise / Skills', form.elements.expertise?.value],
        ['Research Interests', form.elements.interests?.value],
        ['About Me', form.elements.bio?.value]
      ];
      const list = view.querySelector('dl');
      for (const [label, value] of rows) {
        const row = document.createElement('div');
        const term = document.createElement('dt');
        const detail = document.createElement('dd');
        term.textContent = label;
        detail.textContent = value?.trim() || 'Not added yet';
        row.append(term, detail); list.append(row);
      }
      const reset = summary.querySelector('a[href="/forgot-password"]');
      if (reset) { const link = reset.cloneNode(true); link.classList.add('profile-reference-reset'); view.append(link); }
      const close = document.createElement('button');
      close.type = 'button'; close.className = 'button secondary profile-reference-back';
      close.textContent = 'Back to profile';
      panel.prepend(close);
      function show(edit) {
        grid.classList.toggle('profile-reference-editing', edit);
        view.hidden = edit; panel.hidden = !edit; summary.hidden = !edit;
        if (edit) form.querySelector('textarea')?.focus();
        else view.querySelector('[data-profile-edit]').focus();
      }
      view.querySelector('[data-profile-edit]').addEventListener('click', () => show(true));
      view.querySelector('[data-profile-picture]').addEventListener('click', () => {
        show(true); form.querySelector('#profile-photo')?.click();
      });
      close.addEventListener('click', () => show(false));
      grid.prepend(view); panel.hidden = true; summary.hidden = true;
    }
    new MutationObserver(enhance).observe(host, {childList:true});
    enhance();
  }
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', setup, {once:true});
  else setup();
})();
