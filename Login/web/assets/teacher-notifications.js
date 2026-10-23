/* Live in-page notifications. Sound requires an explicit user gesture. */
(() => {
  'use strict';
  let ready = false, highestID = 0, busy = false, stopped = false;
  let audio = null, sound = false, button = null, banner = null, feed = null;
  let timer = null;

  function soundLabel() {
    if (!button) return;
    button.textContent = sound ? 'Sound on · Mute' : 'Enable sound';
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
    text.textContent = items.length > 1 ? `${items.length} new research updates` : 'New research update';
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
  window.RBTeacherNotifications = { update };

  async function poll() {
    if (busy || stopped || !S.user || !teacher()) return;
    busy = true;
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
    button.addEventListener('click', async () => {
      if (sound) { sound = false; soundLabel(); return; }
      try {
        const Audio = window.AudioContext || window.webkitAudioContext;
        if (!Audio) throw new Error('Audio unsupported');
        audio ||= new Audio();
        await audio.resume();
        if (audio.state !== 'running') throw new Error('Audio blocked');
        sound = true; soundLabel(); chime();
      } catch (_) {
        sound = false; soundLabel();
        banner.textContent = 'Sound unavailable. Visual notifications remain active.';
      }
    });
    bar.append(banner, button);
    document.querySelector('.research-top').after(bar);
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
