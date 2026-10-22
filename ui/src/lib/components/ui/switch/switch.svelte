<script>
  import { Switch as SwitchPrimitive } from 'bits-ui';
  import { cn } from '$lib/utils.js';

  let {
    ref = $bindable(null),
    class: className,
    checked = $bindable(false),
    ...restProps
  } = $props();
</script>

<SwitchPrimitive.Root
  bind:ref
  bind:checked
  class={cn(
    'peer inline-flex h-[18px] w-[34px] shrink-0 cursor-pointer items-center border border-foreground bg-background focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-foreground',
    className
  )}
  {...restProps}
>
  <SwitchPrimitive.Thumb
    class="pointer-events-none block h-3.5 w-3.5 bg-foreground ring-0 data-[state=checked]:translate-x-4 data-[state=checked]:bg-background data-[state=unchecked]:translate-x-0"
  />
</SwitchPrimitive.Root>
