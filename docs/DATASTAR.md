# Datastar Docs

Read the full-page docs at [data-star.dev/docs](https://data-star.dev/docs) for the best experience.

## Guide

### Getting Started

Datastar simplifies frontend development, allowing you to build backend-driven, interactive UIs using a [hypermedia-first](https://hypermedia.systems/hypermedia-a-reintroduction/) approach that extends and enhances HTML.

Datastar provides backend reactivity like [htmx](https://htmx.org/) and frontend reactivity like [Alpine.js](https://alpinejs.dev/) in a lightweight frontend framework that doesn't require any npm packages or other dependencies. It provides two primary functions:
- Modify the DOM and state by sending events from your backend.
- Build reactivity into your frontend using standard `data-*` HTML attributes.

## Installation 

The quickest way to use Datastar is to include it using a `script` tag that fetches it from a CDN.

```html
<script type="module" src="https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.0-RC.7/bundles/datastar.js"></script>
```

If you prefer to host the file yourself, download the [script](https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.0-RC.7/bundles/datastar.js) or create your own bundle using the [bundler](https://data-star.dev/bundler), then include it from the appropriate path.

```html
<script type="module" src="/path/to/datastar.js"></script>
```

## `data-*` Attributes

At the core of Datastar are `data-*` HTML attributes (hence the name). They allow you to add reactivity to your frontend and interact with your backend in a declarative way.

The [`data-on`](https://data-star.dev/reference/attributes#data-on) attribute can be used to attach an event listener to an element and execute an expression whenever the event is triggered.

```html
<button data-on:click="alert('Hello!')">
    Click me
</button>
```

## Patching Elements

With Datastar, the backend *drives* the frontend by **patching** (adding, updating and removing) HTML elements in the DOM.

Datastar receives elements from the backend and manipulates the DOM using a morphing strategy (by default). Morphing ensures that only modified parts of the DOM are updated, and that only data attributes that have changed are reapplied.

Datastar provides [actions](https://data-star.dev/reference/actions#backend-actions) for sending requests to the backend. The [`@get()`](https://data-star.dev/reference/actions#get) action sends a `GET` request to the provided URL.

```html
<button data-on:click="@get('/endpoint')">
    Load data
</button>
<div id="target"></div>
```

If the response has a `content-type` of `text/html`, the top-level HTML elements will be morphed into the existing DOM based on the element IDs.

```html
<div id="target">
    New content from server
</div>
```

If the response has a `content-type` of `text/event-stream`, it can contain zero or more SSE events:

```
event: datastar-patch-elements
data: elements <div id="target">
data: elements     Server content
data: elements </div>

```

## Reactive Signals

Datastar uses *signals* to manage frontend state. Signals are denoted using the `$` prefix.

### `data-bind`

Sets up two-way data binding on any HTML element that receives user input.

```html
<input data-bind:foo />
```

This creates a new signal `$foo` and binds it to the element's value.

### `data-text`

Sets the text content of an element to the value of a signal.

```html
<input data-bind:foo />
<div data-text="$foo"></div>
```

### `data-computed`

Creates a read-only signal derived from a reactive expression.

```html
<input data-bind:foo />
<div data-computed:upper="$foo.toUpperCase()" data-text="$upper"></div>
```

### `data-show`

Shows or hides an element based on an expression.

```html
<input data-bind:foo />
<button data-show="$foo != ''">Save</button>
```

### `data-class`

Adds or removes a class based on an expression.

```html
<button data-class:active="$isActive">Click</button>
```

### `data-attr`

Binds the value of any HTML attribute to an expression.

```html
<button data-attr:disabled="$foo == ''">Save</button>
```

### `data-signals`

Patches one or more signals into existing signals.

```html
<div data-signals:foo="1"></div>
<div data-signals="{foo: 1, bar: 2}"></div>
```

### `data-on`

Attaches an event listener to an element.

```html
<button data-on:click="$foo = ''">Reset</button>
```

## Patching Signals

Just like elements, signals can be patched from the backend:

```
event: datastar-patch-signals
data: signals {foo: 'new value'}

```

## Backend Actions

Datastar provides actions for each HTTP method:

- `@get(uri, options)` - GET request
- `@post(uri, options)` - POST request
- `@put(uri, options)` - PUT request
- `@patch(uri, options)` - PATCH request
- `@delete(uri, options)` - DELETE request

### Options

- `contentType` - `json` (default) or `form`
- `filterSignals` - Filter signals sent to backend
- `headers` - Custom headers
- `openWhenHidden` - Keep SSE connection open when page hidden

### `data-indicator`

Sets a signal to `true` while a request is in flight.

```html
<button data-on:click="@get('/endpoint')" data-indicator:loading>
    Load
</button>
<div data-show="$loading">Loading...</div>
```

## SSE Events Reference

### `datastar-patch-elements`

Patches HTML elements into the DOM.

```
event: datastar-patch-elements
data: elements <div id="foo">Hello</div>

```

Options:
- `data: selector #foo` - CSS selector for target
- `data: mode outer|inner|prepend|append|before|after|remove` - Patch mode
- `data: useViewTransition true` - Use View Transition API

### `datastar-patch-signals`

Patches signals.

```
event: datastar-patch-signals
data: signals {foo: 1, bar: 2}

```

Options:
- `data: onlyIfMissing true` - Only patch if signal doesn't exist

## Go SDK

```go
import "github.com/starfederation/datastar-go/datastar"

// Create SSE generator
sse := datastar.NewSSE(w, r)

// Patch elements
sse.PatchElements(`<div id="target">Hello</div>`)

// Patch signals
sse.PatchSignals([]byte(`{foo: 'value'}`))

// Read signals from request
signals := &Signals{}
if err := datastar.ReadSignals(request, signals); err != nil {
    // handle error
}
```

## The Tao of Datastar

- **State in the Right Place**: Most state should live in the backend.
- **Use Signals Sparingly**: Only for user interactions and sending new state.
- **In Morph We Trust**: Let morphing handle DOM updates.
- **SSE Responses**: Use `text/event-stream` for flexibility.
- **Backend Templating**: Keep things DRY.
- **Page Navigation**: Use anchor tags and redirects.
- **Loading Indicators**: Use `data-indicator` attribute.
- **No Optimistic Updates**: Show loading state, confirm from backend.

## Attribute Modifiers

Many attributes support modifiers using `__modifier.tag` syntax:

- `__delay.500ms` - Delay execution
- `__debounce.500ms` - Debounce execution
- `__throttle.500ms` - Throttle execution
- `__once` - Only trigger once
- `__prevent` - Call preventDefault
- `__stop` - Call stopPropagation
- `__window` - Attach to window

```html
<button data-on:click__debounce.300ms="@get('/search')">Search</button>
```

## Security

- Always escape user input in templates
- Signal values are visible in source code
- Implement backend validation for all inputs
