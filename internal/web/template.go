package web

const AppLogoSVG = `<svg width="512" height="512" viewBox="0 0 512 512" fill="none" xmlns="http://www.w3.org/2000/svg"><defs><linearGradient id="iconGradient" x1="0" y1="0" x2="512" y2="512" gradientUnits="userSpaceOnUse"><stop stop-color="#3A8DFF"/><stop offset="1" stop-color="#003366"/></linearGradient><filter id="shadowFilter" x="-5%" y="-5%" width="110%" height="110%"><feDropShadow dx="4" dy="6" stdDeviation="5" flood-opacity="0.2"/></filter></defs><rect x="0" y="0" width="512" height="512" rx="114" fill="url(#iconGradient)"/><g filter="url(#shadowFilter)"><path d="M256 96 L416 384 L336 384 L288 288 L224 288 L176 384 L96 384 Z M256 168 L200 288 L312 288 Z" fill="#FFFFFF" fill-opacity="0.95"/></g></svg>`

const htmlPage = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover">
    <title>115Nexus</title>
    <link rel="manifest" href="/manifest.json">
    <link rel="apple-touch-icon" href="https://img.andp.cc/icons/upload/115Nexus.png">
    <style>
        :root { --primary: #007bff; --bg: #f4f6f9; --card: #ffffff; --text: #333333; --text-sub: #666666; --border: #eeeeee; --input-bg: #ffffff; --input-border: #ddd; }
        html.dark { --primary: #4dabf7; --bg: #141517; --card: #1f2023; --text: #e0e0e0; --text-sub: #a0a0a0; --border: #2c2e33; --input-bg: #25262b; --input-border: #373a40; }
        html { background: var(--bg); transition: background 0.3s; height: 100%; }
        body { font-family: -apple-system, sans-serif; background: var(--bg); color: var(--text); margin: 0; display: flex; flex-direction: column; height: 100vh; overflow: hidden; }
        .navbar { position: fixed; top: 0; left: 0; right: 0; background: var(--card); padding: 12px 20px; padding-top: calc(12px + env(safe-area-inset-top)); display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid var(--border); box-shadow: 0 1px 3px rgba(0,0,0,0.05); z-index: 1000; min-height: 50px; }
        .brand { font-weight: bold; font-size: 18px; cursor: pointer; color: var(--text); display: flex; align-items: center; gap: 8px; }
        .brand svg { width: 28px; height: 28px; border-radius: 6px; }
        .nav-links { display: flex; gap: 5px; }
        .nav-item { cursor: pointer; padding: 6px 12px; border-radius: 15px; font-size: 14px; color: var(--text-sub); }
        .nav-item.active { background: rgba(0, 123, 255, 0.1); color: var(--primary); font-weight: bold; }
        .container { flex: 1; overflow-y: auto; padding: 15px; padding-top: calc(90px + env(safe-area-inset-top)); max-width: 1200px; margin: 0 auto; width: 100%; box-sizing: border-box; }
        @media (max-width: 800px) { .container { max-width: 100%; } }
        .tab-content { display: none; } .tab-content.active { display: block; animation: fadeIn 0.2s; }
        @keyframes fadeIn { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
        .search-box { display: flex; gap: 10px; margin-bottom: 15px; }
        .search-box select, .search-box input { background: var(--card); color: var(--text); border: 1px solid var(--input-border); border-radius: 10px; padding: 10px; font-size: 16px; outline: none; }
        .search-box input { flex: 1; min-width: 0; }
        .btn { background: var(--primary); color: #fff; border: none; padding: 10px 20px; border-radius: 10px; font-weight: bold; cursor: pointer; display: flex; align-items: center; justify-content: center; }
        .btn-green { background: #40c057; } .btn-blue { background: #228be6; } .btn-sm { padding: 4px 10px; font-size: 12px; height: 30px; flex-shrink: 0; white-space: nowrap; }
        .list-item { background: var(--card); padding: 15px; border-radius: 10px; margin-bottom: 10px; border: 1px solid var(--border); cursor: pointer; box-shadow: 0 1px 2px rgba(0,0,0,0.03); overflow: hidden; }
        .list-title { font-size: 16px; font-weight: 600; margin-bottom: 6px; word-break: break-all; }
        .list-meta { font-size: 12px; color: var(--text-sub); display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }
        .badge { padding: 2px 6px; border-radius: 4px; font-size: 11px; background: #f3e5f5; color: #7b1fa2; }
        .modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.4); z-index: 2000; display: none; align-items: center; justify-content: center; backdrop-filter: blur(2px); }
        .modal { background: var(--card); width: 92%; max-width: 550px; max-height: 85vh; border-radius: 16px; display: flex; flex-direction: column; overflow: hidden; border: 1px solid var(--border); }
        .modal-header { padding: 15px; border-bottom: 1px solid var(--border); font-weight: bold; display: flex; justify-content: space-between; align-items: center; }
        .modal-body { flex: 1; overflow-y: auto; padding: 5px 0; }
        .res-item { padding: 15px; border-bottom: 1px solid var(--border); overflow: hidden; }
        .res-top { display: flex; justify-content: space-between; align-items: flex-start; gap: 10px; }
        .res-title { font-size: 14px; line-height: 1.4; color: var(--text); word-break: break-all; flex: 1; }
        .form-card { background: var(--card); padding: 20px; border-radius: 12px; border: 1px solid var(--border); margin-bottom: 15px; }
        .form-group { margin-bottom: 15px; } label { display: block; margin-bottom: 5px; font-size: 14px; color: var(--text-sub); font-weight: bold; }
        .form-control { width: 100%; padding: 12px; border: 1px solid var(--input-border); background: var(--input-bg); color: var(--text); border-radius: 8px; box-sizing: border-box; }
        .log-box { background: #f8f9fa; color: #212529; padding: 15px; border-radius: 12px; font-family: 'Fira Code', monospace; font-size: 13px; height: 60vh; overflow-y: auto; white-space: pre-wrap; line-height: 1.6; border: 1px solid #dee2e6; box-shadow: inset 0 0 5px rgba(0,0,0,0.05); transition: background 0.3s, color 0.3s; }
        html.dark .log-box { background: #0d1117; color: #c9d1d9; border-color: #30363d; box-shadow: inset 0 0 10px rgba(0,0,0,0.5); }
        .log-line-meta { color: #6c757d; font-size: 12px; margin-bottom: 4px; display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
        html.dark .log-line-meta { color: #8b949e; }
        .log-badge { padding: 2px 8px; border-radius: 4px; font-weight: bold; font-size: 10px; text-transform: uppercase; color: #fff; flex-shrink: 0; }
        .log-info { background: #28a745; } .log-error { background: #dc3545; } .log-warn { background: #ffc107; color: #212529 !important; } .log-debug { background: #007bff; }
        .log-msg { color: #212529; font-weight: 500; word-break: break-all; }
        html.dark .log-msg { color: #e6edf3; }
        .log-json { color: #0056b3; font-family: monospace; word-break: break-all; font-size: 11px; }
        html.dark .log-json { color: #79c0ff; }
        .log-source { color: #6c757d; font-style: italic; font-size: 11px; }
        html.dark .log-source { color: #8b949e; }
        .toast { position: fixed; bottom: 40px; left: 50%; transform: translateX(-50%); background: rgba(0,0,0,0.85); color: #fff; padding: 10px 20px; border-radius: 30px; display: none; z-index: 3000; }
        .tag { font-size: 10px; padding: 2px 6px; border-radius: 4px; border: 1px solid rgba(128,128,128,0.2); margin-right: 5px; margin-top: 5px; display: inline-block; }
        .section-title { font-weight: bold; color: var(--primary); margin-bottom: 15px; display: flex; align-items: center; gap: 8px; font-size: 16px; }
        .login-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); backdrop-filter: blur(5px); z-index: 5000; display: flex; align-items: center; justify-content: center; }
        .login-card { background: var(--card); width: 85%; max-width: 320px; padding: 30px; border-radius: 24px; border: 1px solid var(--border); box-shadow: 0 10px 25px rgba(0,0,0,0.1); }
    </style>
</head>
<body>
    <div id="loginModal" class="login-overlay" style="display:none;"><div class="login-card">
        <div style="text-align:center; font-size:20px; font-weight:bold; margin-bottom:20px;">身份验证</div>
        <input class="form-control" id="l_user" placeholder="用户名" style="margin-bottom:12px;">
        <input type="password" class="form-control" id="l_pass" placeholder="密码" style="margin-bottom:20px;">
        <button class="btn" style="height:46px; width:100%;" onclick="doLogin()">登 录</button>
        <div id="loginMsg" style="color:#fa5252; text-align:center; font-size:12px; margin-top:12px; min-height:15px;"></div>
    </div></div>
    <div class="navbar"><div class="brand" onclick="location.reload()">` + AppLogoSVG + ` 115Nexus</div><div class="nav-links">
        <div class="nav-item active" onclick="switchTab('search')">搜索</div><div class="nav-item" onclick="switchTab('logs')">日志</div><div class="nav-item" onclick="switchTab('settings')">设置</div><div class="nav-item" onclick="toggleTheme()" id="themeIcon">🌞</div>
    </div></div>
    <div class="container">
        <div id="tab-search" class="tab-content active"><div class="search-box">
            <select id="searchSource"><option value="tmdb">HDHive</option><option value="pansou">Pansou</option></select>
            <input type="text" id="searchInput" placeholder="搜索片名 / 粘贴115分享链接或磁力链" onkeypress="if(event.keyCode==13) doSearch()"><button class="btn" onclick="doSearch()">🔍</button>
        </div><div id="resultArea" class="list-group"></div><div id="searchStatus" style="text-align:center; color:var(--text-sub); margin-top:50px;"></div></div>
        <div id="tab-logs" class="tab-content"><div class="log-box" id="logContent"></div><div style="margin-top:10px; text-align:right;"><button class="btn btn-sm" onclick="fetchLogs()">刷新日志</button></div></div>
        <div id="tab-settings" class="tab-content">
            <div class="form-card"><div class="section-title">📦 HDHive <button class="btn btn-sm" style="margin-left:auto; background:var(--primary); font-size:11px;" onclick="doManualCheckin(this)">立即签到</button></div>
                <div class="form-group"><label>API Key</label><input class="form-control" id="hdhive_api_key"></div>
                <div style="display:flex; gap:10px; margin-bottom:15px;"><div style="flex:1;"><label>网页账号</label><input class="form-control" id="hdhive_user"></div><div style="flex:1;"><label>网页密码</label><input type="password" class="form-control" id="hdhive_pass"></div></div>
                <div style="display:grid; grid-template-columns: 1fr 1fr 1fr; gap:10px;">
                    <div><label>自动签到</label><select class="form-control" id="hdhive_checkin_enabled"><option value="true">开启</option><option value="false">关闭</option></select></div>
                    <div><label>Cron 表达式</label><input class="form-control" id="hdhive_checkin_cron" placeholder="0 9 * * *"></div>
                    <div><label>赌狗模式</label><select class="form-control" id="hdhive_gambler_mode"><option value="true">开启</option><option value="false">关闭</option></select></div>
                </div>
            </div>
            <div class="form-card"><div class="section-title">🔍 Pansou</div><div class="form-group"><label>Pansou URL</label><input class="form-control" id="pansou_url"></div><div style="display:flex; gap:10px;"><div style="flex:1;"><label>Pansou用户</label><input class="form-control" id="pansou_username"></div><div style="flex:1;"><label>Pansou密码</label><input type="password" class="form-control" id="pansou_password"></div></div></div>
            <div class="form-card"><div class="section-title">🔄 Media302</div><div class="form-group"><label>地址</label><input class="form-control" id="media302_base_url"></div><div class="form-group"><label>Token</label><input class="form-control" id="media302_token"></div><div style="display:flex; gap:10px;"><div style="flex:1;"><label>转存目录</label><input class="form-control" id="media302_folder"></div><div style="flex:1;"><label>磁力目录</label><input class="form-control" id="magnet_folder"></div></div></div>
            <div class="form-card"><div class="section-title">🎞️ 视频过滤</div><div class="form-group"><label>排除关键词 (正则)</label><textarea class="form-control" id="exclude_words" style="height:100px;"></textarea></div><div style="display:grid; grid-template-columns: 1fr 1fr; gap:10px;"><div><label>🎥 Min (MB)</label><input type="number" class="form-control" id="movie_min_size"></div><div><label>🎥 Max (MB)</label><input type="number" class="form-control" id="movie_max_size"></div><div><label>📺 Min (MB)</label><input type="number" class="form-control" id="tv_min_size"></div><div><label>📺 Max (MB)</label><input type="number" class="form-control" id="tv_max_size"></div></div></div>
            <div class="form-card"><div class="section-title">⚙️ 基础设置</div><div class="form-group"><label>TMDB API 密钥</label><input class="form-control" id="tmdb_api_key"></div><div class="form-group"><label>TG Bot Token</label><input class="form-control" id="tg_token"></div><div class="form-group"><label>Webhook 通知</label><input class="form-control" id="webhook_url"></div><div class="form-group"><label>代理地址</label><input class="form-control" id="proxy_url" placeholder="http://127.0.0.1:7890"></div><button class="btn" style="width:100%; height:46px; margin-top:20px;" onclick="saveConfig()">保存全部配置</button></div>
            <div style="text-align:center; color:var(--text-sub); font-size:12px; margin-bottom:30px;">Version v0.2.3</div>
        </div>
    </div>
    <div class="modal-overlay" id="resModal"><div class="modal"><div class="modal-header"><span id="resTitle"></span><span onclick="closeModal()" style="font-size:24px; cursor:pointer;">×</span></div><div class="modal-body" id="resList"></div></div></div>
    <div class="toast" id="toast"></div>
    <script>
        const fields=['tmdb_api_key','pansou_url','pansou_username','pansou_password','media302_base_url','media302_token','media302_folder','magnet_folder','proxy_url','exclude_words','movie_min_size','movie_max_size','tv_min_size','tv_max_size','hdhive_api_key','hdhive_checkin_enabled','hdhive_user','hdhive_pass','hdhive_gambler_mode','webhook_url','hdhive_checkin_hour','hdhive_checkin_cron','tg_token'];
        function switchTab(t){ document.querySelectorAll('.tab-content').forEach(e=>e.classList.remove('active')); document.querySelectorAll('.nav-item').forEach(e=>e.classList.remove('active')); document.getElementById('tab-'+t).classList.add('active'); document.querySelector('[onclick*="'+t+'"]').classList.add('active'); if(t==='settings')loadConfig(); if(t==='logs')fetchLogs(); }
        function loadConfig(){ fetch('/api/config').then(r=>{if(r.status===401){document.getElementById('loginModal').style.display='flex';throw new Error('401');}return r.json();}).then(d=>{fields.forEach(k=>{const e=document.getElementById(k);if(e)e.value=(d[k]!==undefined)?d[k].toString():'';});}).catch(e=>{});}
        async function doLogin(){const u=document.getElementById('l_user').value;const p=document.getElementById('l_pass').value;const btn=document.querySelector('#loginModal button');btn.disabled=true;btn.innerText='登录中...';try{const r=await fetch('/api/login',{method:'POST',body:JSON.stringify({username:u,password:p})});if(r.ok){document.getElementById('loginModal').style.display='none';loadConfig();}else{document.getElementById('loginMsg').innerText='错误';}}catch(e){document.getElementById('loginMsg').innerText='失败';}finally{btn.disabled=false;btn.innerText='登 录';}}
        async function doManualCheckin(b){let old=b.innerText;b.innerText='...';b.disabled=true;try{const r=await fetch('/api/hdhive/checkin');const d=await r.json();showToast(d.message);}finally{b.innerText=old;b.disabled=false;}}
        function saveConfig(){
            const p={}; fields.forEach(k=>{
                const e=document.getElementById(k); if(!e) return;
                const v=e.value;
                if(k.includes('_size')||k.includes('_hour')) p[k]=parseInt(v)||0;
                else if(v==='true')p[k]=true; else if(v==='false')p[k]=false;
                else p[k]=v;
            });
            fetch('/api/config',{method:'POST',body:JSON.stringify(p)}).then(r=>r.json()).then(r=>{
                showToast(r.success?'保存成功':'失败: '+r.message);
                if(r.success) loadConfig();
            }).catch(err => { showToast('保存请求异常'); console.error(err); });
        }
        async function doSearch(){
            const q=document.getElementById('searchInput').value.trim(); const s=document.getElementById('searchSource').value; if(!q)return;
            if(q.includes('115.com/s/') || q.includes('magnet:?xt=')){ pushResource(null, q, 0); return; }
            document.getElementById('searchStatus').innerText='🔍...'; try{
                const r=await fetch('/api/search?q='+encodeURIComponent(q)+'&source='+s); const d=await r.json();
                document.getElementById('searchStatus').innerText=d.results.length?'':'📭 无结果';
                document.getElementById('resultArea').innerHTML=d.results.map(item=>{
                    if(s==='pansou'){
                        let isMag = item.url.startsWith('magnet:?');
                        let btnText = isMag ? '磁力离线' : '115转存';
                        let btnClass = isMag ? 'btn-blue' : 'btn-green';
                        return '<div class="list-item"><div class="res-top"><b class="list-title">'+item.note+'</b><button class="btn btn-sm '+btnClass+'" onclick="pushResource(this,\''+item.url+'\',0)">'+btnText+'</button></div></div>';
                    }
                    let year=(item.release_date||item.first_air_date||'').substring(0,4);
                    return '<div class="list-item" onclick="openResources(\''+encodeURIComponent(item.id)+'\',\''+(item.media_type||'tv')+'\',\''+encodeURIComponent(item.title||item.name)+'\')"><b>'+(item.title||item.name)+'</b><div class="list-meta"><span class="badge">'+(item.media_type==='movie'?'电影':'剧集')+'</span><span>'+year+'</span></div></div>';
                }).join('');
            }catch(e){showToast('搜索失败');}
        }
        async function openResources(id,type,title){document.getElementById('resModal').style.display='flex';document.getElementById('resTitle').innerText=decodeURIComponent(title);document.getElementById('resList').innerHTML='<div style="text-align:center;padding:30px;">⏳</div>';try{const r=await fetch('/api/resources?id='+id+'&type='+type);const d=await r.json();if(!d.items||!d.items.length){document.getElementById('resList').innerHTML='<div style="text-align:center;padding:30px;">📭</div>';return;}document.getElementById('resList').innerHTML=d.items.map(res=>{let tags=(res.tags||[]).map(t=>'<span class="tag '+(t.includes('4K')?'tag-4k':'')+'">'+t+'</span>').join('');return '<div class="res-item"><div class="res-top"><div class="res-title">'+res.display+'</div><button class="btn btn-sm '+(res.hdhive_points>0?'btn-blue':'btn-green')+'" onclick="pushResource(this,\''+res.link+'\','+res.hdhive_points+')">'+(res.hdhive_points>0?res.hdhive_points+'积分转存':'转存')+'</button></div><div style="margin-top:5px;">'+tags+'</div></div>';}).join('');}catch(e){document.getElementById('resList').innerHTML='失败';}}
        async function pushResource(b,l,pts){if(pts>0&&!confirm('消耗积分转存？'))return;let old=b?b.innerText:'';if(b){b.innerText='...';b.disabled=true;}fetch('/api/push',{method:'POST',body:JSON.stringify({link:l})}).then(r=>r.json()).then(d=>{showToast(d.message);if(b){b.innerText=d.success?'✅':'❌';setTimeout(()=>{b.innerText=old;b.disabled=false;},2000);}});}
        function fetchLogs(){const el=document.getElementById('logContent');fetch('/api/logs').then(r=>r.text()).then(t=>{const blocks=t.trim().split('\n\n');const formatted=blocks.map(block=>{if(!block.trim())return '';const lines=block.split('\n');if(lines.length<2)return '';let lvl='INFO',time='';const firstMatch=lines[0].match(/\[(INFO|ERROR|WARN|DEBUG)\] (.*)/);if(firstMatch){lvl=firstMatch[1];time=firstMatch[2];}let source='',msg='',json='';const parts=lines[1].split(' - ');source=parts[0]||'';msg=parts[1]||'';json=parts[2]||'';const badgeClass='log-'+lvl.toLowerCase();let html='<div class="log-line-meta"><span class="log-badge '+badgeClass+'">'+lvl+'</span> <span>'+time+'</span>';if(json)html+=' - <span class="log-json">'+json+'</span>';html+=' - <span class="log-source">'+source+'</span> - <span class="log-msg">'+msg+'</span></div>';return '<div style="margin-bottom:8px; padding-bottom:4px; border-bottom:1px solid #21262d;">'+html+'</div>';}).join('');el.innerHTML=formatted;setTimeout(()=>{el.scrollTop=el.scrollHeight;},100);});}
        function toggleTheme(){const isDark=document.documentElement.classList.toggle('dark');localStorage.setItem('theme',isDark?'dark':'light');document.getElementById('themeIcon').innerText=isDark?'🌙':'🌞';}
        function showToast(m){const t=document.getElementById('toast');t.innerText=m;t.style.display='block';setTimeout(()=>t.style.display='none',3000);}
        function closeModal(){document.getElementById('resModal').style.display='none';}
        (function(){if(localStorage.getItem('theme')==='dark'){document.documentElement.classList.add('dark');document.getElementById('themeIcon').innerText='🌙';}loadConfig();})();
    </script>
</body>
</html>
`
