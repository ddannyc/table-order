# Spec: Dish Selection Page — Add Product Images

## Objective

Add dish display images to order selection page. Reference common ordering mini-program layouts: left thumbnail + text info + right action button.

Current page: text-only rows (name, description, price, add button). Backend Product model already has `Image` field (URL string). No backend changes needed.

## Tech Stack

- WeChat mini-program (WXML + WXSS + JS)
- No framework beyond native mini-program APIs
- `wx.image` component for rendering

## Commands

```
Build: cd frontend && npm run build:mp-weixin
Dev H5: cd frontend && npm run dev:h5
```

## Files Changed

```
frontend/pages/home/index.wxml   — Add <image> to product row
frontend/pages/home/index.wxss   — Add image styles, adjust row layout
```

## Code Style

Follow existing patterns: 2-space indent, `wx:for` for lists, inline `bindtap`. No CSS preprocessor.

## Testing Strategy

Manual visual check in H5 dev server or WeChat dev tools. Verify:
- Image renders when product.image is set
- Layout doesn't break when image is empty string
- Text truncation still works
- Cart +/- buttons remain aligned

## Boundaries

- Always: handle missing image gracefully (show placeholder or hide image slot)
- Always: keep image size reasonable (120-160rpx square)
- Never: change backend API or data model

## Success Criteria

- Each product row shows 120rpx square image on left
- Image uses `mode="aspectFill"` for crop
- Fallback: no image shown when `image` field empty (layout stays clean)
- Row layout: [image] [name/desc/price] [action]
- Existing functionality (cart, categories, scan) unchanged

## Open Questions

None.
