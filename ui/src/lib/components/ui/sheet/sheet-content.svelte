<script module>
  import { tv } from 'tailwind-variants';

  export const sheetVariants = tv({
    base: 'fixed z-50 gap-4 bg-background p-0',
    variants: {
      side: {
        top: 'inset-x-0 top-0 border-b',
        bottom: 'inset-x-0 bottom-0 border-t',
        left: 'inset-y-0 left-0 h-full w-3/4 border-r sm:max-w-sm',
        right: 'inset-y-0 right-0 h-full w-3/4 border-l sm:max-w-sm'
      }
    },
    defaultVariants: {
      side: 'right'
    }
  });
</script>

<script>
  import { Dialog as SheetPrimitive } from 'bits-ui';
  import X from 'lucide-svelte/icons/x';
  import { cn } from '$lib/utils.js';
  import SheetPortal from './sheet-portal.svelte';
  import SheetOverlay from './sheet-overlay.svelte';

  let {
    ref = $bindable(null),
    class: className,
    portalProps,
    side = 'right',
    children,
    showCloseButton = true,
    ...restProps
  } = $props();
</script>

<SheetPortal {...portalProps}>
  <SheetOverlay />
  <SheetPrimitive.Content bind:ref class={cn(sheetVariants({ side }), className)} {...restProps}>
    {@render children?.()}
    {#if showCloseButton}
      <SheetPrimitive.Close
        class="absolute right-3 top-3 border border-input bg-secondary px-1 hover:bg-accent focus:outline-none focus:ring-1 focus:ring-ring disabled:pointer-events-none"
      >
        <X class="size-4" />
        <span class="sr-only">Close</span>
      </SheetPrimitive.Close>
    {/if}
  </SheetPrimitive.Content>
</SheetPortal>
