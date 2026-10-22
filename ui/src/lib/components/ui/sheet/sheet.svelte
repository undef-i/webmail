<script>
  import { Dialog as SheetPrimitive } from 'bits-ui';

  let { open = $bindable(false), children, ...restProps } = $props();
</script>

<SheetPrimitive.Root bind:open {...restProps}>
  {@render children?.()}
</SheetPrimitive.Root>
