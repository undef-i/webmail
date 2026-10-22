<script>
  import { AlertDialog as AlertDialogPrimitive } from 'bits-ui';
  import { cn } from '$lib/utils.js';

  let { ref = $bindable(null), class: className, children, ...restProps } = $props();
</script>

<AlertDialogPrimitive.Description bind:ref class={cn('text-sm text-muted-foreground', className)} {...restProps}>
  {@render children?.()}
</AlertDialogPrimitive.Description>
