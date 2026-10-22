<script>
  import { Dialog as SheetPrimitive } from 'bits-ui';

  let { children, ...restProps } = $props();
</script>

<SheetPrimitive.Portal {...restProps}>
  {@render children?.()}
</SheetPrimitive.Portal>
