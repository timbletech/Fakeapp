package api

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"device_only/internal/sim/model"
	"device_only/internal/sim/service"

	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
)

// AuthHandler exposes the two Timble authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Start handles POST /v1/auth/start.
func (h *AuthHandler) Start(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST is supported.", requestID)
		return
	}

	var req model.StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.", requestID)
		return
	}

	log.Printf("[INFO] POST /v1/sim/start client_id=%s user_ref=%s", req.ClientID, req.UserRef)

	if strings.TrimSpace(req.ClientID) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Field 'client_id' is required.", requestID)
		return
	}
	if strings.TrimSpace(req.UserRef) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Field 'user_ref' is required.", requestID)
		return
	}
	if strings.TrimSpace(req.MSISDN) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Field 'msisdn' is required.", requestID)
		return
	}
	if !validMSISDN(req.MSISDN) {
		writeError(w, http.StatusBadRequest, "validation_error", "Field 'msisdn' must be numeric and at least 10 digits.", requestID)
		return
	}

	resp, err := h.authService.StartAuth(&req)
	if err != nil {
		log.Printf("[ERROR] StartAuth failed: %v", err)
		writeError(w, http.StatusBadGateway, "upstream_error", "Failed to initialize verification with upstream provider.", requestID)
		return
	}

	resp.RequestID = requestID
	writeJSON(w, http.StatusOK, resp)
}

// Complete handles POST /v1/auth/complete.
func (h *AuthHandler) Complete(w http.ResponseWriter, r *http.Request) {
	requestID := "req_" + uuid.New().String()

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST is supported.", requestID)
		return
	}

	var req model.CompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.", requestID)
		return
	}

	log.Printf("[INFO] POST /v1/sim/complete auth_session_id=%s", req.AuthSessionID)

	if strings.TrimSpace(req.AuthSessionID) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Field 'auth_session_id' is required.", requestID)
		return
	}

	result, sessErr := h.authService.CompleteAuth(req.AuthSessionID)
	if sessErr != nil {
		log.Printf("[WARN] CompleteAuth session error: code=%s session_id=%s", sessErr.Code, req.AuthSessionID)
		writeError(w, sessErr.HTTP, strings.ToLower(sessErr.Code), sessErr.Message, requestID)
		return
	}

	if result.IsPending {
		result.Pending.RequestID = requestID
		writeJSON(w, http.StatusAccepted, result.Pending)
		return
	}

	result.Complete.RequestID = requestID
	writeJSON(w, http.StatusOK, result.Complete)
}

// Redirect handles GET /v1/sim/redirect/{session_id}.
// It looks up the session and issues a 302 Found redirect to the upstream SessionURI.
//
// Pass ?probe=1 to skip the 302 and return the resolved upstream URL as JSON.
// Useful for debugging from a desktop browser/curl without consuming the
// single-use upstream session_uri.
func (h *AuthHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	// Simple path extraction assuming path is exactly /v1/sim/redirect/{id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid redirect URL", http.StatusBadRequest)
		return
	}
	sessionID := pathParts[4]

	// HTTPS welcome gate: first HTTPS hit (typical QR scan landing) shows a
	// no-consumption welcome page. After 4 s the welcome JS navigates to the
	// same URL with ?proceed=1, which skips this gate and runs the existing
	// verify interstitial (which itself navigates the device cross-origin to
	// Sekura over HTTP for operator enrichment). ?force=1 and ?probe=1 also
	// bypass the gate for direct/debug callers.
	if isHTTPS(r) && r.URL.Query().Get("proceed") == "" && r.URL.Query().Get("force") == "" && r.URL.Query().Get("probe") == "" {
		serveWelcomePage(w, r, sessionID)
		return
	}

	// Probe mode is a non-consuming peek for debugging.
	if r.URL.Query().Get("probe") != "" {
		sessURI, err := h.authService.GetUpstreamSessionURI(sessionID)
		if err != nil {
			http.Error(w, "Session not found or expired", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"session_id":   sessionID,
			"upstream_uri": sessURI,
			"note":         "probe mode: did NOT redirect, session_uri not consumed",
		})
		return
	}

	// Real redirect: atomically mark the session as consumed so a refresh /
	// browser-back / second tap returns a clean 410 instead of letting the
	// browser walk into Sekura's INVALID_REQUEST response (which has no
	// useful CORS or status semantics for a frontend).
	sessURI, prior, err := h.authService.ConsumeSessionURI(sessionID)
	if err != nil {
		http.Error(w, "Session not found or expired", http.StatusNotFound)
		return
	}

	// Force the upstream Sekura URL to plain HTTP across every downstream
	// path (interstitial nav AND raw 302). Operator header-enrichment only
	// injects MSISDN on http:// requests, so an https:// URI here would
	// silently break SIM verification.
	if strings.HasPrefix(sessURI, "https://") {
		sessURI = "http://" + strings.TrimPrefix(sessURI, "https://")
	}

	accept := r.Header.Get("Accept")
	wantsHTML := strings.Contains(accept, "text/html")

	if prior != nil {
		// Already consumed — refuse to redirect again.
		if wantsHTML {
			serveAlreadyConsumed(w, sessionID, *prior)
		} else {
			writeJSON(w, http.StatusGone, map[string]string{
				"error":           "session_already_consumed",
				"session_id":      sessionID,
				"first_used_at":   prior.Format("2006-01-02T15:04:05Z"),
				"message":         "This session_uri has already been used. Sekura sessions are single-use; start a fresh /v1/auth/start.",
			})
		}
		return
	}

	// First-time redirect.
	// Browsers (Accept: text/html) get an HTML interstitial that shows a 5-sec
	// countdown then sets window.location.href directly to the upstream
	// session_uri (top-level navigation, HTTPS→HTTP). The countdown gives the
	// user a moment to see what's happening; the direct nav lets the device
	// hit the upstream URL on cellular so the operator can identify the SIM.
	// WebViews / non-browser clients (Accept: */*) and ?force=1 still get the
	// raw 302 so existing production callers aren't disturbed.
	forceRedirect := r.URL.Query().Get("force") != ""
	if !forceRedirect && wantsHTML {
		serveRedirectInterstitial(w, sessionID, sessURI)
		return
	}

	http.Redirect(w, r, sessURI, http.StatusFound)
}

// simBrandCSS is the shared on-brand stylesheet for every SIM-flow page
// (redirect, bridge, result, already-consumed). Kept inline + system fonts
// because the bridge page is served over HTTP and must not load HTTPS
// subresources (would mixed-content block) or HTTP fonts on the HTTPS pages.
const simBrandCSS = `
*{box-sizing:border-box;margin:0;padding:0}
html,body{height:100%}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;background:#f0f2f5;color:#111827;line-height:1.5;-webkit-font-smoothing:antialiased}
.shell{min-height:100%;display:flex;flex-direction:column;align-items:center;padding:24px 16px 32px}
.brand-bar{display:flex;align-items:center;gap:10px;margin:4px 0 22px}
.brand-mark{width:10px;height:10px;border-radius:50%;background:#0f766e;display:inline-block;box-shadow:0 0 0 4px rgba(15,118,110,.12)}
.brand-text{color:#0f766e;font-size:18px;font-weight:700;letter-spacing:.2px}
.brand-sub{color:#6b7280;font-size:13px;font-weight:500;border-left:1px solid #d1d5db;padding-left:10px}
.card{width:100%;max-width:460px;background:#fff;border:1px solid #e5e7eb;border-radius:14px;box-shadow:0 4px 16px rgba(0,0,0,.06);padding:28px 24px 24px;text-align:center}
.icon{width:68px;height:68px;border-radius:50%;display:flex;align-items:center;justify-content:center;margin:0 auto 18px;font-size:32px;line-height:1;font-weight:600}
.icon.ok{background:#dcfce7;color:#15803d}
.icon.err{background:#fee2e2;color:#b91c1c}
.icon.warn{background:#fef3c7;color:#b45309}
.spin{width:60px;height:60px;border-radius:50%;margin:0 auto 18px;border:4px solid #e6f4f3;border-top-color:#0f766e;animation:spin .9s linear infinite}
@keyframes spin{to{transform:rotate(360deg)}}
h1{font-size:20px;font-weight:600;margin:0 0 8px;color:#111827;letter-spacing:-.1px}
.lede{color:#374151;font-size:15px;margin:0}
.muted{color:#6b7280;font-size:13px;margin-top:6px}
.note{margin-top:16px;color:#4b5563;font-size:13.5px;line-height:1.55}
.warn-box{display:none;margin-top:16px;padding:12px 14px;background:#fef3c7;border:1px solid #fde68a;border-radius:10px;text-align:left;color:#78350f;font-size:13px;line-height:1.55}
.warn-box.show{display:block}
.warn-box strong{color:#7c2d12}
.btn{display:inline-block;margin-top:16px;padding:11px 18px;background:#0f766e;color:#fff;border-radius:10px;text-decoration:none;font-weight:600;font-size:14px;letter-spacing:.1px;border:1px solid #0f766e}
.btn:active{background:#043f3b;border-color:#043f3b}
.btn.ghost{background:#fff;color:#0f766e}
.foot{margin-top:18px;color:#9ca3af;font-size:12px;display:flex;align-items:center;gap:6px;flex-wrap:wrap;justify-content:center}
.foot code{background:#fff;border:1px solid #e5e7eb;padding:2px 6px;border-radius:6px;font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:11px;color:#4b5563;word-break:break-all}
.dots{display:inline-flex;gap:4px;margin-left:4px}
.dots span{width:4px;height:4px;background:#9ca3af;border-radius:50%;animation:dots 1.2s infinite}
.dots span:nth-child(2){animation-delay:.2s}
.dots span:nth-child(3){animation-delay:.4s}
@keyframes dots{0%,80%,100%{opacity:.2}40%{opacity:1}}
`

// isHTTPS reports whether the inbound request was made over TLS, accounting
// for TLS-terminating proxies that forward via HTTP and set X-Forwarded-Proto.
func isHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		return true
	}
	return false
}

// serveWelcomePage is the HTTPS landing rendered on first QR scan. It does
// not consume the session_uri. After 4 seconds a script does
// window.location.href = nextURL, where nextURL is the HTTP /v1/sim/verify
// alias on the app's direct port (8097). Going via :8097 bypasses the front
// proxy's HTTP->HTTPS 301 redirect for this domain, which otherwise looped
// the user back to the welcome page.
func serveWelcomePage(w http.ResponseWriter, r *http.Request, sessionID string) {
	host := r.Host
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}
	// Hard-coded to match SERVER_PORT in the systemd unit / .env. Update both
	// if the listener moves.
	nextURL := "http://" + host + ":8097/v1/sim/verify/" + sessionID

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	safeID := html.EscapeString(sessionID)

	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="theme-color" content="#0f766e">
<title>Welcome | Vishwas</title>
<style>%s
.count{display:inline-flex;align-items:center;justify-content:center;width:64px;height:64px;border-radius:50%%;background:#e6f4f3;color:#0f766e;font-size:28px;font-weight:700;margin:0 auto 18px;font-variant-numeric:tabular-nums}
</style>
</head>
<body>
<div class="shell">
<header class="brand-bar"><span class="brand-mark"></span><span class="brand-text">Vishwas</span><span class="brand-sub">SIM Verification</span></header>
<main class="card">
<div class="count" id="count" aria-live="polite">4</div>
<h1>Welcome</h1>
<p class="lede">Moving to verification in <span id="countdown">4</span> seconds.</p>
</main>
<div class="foot"><span>Session</span><code>%s</code></div>
</div>
<script>
(function(){
  var next = %q;
  var n = 4;
  var bigEl = document.getElementById('count');
  var inlineEl = document.getElementById('countdown');
  var t = setInterval(function(){
    n--;
    if (bigEl) bigEl.textContent = n;
    if (inlineEl) inlineEl.textContent = n;
    if (n <= 0) {
      clearInterval(t);
      window.location.href = next;
    }
  }, 1000);
})();
</script>
</body>
</html>`, simBrandCSS, safeID, nextURL)
}

// serveAlreadyConsumed renders an HTML 410 page when a session_uri has been
// reloaded after first use. Without this the browser would 302 into a stale
// Sekura session and end on INVALID_REQUEST/unavailable, which looks like a
// generic failure.
func serveAlreadyConsumed(w http.ResponseWriter, sessionID string, firstUsed time.Time) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusGone)
	safeID := html.EscapeString(sessionID)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="theme-color" content="#0f766e">
<title>Session already used | Vishwas</title>
<style>%s</style>
</head>
<body>
<div class="shell">
<header class="brand-bar"><span class="brand-mark"></span><span class="brand-text">Vishwas</span><span class="brand-sub">SIM Verification</span></header>
<main class="card">
<div class="icon warn" aria-hidden="true">!</div>
<h1>This session has already been used</h1>
<p class="lede">SIM verification links are single-use for security.</p>
<p class="note">Reloading or sharing this link won't work. Open the bank app and request a new SIM verification to receive a fresh link.</p>
<p class="muted">First opened at %s</p>
</main>
<div class="foot"><span>Session</span><code>%s</code></div>
</div>
</body>
</html>`, simBrandCSS, firstUsed.Format("2006-01-02 15:04:05 UTC"), safeID)
}

// serveRedirectInterstitial renders the SIM verification entry page reached
// by the QR scan. The flow:
//
//   QR → HTTPS interstitial (this page)
//        ↓ 4-second countdown
//        ↓ window.open(upstream, '_blank')   ← phone hits Sekura
//                                               with its cellular IP
//
// No server-side fetch. After the popup opens, the original tab stays
// where it is — the user can read the page, the popup-tab handles Sekura.
//
// Popup-blocker fallback: if window.open returns null (Chrome/Android often
// blocks timer-driven popups), we top-level-navigate to upstream instead.
// SIM call still happens; user sees Sekura's response and presses back.
func serveRedirectInterstitial(w http.ResponseWriter, sessionID, upstreamURI string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	// Don't let any proxy/CDN upgrade this path to HTTPS — the whole flow
	// must stay cleartext so the operator can enrich the redirect target.
	// HSTS would break SIM verification on iOS by forcing the interstitial
	// to HTTPS, which then makes the redirect to http://sekura a
	// downgrade and triggers Safari's "Connection Is Not Private" warning.
	w.Header().Set("Strict-Transport-Security", "max-age=0")
	w.WriteHeader(http.StatusOK)

	// Force cleartext on the upstream too: operator header-enrichment only
	// injects MSISDN on http:// requests, so https:// would silently break
	// SIM verification.
	if strings.HasPrefix(upstreamURI, "https://") {
		upstreamURI = "http://" + strings.TrimPrefix(upstreamURI, "https://")
	}

	safeID := html.EscapeString(sessionID)
	safeURI := html.EscapeString(upstreamURI)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="theme-color" content="#0f766e">
<title>SIM Verification | Vishwas</title>
<style>%s
.count{display:inline-flex;align-items:center;justify-content:center;width:64px;height:64px;border-radius:50%%;background:#e6f4f3;color:#0f766e;font-size:28px;font-weight:700;margin:0 auto 18px;font-variant-numeric:tabular-nums}
</style>
</head>
<body>
<div class="shell">
<header class="brand-bar"><span class="brand-mark"></span><span class="brand-text">Vishwas</span><span class="brand-sub">SIM Verification</span></header>
<main class="card">
<div class="count" id="count" aria-live="polite">4</div>
<h1>Starting verification</h1>
<p class="lede">Verifying SIM in <span id="countdown">4</span> seconds.</p>
<p class="note">Make sure you're on cellular data (not WiFi). SIM verification only works over the operator's network.</p>
<a class="btn primary" id="manual-link" href="%s" rel="noopener" style="display:none;margin-top:14px">Tap to continue verification</a>
<div class="warn-box" id="warn"><strong>The redirect didn't complete.</strong> Switch off WiFi, stay on cellular data, then tap the button above.</div>
</main>
<div class="foot"><span>Session</span><code>%s</code></div>
</div>
<script>
(function(){
  var upstream = %q;
  var n = 4;
  var bigEl = document.getElementById('count');
  var inlineEl = document.getElementById('countdown');
  var manualLink = document.getElementById('manual-link');

  function go(){
    // HTTP-to-HTTP top-level navigation. Works on both iOS and Android
    // because there is no protocol downgrade — the interstitial itself is
    // served over HTTP, so Safari has no reason to warn.
    window.location.href = upstream;
  }

  var t = setInterval(function(){
    n--;
    if (bigEl) bigEl.textContent = n;
    if (inlineEl) inlineEl.textContent = n;
    if (n <= 0) {
      clearInterval(t);
      go();
      // Recovery path: if the user is still on this page 3 seconds after
      // the redirect attempt, show a manual retry button. Most common
      // cause is WiFi being on — the cellular GET never goes out.
      setTimeout(function(){
        if (manualLink) manualLink.style.display = 'inline-block';
        var w = document.getElementById('warn');
        if (w) w.classList.add('show');
      }, 3000);
    }
  }, 1000);

  if (manualLink) {
    manualLink.addEventListener('click', function(e){
      e.preventDefault();
      go();
    });
  }
})();
</script>
</body>
</html>`, simBrandCSS, safeURI, safeID, upstreamURI)
}

// serveResultPage renders the success/error page the user lands on after the
// background SIM verification call completes. ALLOW/DENY is still decided by
// the bank app via /v1/sim/complete; this page is purely user-facing UX
// telling the human "done, return to the app."
func serveResultPage(w http.ResponseWriter, sessionID, status string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	safeID := html.EscapeString(sessionID)

	iconClass := "ok"
	iconChar := "&#10003;"
	heading := "Verification successful"
	lede := "Your SIM has been verified."
	note := "You can close this page and return to the bank app. The app will continue with the next step automatically."
	if status == "error" {
		iconClass = "err"
		iconChar = "&#33;"
		heading = "Couldn't reach verification network"
		lede = "We couldn't contact the SIM verification service."
		note = "Make sure you're on cellular data (not WiFi), then ask the bank app to retry. SIM verification needs the mobile operator's network."
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="theme-color" content="#0f766e">
<title>%s | Vishwas</title>
<style>%s</style>
</head>
<body>
<div class="shell">
<header class="brand-bar"><span class="brand-mark"></span><span class="brand-text">Vishwas</span><span class="brand-sub">SIM Verification</span></header>
<main class="card">
<div class="icon %s" aria-hidden="true">%s</div>
<h1>%s</h1>
<p class="lede">%s</p>
<p class="note">%s</p>
</main>
<div class="foot"><span>Session</span><code>%s</code></div>
</div>
</body>
</html>`, heading, simBrandCSS, iconClass, iconChar, heading, lede, note, safeID)
}

// Result handles GET /v1/sim/result/{session_id}.
//
// The user lands here after the redirect interstitial's background fetch to
// Sekura settles. ?status=ok|error controls the success/failure UI, but
// the authoritative ALLOW/DENY decision is still produced by the bank app
// via POST /v1/sim/complete - this page just acknowledges the device-side
// flow finished.
func (h *AuthHandler) Result(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is supported.", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid result URL", http.StatusBadRequest)
		return
	}
	sessionID := pathParts[4]

	status := r.URL.Query().Get("status")
	if status != "ok" && status != "error" {
		status = "ok"
	}

	serveResultPage(w, sessionID, status)
}

// QR handles GET /v1/sim/qr.png?text=<url>&size=<px>.
// Encodes arbitrary text (typically the wrapped session_uri) into a PNG QR
// code so the demo UI can render `<img src="/v1/sim/qr.png?text=...">` and
// the user can scan it from a phone. Server-side rendering keeps the static
// page free of QR-library JS dependencies.
func (h *AuthHandler) QR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is supported.", http.StatusMethodNotAllowed)
		return
	}
	text := r.URL.Query().Get("text")
	if text == "" {
		http.Error(w, "Missing 'text' query parameter", http.StatusBadRequest)
		return
	}
	if len(text) > 2048 {
		http.Error(w, "Text too long (max 2048 bytes)", http.StatusBadRequest)
		return
	}

	size := 240
	if s := r.URL.Query().Get("size"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 64 && n <= 1024 {
			size = n
		}
	}

	png, err := qrcode.Encode(text, qrcode.Medium, size)
	if err != nil {
		log.Printf("[ERROR] QR encode failed: %v", err)
		http.Error(w, "Failed to encode QR", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Length", strconv.Itoa(len(png)))
	_, _ = w.Write(png)
}

// Poll handles GET /v1/sim/poll/{session_id}.
// It proxies the upstream polling result through the current server domain.
func (h *AuthHandler) Poll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is supported.", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid polling URL", http.StatusBadRequest)
		return
	}
	sessionID := pathParts[4]

	pollResp, ready, err := h.authService.PollBySessionID(sessionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "expired") {
			http.Error(w, "Session not found or expired", http.StatusNotFound)
			return
		}
		log.Printf("[ERROR] Poll session_id=%s upstream error: %v", sessionID, err)
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"status":     "ERROR",
			"session_id": sessionID,
			"message":    "Upstream polling failed",
			"upstream":   err.Error(),
		})
		return
	}

	if !ready {
		writeJSON(w, http.StatusAccepted, map[string]any{
			"status":     "PENDING",
			"message":    "Device verification not yet complete. Ensure session_uri has been loaded on device over mobile data, then retry.",
			"session_id": sessionID,
		})
		return
	}

	writeJSON(w, http.StatusOK, pollResp)
}

// validMSISDN returns true when s is all digits and at least 10 characters long.
func validMSISDN(s string) bool {
	if len(s) < 10 {
		return false
	}
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, errCode, message, requestID string) {
	writeJSON(w, status, model.ErrorResponse{
		Error:     errCode,
		Message:   message,
		RequestID: requestID,
	})
}
