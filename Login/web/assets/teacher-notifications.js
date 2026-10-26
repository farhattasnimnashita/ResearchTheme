/* Live in-page notifications. Sound requires an explicit user gesture. */
(() => {
  'use strict';
  let ready = false, highestID = 0, busy = false, stopped = false;
  let audio = null, sound = false, button = null, banner = null, feed = null;
  let timer = null, preferenceKey = '', wantsSound = false;
  let reminder = null, reminderBusy = false, requestVersion = 0, requestSignature = null;
  function rememberSound(value) { wantsSound=value;try{localStorage.setItem(preferenceKey,value?'on':'off')}catch(_){} }
  async function enableSound(test=false) {
    const Audio=window.AudioContext||window.webkitAudioContext;
    if(!Audio)throw new Error('Audio unsupported');
    audio ||= new Audio();await audio.resume();
    if(audio.state!=='running')throw new Error('Audio blocked');
    sound=true;soundLabel();if(test)chime();
  }

  function soundLabel() {
    if (!button) return;
    button.textContent = sound ? '🔊 Sound on · Mute' : wantsSound ? '🔔 Click to resume sound' : '🔔 Enable notification sound';
    button.setAttribute('aria-pressed', String(sound));
  }
  function chime() {
    if (!sound || !audio || audio.state !== 'running') return;
    const start = audio.currentTime;
    [660, 880].forEach((frequency, index) => {
      const oscillator = audio.createOscillator(), gain = audio.createGain();
      const at = start + index * 0.16;
      oscillator.type = 'sine'; oscillator.frequency.value = frequency;
      gain.gain.setValueAtTime(0, at);
      gain.gain.linearRampToValueAtTime(0.11, at + 0.015);
      gain.gain.exponentialRampToValueAtTime(0.001, at + 0.24);
      oscillator.connect(gain); gain.connect(audio.destination);
      oscillator.start(at); oscillator.stop(at + 0.26);
      oscillator.onended = () => { oscillator.disconnect(); gain.disconnect(); };
    });
  }
  function showNew(items) {
    const newest = items[0];
    const card = document.createElement('div'); card.className = 'rb-live-notice';
    const text = document.createElement('strong');
    text.textContent = items.length > 1 ? `${items.length} new research updates` : newest.kind === 'request' ? 'New supervision request' : 'New research update';
    const message = document.createElement('p'); message.textContent = newest.message;
    const link = document.createElement('a');
    link.href = teacherNotificationLink(newest); link.textContent = 'View update →';
    link.addEventListener('click', () => {
      card.remove();
      if (location.hash === link.hash && !editor.open) refresh().catch(error => notice(error.message, true));
    });
    const close = document.createElement('button'); close.type = 'button';
    close.textContent = '×'; close.setAttribute('aria-label', 'Dismiss notification');
    close.addEventListener('click', () => card.remove());
    card.append(text, message, link, close); feed.replaceChildren(card);
    const bell=document.querySelector('.top-notification');
    if(bell){bell.classList.remove('rb-bell-ring');void bell.offsetWidth;bell.classList.add('rb-bell-ring');setTimeout(()=>bell.classList.remove('rb-bell-ring'),1800)}
    chime();
  }
  function update(items) {
    if (!Array.isArray(items) || !S.user || !teacher()) return;
    const list = items.filter(n => Number.isSafeInteger(Number(n.id)) && Number(n.id) > 0)
      .sort((a, b) => Number(b.id) - Number(a.id));
    const fresh = ready ? list.filter(n => Number(n.id) > highestID && !n.read_at) : [];
    highestID = Math.max(highestID, ...list.map(n => Number(n.id)), 0);
    ready = true;
    S.notifications = list;
    const count = list.filter(n => !n.read_at).length;
    for (const id of ['teacher-notification-count', 'pending-badge']) {
      const badge = document.getElementById(id);
      if (badge) { badge.textContent = count ? String(count) : ''; badge.hidden = !count; }
    }
    const bell = document.querySelector('.top-notification');
    if (bell) bell.setAttribute('aria-label', `Teacher notifications: ${count} unread`);
    if (fresh.length && feed) showNew(fresh);
  }
  function updateRequests(items) {
    if (!Array.isArray(items) || !reminder || !teacher()) return;
    requestVersion++;
    const pending = items.filter(p => p.status === 'submitted');
    const signature = pending.map(p => p.id).sort().join(',');
    if (signature === requestSignature) return;
    requestSignature = signature;
    reminder.replaceChildren();
    reminder.hidden = pending.length === 0;
    if (!pending.length) return;
    const copy = document.createElement('div');
    const title = document.createElement('strong');
    title.textContent = `Reminder: ${pending.length} supervision request${pending.length === 1 ? '' : 's'} awaiting your response`;
    const detail = document.createElement('p');
    detail.textContent = 'Review the proposal and accept, ask for changes, or decline. Reading a notification does not resolve the request.';
    const link = document.createElement('a');
    link.className = 'rb-reminder-link'; link.href = '#requests'; link.textContent = 'Review requests →';
    link.addEventListener('click', () => {
      if (location.hash === '#requests' && !editor.open) refresh().catch(error => notice(error.message, true));
    });
    copy.append(title, detail); reminder.append(copy, link);
  }
  async function pollRequests() {
    if (reminderBusy || stopped || !S.user || !teacher()) return;
    reminderBusy = true;
    const version = requestVersion, controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 12000);
    try {
      const response = await fetch('/api/projects', {credentials:'same-origin', cache:'no-store', signal:controller.signal});
      if (!response.ok) throw new Error('Requests unavailable');
      const items = await response.json();
      if (!Array.isArray(items)) throw new Error('Invalid requests');
      // A newer review result must not be overwritten by an older poll.
      if (version === requestVersion) updateRequests(items);
    } catch (_) {
      // Retain the last confirmed reminder until the server can be reached.
    } finally { clearTimeout(timeout); reminderBusy = false; }
  }
  window.RBTeacherNotifications = { update, updateRequests };

  async function poll() {
    if (busy || stopped || !S.user || !teacher()) return;
    busy = true;
    pollRequests();
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 12000);
    try {
      const response = await fetch('/api/notifications', {
        credentials: 'same-origin', cache: 'no-store', signal: controller.signal
      });
      if (response.status === 401) {
        stopped = true; clearInterval(timer);
        banner.textContent = 'Your session expired. Sign in again to receive updates.';
        return;
      }
      if (!response.ok) throw new Error('Notifications unavailable');
      const items = await response.json();
      if (!Array.isArray(items)) throw new Error('Invalid notification response');
      update(items);
      banner.textContent = 'Live updates · checks every 10 seconds';
    } catch (_) {
      banner.textContent = 'Updates paused · reconnecting automatically';
    } finally {
      clearTimeout(timeout); busy = false;
    }
  }
  function setup() {
    if (!S.user) { setTimeout(setup, 700); return; }
    if (!teacher()) return;
    const bar = document.createElement('div'); bar.className = 'rb-live-bar';
    banner = document.createElement('span'); banner.textContent = 'Connecting to live updates…';
    button = document.createElement('button'); button.type = 'button'; soundLabel();
    preferenceKey='rb-teacher-notification-sound-'+S.user.id;
    try{wantsSound=localStorage.getItem(preferenceKey)==='on'}catch(_){}
    soundLabel();
    button.addEventListener('click', async () => {
      if(sound){sound=false;rememberSound(false);soundLabel();return}
      try{await enableSound(true);rememberSound(true)}catch(_){sound=false;soundLabel();banner.textContent='Sound unavailable. Visual notifications remain active.'}
    });
    // Restore an opted-in preference only after a real user gesture.
    const restore=()=>{if(wantsSound&&(!sound||audio?.state!=='running'))enableSound().catch(()=>{sound=false;soundLabel()})};
    document.addEventListener('pointerdown',event=>{if(!button.contains(event.target))restore()});
    document.addEventListener('keydown',event=>{if(!button.contains(event.target)&&['Enter',' '].includes(event.key))restore()});
    bar.append(banner, button);
    document.querySelector('.research-top').after(bar);
    reminder = document.createElement('section'); reminder.className = 'rb-request-reminder'; reminder.hidden = true;
    reminder.setAttribute('role', 'status'); reminder.setAttribute('aria-live', 'polite');
    bar.after(reminder);
    if (Array.isArray(S.proposals)) updateRequests(S.proposals);
    feed = document.createElement('div'); feed.className = 'rb-live-feed';
    feed.setAttribute('role', 'status'); feed.setAttribute('aria-live', 'polite');
    document.body.append(feed);
    poll(); timer = setInterval(poll, 10000);
    window.addEventListener('online', poll);
    document.addEventListener('visibilitychange', () => { if (!document.hidden) poll(); });
    window.addEventListener('pagehide', () => { stopped = true; clearInterval(timer); });
    window.addEventListener('pageshow', event => {
      if (event.persisted) { stopped = false; timer = setInterval(poll, 10000); poll(); }
    });
  }
  setup();
})();
