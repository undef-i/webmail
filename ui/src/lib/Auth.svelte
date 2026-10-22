<script>
  import * as api from './api.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

  let { mode = 'login', onauthed } = $props();
  let password = $state('');
  let confirm = $state('');
  let error = $state('');
  let busy = $state(false);

  async function submit(e) {
    e.preventDefault();
    error = '';
    if (mode === 'setup' && password !== confirm) {
      error = '两次输入的密码不一致';
      return;
    }
    busy = true;
    try {
      if (mode === 'setup') await api.setup(password);
      else await api.login(password);
      onauthed();
    } catch (err) {
      error = err.message;
    } finally {
      busy = false;
    }
  }
</script>

<div class="flex min-h-dvh items-center justify-center p-4">
  <Card class="w-full max-w-sm">
    <CardHeader>
      <CardTitle class="text-xl">Webmail</CardTitle>
      <CardDescription>{mode === 'setup' ? '首次使用：设置主密码' : '输入主密码登录'}</CardDescription>
    </CardHeader>
    <CardContent>
      <form onsubmit={submit} class="flex flex-col gap-3">
        <div class="flex flex-col gap-1.5">
          <Label for="pw">主密码</Label>
          <Input
            id="pw"
            type="password"
            bind:value={password}
            autocomplete={mode === 'setup' ? 'new-password' : 'current-password'}
            required
          />
        </div>
        {#if mode === 'setup'}
          <div class="flex flex-col gap-2">
            <Label for="pw2">再次输入</Label>
            <Input id="pw2" type="password" bind:value={confirm} autocomplete="new-password" required />
          </div>
        {/if}
        {#if error}<p class="text-sm font-medium text-destructive">{error}</p>{/if}
        <Button type="submit" class="w-full" disabled={busy}>
          {mode === 'setup' ? '设置并进入' : '登录'}
        </Button>
        {#if mode === 'setup'}
          <p class="text-xs text-muted-foreground">
            主密码用于解锁所有邮箱账号的密码，服务器重启后需要重新输入。遗忘后无法恢复，只能重置数据。
          </p>
        {/if}
      </form>
    </CardContent>
  </Card>
</div>
