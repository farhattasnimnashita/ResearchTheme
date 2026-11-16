(() => {
"use strict";
const form=document.getElementById("academic-profile-form");
if(!form)return;
const msg=document.getElementById("academic-profile-message");
const request=async(url,options={})=>{const r=await fetch(url,{credentials:"same-origin",...options,headers:{"Content-Type":"application/json",...(options.headers||{})}});const d=await r.json().catch(()=>({}));if(!r.ok)throw new Error(d.error||d.message||"Request failed");return d;};
async function load(){
 try{const d=await request("/api/student-academic-profile");(d.records||[]).forEach(x=>{const el=document.getElementById("cgpa-year-"+x.year);if(el)el.value=x.cgpa;});document.getElementById("academic-skills").value=(d.skills||[]).join(", ");document.getElementById("academic-interests").value=d.interests||"";}
 catch(e){msg.textContent=e.message;}
}
form.addEventListener("submit",async e=>{
 e.preventDefault();msg.textContent="Saving...";
 const records=[];for(let year=1;year<=6;year++){const v=document.getElementById("cgpa-year-"+year).value.trim();if(v!=="")records.push({year,cgpa:Number(v)});}
 const skills=document.getElementById("academic-skills").value.split(",").map(x=>x.trim()).filter(Boolean);
 try{const d=await request("/api/student-academic-profile",{method:"PUT",body:JSON.stringify({records,skills,interests:document.getElementById("academic-interests").value})});msg.textContent=d.message||"Saved.";}
 catch(err){msg.textContent=err.message;}
});
load();
})();