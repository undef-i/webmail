<script>
  import { cn } from '$lib/utils.js';

  let { class: className, children, ...restProps } = $props();
</script>

<div class={cn('flex flex-col gap-1.5 text-left', className)} {...restProps}>
  {@render children?.()}
</div>
