<script>
  import { AlertDialog as AlertDialogPrimitive } from 'bits-ui';
  import { cn } from '$lib/utils.js';
  import AlertDialogPortal from './alert-dialog-portal.svelte';
  import AlertDialogOverlay from './alert-dialog-overlay.svelte';

  let { ref = $bindable(null), class: className, portalProps, children, ...restProps } = $props();
</script>

<AlertDialogPortal {...portalProps}>
  <AlertDialogOverlay />
  <AlertDialogPrimitive.Content
    bind:ref
    class={cn(
      'fixed left-[50%] top-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] gap-3 border-2 border-foreground bg-background p-4 sm:max-w-lg',
      className
    )}
    {...restProps}
  >
    {@render children?.()}
  </AlertDialogPrimitive.Content>
</AlertDialogPortal>
