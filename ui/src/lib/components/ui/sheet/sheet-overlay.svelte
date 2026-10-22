<script>
  import { Dialog as SheetPrimitive } from 'bits-ui';
  import { cn } from '$lib/utils.js';

  let { ref = $bindable(null), class: className, ...restProps } = $props();
</script>

<SheetPrimitive.Overlay
  bind:ref
  class={cn(
    'fixed inset-0 z-50 bg-black/40',
    className
  )}
  {...restProps}
/>
