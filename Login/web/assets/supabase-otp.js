'use strict';
const form=document.querySelector('#otp-form'), status=document.querySelector('#status');
const forgot=location.pathname==='/forgot-password', reset=location.pathname==='/reset-password';
const title=forgot?'Reset your password':reset?'Choose a new password':'Verify your email';
document.querySelector('#title').textContent=title;document.title=title+' | ResearchTheme';
document.querySelector('#submit').textContent=forgot?'Send recovery code':reset?'Update password':'Verify email';
document.querySelector('#intro').textContent=forgot?'Enter your account email to receive a recovery code.':'Enter your email and the code from your latest ResearchTheme email.';
document.querySelector('#code-field').hidden=forgot;form.elements.token.disabled=forgot;
document.querySelector('#password-field').hidden=!reset;form.elements.password.disabled=!reset;form.elements.password.required=reset;
document.querySelector('#resend').hidden=forgot||reset;document.querySelector('#have-code').hidden=!forgot;
async function request(path,data){const r=await fetch(path,{method:'POST',headers:{'Content-Type':'application/json','X-Requested-With':'ResearchBridge'},body:JSON.stringify(data)});const body=await r.json();if(!r.ok)throw Error(body.error||'Please try again.');return body;}
form.addEventListener('submit',async e=>{e.preventDefault();const button=document.querySelector('#submit');button.disabled=true;status.textContent='Please wait…';try{const result=await request(forgot?'/api/forgot-password':reset?'/api/reset-password':'/api/verify',Object.fromEntries(new FormData(form)));status.textContent=result.message;if(result.redirect){location.assign(result.redirect);return;}form.reset();form.hidden=true;document.querySelector('#resend').hidden=true;}catch(err){status.textContent=err.message;}finally{button.disabled=false;}});
document.querySelector('#resend').addEventListener('click',async e=>{if(!form.elements.email.reportValidity())return;e.target.disabled=true;try{const r=await request('/api/resend',{email:form.elements.email.value});status.textContent=r.message;}catch(err){status.textContent=err.message;}finally{e.target.disabled=false;}});
