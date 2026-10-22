<script>
  import { Dialog as SheetPrimitive } from 'bits-ui';
  import { cn } from '$lib/utils.js';

  let { ref = $bindable(null), class: className, children, ...restProps } = $props();
</script>

<SheetPrimitive.Description bind:ref class={cn('text-sm text-muted-foreground', className)} {...restProps}>
  {@render children?.()}
</SheetPrimitive.Description>
