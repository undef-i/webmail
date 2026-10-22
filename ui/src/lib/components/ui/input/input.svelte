<script>
  import { cn } from '$lib/utils.js';

  let {
    ref = $bindable(null),
    value = $bindable(),
    class: className,
    type = 'text',
    ...restProps
  } = $props();
</script>

<input
  bind:this={ref}
  bind:value
  class={cn(
    'flex h-8 w-full border border-input bg-background px-2 py-1 text-base placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 md:text-sm',
    className
  )}
  {type}
  {...restProps}
/>
