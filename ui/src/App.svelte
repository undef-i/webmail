<script>
  import { onMount } from 'svelte';
  import {
    Inbox,
    Send,
    FileEdit,
    Archive,
    OctagonAlert,
    Trash2,
    Settings,
    LogOut,
    Menu,
    ChevronLeft,
    RefreshCw,
    Paperclip,
    Star,
    Mail,
    MailOpen,
    PenLine,
    Reply,
    Forward,
    Search,
    X
  } from 'lucide-svelte';
  import Auth from './lib/Auth.svelte';
  import Accounts from './lib/Accounts.svelte';
  import Compose from './lib/Compose.svelte';
  import * as api from './lib/api.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Separator } from '$lib/components/ui/separator/index.js';
  import { Switch } from '$lib/components/ui/switch/index.js';
  import { Sheet, SheetContent } from '$lib/components/ui/sheet/index.js';
  import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle
  } from '$lib/components/ui/dialog/index.js';

  const SCOPES = [
    { key: 'inbox', label: '收件箱', icon: Inbox },
    { key: 'sent', label: '已发送', icon: Send },
    { key: 'drafts', label: '草稿箱', icon: FileEdit },
    { key: 'archive', label: '归档', icon: Archive },
    { key: 'junk', label: '垃圾邮件', icon: OctagonAlert },
    { key: 'trash', label: '回收站', icon: Trash2 }
  ];

  const PAGE_SIZE = 60;

  let view = $state('loading');
  let authMode = $state('login');
  let accounts = $state([]);
  let folders = $state([]);
  let messages = $state([]);
  let statuses = $state({});
  let page = $state('mail');
  let scope = $state('inbox');
  let selected = $state(null);
  let detail = $state(null);
  let detailLoading = $state(false);
  let detailError = $state('');
  let opBusy = $state(false);
  let detailToken = 0;
  let frameEl = $state(null);
  let drawerOpen = $state(false);
  let listLoading = $state(false);
  let loadingMore = $state(false);
  let listHasMore = $state(false);
  let dark = $state(false);
  let composeOpen = $state(false);
  let composeInit = $state({});
  let searching = $state(false);
  let searchQuery = $state('');
  let searchTimer = null;
  let threaded = $state(false);
  let expandedKey = $state('');
  let threadMsgs = $state([]);
  let cursorIdx = $state(0);
  let helpOpen = $state(false);
  let searchInputEl = $state(null);
  let listEl = $state(null);
  let pendingReply = false;
  let listOffset = 0;
  const detailCache = new Map();
  let es = null;
  let esTimer = null;
  let esDelay = 1000;
  let statusTimer = null;

  const scopeLabel = $derived(SCOPES.find((s) => s.key === scope)?.label ?? scope);
  const scopeStats = $derived.by(() => {
    const map = {};
    for (const f of folders) {
      if (!f.special_use) continue;
      const cur = map[f.special_use] ?? { total: 0, unread: 0, present: false };
      cur.total += f.total;
      cur.unread += f.unread;
      cur.present = true;
      map[f.special_use] = cur;
    }
    return map;
  });
  const visibleScopes = $derived(SCOPES.filter((s) => scopeStats[s.key]?.present));

  function applyDark() {
    document.documentElement.classList.toggle('dark', dark);
    localStorage.setItem('wm_dark', dark ? '1' : '0');
  }

  function fmtListTime(ts) {
    const d = new Date(ts * 1000);
    const now = new Date();
    const pad = (n) => String(n).padStart(2, '0');
    if (d.toDateString() === now.toDateString()) return pad(d.getHours()) + ':' + pad(d.getMinutes());
    if (d.getFullYear() === now.getFullYear()) return d.getMonth() + 1 + '/' + d.getDate();
    return d.getFullYear() + '/' + (d.getMonth() + 1) + '/' + d.getDate();
  }

  function fmtFull(ts) {
    const d = new Date(ts * 1000);
    const pad = (n) => String(n).padStart(2, '0');
    return (
      d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) + ' ' +
      pad(d.getHours()) + ':' + pad(d.getMinutes())
    );
  }

  async function loadMessages() {
    listLoading = true;
    listOffset = 0;
    expandedKey = '';
    threadMsgs = [];
    cursorIdx = 0;
    try {
      if (searching) {
        messages = await api.searchMessages(searchQuery);
        listHasMore = false;
      } else {
        const batch = await api.getMessages(scope, 0, PAGE_SIZE, threaded);
        messages = batch;
        listOffset = batch.length;
        listHasMore = batch.length === PAGE_SIZE;
      }
    } catch {
      messages = [];
      listHasMore = false;
    } finally {
      listLoading = false;
    }
  }

  async function loadMore() {
    if (searching || listLoading || loadingMore || !listHasMore) return;
    loadingMore = true;
    try {
      const batch = await api.getMessages(scope, listOffset, PAGE_SIZE, threaded);
      const seen = new Set(messages.map((m) => m.id));
      messages = [...messages, ...batch.filter((m) => !seen.has(m.id))];
      listOffset += batch.length;
      listHasMore = batch.length === PAGE_SIZE;
    } catch {
    } finally {
      loadingMore = false;
    }
  }

  async function reloadLoaded() {
    if (searching || listLoading) return;
    try {
      const n = Math.max(PAGE_SIZE, messages.length);
      const batch = await api.getMessages(scope, 0, n, threaded);
      messages = batch;
      listOffset = batch.length;
      listHasMore = batch.length === n;
    } catch {
    }
  }

  function onListScroll() {
    if (!listEl) return;
    if (listEl.scrollTop + listEl.clientHeight > listEl.scrollHeight - 500) loadMore();
  }

  function onSearchInput() {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => {
      searching = searchQuery.trim().length > 0;
      selected = null;
      detail = null;
      loadMessages();
    }, 300);
  }

  function clearSearch() {
    searchQuery = '';
    searching = false;
    loadMessages();
  }

  function openCompose(init) {
    composeInit = init;
    drawerOpen = false;
    composeOpen = true;
  }

  function fmtAddrList(list) {
    return (list ?? [])
      .map((a) => (a.name ? a.name + ' <' + a.email + '>' : a.email))
      .join(', ');
  }

  async function loadFolders() {
    try {
      folders = await api.getFolders();
    } catch {
      folders = [];
    }
  }

  async function loadStatuses() {
    const entries = await Promise.all(
      accounts.map(async (a) => {
        try {
          return [a.id, await api.getSyncStatus(a.id)];
        } catch {
          return [a.id, null];
        }
      })
    );
    statuses = Object.fromEntries(entries);
  }

  async function refreshData() {
    await Promise.all([loadFolders(), loadMessages(), loadStatuses()]);
  }


  function scheduleStatusRefresh() {
    clearTimeout(statusTimer);
    statusTimer = setTimeout(() => {
      if (view === 'app') loadStatuses();
    }, 600);
  }

  function rowMatchesScope(item) {
    return searching ? false : item.folder_special === scope;
  }

  function insertFlat(item) {
    const i = messages.findIndex((m) => m.id === item.id);
    if (i >= 0) {
      messages[i] = { ...messages[i], ...item };
      messages = [...messages];
      return;
    }
    if (!rowMatchesScope(item)) return;
    messages = [...messages, item].sort((a, b) => b.date_ts - a.date_ts || b.id - a.id);
  }

  function insertThreaded(item) {
    const key = item.thread_key || 'id:' + item.id;
    const i = messages.findIndex((m) => m.account_id === item.account_id && (m.thread_key || 'id:' + m.id) === key);
    if (i < 0) {
      if (!rowMatchesScope(item)) return;
      const row = { ...item, thread_key: key, thread_count: 1, thread_unread: item.flags & 1 ? 0 : 1 };
      messages = [...messages, row].sort((a, b) => b.date_ts - a.date_ts || b.id - a.id);
      return;
    }
    const row = { ...messages[i] };
    if (messages.some((m) => m.id === item.id && m !== messages[i])) return;
    row.thread_count = (row.thread_count ?? 1) + 1;
    if (!(item.flags & 1)) row.thread_unread = (row.thread_unread ?? 0) + 1;
    if (item.date_ts > row.date_ts || (item.date_ts === row.date_ts && item.id > row.id)) {
      const counts = { thread_count: row.thread_count, thread_unread: row.thread_unread };
      Object.assign(row, item, counts);
    }
    messages[i] = row;
    messages = [...messages].sort((a, b) => b.date_ts - a.date_ts || b.id - a.id);
    if (expandedKey === item.account_id + ':' + key) {
      threadMsgs = [...threadMsgs, item].sort((a, b) => a.date_ts - b.date_ts || a.id - b.id);
    }
  }

  function applyFlagsEvent(items) {
    const byId = new Map(items.map((it) => [it.id, it.flags]));
    const patch = (m) => (byId.has(m.id) ? { ...m, flags: byId.get(m.id) } : m);
    messages = messages.map((m) => {
      if (!byId.has(m.id)) return m;
      const next = patch(m);
      if (threaded && m.thread_count > 1 && (m.flags & 1) !== (next.flags & 1)) {
        reloadLoaded();
      }
      return next;
    });
    threadMsgs = threadMsgs.map(patch);
    if (selected && byId.has(selected.id)) {
      selected = { ...selected, flags: byId.get(selected.id) };
      if (detail) detail = { ...detail, flags: selected.flags };
    }
  }

  function applyRemovedEvent(ev) {
    const ids = new Set(ev.ids ?? []);
    if (ids.size === 0) return;
    if (threaded && messages.some((m) => ids.has(m.id))) {
      reloadLoaded();
    } else {
      messages = messages.filter((m) => !ids.has(m.id));
    }
    threadMsgs = threadMsgs.filter((m) => !ids.has(m.id));
    if (selected && ids.has(selected.id)) {
      selected = null;
      detail = null;
    }
  }

  function applyFolderEvent(folder) {
    const i = folders.findIndex((f) => f.id === folder.id);
    if (i >= 0) {
      folders[i] = folder;
      folders = [...folders];
    } else {
      folders = [...folders, folder];
    }
  }

  function onServerEvent(ev) {
    if (view !== 'app' || page !== 'mail') return;
    if (ev.type === 'messages') {
      for (const item of ev.items ?? []) {
        if (threaded && !searching) insertThreaded(item);
        else insertFlat(item);
      }
      scheduleStatusRefresh();
    } else if (ev.type === 'flags') {
      applyFlagsEvent(ev.items ?? []);
    } else if (ev.type === 'removed') {
      applyRemovedEvent(ev);
    } else if (ev.type === 'folder') {
      applyFolderEvent(ev.folder);
    }
  }

  function connectEvents() {
    disconnectEvents();
    try {
      es = new EventSource('/api/events');
    } catch {
      scheduleReconnect();
      return;
    }
    es.onopen = () => {
      const wasReconnect = esDelay > 1000;
      esDelay = 1000;
      if (wasReconnect && view === 'app') {
        loadFolders();
        reloadLoaded();
        loadStatuses();
      }
    };
    es.onmessage = (e) => {
      try {
        onServerEvent(JSON.parse(e.data));
      } catch {
      }
    };
    es.onerror = () => {
      es?.close();
      es = null;
      scheduleReconnect();
    };
  }

  function scheduleReconnect() {
    clearTimeout(esTimer);
    esTimer = setTimeout(() => {
      if (view === 'app') connectEvents();
    }, esDelay);
    esDelay = Math.min(esDelay * 2, 30000);
  }

  function disconnectEvents() {
    clearTimeout(esTimer);
    if (es) {
      es.close();
      es = null;
    }
  }

  async function refreshAccounts() {
    try {
      accounts = await api.listAccounts();
    } catch {
      accounts = [];
    }
    await refreshData();
  }

  function statusText(a) {
    const st = statuses[a.id];
    if (!st) return '';
    if (st.state === 'idle') return '已同步 ' + fmtListTime(st.last_sync_at);
    if (st.state === 'syncing') return '同步中';
    if (st.state === 'connecting') return '连接中';
    if (st.state === 'error') return '同步失败';
    return '未同步';
  }

  function selectScope(key) {
    scope = key;
    selected = null;
    detail = null;
    drawerOpen = false;
    page = 'mail';
    searchQuery = '';
    searching = false;
    loadMessages();
  }

  function fmtSize(n) {
    if (n < 1024) return n + ' B';
    if (n < 1024 * 1024) return (n / 1024).toFixed(1) + ' KB';
    return (n / 1024 / 1024).toFixed(1) + ' MB';
  }

  const FRAME_CSS =
    '<style>html,body{margin:0;padding:0;background:#fff;color:#000;' +
    'font:14px/1.5 Arial,Helvetica,"PingFang SC","Microsoft YaHei",sans-serif;' +
    'overflow-wrap:anywhere}img{max-width:100%;height:auto}table{max-width:100%}a{color:#0000ee}</style>';

  function frameLoaded() {
    try {
      const doc = frameEl?.contentDocument;
      if (doc) frameEl.style.height = doc.documentElement.scrollHeight + 'px';
    } catch {
    }
  }

  function patchLocalFlags(id, flags) {
    messages = messages.map((x) => (x.id === id ? { ...x, flags } : x));
    threadMsgs = threadMsgs.map((x) => (x.id === id ? { ...x, flags } : x));
    if (selected?.id === id) {
      selected = { ...selected, flags };
      if (detail) detail = { ...detail, flags };
    }
  }

  async function markSeen(m) {
    if (m.flags & 1) return;
    try {
      const res = await api.setFlags(m.id, { seen: true });
      patchLocalFlags(m.id, res.flags);
      loadFolders();
    } catch {
    }
  }

  function prefetchNeighbors(m) {
    const pool =
      threadMsgs.length && threadMsgs.some((x) => x.id === m.id) ? threadMsgs : messages;
    const i = pool.findIndex((x) => x.id === m.id);
    if (i < 0) return;
    for (const n of [pool[i - 1], pool[i + 1]]) {
      if (n && !detailCache.has(n.id)) {
        api
          .getMessage(n.id)
          .then((d) => detailCache.set(n.id, d))
          .catch(() => {});
      }
    }
  }

  function afterOpen(m) {
    if (pendingReply) {
      pendingReply = false;
      replySelected();
    }
    prefetchNeighbors(m);
  }

  let booted = false;

  async function openMessageById(id) {
    try {
      const d = await api.getMessage(id);
      detailCache.set(id, d);
      await openMessage({
        id: d.id,
        account_id: d.account_id,
        account_label: d.account_label,
        folder_id: d.folder_id,
        folder_name: d.folder_name,
        folder_special: '',
        thread_key: '',
        subject: d.subject || '',
        from_name: d.from?.name || '',
        from_email: d.from?.email || '',
        date_ts: d.date_ts,
        flags: d.flags,
        has_attach: !!d.has_attach,
        snippet: '',
        size: d.size
      });
    } catch {
    }
  }

  function onHashChange() {
    if (!booted) return;
    const m = /^#\/mail\/(\d+)$/.exec(location.hash) || /^#m-(\d+)$/.exec(location.hash);
    if (m) {
      const id = +m[1];
      if (selected?.id !== id) openMessageById(id);
    } else if (selected) {
      selected = null;
      detail = null;
    }
  }

  $effect(() => {
    const want = selected ? '#/mail/' + selected.id : '';
    if (!booted) return;
    const cur = location.hash.startsWith('#/mail/') || location.hash.startsWith('#m-') ? location.hash : '';
    if (want !== cur) {
      history.pushState(null, '', want === '' ? location.pathname + location.search : want);
    }
  });

  async function openMessage(m) {
    if (scope === 'drafts' && !searching) {
      try {
        const d = await api.getMessage(m.id);
        openCompose({
          account_id: m.account_id,
          to: fmtAddrList(d.to),
          cc: fmtAddrList(d.cc),
          subject: d.subject || m.subject || '',
          text: d.text || '',
          draft_id: m.id
        });
        return;
      } catch {
      }
    }
    selected = m;
    detailError = '';
    const cached = detailCache.get(m.id);
    if (cached) {
      detail = cached;
      detailLoading = false;
      markSeen(m);
      afterOpen(m);
      return;
    }
    detail = null;
    detailLoading = true;
    const tk = ++detailToken;
    try {
      const d = await api.getMessage(m.id);
      if (tk !== detailToken) return;
      detail = d;
      detailCache.set(m.id, d);
      markSeen(m);
      afterOpen(m);
    } catch (err) {
      if (tk === detailToken) detailError = err.message;
    } finally {
      if (tk === detailToken) detailLoading = false;
    }
  }

  async function toggleThread(m) {
    const key = m.account_id + ':' + (m.thread_key || 'id:' + m.id);
    if (expandedKey === key) {
      expandedKey = '';
      threadMsgs = [];
      return;
    }
    expandedKey = key;
    threadMsgs = [];
    try {
      threadMsgs = await api.getThreadMessages(scope, m.account_id, m.thread_key || 'id:' + m.id);
    } catch {
      threadMsgs = [];
    }
  }

  function rowClick(m, idx) {
    cursorIdx = idx;
    if (threaded && !searching && (m.thread_count ?? 1) > 1) {
      toggleThread(m);
      return;
    }
    openMessage(m);
  }

  async function toggleSeen() {
    if (!selected || opBusy) return;
    opBusy = true;
    try {
      const res = await api.setFlags(selected.id, { seen: !(selected.flags & 1) });
      patchLocalFlags(selected.id, res.flags);
      loadFolders();
    } finally {
      opBusy = false;
    }
  }

  async function toggleStar() {
    if (!selected || opBusy) return;
    opBusy = true;
    try {
      const res = await api.setFlags(selected.id, { flagged: !(selected.flags & 2) });
      patchLocalFlags(selected.id, res.flags);
    } finally {
      opBusy = false;
    }
  }

  async function starFor(m) {
    try {
      const res = await api.setFlags(m.id, { flagged: !(m.flags & 2) });
      patchLocalFlags(m.id, res.flags);
    } catch {
    }
  }

  function quoteText() {
    const from = detail.from
      ? (detail.from.name ? detail.from.name + ' ' : '') + '<' + detail.from.email + '>'
      : '';
    return (
      '\n\n----- 原邮件 -----\n发件人：' +
      from +
      '\n时间：' +
      fmtFull(selected.date_ts) +
      '\n主题：' +
      (selected.subject || '') +
      '\n\n' +
      (detail.text || '')
    );
  }

  function replySelected() {
    if (!selected || !detail) return;
    const subj = selected.subject || '';
    openCompose({
      account_id: selected.account_id,
      to: detail.from
        ? detail.from.name
          ? detail.from.name + ' <' + detail.from.email + '>'
          : detail.from.email
        : '',
      subject: /^re:/i.test(subj) ? subj : 'Re: ' + subj,
      text: quoteText(),
      reply_to_id: selected.id
    });
  }

  function forwardSelected() {
    if (!selected || !detail) return;
    const subj = selected.subject || '';
    openCompose({
      account_id: selected.account_id,
      subject: /^(fwd|fw):/i.test(subj) ? subj : 'Fwd: ' + subj,
      text: quoteText()
    });
  }

  function afterRemoval(id) {
    threadMsgs = threadMsgs.filter((m) => m.id !== id);
    if (selected?.id === id) {
      selected = null;
      detail = null;
    }
    if (threaded && !searching) {
      reloadLoaded();
    } else {
      messages = messages.filter((m) => m.id !== id);
    }
    loadFolders();
  }

  async function archiveFor(m) {
    if (!m || opBusy) return;
    opBusy = true;
    try {
      await api.moveMessage(m.id, 'archive');
      afterRemoval(m.id);
    } finally {
      opBusy = false;
    }
  }

  async function deleteFor(m) {
    if (!m || opBusy) return;
    opBusy = true;
    try {
      await api.deleteMessage(m.id);
      afterRemoval(m.id);
    } finally {
      opBusy = false;
    }
  }

  function archiveSelected() {
    archiveFor(selected);
  }

  function deleteSelected() {
    deleteFor(selected);
  }

  async function markAllRead() {
    if (searching || opBusy) return;
    opBusy = true;
    try {
      await api.markAllRead(scope);
      messages = messages.map((m) =>
        threaded ? { ...m, flags: m.flags | 1, thread_unread: 0 } : { ...m, flags: m.flags | 1 }
      );
      threadMsgs = threadMsgs.map((m) => ({ ...m, flags: m.flags | 1 }));
      if (selected) patchLocalFlags(selected.id, selected.flags | 1);
      loadFolders();
    } catch {
    } finally {
      opBusy = false;
    }
  }


  function activeMessage() {
    if (selected) return selected;
    return messages[cursorIdx] ?? null;
  }

  function replyToMessage(m) {
    if (selected?.id === m.id && detail) {
      replySelected();
      return;
    }
    pendingReply = true;
    openMessage(m);
  }

  function moveCursor(d) {
    if (messages.length === 0) return;
    cursorIdx = Math.max(0, Math.min(messages.length - 1, cursorIdx + d));
    requestAnimationFrame(() => {
      listEl
        ?.querySelector('[data-rowidx="' + cursorIdx + '"]')
        ?.scrollIntoView({ block: 'nearest' });
    });
  }

  function onGlobalKey(e) {
    if (view !== 'app' || page !== 'mail') return;
    if (composeOpen || helpOpen) return;
    const t = e.target;
    if (
      t &&
      (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT' || t.isContentEditable)
    ) {
      return;
    }
    const m = activeMessage();
    switch (e.key) {
      case 'j':
        moveCursor(1);
        break;
      case 'k':
        moveCursor(-1);
        break;
      case 'Enter': {
        const row = messages[cursorIdx];
        if (row) rowClick(row, cursorIdx);
        break;
      }
      case 'e':
        if (m) archiveFor(m);
        break;
      case 'Delete':
        if (m) deleteFor(m);
        break;
      case 's':
        if (m) starFor(m);
        break;
      case 'r':
        if (m) replyToMessage(m);
        break;
      case 'a':
        markAllRead();
        break;
      case 'c':
        openCompose({});
        break;
      case '/':
        e.preventDefault();
        document.querySelector('#wm-search')?.focus();
        break;
      case '?':
        helpOpen = true;
        break;
      default:
        return;
    }
  }

  function openAccounts() {
    page = 'accounts';
    drawerOpen = false;
  }

  async function enter() {
    view = 'app';
    threaded = localStorage.getItem('wm_threaded') === '1';
    await refreshAccounts();
    connectEvents();
    booted = true;
    onHashChange();
  }

  async function boot() {
    window.addEventListener('hashchange', onHashChange);
    dark = localStorage.getItem('wm_dark') === '1';
    applyDark();
    try {
      const st = await api.getState();
      if (!st.setup) {
        authMode = 'setup';
        view = 'auth';
      } else if (!st.authed) {
        authMode = 'login';
        view = 'auth';
      } else {
        await enter();
      }
    } catch {
      view = 'auth';
      authMode = 'login';
    }
  }

  async function doLogout() {
    booted = false;
    disconnectEvents();
    await api.logout().catch(() => {});
    accounts = [];
    folders = [];
    messages = [];
    selected = null;
    page = 'mail';
    authMode = 'login';
    view = 'auth';
  }

  onMount(boot);
</script>

{#snippet sidebarContent()}
  <div class="flex h-full flex-col gap-2 overflow-y-auto p-2">
    <div class="border-b px-1 pb-2 text-lg font-bold">Webmail</div>
    <Button onclick={() => openCompose({})}><PenLine /> 写信</Button>
    <nav class="flex flex-col gap-1">
      <p class="px-2 pb-1 text-xs font-medium text-muted-foreground">聚合文件夹</p>
      {#each visibleScopes as s (s.key)}
        <Button
          variant={page === 'mail' && scope === s.key ? 'secondary' : 'ghost'}
          class="justify-start"
          onclick={() => selectScope(s.key)}
        >
          <s.icon />
          {s.label}
          {#if scopeStats[s.key].unread > 0}
            <Badge class="ml-auto">{scopeStats[s.key].unread}</Badge>
          {/if}
        </Button>
      {/each}
      {#if visibleScopes.length === 0}
        <p class="px-2 text-sm text-muted-foreground">暂无文件夹（同步后出现）</p>
      {/if}
    </nav>
    <Separator />
    <div class="flex flex-col gap-1">
      <p class="px-2 pb-1 text-xs font-medium text-muted-foreground">账号</p>
      {#if accounts.length === 0}
        <p class="px-2 text-sm text-muted-foreground">未添加账号</p>
      {:else}
        {#each accounts as a (a.id)}
          <div class="min-w-0 px-2 py-1">
            <div class="truncate text-sm font-medium">{a.label}</div>
            <div class="truncate text-xs text-muted-foreground">{a.email}</div>
            <div class="text-xs {statuses[a.id]?.state === 'error' ? 'font-medium text-destructive' : 'text-muted-foreground'}">
              {statusText(a)}
            </div>
            {#if statuses[a.id]?.cache_bodies_total > 0}
              <div class="text-xs text-muted-foreground">
                正文缓存 {statuses[a.id].cache_bodies_cached}/{statuses[a.id].cache_bodies_total}，附件 {statuses[a.id].cache_attachments_cached}/{statuses[a.id].cache_attachments_total}
              </div>
            {/if}
          </div>
        {/each}
      {/if}
      <Button variant={page === 'accounts' ? 'secondary' : 'ghost'} class="justify-start" onclick={openAccounts}>
        <Settings /> 账号管理
      </Button>
    </div>
    <div class="mt-auto flex flex-col gap-1 border-t pt-2">
      <div class="flex items-center gap-2 px-2 py-1.5">
        <Switch
          bind:checked={dark}
          onCheckedChange={(v) => {
            document.documentElement.classList.toggle('dark', v);
            localStorage.setItem('wm_dark', v ? '1' : '0');
          }}
          aria-label="深色模式"
        />
        <span class="text-sm">深色模式</span>
      </div>
      <Button variant="ghost" class="justify-start" onclick={doLogout}><LogOut /> 退出登录</Button>
    </div>
  </div>
{/snippet}

<svelte:window onkeydown={onGlobalKey} />

{#if view === 'loading'}
  <div class="flex min-h-dvh items-center justify-center">
    <p class="text-sm text-muted-foreground">加载中</p>
  </div>
{:else if view === 'auth'}
  <Auth mode={authMode} onauthed={enter} />
{:else}
  <div class="flex h-dvh w-full max-w-full overflow-hidden">
    <aside class="hidden w-64 shrink-0 border-r lg:block">
      {@render sidebarContent()}
    </aside>
    <Sheet bind:open={drawerOpen}>
      <SheetContent side="left" class="w-72 p-0" showCloseButton={false}>
        {@render sidebarContent()}
      </SheetContent>
    </Sheet>

    {#if page === 'accounts'}
      <main class="min-w-0 flex-1 overflow-y-auto">
        <div class="flex items-center gap-1 border-b px-2 py-2 lg:hidden">
          <Button variant="ghost" size="icon" aria-label="菜单" onclick={() => (drawerOpen = true)}><Menu /></Button>
          <span class="text-base font-bold">账号管理</span>
        </div>
        <Accounts {accounts} onchanged={refreshAccounts} />
      </main>
    {:else}
      <section
        class="min-w-0 flex-1 flex-col border-r lg:max-w-sm lg:flex-none lg:basis-[380px] {selected
          ? 'hidden lg:flex'
          : 'flex'}"
      >
        <div class="flex items-center gap-1 border-b px-2 py-2">
          <Button variant="ghost" size="icon" class="lg:hidden" aria-label="菜单" onclick={() => (drawerOpen = true)}>
            <Menu />
          </Button>
          <h1 class="truncate px-1 text-base font-bold">{searching ? '搜索结果' : scopeLabel}</h1>
          <span class="shrink-0 whitespace-nowrap text-xs text-muted-foreground">
            {#if searching}
              {messages.length} 封
            {:else}
              {scopeStats[scope]?.total ?? 0} 封{#if scopeStats[scope]?.unread}（{scopeStats[scope].unread} 未读）{/if}
            {/if}
          </span>
          <span class="flex-1"></span>
          <label class="flex shrink-0 items-center gap-1 px-1 text-xs whitespace-nowrap">
            <Switch
              bind:checked={threaded}
              onCheckedChange={(v) => {
                localStorage.setItem('wm_threaded', v ? '1' : '0');
                if (!searching) loadMessages();
              }}
              aria-label="按线程折叠"
            />
            按线程
          </label>
          <Button variant="ghost" size="icon" aria-label="刷新" onclick={refreshData}><RefreshCw /></Button>
        </div>
        <div class="flex items-center gap-1 border-b px-2 py-2">
          <div class="relative min-w-0 flex-1">
            <Search
              class="pointer-events-none absolute top-1/2 left-2 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              id="wm-search"
              class="pl-8"
              placeholder="搜索主题、发件人、正文…"
              bind:value={searchQuery}
              oninput={onSearchInput}
            />
          </div>
          {#if searching}
            <Button variant="ghost" size="icon" aria-label="清除搜索" onclick={clearSearch}><X /></Button>
          {/if}
        </div>
        <div class="flex-1 overflow-y-auto" bind:this={listEl} onscroll={onListScroll}>
          {#if messages.length === 0}
            <p class="p-6 text-sm text-muted-foreground">
              {listLoading ? '加载中' : '这里还没有邮件。添加账号后会自动同步进来。'}
            </p>
          {:else}
            {#each messages as m, idx (m.id)}
              <button
                data-rowidx={idx}
                class="flex w-full max-w-full flex-col gap-0.5 overflow-hidden border-b px-2 py-1.5 text-left hover:bg-accent {selected?.id === m.id || idx === cursorIdx
                  ? 'bg-accent'
                  : ''}"
                onclick={() => rowClick(m, idx)}
              >
                <span class="flex w-full items-center gap-2">
                  {#if !(m.flags & 1)}<span class="size-2 shrink-0 bg-foreground"></span>{/if}
                  <span class="min-w-0 flex-1 truncate text-sm {!(m.flags & 1) ? 'font-semibold' : ''}">
                    {m.from_name || m.from_email || '(未知发件人)'}
                  </span>
                  {#if threaded && !searching && (m.thread_count ?? 1) > 1}
                    <span class="shrink-0 text-xs text-muted-foreground">
                      共 {m.thread_count} 封{#if m.thread_unread}，{m.thread_unread} 未读{/if}
                    </span>
                  {/if}
                  {#if m.has_attach}<Paperclip class="size-3.5 shrink-0 text-muted-foreground" />{/if}
                  <span class="shrink-0 text-xs text-muted-foreground">{fmtListTime(m.date_ts)}</span>
                </span>
                <span class="w-full truncate text-sm {!(m.flags & 1) ? 'font-semibold' : ''}">
                  {m.subject || '(无主题)'}
                </span>
                <span class="w-full truncate text-xs text-muted-foreground">
                  {m.account_label}{searching ? '，' + m.folder_name : ''}{m.snippet ? '，' + m.snippet : ''}
                </span>
              </button>
              {#if threaded && !searching && expandedKey === m.account_id + ':' + (m.thread_key || 'id:' + m.id)}
                {#each threadMsgs as sub (sub.id)}
                  <button
                    class="flex w-full max-w-full flex-col gap-0.5 overflow-hidden border-b bg-secondary/40 py-1.5 pr-2 pl-6 text-left hover:bg-accent {selected?.id === sub.id
                      ? 'bg-accent'
                      : ''}"
                    onclick={() => openMessage(sub)}
                  >
                    <span class="flex w-full items-center gap-2">
                      {#if !(sub.flags & 1)}<span class="size-2 shrink-0 bg-foreground"></span>{/if}
                      <span class="min-w-0 flex-1 truncate text-sm {!(sub.flags & 1) ? 'font-semibold' : ''}">
                        {sub.from_name || sub.from_email || '(未知发件人)'}
                      </span>
                      <span class="shrink-0 text-xs text-muted-foreground">{fmtListTime(sub.date_ts)}</span>
                    </span>
                    <span class="w-full truncate text-sm {!(sub.flags & 1) ? 'font-semibold' : ''}">
                      {sub.subject || '(无主题)'}
                    </span>
                  </button>
                {/each}
              {/if}
            {/each}
            {#if listHasMore && !searching}
              <button
                class="w-full py-2 text-center text-sm underline"
                disabled={loadingMore}
                onclick={loadMore}
              >
                {loadingMore ? '加载中' : '加载更多'}
              </button>
            {/if}
          {/if}
        </div>
      </section>

      <section class="min-w-0 flex-1 flex-col {selected ? 'flex' : 'hidden lg:flex'}">
        {#if !selected}
          <div class="flex h-full items-center justify-center">
            <p class="text-sm text-muted-foreground">未选择邮件</p>
          </div>
        {:else}
          <div class="min-w-0 flex-1 overflow-x-hidden overflow-y-auto p-3 sm:p-4">
            <Button
              variant="ghost"
              class="mb-3 -ml-2 lg:hidden"
              onclick={() => {
                selected = null;
                detail = null;
              }}
            >
              <ChevronLeft /> 返回列表
            </Button>
            <h2 class="mb-2 break-words text-lg font-bold">{selected.subject || '(无主题)'}</h2>
            <dl class="mb-3 grid grid-cols-[52px_minmax(0,1fr)] gap-y-1 text-sm">
              <dt class="text-muted-foreground">发件人</dt>
              <dd class="min-w-0 break-words">
                {detail?.from
                  ? detail.from.name
                    ? detail.from.name + ' <' + detail.from.email + '>'
                    : detail.from.email
                  : selected.from_name
                    ? selected.from_name + ' <' + selected.from_email + '>'
                    : selected.from_email}
              </dd>
              {#if detail?.to?.length}
                <dt class="text-muted-foreground">收件人</dt>
                <dd class="min-w-0 break-words">{detail.to.map((a) => (a.name ? a.name + ' <' + a.email + '>' : a.email)).join(', ')}</dd>
              {/if}
              {#if detail?.cc?.length}
                <dt class="text-muted-foreground">抄送</dt>
                <dd class="min-w-0 break-words">{detail.cc.map((a) => (a.name ? a.name + ' <' + a.email + '>' : a.email)).join(', ')}</dd>
              {/if}
              <dt class="text-muted-foreground">时间</dt>
              <dd class="min-w-0">{fmtFull(selected.date_ts)}</dd>
              <dt class="text-muted-foreground">账号</dt>
              <dd class="min-w-0 break-words">{selected.account_label}，{selected.folder_name}</dd>
            </dl>

            <div class="flex flex-wrap items-center gap-2 border-y py-2">
              <Button variant="outline" size="sm" disabled={opBusy} onclick={toggleSeen}>
                {#if selected.flags & 1}<Mail /> 标为未读{:else}<MailOpen /> 标为已读{/if}
              </Button>
              <Button variant="outline" size="sm" disabled={opBusy || !detail} onclick={replySelected}>
                <Reply /> 回复
              </Button>
              <Button variant="outline" size="sm" disabled={opBusy || !detail} onclick={forwardSelected}>
                <Forward /> 转发
              </Button>
              <Button variant="outline" size="sm" disabled={opBusy} onclick={toggleStar}>
                <Star class={selected.flags & 2 ? 'fill-foreground' : ''} />
                {selected.flags & 2 ? '取消星标' : '星标'}
              </Button>
              {#if scope !== 'archive' && scope !== 'trash'}
                <Button variant="outline" size="sm" disabled={opBusy} onclick={archiveSelected}>
                  <Archive /> 归档
                </Button>
              {/if}
              <Button variant="outline" size="sm" disabled={opBusy} onclick={deleteSelected}>
                <Trash2 /> {scope === 'trash' || scope === 'junk' ? '彻底删除' : '删除'}
              </Button>
            </div>

            {#if detailLoading}
              <p class="py-8 text-sm text-muted-foreground">正在从服务器取正文…</p>
            {:else if detailError}
              <p class="py-8 text-sm font-medium text-destructive">正文加载失败：{detailError}</p>
            {:else if detail}
              {#if detail.attachments.length > 0}
                <div class="flex flex-wrap gap-2 py-4">
                  {#each detail.attachments as a (a.id)}
                    <Button variant="outline" size="sm" href={api.attachmentURL(detail.id, a.id)}>
                      <Paperclip class="shrink-0" />
                      <span class="max-w-[55vw] truncate sm:max-w-xs">{a.filename}</span>
                      <span class="shrink-0 text-muted-foreground">{fmtSize(a.size)}</span>
                    </Button>
                  {/each}
                </div>
              {/if}
              {#if detail.html}
                <div class="overflow-hidden">
                  <iframe
                    title="邮件正文"
                    class="mt-2 block w-full border-0"
                    style="height: 480px"
                    sandbox="allow-same-origin"
                    srcdoc={FRAME_CSS + detail.html}
                    bind:this={frameEl}
                    onload={frameLoaded}
                  ></iframe>
                </div>
              {:else if detail.text}
                <div class="whitespace-pre-wrap py-4 text-sm leading-relaxed">{detail.text}</div>
              {:else}
                <p class="py-8 text-sm text-muted-foreground">这封邮件没有可显示的正文。</p>
              {/if}
            {/if}
          </div>
        {/if}
      </section>

      {#if !selected}
        <Button
          class="fixed right-4 bottom-4 z-40 size-11 lg:hidden"
          size="icon"
          aria-label="写信"
          onclick={() => openCompose({})}
        >
          <PenLine class="size-5" />
        </Button>
      {/if}
    {/if}
    <Dialog bind:open={helpOpen}>
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>键盘快捷键</DialogTitle>
        </DialogHeader>
        <table class="w-full text-sm">
          <tbody>
            <tr class="border-b"><td class="py-1 pr-4 font-mono">j / k</td><td class="py-1">下移 / 上移</td></tr>
            <tr class="border-b"><td class="py-1 pr-4 font-mono">Enter</td><td class="py-1">打开邮件（线程则展开）</td></tr>
            <tr class="border-b"><td class="py-1 pr-4 font-mono">e</td><td class="py-1">归档</td></tr>
            <tr class="border-b"><td class="py-1 pr-4 font-mono">Delete</td><td class="py-1">删除</td></tr>
            <tr class="border-b"><td class="py-1 pr-4 font-mono">s</td><td class="py-1">星标 / 取消星标</td></tr>
            <tr class="border-b"><td class="py-1 pr-4 font-mono">r</td><td class="py-1">回复</td></tr>
            <tr class="border-b"><td class="py-1 pr-4 font-mono">a</td><td class="py-1">当前文件夹全部标为已读</td></tr>
            <tr class="border-b"><td class="py-1 pr-4 font-mono">c</td><td class="py-1">写新邮件</td></tr>
            <tr class="border-b"><td class="py-1 pr-4 font-mono">/</td><td class="py-1">聚焦搜索框</td></tr>
            <tr><td class="py-1 pr-4 font-mono">?</td><td class="py-1">本帮助</td></tr>
          </tbody>
        </table>
      </DialogContent>
    </Dialog>
    <Compose bind:open={composeOpen} {accounts} initial={composeInit} ondone={refreshData} />
  </div>
{/if}
