<script>
  import { cn } from '$lib/utils.js';

  let { ref = $bindable(null), class: className, children, ...restProps } = $props();
</script>

<div bind:this={ref} class={cn('font-bold leading-none', className)} {...restProps}>
  {@render children?.()}
</div>
