<script>
  import { onMount } from 'svelte';
  import { Plus, Pencil, Trash2, PlugZap } from 'lucide-svelte';
  import * as api from './api.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Switch } from '$lib/components/ui/switch/index.js';
  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table/index.js';
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle
  } from '$lib/components/ui/dialog/index.js';
  import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle
  } from '$lib/components/ui/alert-dialog/index.js';

  let { accounts = [], onchanged } = $props();

  const emptyForm = () => ({
    label: '',
    email: '',
    imap_host: '',
    imap_port: 993,
    imap_tls: 'tls',
    smtp_host: '',
    smtp_port: 465,
    smtp_tls: 'tls',
    username: '',
    password: ''
  });

  const selectClass =
    'flex h-8 w-full min-w-0 border border-input bg-background px-2 py-1 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50';

  let dialogOpen = $state(false);
  let editing = $state(null);
  let form = $state(emptyForm());
  let formError = $state('');
  let deleteTarget = $state(null);
  let alertOpen = $state(false);
  let testState = $state({});

  let tg = $state({ enabled: false, chat_id: '', bot_token: '', has_token: false });
  let tgChats = $state([]);
  let tgMsg = $state('');
  let tgBusy = $state('');

  async function loadTG() {
    try {
      const s = await api.getTelegram();
      tg.enabled = !!s.enabled;
      tg.chat_id = s.chat_id || '';
      tg.has_token = !!s.has_token;
      tg.bot_token = '';
    } catch {
    }
  }
  onMount(loadTG);

  async function saveTG() {
    tgBusy = 'save';
    tgMsg = '';
    try {
      const s = await api.saveTelegram({
        enabled: tg.enabled,
        chat_id: tg.chat_id,
        bot_token: tg.bot_token
      });
      tg.has_token = !!s.has_token;
      tg.bot_token = '';
      tgMsg = '已保存。';
    } catch (err) {
      tgMsg = '保存失败：' + err.message;
    } finally {
      tgBusy = '';
    }
  }

  async function discoverTG() {
    tgBusy = 'discover';
    tgMsg = '';
    tgChats = [];
    try {
      const res = await api.discoverTelegram(tg.bot_token);
      if (res.chats.length === 0) {
        tgMsg = '暂无记录。';
      } else if (res.chats.length === 1) {
        tg.chat_id = res.chats[0].id;
        tgMsg = '已填入 Chat ID（' + (res.chats[0].title || res.chats[0].type) + '）。';
      } else {
        tgChats = res.chats;
        tgMsg = '找到多个会话，点选一个填入。';
      }
    } catch (err) {
      tgMsg = err.message;
    } finally {
      tgBusy = '';
    }
  }

  async function testTG() {
    tgBusy = 'test';
    tgMsg = '';
    try {
      await api.testTelegram({ bot_token: tg.bot_token, chat_id: tg.chat_id });
      tgMsg = '测试消息已发送，去 Telegram 看一下。';
    } catch (err) {
      tgMsg = '发送失败：' + err.message;
    } finally {
      tgBusy = '';
    }
  }

  function openCreate() {
    editing = null;
    form = emptyForm();
    formError = '';
    dialogOpen = true;
  }

  function openEdit(a) {
    editing = a.id;
    form = { ...a, password: '' };
    formError = '';
    dialogOpen = true;
  }

  async function save(e) {
    e.preventDefault();
    formError = '';
    const payload = { ...form, imap_port: Number(form.imap_port), smtp_port: Number(form.smtp_port) };
    try {
      if (editing) await api.updateAccount(editing, payload);
      else await api.createAccount(payload);
      dialogOpen = false;
      onchanged();
    } catch (err) {
      formError = err.message;
    }
  }

  function askDelete(a) {
    deleteTarget = a;
    alertOpen = true;
  }

  async function confirmDelete() {
    if (!deleteTarget) return;
    await api.deleteAccount(deleteTarget.id);
    deleteTarget = null;
    onchanged();
  }

  async function runTest(a) {
    testState[a.id] = { busy: true, ok: false, text: '连接中' };
    try {
      const res = await api.testAccount(a.id);
      testState[a.id] = {
        busy: false,
        ok: true,
        text: '连接成功：' + res.folders.length + ' 个文件夹，能力 ' + res.capabilities.length + ' 项（含 ' +
          res.capabilities.filter((c) => ['IDLE', 'CONDSTORE', 'QRESYNC', 'MOVE', 'SPECIAL-USE'].includes(c)).join('/') + '）'
      };
    } catch (err) {
      testState[a.id] = { busy: false, ok: false, text: '连接失败：' + err.message };
    }
  }
</script>

<div class="mx-auto w-full min-w-0 max-w-3xl p-3 sm:p-4">
  <div class="mb-3 flex items-center justify-between">
    <h2 class="text-lg font-bold">邮箱账号</h2>
    <Button onclick={openCreate}><Plus /> 添加账号</Button>
  </div>

  {#if accounts.length === 0}
    <p class="text-sm text-muted-foreground">还没有账号。添加第一个邮箱后，它的邮件会自动同步进聚合收件箱。</p>
  {:else}
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>名称</TableHead>
          <TableHead>邮箱</TableHead>
          <TableHead>IMAP</TableHead>
          <TableHead>SMTP</TableHead>
          <TableHead class="w-[130px]">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {#each accounts as a (a.id)}
          <TableRow>
            <TableCell class="font-medium">{a.label}</TableCell>
            <TableCell>{a.email}</TableCell>
            <TableCell class="text-muted-foreground">{a.imap_host}:{a.imap_port} ({a.imap_tls})</TableCell>
            <TableCell class="text-muted-foreground">{a.smtp_host}:{a.smtp_port} ({a.smtp_tls})</TableCell>
            <TableCell>
              <div class="flex gap-1">
                <Button variant="ghost" size="icon" title="测试连接" onclick={() => runTest(a)}><PlugZap /></Button>
                <Button variant="ghost" size="icon" title="编辑" onclick={() => openEdit(a)}><Pencil /></Button>
                <Button variant="ghost" size="icon" title="删除" onclick={() => askDelete(a)}><Trash2 /></Button>
              </div>
            </TableCell>
          </TableRow>
          {#if testState[a.id]}
            <TableRow>
              <TableCell
                colspan="5"
                class={(testState[a.id].ok ? 'text-sm' : 'text-sm font-medium text-destructive') +
                  ' max-w-[80vw] break-words whitespace-normal'}
              >
                {testState[a.id].text}
              </TableCell>
            </TableRow>
          {/if}
        {/each}
      </TableBody>
    </Table>
  {/if}

  <div class="mt-6 border-t pt-4">
    <h2 class="mb-1 text-lg font-bold">Telegram 推送</h2>
    <p class="mb-3 text-sm text-muted-foreground">
      新邮件到达收件箱时，通过 Telegram 机器人推送提醒。用 @BotFather 建机器人拿到 Token，填进来保存后，在 Telegram 里给机器人发
      /id，它会回复你的 Chat ID，把它填到下面保存即完成绑定。发过消息后也可以点自动获取。
    </p>
    <div class="grid max-w-xl gap-3">
      <div class="flex items-center gap-2">
        <Switch id="tg-enabled" bind:checked={tg.enabled} />
        <Label for="tg-enabled">启用新邮件推送</Label>
      </div>
      <div class="grid gap-1">
        <Label for="tg-token">Bot Token</Label>
        <Input
          id="tg-token"
          type="password"
          autocomplete="off"
          placeholder={tg.has_token ? '已保存，留空不改' : '如 123456:ABC-DEF...'}
          bind:value={tg.bot_token}
        />
      </div>
      <div class="grid gap-1">
        <Label for="tg-chat">Chat ID</Label>
        <div class="flex gap-2">
          <Input id="tg-chat" autocomplete="off" placeholder="如 123456789" bind:value={tg.chat_id} />
          <Button variant="outline" disabled={tgBusy !== ''} onclick={discoverTG}>自动获取</Button>
        </div>
        {#if tgChats.length > 0}
          <div class="flex flex-wrap gap-1">
            {#each tgChats as c (c.id)}
              <Button variant="outline" size="sm" onclick={() => (tg.chat_id = c.id)}>
                {c.title || c.type} ({c.id})
              </Button>
            {/each}
          </div>
        {/if}
      </div>
      <div class="flex items-center gap-2">
        <Button disabled={tgBusy !== ''} onclick={saveTG}>保存</Button>
        <Button variant="outline" disabled={tgBusy !== ''} onclick={testTG}>发送测试消息</Button>
        {#if tgMsg}<span class="text-sm">{tgMsg}</span>{/if}
      </div>
    </div>
  </div>
</div>

<Dialog bind:open={dialogOpen}>
  <DialogContent class="sm:max-w-lg">
    <form onsubmit={save}>
      <DialogHeader>
        <DialogTitle>{editing ? '编辑账号' : '添加账号'}</DialogTitle>
      </DialogHeader>
      <div class="grid gap-3 py-3">
        <div class="grid gap-2">
          <Label for="f-label">名称</Label>
          <Input id="f-label" bind:value={form.label} placeholder="如 工作邮箱" required />
        </div>
        <div class="grid gap-2">
          <Label for="f-email">邮箱</Label>
          <Input id="f-email" type="email" bind:value={form.email} required />
        </div>
        <div class="grid gap-2">
          <Label for="f-user">用户名</Label>
          <Input id="f-user" bind:value={form.username} required />
        </div>
        <div class="grid gap-2">
          <Label for="f-pass">密码</Label>
          <Input
            id="f-pass"
            type="password"
            bind:value={form.password}
            placeholder={editing ? '留空 = 不修改' : ''}
            required={!editing}
          />
        </div>
        <div class="grid gap-2">
          <Label for="f-ihost">IMAP 服务器</Label>
          <div class="flex flex-wrap gap-2">
            <Input id="f-ihost" class="min-w-0 flex-1 basis-36" bind:value={form.imap_host} placeholder="imap.example.com" required />
            <Input class="w-20" type="number" bind:value={form.imap_port} required />
            <select class={selectClass + ' w-28 flex-none'} bind:value={form.imap_tls}>
              <option value="tls">TLS</option>
              <option value="starttls">STARTTLS</option>
              <option value="none">明文</option>
            </select>
          </div>
        </div>
        <div class="grid gap-2">
          <Label for="f-shost">SMTP 服务器</Label>
          <div class="flex flex-wrap gap-2">
            <Input id="f-shost" class="min-w-0 flex-1 basis-36" bind:value={form.smtp_host} placeholder="smtp.example.com" required />
            <Input class="w-20" type="number" bind:value={form.smtp_port} required />
            <select class={selectClass + ' w-28 flex-none'} bind:value={form.smtp_tls}>
              <option value="tls">TLS</option>
              <option value="starttls">STARTTLS</option>
              <option value="none">明文</option>
            </select>
          </div>
        </div>
        {#if formError}<p class="text-sm font-medium text-destructive">{formError}</p>{/if}
      </div>
      <DialogFooter>
        <Button type="button" variant="outline" onclick={() => (dialogOpen = false)}>取消</Button>
        <Button type="submit">保存</Button>
      </DialogFooter>
    </form>
  </DialogContent>
</Dialog>

<AlertDialog bind:open={alertOpen} onOpenChange={(o) => { if (!o) deleteTarget = null; }}>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>删除账号</AlertDialogTitle>
      <AlertDialogDescription>
        删除账号「{deleteTarget?.label}」？本地同步的邮件会一并删除，服务器上的邮件不受影响。
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <AlertDialogCancel>取消</AlertDialogCancel>
      <AlertDialogAction onclick={confirmDelete}>删除</AlertDialogAction>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
