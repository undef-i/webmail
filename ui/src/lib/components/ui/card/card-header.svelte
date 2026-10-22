<script>
  import { cn } from '$lib/utils.js';

  let { ref = $bindable(null), class: className, children, ...restProps } = $props();
</script>

<div bind:this={ref} class={cn('flex flex-col gap-1 border-b px-4 py-3', className)} {...restProps}>
  {@render children?.()}
</div>
