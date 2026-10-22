<script module>
  import { tv } from 'tailwind-variants';

  export const badgeVariants = tv({
    base: 'inline-flex items-center border px-1.5 py-0 text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-ring',
    variants: {
      variant: {
        default: 'border-foreground bg-foreground text-background',
        secondary: 'border-input bg-secondary text-secondary-foreground',
        destructive: 'border-destructive bg-destructive text-destructive-foreground',
        outline: 'text-foreground'
      }
    },
    defaultVariants: {
      variant: 'default'
    }
  });
</script>

<script>
  import { cn } from '$lib/utils.js';

  let { class: className, variant = 'default', children, ...restProps } = $props();
</script>

<span class={cn(badgeVariants({ variant }), className)} {...restProps}>
  {@render children?.()}
</span>
