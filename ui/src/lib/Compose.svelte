<script>
  import { untrack } from 'svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle
  } from '$lib/components/ui/dialog/index.js';
  import * as api from './api.js';

  const MAX_ATTACH_TOTAL = 20 * 1024 * 1024;

  let { open = $bindable(false), accounts = [], initial = {}, ondone } = $props();

  let accountId = $state('');
  let to = $state('');
  let cc = $state('');
  let bcc = $state('');
  let showCc = $state(false);
  let subject = $state('');
  let text = $state('');
  let attachments = $state([]);
  let draftId = $state(0);
  let busy = $state(false);
  let sending = $state(false);
  let error = $state('');
  let attachError = $state('');
  let warning = $state('');
  let done = $state(false);
  let savedAt = $state('');
  let fileInput = $state(null);
  let saveTimer = null;
  let lastSavedJSON = '';

  $effect(() => {
    if (open) {
      accountId = String(initial.account_id ?? accounts[0]?.id ?? '');
      to = initial.to ?? '';
      cc = initial.cc ?? '';
      bcc = '';
      showCc = !!cc;
      subject = initial.subject ?? '';
      text = initial.text ?? '';
      attachments = [];
      draftId = initial.draft_id ?? 0;
      error = '';
      attachError = '';
      warning = '';
      done = false;
      busy = false;
      sending = false;
      savedAt = '';
      lastSavedJSON = untrack(() => snapshot());
    } else {
      clearTimeout(saveTimer);
    }
  });

  const title = $derived(
    initial.draft_id ? '编辑草稿' : initial.reply_to_id ? '回复邮件' : '写信'
  );

  function snapshot() {
    return JSON.stringify({
      account_id: Number(accountId),
      to,
      cc,
      bcc,
      subject,
      text,
      atts: attachments.map((a) => a.filename + ':' + a.size)
    });
  }

  function payload() {
    return {
      account_id: Number(accountId),
      to,
      cc,
      bcc,
      subject,
      text,
      reply_to_id: initial.reply_to_id ?? 0,
      draft_id: draftId,
      attachments: attachments.map((a) => ({
        filename: a.filename,
        content_type: a.content_type,
        data_b64: a.data_b64
      }))
    };
  }

  function fmtSize(n) {
    if (n < 1024) return n + ' B';
    if (n < 1024 * 1024) return (n / 1024).toFixed(1) + ' KB';
    return (n / 1024 / 1024).toFixed(1) + ' MB';
  }

  function fmtClock(d) {
    const pad = (n) => String(n).padStart(2, '0');
    return pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds());
  }

  async function onFilesChosen(e) {
    attachError = '';
    const files = [...(e.target.files ?? [])];
    e.target.value = '';
    let total = attachments.reduce((s, a) => s + a.size, 0);
    for (const f of files) {
      if (total + f.size > MAX_ATTACH_TOTAL) {
        attachError = '附件总大小超过 20 MB 上限，已跳过 ' + f.name;
        continue;
      }
      const dataUrl = await new Promise((resolve, reject) => {
        const rd = new FileReader();
        rd.onload = () => resolve(rd.result);
        rd.onerror = reject;
        rd.readAsDataURL(f);
      });
      attachments = [
        ...attachments,
        {
          filename: f.name,
          size: f.size,
          content_type: f.type || 'application/octet-stream',
          data_b64: String(dataUrl).split(',')[1] ?? ''
        }
      ];
      total += f.size;
    }
  }

  function removeAttachment(i) {
    attachments = attachments.filter((_, j) => j !== i);
  }

  async function autosave() {
    if (!open || done || busy || sending) return;
    if (!accountId) return;
    if (!to && !subject && !text && attachments.length === 0) return;
    const snap = snapshot();
    if (snap === lastSavedJSON) return;
    busy = true;
    try {
      const res = await api.saveDraft(payload());
      draftId = res.id;
      lastSavedJSON = snapshot();
      savedAt = fmtClock(new Date());
      if (res.warning) warning = res.warning;
    } catch {
    } finally {
      busy = false;
    }
  }

  $effect(() => {
    void [accountId, to, cc, bcc, subject, text, attachments];
    if (!open) return;
    clearTimeout(saveTimer);
    saveTimer = setTimeout(autosave, 2000);
  });

  async function doSend() {
    sending = true;
    busy = true;
    error = '';
    clearTimeout(saveTimer);
    try {
      const res = await api.sendMail(payload());
      if (res.warning) {
        warning = res.warning;
        done = true;
      } else {
        open = false;
        ondone?.();
      }
    } catch (e) {
      error = e.message;
    } finally {
      busy = false;
      sending = false;
    }
  }

  async function doSaveDraft() {
    busy = true;
    error = '';
    clearTimeout(saveTimer);
    try {
      const res = await api.saveDraft(payload());
      draftId = res.id;
      lastSavedJSON = snapshot();
      if (res.warning) {
        warning = res.warning;
        done = true;
      } else {
        open = false;
        ondone?.();
      }
    } catch (e) {
      error = e.message;
    } finally {
      busy = false;
    }
  }

  function closeAfterDone() {
    open = false;
    ondone?.();
  }
</script>

<Dialog bind:open>
  <DialogContent class="max-h-[92dvh] overflow-y-auto sm:max-w-lg">
    <DialogHeader>
      <DialogTitle>{title}</DialogTitle>
    </DialogHeader>
    {#if done}
      <p class="py-2 text-sm break-words">{warning}</p>
      <DialogFooter>
        <Button onclick={closeAfterDone}>关闭</Button>
      </DialogFooter>
    {:else}
      <div class="flex flex-col gap-3">
        <div class="flex flex-col gap-1.5">
          <Label for="c-acct">发信账号</Label>
          <select
            id="c-acct"
            bind:value={accountId}
            class="h-8 min-w-0 border border-input bg-background px-2 text-sm"
          >
            {#each accounts as a (a.id)}
              <option value={String(a.id)}>{a.label}（{a.email}）</option>
            {/each}
          </select>
        </div>
        <div class="flex flex-col gap-1.5">
          <Label for="c-to">收件人</Label>
          <Input id="c-to" bind:value={to} placeholder="name@example.com，多人用逗号分隔" />
        </div>
        {#if showCc}
          <div class="flex flex-col gap-1.5">
            <Label for="c-cc">抄送</Label>
            <Input id="c-cc" bind:value={cc} />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label for="c-bcc">密送</Label>
            <Input id="c-bcc" bind:value={bcc} />
          </div>
        {:else}
          <button
            type="button"
            class="self-start text-xs text-muted-foreground underline"
            onclick={() => (showCc = true)}
          >
            添加抄送 / 密送
          </button>
        {/if}
        <div class="flex flex-col gap-1.5">
          <Label for="c-subj">主题</Label>
          <Input id="c-subj" bind:value={subject} />
        </div>
        <div class="flex flex-col gap-1.5">
          <Label for="c-text">正文</Label>
          <textarea
            id="c-text"
            bind:value={text}
            rows="10"
            class="min-w-0 border border-input bg-background px-2 py-1.5 text-sm leading-normal"
          ></textarea>
        </div>
        <div class="flex flex-col gap-1.5">
          <div class="flex items-center gap-2">
            <Button variant="outline" size="sm" disabled={busy} onclick={() => fileInput?.click()}>
              添加附件
            </Button>
            <input
              bind:this={fileInput}
              type="file"
              multiple
              class="hidden"
              onchange={onFilesChosen}
            />
            {#if savedAt}
              <span class="text-xs text-muted-foreground">草稿已自动保存 {savedAt}</span>
            {/if}
          </div>
          {#if attachments.length > 0}
            <ul class="flex flex-col gap-1">
              {#each attachments as a, i (a.filename + ':' + i)}
                <li class="flex items-center gap-2 border px-2 py-1 text-sm">
                  <span class="min-w-0 flex-1 truncate">{a.filename}</span>
                  <span class="shrink-0 text-xs text-muted-foreground">{fmtSize(a.size)}</span>
                  <button
                    type="button"
                    class="shrink-0 text-xs underline"
                    disabled={busy}
                    onclick={() => removeAttachment(i)}
                  >
                    移除
                  </button>
                </li>
              {/each}
            </ul>
          {/if}
          {#if attachError}
            <p class="text-sm font-medium break-words text-destructive">{attachError}</p>
          {/if}
        </div>
        {#if error}
          <p class="text-sm font-medium break-words text-destructive">{error}</p>
        {/if}
      </div>
      <DialogFooter>
        <Button variant="ghost" disabled={busy} onclick={() => (open = false)}>取消</Button>
        <Button variant="outline" disabled={busy} onclick={doSaveDraft}>存草稿</Button>
        <Button disabled={busy} onclick={doSend}>{sending ? '发送中…' : '发送'}</Button>
      </DialogFooter>
    {/if}
  </DialogContent>
</Dialog>
