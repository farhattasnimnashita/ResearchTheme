(function () {
 "use strict";
 let choice;
 try {choice = localStorage.getItem('researchbridge-theme');} catch {}
 const night = choice ? choice === 'dark' : window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
 document.documentElement.dataset.theme = night ? 'dark' : 'light';
 function update() {
  document.querySelectorAll('[data-theme-toggle]').forEach(button => {
   const dark = document.documentElement.dataset.theme === 'dark';
   button.textContent = dark ? '☀ Light mode' : '☾ Dark mode';
   button.setAttribute('aria-label', dark ? 'Switch to light mode' : 'Switch to dark mode');
   button.setAttribute('aria-pressed', String(dark));
  });
 }
 document.addEventListener('DOMContentLoaded', () => {
  if (!document.querySelector('[data-theme-toggle]')) {
   const button=document.createElement('button');button.type='button';button.dataset.themeToggle='';button.className='floating-theme';document.body.append(button);
  }
  update();
 });
 document.addEventListener('click', event => {
  if (!event.target.closest('[data-theme-toggle]')) return;
  const choice = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark';
  document.documentElement.dataset.theme=choice;
  try {localStorage.setItem('researchbridge-theme',choice);} catch {}
  update();
 });
})();
