const fields=['tmdb_api_key','pansou_url','pansou_username','pansou_password','media302_base_url','media302_token','media302_folder','magnet_folder','proxy_url','exclude_words','movie_min_size','movie_max_size','tv_min_size','tv_max_size','hdhive_api_key','hdhive_checkin_enabled','hdhive_user','hdhive_pass','hdhive_gambler_mode','webhook_url','hdhive_checkin_hour','hdhive_checkin_cron','tg_token'];

function switchTab(t){ 
    document.querySelectorAll('.tab-content').forEach(e=>e.classList.remove('active')); 
    document.querySelectorAll('.nav-item').forEach(e=>e.classList.remove('active')); 
    const targetTab = document.getElementById('tab-'+t);
    if (targetTab) targetTab.classList.add('active'); 
    
    const navItem = document.querySelector(`[onclick*="switchTab('${t}')"]`);
    if (navItem) {
        navItem.classList.add('active');
        updateNavPill(navItem);
    }

    if(t==='settings') loadConfig(); 
    if(t==='logs') fetchLogs(); 
}

function updateNavPill(target) {
    const pill = document.getElementById('navPill');
    if (!pill || !target) return;
    
    // 液态拉伸效果：移动时增加拉伸类
    pill.classList.add('stretching');
    
    const rect = target.getBoundingClientRect();
    const parentRect = target.parentElement.getBoundingClientRect();
    
    pill.style.width = rect.width + 'px';
    pill.style.left = (rect.left - parentRect.left) + 'px';
    
    setTimeout(() => {
        pill.classList.remove('stretching');
    }, 400);
}

async function loadConfig(){ 
    try {
        const r = await fetch('/api/config');
        if(r.status === 401){
            document.getElementById('loginModal').style.display='flex';
            return;
        }
        const res = await r.json();
        if (res.success && res.data) {
            fields.forEach(k => {
                const e = document.getElementById(k);
                if(e) e.value = (res.data[k] !== undefined) ? res.data[k].toString() : '';
            });
        }
    } catch(e) {
        showToast('加载配置失败');
    }
}

async function doLogin(){
    const u=document.getElementById('l_user').value;
    const p=document.getElementById('l_pass').value;
    const btn=document.querySelector('#loginModal button');
    btn.disabled=true;
    btn.innerText='登录中...';
    try{
        const r = await fetch('/api/login',{
            method:'POST',
            headers: {'Content-Type': 'application/json'},
            body:JSON.stringify({username:u,password:p})
        });
        const res = await r.json();
        if(res.success){
            document.getElementById('loginModal').style.display='none';
            showToast('登录成功');
            loadConfig();
        }else{
            document.getElementById('loginMsg').innerText = res.message || '登录失败';
        }
    }catch(e){
        document.getElementById('loginMsg').innerText='网络请求失败';
    }finally{
        btn.disabled=false;
        btn.innerText='登 录';
    }
}

async function doManualCheckin(b){
    let old=b.innerText;
    b.innerText='...';
    b.disabled=true;
    try{
        const r=await fetch('/api/hdhive/checkin');
        const res=await r.json();
        showToast(res.message || (res.success ? '签到成功' : '签到失败'));
    }catch(e){
        showToast('请求异常');
    }finally{
        b.innerText=old;
        b.disabled=false;
    }
}

async function saveConfig(){
    const p={}; 
    fields.forEach(k=>{
        const e=document.getElementById(k); if(!e) return;
        const v=e.value;
        if(k.includes('_size')||k.includes('_hour')) p[k]=parseInt(v)||0;
        else if(v==='true')p[k]=true; 
        else if(v==='false')p[k]=false;
        else p[k]=v;
    });
    try {
        const r = await fetch('/api/config',{
            method:'POST',
            headers: {'Content-Type': 'application/json'},
            body:JSON.stringify(p)
        });
        const res = await r.json();
        showToast(res.message);
        if(res.success) loadConfig();
    } catch(err) {
        showToast('保存请求异常');
    }
}

async function doSearch(){
    const q=document.getElementById('searchInput').value.trim(); 
    const s=document.getElementById('searchSource').value; 
    if(!q)return;
    
    if(q.includes('115.com/s/') || q.includes('magnet:?xt=')){ 
        pushResource(null, q, 0); 
        return; 
    }
    
    const statusEl = document.getElementById('searchStatus');
    const resultArea = document.getElementById('resultArea');
    
    statusEl.innerText='🔍 正在搜索...'; 
    resultArea.innerHTML = '';

    try{
        const r = await fetch('/api/search?q='+encodeURIComponent(q)+'&source='+s); 
        const res = await r.json();
        
        if (!res.success) {
            statusEl.innerText = '❌ ' + (res.message || '搜索失败');
            return;
        }

        const items = res.data.results || [];
        statusEl.innerText = items.length ? '' : '📭 无结果';
        
        resultArea.innerHTML = items.map(item => {
            if(s==='pansou'){
                let isMag = item.url.startsWith('magnet:?');
                let btnText = isMag ? '磁力离线' : '115转存';
                let btnClass = isMag ? 'btn-blue' : 'btn-green';
                return `<div class="list-item">
                            <div class="res-top">
                                <b class="list-title">${item.note}</b>
                                <button class="btn btn-sm ${btnClass}" onclick="pushResource(this,'${item.url}',0)">${btnText}</button>
                            </div>
                        </div>`;
            }
            const title = item.title || item.name;
            const year = (item.release_date || item.first_air_date || '').substring(0, 4);
            return `<div class="list-item search-result-item" onclick="openResources('${encodeURIComponent(item.id)}','${item.media_type||'tv'}','${encodeURIComponent(title)}')">
                        ${item.poster_path ? `<img class="search-item-poster" src="https://image.tmdb.org/t/p/w200${item.poster_path}">` : '<div class="search-item-poster no-poster">🎬</div>'}
                        <div class="search-item-info">
                            <b class="list-title">${title}</b>
                            <div class="list-meta">
                                <span class="badge">${item.media_type==='movie'?'电影':'剧集'}</span>
                                <span>${year}</span>
                            </div>
                        </div>
                    </div>`;
        }).join('');
    }catch(e){
        statusEl.innerText = '❌ 搜索请求异常';
    }
}

function showSelectionModal(items, q) {
    const modal = document.getElementById('resModal');
    const listEl = document.getElementById('resList');
    const detailEl = document.getElementById('movieDetail');
    document.getElementById('resTitle').innerText = `搜索: ${q}`;
    
    detailEl.style.display = 'none';
    modal.style.display = 'flex';
    
    listEl.innerHTML = `
        <div style="padding:15px 24px; font-size:14px; color:var(--text-sub); border-bottom:1px solid var(--border);">请选择正确的影视项目：</div>
        ${items.map(item => {
            const title = item.title || item.name;
            const year = (item.release_date || item.first_air_date || '').substring(0, 4);
            return `<div class="res-item search-result-item" onclick="openResources('${encodeURIComponent(item.id)}','${item.media_type||'tv'}','${encodeURIComponent(title)}', true)">
                        ${item.poster_path ? `<img class="search-item-poster" src="https://image.tmdb.org/t/p/w200${item.poster_path}">` : '<div class="search-item-poster no-poster">🎬</div>'}
                        <div class="search-item-info">
                            <b class="list-title">${title}</b>
                            <div class="list-meta">
                                <span class="badge">${item.media_type==='movie'?'电影':'剧集'}</span>
                                <span>${year}</span>
                            </div>
                        </div>
                    </div>`;
        }).join('')}
    `;
}

async function openResources(id,type,title){
    const modal = document.getElementById('resModal');
    const listEl = document.getElementById('resList');
    const detailEl = document.getElementById('movieDetail');
    
    document.getElementById('resTitle').innerText = decodeURIComponent(title);
    
    modal.style.display='flex';
    listEl.innerHTML='<div style="text-align:center;padding:30px;">⏳ 正在获取资源...</div>';
    detailEl.style.display='none';
    
    try{
        const r = await fetch(`/api/resources?id=${id}&type=${type}`);
        const res = await r.json();
        
        if(!res.success) {
            listEl.innerHTML = `<div style="text-align:center;padding:30px;color:red;">❌ ${res.message}</div>`;
            return;
        }

        const items = res.data.items || [];
        if(!items.length){
            listEl.innerHTML='<div style="text-align:center;padding:30px;">📭 该影视暂无可用资源</div>';
            return;
        }

        listEl.innerHTML = items.map(item => {
            let tags = (item.tags||[]).map(t=>`<span class="tag ${t.includes('4K')?'tag-4k':''}">${t}</span>`).join('');
            const btnClass = item.hdhive_points > 0 ? 'btn-blue' : 'btn-green';
            const btnText = item.hdhive_points > 0 ? `${item.hdhive_points}pt 转存` : '一键转存';
            
            return `<div class="res-item">
                        <div class="res-top">
                            <div class="res-title">${item.display}</div>
                            <button class="btn btn-sm ${btnClass}" onclick="pushResource(this,'${item.link}',${item.hdhive_points})">${btnText}</button>
                        </div>
                        <div class="res-line">
                            <div style="flex:1;">${tags}</div>
                            <button class="btn btn-sm btn-link" onclick="getLink(this, '${item.link}')">🔗 获取分享链</button>
                        </div>
                    </div>`;
        }).join('');
    }catch(e){
        listEl.innerHTML='<div style="text-align:center;padding:30px;">❌ 请求失败</div>';
    }
}

async function getLink(b, l) {
    let oldText = b.innerText;
    b.innerText = '获取中...';
    b.disabled = true;
    try {
        const r = await fetch('/api/get-link', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({link: l})
        });
        const res = await r.json();
        if(res.success) {
            b.innerText = '📋 点击复制';
            b.disabled = false;
            b.onclick = () => {
                const el = document.createElement('textarea');
                el.value = res.data.link;
                document.body.appendChild(el);
                el.select();
                document.execCommand('copy');
                document.body.removeChild(el);
                showToast('已复制到剪贴板');
                b.innerText = '✅ 已复制';
                setTimeout(() => { b.innerText = '📋 点击复制'; }, 2000);
            };
        } else {
            showToast(res.message);
            b.innerText = oldText;
            b.disabled = false;
        }
    } catch(e) {
        showToast('请求失败');
        b.innerText = oldText;
        b.disabled = false;
    }
}

async function pushResource(b,l,pts){
    if(pts > 0 && !confirm(`解锁此资源将消耗 ${pts} 积分，确定继续吗？`)) return;
    
    let oldText = b ? b.innerText : '';
    if(b){
        b.innerText = '提交中...';
        b.disabled = true;
    }

    try {
        const r = await fetch('/api/push',{
            method:'POST',
            headers: {'Content-Type': 'application/json'},
            body:JSON.stringify({link:l})
        });
        const res = await r.json();
        showToast(res.message);
        
        if(b){
            b.innerText = res.success ? '✅ 已提交' : '❌ 失败';
            setTimeout(() => {
                b.innerText = oldText;
                b.disabled = false;
            }, 2000);
        }
    } catch(e) {
        showToast('推送请求异常');
        if(b) {
            b.innerText = oldText;
            b.disabled = false;
        }
    }
}

let logEventSource = null;

function fetchLogs(){
    if(logEventSource) {
        logEventSource.close();
    }

    fetch('/api/logs').then(r=>r.text()).then(t=>{
        renderLogBlocks(t, true);
        setupLogStream();
    }).catch(() => showToast('无法加载历史日志'));
}

function setupLogStream() {
    logEventSource = new EventSource('/api/logs');

    logEventSource.onmessage = function(event) {
        if (event.data) {
            renderLogBlocks(event.data, false);
        }
    };

    logEventSource.onerror = function() {
        console.warn("日志连接断开，5秒后重试...");
        logEventSource.close();
        setTimeout(setupLogStream, 5000);
    };
}

function renderLogBlocks(text, isFull) {
    const el = document.getElementById('logContent');
    if (!el) return;

    const blocks = text.trim().split('\n\n');
    const html = blocks.map(block => {
        if(!block.trim()) return '';
        const lines = block.split('\n');
        if(lines.length < 2) return '';

        let lvl='INFO', time='';
        const firstLine = lines[0].match(/\[(INFO|ERROR|WARN|DEBUG)\] (.*)/);
        if(firstLine){ lvl=firstLine[1]; time=firstLine[2]; }

        const parts = lines[1].split(' - ');
        const source = parts[0] || '';
        const msg = parts[1] || '';
        const json = parts[2] || '';

        return `<div class="log-line">
                    <div class="log-line-meta">
                        <span class="log-badge log-${lvl.toLowerCase()}">${lvl}</span>
                        <span>${time}</span>
                        <span style="margin-left:auto; opacity:0.5;">${source}</span>
                    </div>
                    <div class="log-msg">${msg}</div>
                    ${json ? `<div class="log-json">${json}</div>` : ''}
                </div>`;
    }).join('');

    if (isFull) {
        el.innerHTML = html;
    } else {
        el.insertAdjacentHTML('beforeend', html);
    }
    
    // 自动滚动到底部
    setTimeout(() => { el.scrollTop = el.scrollHeight; }, 50);
}

function showToast(m){
    const t = document.getElementById('toast');
    if(!t) return;
    t.innerText = m;
    t.style.display = 'block';
    setTimeout(() => { t.style.display = 'none'; }, 3000);
}

function closeModal(){
    document.getElementById('resModal').style.display='none';
}

(function(){
    document.documentElement.classList.add('dark');
    loadConfig();
    
    // 初始化药丸位置
    setTimeout(() => {
        const active = document.querySelector('.nav-item.active');
        if(active) updateNavPill(active);
    }, 100);
})();
