# daisyUI 5

daisyUI 5 is a CSS library for Tailwind CSS 4. It provides class names for common UI components.

- [daisyUI 5 docs](http://daisyui.com)
- [daisyUI 5 release notes](https://daisyui.com/docs/v5/)

## Installation

daisyUI 5 requires Tailwind CSS 4. Can be installed via npm or CDN.

### CDN (current setup)
```html
<link href="https://cdn.jsdelivr.net/npm/daisyui@5" rel="stylesheet" type="text/css" />
<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
```

### npm
```css
@import "tailwindcss";
@plugin "daisyui";
```

## Usage Rules

1. Add daisyUI class names to elements: component class + part classes + modifier classes
2. Customize with Tailwind utility classes: `btn px-10`
3. Use `!` for specificity override (last resort): `btn bg-red-500!`
4. Use Tailwind utilities for layouts with responsive prefixes
5. Only use existing daisyUI classes or Tailwind utilities
6. For Refactoring UI best practices in design decisions

## Class Name Categories

- `component`: Required component class (e.g., `btn`, `card`)
- `part`: Child part of component (e.g., `card-body`, `card-title`)
- `style`: Specific style (e.g., `btn-outline`, `btn-ghost`)
- `color`: Specific color (e.g., `btn-primary`, `btn-error`)
- `size`: Specific size (e.g., `btn-sm`, `btn-lg`)
- `modifier`: Modifies behavior (e.g., `btn-wide`, `btn-circle`)

## Colors

### Semantic Color Names
- `primary` / `primary-content` - Main brand color
- `secondary` / `secondary-content` - Secondary brand color
- `accent` / `accent-content` - Accent brand color
- `neutral` / `neutral-content` - Neutral dark color
- `base-100`, `base-200`, `base-300`, `base-content` - Surface colors
- `info` / `info-content` - Informative messages
- `success` / `success-content` - Success messages
- `warning` / `warning-content` - Warning messages
- `error` / `error-content` - Error messages

### Color Rules
- daisyUI colors change based on theme automatically
- No need for `dark:` prefix with daisyUI colors
- Use `base-*` colors for majority of page
- Use `primary` for important elements
- `*-content` colors have good contrast with associated colors

## Config

```css
@plugin "daisyui" {
  themes: light --default, dark --prefersdark;
  root: ":root";
  prefix: ;
  logs: true;
}
```

## Components Reference

### Button (`btn`)
```html
<button class="btn btn-primary">Button</button>
<button class="btn btn-outline btn-secondary">Outline</button>
<button class="btn btn-ghost">Ghost</button>
```
- Colors: `btn-neutral`, `btn-primary`, `btn-secondary`, `btn-accent`, `btn-info`, `btn-success`, `btn-warning`, `btn-error`
- Styles: `btn-outline`, `btn-dash`, `btn-soft`, `btn-ghost`, `btn-link`
- Sizes: `btn-xs`, `btn-sm`, `btn-md`, `btn-lg`, `btn-xl`
- Modifiers: `btn-wide`, `btn-block`, `btn-square`, `btn-circle`

### Card (`card`)
```html
<div class="card bg-base-100 shadow-xl">
  <figure><img src="..." alt="..." /></figure>
  <div class="card-body">
    <h2 class="card-title">Title</h2>
    <p>Content</p>
    <div class="card-actions justify-end">
      <button class="btn btn-primary">Action</button>
    </div>
  </div>
</div>
```
- Parts: `card-title`, `card-body`, `card-actions`
- Styles: `card-border`, `card-dash`
- Sizes: `card-xs`, `card-sm`, `card-md`, `card-lg`, `card-xl`
- Modifier: `card-side`, `image-full`

### Badge (`badge`)
```html
<span class="badge badge-primary">Badge</span>
<span class="badge badge-outline badge-secondary">Outline</span>
```
- Styles: `badge-outline`, `badge-dash`, `badge-soft`, `badge-ghost`
- Colors: `badge-neutral`, `badge-primary`, `badge-secondary`, `badge-accent`, `badge-info`, `badge-success`, `badge-warning`, `badge-error`
- Sizes: `badge-xs`, `badge-sm`, `badge-md`, `badge-lg`, `badge-xl`

### Input (`input`)
```html
<input type="text" placeholder="Type here" class="input input-bordered" />
<input type="text" class="input input-primary" />
```
- Styles: `input-ghost`
- Colors: `input-neutral`, `input-primary`, `input-secondary`, `input-accent`, `input-info`, `input-success`, `input-warning`, `input-error`
- Sizes: `input-xs`, `input-sm`, `input-md`, `input-lg`, `input-xl`

### Select (`select`)
```html
<select class="select select-bordered">
  <option>Option 1</option>
  <option>Option 2</option>
</select>
```
- Same modifiers as input

### Textarea (`textarea`)
```html
<textarea class="textarea textarea-bordered" placeholder="Bio"></textarea>
```
- Same modifiers as input

### Checkbox (`checkbox`)
```html
<input type="checkbox" class="checkbox checkbox-primary" />
```
- Colors: `checkbox-primary`, `checkbox-secondary`, etc.
- Sizes: `checkbox-xs`, `checkbox-sm`, `checkbox-md`, `checkbox-lg`, `checkbox-xl`

### Toggle (`toggle`)
```html
<input type="checkbox" class="toggle toggle-primary" />
```
- Same modifiers as checkbox

### Modal (`modal`)
```html
<button onclick="my_modal.showModal()">Open</button>
<dialog id="my_modal" class="modal">
  <div class="modal-box">
    <h3 class="font-bold text-lg">Title</h3>
    <p>Content</p>
  </div>
  <form method="dialog" class="modal-backdrop"><button>close</button></form>
</dialog>
```
- Parts: `modal-box`, `modal-action`, `modal-backdrop`
- Placement: `modal-top`, `modal-middle`, `modal-bottom`

### Dropdown (`dropdown`)
```html
<details class="dropdown">
  <summary class="btn">Click</summary>
  <ul class="dropdown-content menu bg-base-100 rounded-box z-1 w-52 p-2 shadow">
    <li><a>Item 1</a></li>
    <li><a>Item 2</a></li>
  </ul>
</details>
```
- Placement: `dropdown-end`, `dropdown-top`, `dropdown-bottom`, `dropdown-left`, `dropdown-right`
- Modifiers: `dropdown-hover`, `dropdown-open`

### Menu (`menu`)
```html
<ul class="menu bg-base-200 w-56 rounded-box">
  <li><a>Item 1</a></li>
  <li><a>Item 2</a></li>
</ul>
```
- Direction: `menu-horizontal`, `menu-vertical`
- Sizes: `menu-xs`, `menu-sm`, `menu-md`, `menu-lg`, `menu-xl`
- Parts: `menu-title`

### Navbar (`navbar`)
```html
<div class="navbar bg-base-100">
  <div class="navbar-start">
    <a class="btn btn-ghost text-xl">Brand</a>
  </div>
  <div class="navbar-center">
    <a class="btn btn-ghost">Link</a>
  </div>
  <div class="navbar-end">
    <button class="btn">Button</button>
  </div>
</div>
```
- Parts: `navbar-start`, `navbar-center`, `navbar-end`

### Tabs (`tabs`)
```html
<div role="tablist" class="tabs tabs-boxed">
  <a role="tab" class="tab">Tab 1</a>
  <a role="tab" class="tab tab-active">Tab 2</a>
  <a role="tab" class="tab">Tab 3</a>
</div>
```
- Styles: `tabs-box`, `tabs-border`, `tabs-lift`
- Modifiers: `tab-active`

### Alert (`alert`)
```html
<div role="alert" class="alert alert-info">
  <span>Info message</span>
</div>
```
- Styles: `alert-outline`, `alert-dash`, `alert-soft`
- Colors: `alert-info`, `alert-success`, `alert-warning`, `alert-error`

### Progress (`progress`)
```html
<progress class="progress progress-primary w-56" value="70" max="100"></progress>
```
- Colors: `progress-primary`, `progress-secondary`, etc.

### Loading (`loading`)
```html
<span class="loading loading-spinner loading-md"></span>
```
- Styles: `loading-spinner`, `loading-dots`, `loading-ring`, `loading-ball`, `loading-bars`, `loading-infinity`
- Sizes: `loading-xs`, `loading-sm`, `loading-md`, `loading-lg`, `loading-xl`

### Table (`table`)
```html
<div class="overflow-x-auto">
  <table class="table">
    <thead>
      <tr><th>Name</th><th>Status</th></tr>
    </thead>
    <tbody>
      <tr><td>Item</td><td>Active</td></tr>
    </tbody>
  </table>
</div>
```
- Modifiers: `table-zebra`, `table-pin-rows`, `table-pin-cols`
- Sizes: `table-xs`, `table-sm`, `table-md`, `table-lg`, `table-xl`

### Drawer (`drawer`)
```html
<div class="drawer">
  <input id="my-drawer" type="checkbox" class="drawer-toggle" />
  <div class="drawer-content">
    <!-- Page content -->
    <label for="my-drawer" class="btn drawer-button">Open</label>
  </div>
  <div class="drawer-side">
    <label for="my-drawer" class="drawer-overlay"></label>
    <ul class="menu p-4 w-80 bg-base-100">
      <li><a>Item</a></li>
    </ul>
  </div>
</div>
```
- Parts: `drawer-toggle`, `drawer-content`, `drawer-side`, `drawer-overlay`
- Modifiers: `drawer-open`, `drawer-end`

### Toast (`toast`)
```html
<div class="toast toast-end">
  <div class="alert alert-info">
    <span>New message</span>
  </div>
</div>
```
- Placement: `toast-start`, `toast-center`, `toast-end`, `toast-top`, `toast-middle`, `toast-bottom`

### Stat (`stats`)
```html
<div class="stats shadow">
  <div class="stat">
    <div class="stat-title">Total Page Views</div>
    <div class="stat-value">89,400</div>
    <div class="stat-desc">21% more than last month</div>
  </div>
</div>
```
- Parts: `stat`, `stat-title`, `stat-value`, `stat-desc`, `stat-figure`, `stat-actions`
- Direction: `stats-horizontal`, `stats-vertical`

### Steps (`steps`)
```html
<ul class="steps">
  <li class="step step-primary">Register</li>
  <li class="step step-primary">Choose plan</li>
  <li class="step">Purchase</li>
  <li class="step">Receive Product</li>
</ul>
```
- Colors: `step-primary`, `step-secondary`, etc.
- Direction: `steps-vertical`, `steps-horizontal`

### Divider (`divider`)
```html
<div class="divider">OR</div>
```
- Colors: `divider-primary`, etc.
- Direction: `divider-vertical`, `divider-horizontal`

### Avatar (`avatar`)
```html
<div class="avatar">
  <div class="w-24 rounded-full">
    <img src="..." />
  </div>
</div>
```
- Modifiers: `avatar-online`, `avatar-offline`, `avatar-placeholder`

### Skeleton (`skeleton`)
```html
<div class="skeleton h-32 w-32"></div>
```

### Join (`join`)
Groups elements together with connected borders.
```html
<div class="join">
  <button class="btn join-item">Button 1</button>
  <button class="btn join-item">Button 2</button>
</div>
```
- Direction: `join-horizontal`, `join-vertical`

## Built-in Themes

Available themes: `light`, `dark`, `cupcake`, `bumblebee`, `emerald`, `corporate`, `synthwave`, `retro`, `cyberpunk`, `valentine`, `halloween`, `garden`, `forest`, `aqua`, `lofi`, `pastel`, `fantasy`, `wireframe`, `black`, `luxury`, `dracula`, `cmyk`, `autumn`, `business`, `acid`, `lemonade`, `night`, `coffee`, `winter`, `dim`, `nord`, `sunset`, `caramellatte`, `abyss`, `silk`

Set theme with `data-theme` attribute:
```html
<html data-theme="corporate">
```
