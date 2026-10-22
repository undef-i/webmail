<script>
  import { Dialog as DialogPrimitive } from 'bits-ui';
  import X from 'lucide-svelte/icons/x';
  import { cn } from '$lib/utils.js';
  import DialogPortal from './dialog-portal.svelte';
  import DialogOverlay from './dialog-overlay.svelte';

  let {
    ref = $bindable(null),
    class: className,
    portalProps,
    children,
    showCloseButton = true,
    ...restProps
  } = $props();
</script>

<DialogPortal {...portalProps}>
  <DialogOverlay />
  <DialogPrimitive.Content
    bind:ref
    class={cn(
      'fixed left-[50%] top-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] gap-3 border-2 border-foreground bg-background p-4 sm:max-w-lg',
      className
    )}
    {...restProps}
  >
    {@render children?.()}
    {#if showCloseButton}
      <DialogPrimitive.Close
        class="absolute right-3 top-3 border border-input bg-secondary px-1 hover:bg-accent focus:outline-none focus:ring-1 focus:ring-ring disabled:pointer-events-none"
      >
        <X class="size-4" />
        <span class="sr-only">Close</span>
      </DialogPrimitive.Close>
    {/if}
  </DialogPrimitive.Content>
</DialogPortal>
