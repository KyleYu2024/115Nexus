package web

const AppLogoSVG = `<svg width="512" height="512" viewBox="0 0 512 512" fill="none" xmlns="http://www.w3.org/2000/svg"><defs><linearGradient id="iconGradient" x1="0" y1="0" x2="512" y2="512" gradientUnits="userSpaceOnUse"><stop stop-color="#3A8DFF"/><stop offset="1" stop-color="#003366"/></linearGradient><filter id="shadowFilter" x="-5%" y="-5%" width="110%" height="110%"><feDropShadow dx="4" dy="6" stdDeviation="5" flood-opacity="0.2"/></filter></defs><rect x="0" y="0" width="512" height="512" rx="114" fill="url(#iconGradient)"/><g filter="url(#shadowFilter)"><path d="M256 96 L416 384 L336 384 L288 288 L224 288 L176 384 L96 384 Z M256 168 L200 288 L312 288 Z" fill="#FFFFFF" fill-opacity="0.95"/></g></svg>`

const htmlPage = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover">
    <title>115Nexus</title>
    <link rel="manifest" href="/manifest.json">
    <link rel="icon" href="https://img.andp.cc/icons/upload/115Nexus.png">
    <link rel="apple-touch-icon" href="https://img.andp.cc/icons/upload/115Nexus.png">
    <style>
        :root {
            --primary: #007aff;
            --bg: #f9fafb;
            --card: #ffffff;
            --text: #111827;
            --text-sub: #6b7280;
            --border: #e5e7eb;
            --input-bg: #ffffff;
            --input-border: #d1d5db;
            --navbar-bg: rgba(255, 255, 255, 0.8);
            --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
            --shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
            --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
        }
        html.dark {
            --primary: #3b82f6;
            --bg: #09090b;
            --card: #18181b;
            --text: #f4f4f5;
            --text-sub: #a1a1aa;
            --border: #27272a;
            --input-bg: #18181b;
            --input-border: #3f3f46;
            --navbar-bg: rgba(9, 9, 11, 0.8);
        }
        * { box-sizing: border-box; -webkit-tap-highlight-color: transparent; }
        html { background: var(--bg); transition: background 0.3s; height: 100%; scroll-behavior: smooth; }
        body { 
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; 
            background: var(--bg); color: var(--text); margin: 0; display: flex; flex-direction: column; 
            height: 100vh; overflow: hidden; -webkit-font-smoothing: antialiased; 
        }
        
        .navbar { 
            flex-shrink: 0; position: sticky; top: 0; 
            background: var(--navbar-bg); 
            backdrop-filter: saturate(180%) blur(20px); -webkit-backdrop-filter: saturate(180%) blur(20px);
            padding: 10px 20px; padding-top: calc(10px + env(safe-area-inset-top)); 
            display: flex; justify-content: space-between; align-items: center; 
            border-bottom: 1px solid var(--border); z-index: 1000; min-height: 60px;
        }
        .brand { font-weight: 700; font-size: 1.15rem; cursor: pointer; color: var(--text); display: flex; align-items: center; gap: 10px; letter-spacing: -0.025em; }
        .brand svg { width: 32px; height: 32px; border-radius: 8px; box-shadow: var(--shadow-sm); }
        
        .nav-links { display: flex; gap: 4px; background: var(--bg); padding: 4px; border-radius: 12px; border: 1px solid var(--border); }
        .nav-item { 
            cursor: pointer; padding: 6px 16px; border-radius: 8px; font-size: 14px; font-weight: 500;
            color: var(--text-sub); transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
        }
        .nav-item:hover { color: var(--text); background: rgba(0,0,0,0.03); }
        html.dark .nav-item:hover { background: rgba(255,255,255,0.05); }
        .nav-item.active { background: var(--card); color: var(--primary); box-shadow: var(--shadow-sm); }
        
        .container { 
            flex: 1; overflow-y: auto; padding: 20px; padding-bottom: calc(30px + env(safe-area-inset-bottom)); 
            max-width: 1000px; margin: 0 auto; width: 100%; -webkit-overflow-scrolling: touch; 
        }
        
        .tab-content { display: none; } 
        .tab-content.active { display: block; animation: slideUp 0.3s cubic-bezier(0, 0, 0.2, 1); }
        @keyframes slideUp { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
        
        .search-box { display: flex; gap: 12px; margin-bottom: 24px; position: sticky; top: 0; z-index: 10; padding: 4px 0; background: var(--bg); }
        .search-box select, .search-box input { 
            background: var(--card); color: var(--text); border: 1px solid var(--input-border); 
            border-radius: 12px; padding: 12px 16px; font-size: 16px; outline: none;
            transition: all 0.2s; box-shadow: var(--shadow-sm);
        }
        .search-box input { flex: 1; min-width: 0; }
        .search-box input:focus, .search-box select:focus { border-color: var(--primary); box-shadow: 0 0 0 3px rgba(0, 122, 255, 0.15); }
        
        .btn { 
            background: var(--primary); color: #fff; border: none; padding: 12px 24px; border-radius: 12px; 
            font-weight: 600; cursor: pointer; display: flex; align-items: center; justify-content: center;
            transition: all 0.2s; gap: 8px; font-size: 15px; box-shadow: var(--shadow-sm);
        }
        .btn:hover { filter: brightness(1.05); transform: translateY(-1px); box-shadow: var(--shadow); }
        .btn:active { transform: scale(0.98); }
        .btn-green { background: #10b981; } .btn-blue { background: #3b82f6; } 
        .btn-sm { padding: 6px 12px; font-size: 13px; border-radius: 8px; }
        
        .list-item { 
            background: var(--card); padding: 20px; border-radius: 16px; margin-bottom: 12px; 
            border: 1px solid var(--border); transition: all 0.2s; box-shadow: var(--shadow-sm);
        }
        .list-item:hover { border-color: var(--primary); transform: translateY(-2px); box-shadow: var(--shadow); }
        .list-title { font-size: 1.05rem; font-weight: 600; line-height: 1.4; word-break: break-word; overflow-wrap: break-word; flex: 1; }
        .list-meta { font-size: 13px; color: var(--text-sub); display: flex; gap: 12px; align-items: center; }
        
        .res-top { display: flex; justify-content: space-between; align-items: center; gap: 12px; }
        .res-title { font-weight: 600; font-size: 15px; color: var(--text); line-height: 1.4; flex: 1; word-break: break-word; overflow-wrap: break-word; }
        
        .badge { 
            padding: 2px 8px; border-radius: 6px; font-size: 11px; font-weight: 600; 
            background: rgba(0, 122, 255, 0.1); color: var(--primary); text-transform: uppercase; 
        }
        
        .modal-overlay { 
            position: fixed; inset: 0; background: rgba(0,0,0,0.4); z-index: 2000; 
            display: none; align-items: center; justify-content: center; backdrop-filter: blur(8px); 
            animation: fadeIn 0.2s ease-out;
        }
        .modal { 
            background: var(--card); width: 95%; max-width: 600px; max-height: 80vh; 
            border-radius: 24px; display: flex; flex-direction: column; overflow: hidden; 
            border: 1px solid var(--border); box-shadow: var(--shadow-lg);
        }
        .modal-header { padding: 20px 24px; border-bottom: 1px solid var(--border); font-weight: 700; font-size: 1.1rem; display: flex; justify-content: space-between; align-items: center; }
        .modal-body { flex: 1; overflow-y: auto; padding: 10px 0; }
        
        .res-item { padding: 16px 24px; border-bottom: 1px solid var(--border); transition: background 0.2s; }
        .res-item:last-child { border-bottom: none; }
        .res-item:hover { background: rgba(0,0,0,0.01); }
        html.dark .res-item:hover { background: rgba(255,255,255,0.02); }
        
        .form-card { background: var(--card); padding: 24px; border-radius: 20px; border: 1px solid var(--border); margin-bottom: 20px; box-shadow: var(--shadow-sm); }
        .section-title { font-weight: 700; color: var(--text); margin-bottom: 20px; display: flex; align-items: center; gap: 8px; font-size: 17px; }
        .form-group { margin-bottom: 20px; } 
        label { display: block; margin-bottom: 8px; font-size: 14px; color: var(--text-sub); font-weight: 600; }
        .form-control { 
            width: 100%; padding: 12px 16px; border: 1px solid var(--input-border); background: var(--input-bg); 
            color: var(--text); border-radius: 12px; transition: all 0.2s; font-size: 15px; 
        }
        .form-control:focus { border-color: var(--primary); outline: none; box-shadow: 0 0 0 3px rgba(0, 122, 255, 0.15); }
        
        .log-box { 
            background: var(--input-bg); color: var(--text); padding: 20px; border-radius: 20px; 
            font-family: 'JetBrains Mono', 'Fira Code', ui-monospace, monospace; 
            font-size: 13px; height: 65vh; overflow-y: auto; overflow-x: hidden; line-height: 1.6; 
            border: 1px solid var(--border); box-shadow: inset 0 2px 10px rgba(0,0,0,0.05); 
        }
        html.dark .log-box { background: #09090b; color: #e4e4e7; border-color: #27272a; box-shadow: inset 0 2px 10px rgba(0,0,0,0.5); }
        .log-line { margin-bottom: 12px; padding-bottom: 12px; border-bottom: 1px solid var(--border); width: 100%; }
        html.dark .log-line { border-bottom-color: #18181b; }
        .log-line-meta { display: flex; align-items: center; gap: 10px; margin-bottom: 6px; font-size: 12px; color: var(--text-sub); flex-wrap: wrap; }
        html.dark .log-line-meta { color: #71717a; }
        .log-badge { 
            padding: 1px 10px; border-radius: 4px; font-weight: 700; font-size: 10px; 
            text-transform: uppercase; border: 1px solid currentColor; background: transparent; 
        }
        .log-info { color: #10b981; } .log-error { color: #ef4444; } .log-warn { color: #f59e0b; } .log-debug { color: #3b82f6; }
        .log-msg { color: var(--text); font-weight: 500; word-break: break-word; overflow-wrap: break-word; }
        html.dark .log-msg { color: #f4f4f5; }
        .log-json { color: var(--primary); font-size: 11px; opacity: 0.8; word-break: break-all; white-space: pre-wrap; }
        html.dark .log-json { color: #a5b4fc; }
        
        .toast { 
            position: fixed; bottom: 40px; left: 50%; transform: translateX(-50%); 
            background: var(--text); color: var(--bg); padding: 12px 24px; border-radius: 50px; 
            font-weight: 600; pointer-events: none; z-index: 3000; display: none;
            box-shadow: var(--shadow-lg); animation: slideUp 0.3s ease-out;
        }
        .tag { font-size: 11px; padding: 3px 8px; border-radius: 6px; background: rgba(128,128,128,0.1); color: var(--text-sub); margin: 4px 4px 0 0; display: inline-block; font-weight: 500; }
        @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
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
            <div style="text-align:center; color:var(--text-sub); font-size:12px; margin-bottom:30px;">Version 0.2.6</div>
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
        function fetchLogs(){
            const el=document.getElementById('logContent');
            fetch('/api/logs').then(r=>r.text()).then(t=>{
                const blocks=t.trim().split('\n\n');
                const formatted=blocks.map(block=>{
                    if(!block.trim())return '';
                    const lines=block.split('\n');
                    if(lines.length<2)return '';
                    let lvl='INFO',time='';
                    const firstMatch=lines[0].match(/\[(INFO|ERROR|WARN|DEBUG)\] (.*)/);
                    if(firstMatch){lvl=firstMatch[1];time=firstMatch[2];}
                    let source='',msg='',json='';
                    const parts=lines[1].split(' - ');
                    source=parts[0]||'';
                    msg=parts[1]||'';
                    json=parts[2]||'';
                    const badgeClass='log-'+lvl.toLowerCase();
                    return '<div class="log-line">' +
                                '<div class="log-line-meta">' +
                                    '<span class="log-badge ' + badgeClass + '">' + lvl + '</span>' +
                                    '<span>' + time + '</span>' +
                                    '<span style="margin-left:auto; opacity:0.5;">' + source + '</span>' +
                                '</div>' +
                                '<div class="log-msg">' + msg + '</div>' +
                                (json ? '<div class="log-json">' + json + '</div>' : '') +
                            '</div>';
                }).join('');
                el.innerHTML=formatted;
                setTimeout(()=>{el.scrollTop=el.scrollHeight;},100);
            });
        }
        function toggleTheme(){const isDark=document.documentElement.classList.toggle('dark');localStorage.setItem('theme',isDark?'dark':'light');document.getElementById('themeIcon').innerText=isDark?'🌙':'🌞';}
        function showToast(m){const t=document.getElementById('toast');t.innerText=m;t.style.display='block';setTimeout(()=>t.style.display='none',3000);}
        function closeModal(){document.getElementById('resModal').style.display='none';}
        (function(){if(localStorage.getItem('theme')==='dark'){document.documentElement.classList.add('dark');document.getElementById('themeIcon').innerText='🌙';}loadConfig();})();
    </script>
</body>
</html>
`