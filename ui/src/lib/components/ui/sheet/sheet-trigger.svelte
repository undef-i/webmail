<script>
  import { Dialog as SheetPrimitive } from 'bits-ui';

  let { ref = $bindable(null), children, ...restProps } = $props();
</script>

<SheetPrimitive.Trigger bind:ref {...restProps}>
  {@render children?.()}
</SheetPrimitive.Trigger>
