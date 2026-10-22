<script>
  import { cn } from '$lib/utils.js';

  let { class: className, children, ...restProps } = $props();
</script>

<div class={cn('flex flex-col-reverse sm:flex-row sm:justify-end sm:gap-2', className)} {...restProps}>
  {@render children?.()}
</div>
