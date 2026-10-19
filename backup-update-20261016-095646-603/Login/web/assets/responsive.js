/* Responsive navigation shared by both dashboards and the proposal workspace. */
(()=>{
 'use strict';
 const compact=window.matchMedia('(max-width: 1024px)');
 function attach(){
  document.querySelectorAll('aside.sidebar,aside.research-sidebar').forEach(sidebar=>{
   if(sidebar.dataset.responsiveReady)return;
   sidebar.dataset.responsiveReady='true';
   const children=Array.from(sidebar.children);
   if(children.length<2)return;
   const menu=document.createElement('details');menu.className='device-menu';menu.open=!compact.matches;
   const summary=document.createElement('summary');summary.textContent='Menu';summary.setAttribute('aria-label','Dashboard navigation menu');
   const body=document.createElement('div');body.className='device-menu-body';
   menu.append(summary,body);children.slice(1).forEach(child=>body.append(child));sidebar.append(menu);
   menu.addEventListener('click',event=>{if(compact.matches&&event.target.closest('a,button[data-view]'))menu.open=false});
  });
 }
 function resize(){document.querySelectorAll('.device-menu').forEach(menu=>menu.open=!compact.matches)}
 compact.addEventListener('change',resize);
 window.addEventListener('hashchange',()=>{if(compact.matches)document.querySelectorAll('.device-menu').forEach(menu=>menu.open=false)});
 const observer=new MutationObserver(attach);
 function start(){attach();observer.observe(document.body,{childList:true,subtree:true})}
 start();window.addEventListener('pagehide',()=>observer.disconnect());window.addEventListener('pageshow',start);
})();
