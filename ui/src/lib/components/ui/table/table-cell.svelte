<script>
  import { cn } from '$lib/utils.js';

  let { ref = $bindable(null), class: className, children, ...restProps } = $props();
</script>

<td bind:this={ref} class={cn('border px-2 py-1.5 align-middle [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]', className)} {...restProps}>
  {@render children?.()}
</td>
