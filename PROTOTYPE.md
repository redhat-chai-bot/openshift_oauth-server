# Local external-script prototype

This is a scope experiment, not a release-ready fix. No live PR, Jira or cluster is changed.

## Design and boundaries

- Built-in OKD login, provider and error HTML comes from the companion oauth-templates Eleventy build. Theme code runs synchronously in the head before body paint; login behavior loads with defer and retains validity, submit progress and pageshow reset behavior. CSS, logos, data fonts and favicon remain inline and are not restricted by this deliberately narrow policy.
- Built-in renderers and token display/confirmation responses send `Content-Security-Policy: script-src 'self'; object-src 'none'; base-uri 'none'`. No unsafe-inline, nonce, default-src, style-src or universal CSP claim is made. Existing standard headers, including X-XSS-Protection, remain untouched to avoid overlap with PR196.
- Public fixed scripts are embedded and served at `/oauth/assets/v1/{theme,login,token}.js` from the existing OAuth mux. This is inside the existing `/oauth/*` anonymous authentication/authorization exception; the generic authentication/authorization chain is neither bypassed nor broadened. Token exchange, CSRF, login and session logic are unchanged. Only three exact filenames are served, GET/HEAD only, with JavaScript MIME, nosniff and no-store. Unmatched filenames return 404.
- Token/server strings remain ordinary Go strings and pass through html/template into hidden quoted data attributes. The external token script reads dataset values and creates text spans, retaining show/copy buttons and clipboard behavior. It contains no per-response data, eval or inline event attributes. Tokens are never placed in script URLs or static files; token/form HTML is no-store. Clipboard failure does not log the snippet or exception.
- **Exact compatibility exception:** constructors passed any custom file continue to use ParseFiles and do not add this CSP, even if their contents resemble the built-in template. Existing CSP from another middleware is not removed. This preserves existing inline/custom remote scripts and legacy operator-supplied branding. Such responses receive no new script restriction from this prototype: they retain their current XSS risk. This is renderer-scoped, not a global configuration switch or upstream API proposal. Token pages have no custom-template entry point and always receive the policy.
- Companion shared layouts externalize scripts for all five generated brands; deploying those updated custom branding templates requires the matching server assets. Existing deployed custom templates stay compatible. Cluster-authentication-operator and managed-cluster-config generated secrets are intentionally not updated here.

## Regeneration

In companion oauth-templates, run `yarn install`, `yarn build`, then `node sync-server.mjs ../oauth-server`. The build performs the actual Liquid, HTML minification and CSS purge transforms and emits `_site/_okd/{login,errors,providers}/index.html` (README's historical flat paths are stale). The synchronization step follows the documented consumer copy into the three Go raw-string template constants and copies built JavaScript assets. `_site` is not committed. Never hand-edit generated Go templates or copied theme/login scripts.

## Focused checks and remaining work

Use `make test-unit GO_TEST_PACKAGES='./pkg/server/assets ./pkg/server/login ./pkg/server/errorpage ./pkg/server/selectprovider ./pkg/server/tokenrequest ./pkg/server/headers'`, scoped `make verify-govet`, and `make build` for the real binary consumer. Standard-library tests inspect rendered script elements, local delivery/MIME, hostile data round trips and custom-file compatibility; they are not a browser HTML parser or CSP execution test.

Before production: run full repository checks and integration suites; browser-test actual CSP enforcement, accessibility, first paint/theme, bfcache, clipboard permission/failure/HTTPs and localization; verify anonymous assets through the full apiserver chain and Route/reverse proxy; test all operator/managed branded artifacts and mixed-version rollout compatibility; decide custom-template policy and migration ownership; examine shell quoting of pasted oc/curl snippets (escaping HTML is not shell escaping, pre-existing behavior); consider CSP reporting and complementary directives separately. Rendering errors retain existing partial-response/log handling and require dedicated integration coverage.

Estimated scope from this prototype: server/source changes and unit coverage are small (roughly 2–4 engineering days); brand consumer coordination and browser/integration/upgrade validation are the dominant risk (roughly another 1–2 weeks across owners). Estimates are planning judgments, not measured delivery guarantees.
