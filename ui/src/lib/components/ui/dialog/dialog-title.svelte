<script>
  import { Dialog as DialogPrimitive } from 'bits-ui';
  import { cn } from '$lib/utils.js';

  let { ref = $bindable(null), class: className, children, ...restProps } = $props();
</script>

<DialogPrimitive.Title
  bind:ref
  class={cn('text-base font-bold leading-none', className)}
  {...restProps}
>
  {@render children?.()}
</DialogPrimitive.Title>
