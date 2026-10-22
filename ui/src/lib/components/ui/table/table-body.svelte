<script>
  import { cn } from '$lib/utils.js';

  let { ref = $bindable(null), class: className, children, ...restProps } = $props();
</script>

<tbody bind:this={ref} class={cn('', className)} {...restProps}>
  {@render children?.()}
</tbody>
