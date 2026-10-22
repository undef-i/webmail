<script>
  import { cn } from '$lib/utils.js';

  let { ref = $bindable(null), class: className, children, ...restProps } = $props();
</script>

<div bind:this={ref} class={cn('text-sm text-muted-foreground', className)} {...restProps}>
  {@render children?.()}
</div>
