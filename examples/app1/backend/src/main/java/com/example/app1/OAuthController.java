package com.example.app1;

import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import jakarta.servlet.http.HttpSession;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.io.IOException;
import java.net.URI;
import java.net.URLEncoder;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.security.SecureRandom;
import java.util.Base64;
import java.util.HexFormat;
import java.util.Map;

/**
 * OAuth Authorization Code Flow 클라이언트.
 * 백엔드 세션 패턴: 토큰을 HttpSession 에 저장, 브라우저에는 세션 ID 쿠키만 전달.
 */
@RestController
public class OAuthController {

    @Value("${oauth.server-url}") private String oauthServerUrl;
    @Value("${oauth.client-id}") private String clientId;
    @Value("${oauth.client-secret}") private String clientSecret;
    @Value("${oauth.redirect-uri}") private String redirectUri;
    @Value("${frontend.url}") private String frontendUrl;

    private final HttpClient http = HttpClient.newHttpClient();
    private final ObjectMapper json = new ObjectMapper();
    private final SecureRandom random = new SecureRandom();

    // ──────────────────────────────────────────
    // GET /login
    // ──────────────────────────────────────────
    @GetMapping("/login")
    public void login(HttpServletRequest req, HttpServletResponse res) throws IOException {
        String state = randomHex(16);
        req.getSession(true).setAttribute("oauthState", state);

        String authorizeUrl = oauthServerUrl + "/oauth/authorize"
                + "?response_type=code"
                + "&client_id=" + urlEncode(clientId)
                + "&redirect_uri=" + urlEncode(redirectUri)
                + "&scope=" + urlEncode("openid profile")
                + "&state=" + urlEncode(state);

        res.sendRedirect(authorizeUrl);
    }

    // ──────────────────────────────────────────
    // GET /callback
    // ──────────────────────────────────────────
    @GetMapping("/callback")
    public void callback(@RequestParam(required = false) String code,
                         @RequestParam(required = false) String state,
                         @RequestParam(required = false) String error,
                         HttpServletRequest req, HttpServletResponse res)
            throws IOException, InterruptedException {

        if (error != null) {
            res.sendRedirect(frontendUrl + "?error=" + urlEncode(error));
            return;
        }

        HttpSession session = req.getSession(false);
        String sessionState = session == null ? null : (String) session.getAttribute("oauthState");
        if (state == null || !state.equals(sessionState)) {
            res.sendRedirect(frontendUrl + "?error=invalid_state");
            return;
        }
        session.removeAttribute("oauthState");

        Map<String, Object> tokens = exchangeToken(code);
        if (tokens == null) {
            res.sendRedirect(frontendUrl + "?error=token_exchange_failed");
            return;
        }

        session.setAttribute("accessToken", (String) tokens.get("access_token"));
        session.setAttribute("refreshToken", (String) tokens.get("refresh_token"));

        res.sendRedirect(frontendUrl);
    }

    // ──────────────────────────────────────────
    // GET /api/me — access token 만료 시 refresh
    // ──────────────────────────────────────────
    @GetMapping("/api/me")
    public ResponseEntity<?> me(HttpServletRequest req) throws IOException, InterruptedException {
        HttpSession session = req.getSession(false);
        if (session == null || session.getAttribute("accessToken") == null) {
            return ResponseEntity.status(401).body(Map.of("error", "Not authenticated"));
        }

        String accessToken = (String) session.getAttribute("accessToken");
        Map<String, Object> userInfo = fetchUserInfo(accessToken);

        if (userInfo == null) {
            // access token 만료 → refresh
            if (!refreshAccessToken(session)) {
                session.invalidate();
                return ResponseEntity.status(401).body(Map.of("error", "Session expired"));
            }
            accessToken = (String) session.getAttribute("accessToken");
            userInfo = fetchUserInfo(accessToken);
        }

        if (userInfo == null) {
            return ResponseEntity.status(500).body(Map.of("error", "Failed to fetch user info"));
        }
        return ResponseEntity.ok(userInfo);
    }

    // ──────────────────────────────────────────
    // POST /api/logout
    // ──────────────────────────────────────────
    @PostMapping("/api/logout")
    public Map<String, Boolean> logout(HttpServletRequest req) {
        HttpSession session = req.getSession(false);
        if (session != null) {
            session.invalidate();
        }
        return Map.of("ok", true);
    }

    // ──────────────────────────────────────────
    // 헬퍼
    // ──────────────────────────────────────────

    private Map<String, Object> exchangeToken(String code) throws IOException, InterruptedException {
        String body = "grant_type=authorization_code"
                + "&code=" + urlEncode(code)
                + "&redirect_uri=" + urlEncode(redirectUri);

        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(oauthServerUrl + "/oauth/token"))
                .header("Content-Type", "application/x-www-form-urlencoded")
                .header("Authorization", "Basic " + basicAuth(clientId, clientSecret))
                .POST(HttpRequest.BodyPublishers.ofString(body))
                .build();

        HttpResponse<String> resp = http.send(req, HttpResponse.BodyHandlers.ofString());
        if (resp.statusCode() != 200) return null;
        return parseJson(resp.body());
    }

    private boolean refreshAccessToken(HttpSession session) throws IOException, InterruptedException {
        String refreshToken = (String) session.getAttribute("refreshToken");
        if (refreshToken == null) return false;

        String body = "grant_type=refresh_token"
                + "&refresh_token=" + urlEncode(refreshToken);

        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(oauthServerUrl + "/oauth/token"))
                .header("Content-Type", "application/x-www-form-urlencoded")
                .header("Authorization", "Basic " + basicAuth(clientId, clientSecret))
                .POST(HttpRequest.BodyPublishers.ofString(body))
                .build();

        HttpResponse<String> resp = http.send(req, HttpResponse.BodyHandlers.ofString());
        if (resp.statusCode() != 200) return false;
        Map<String, Object> tokens = parseJson(resp.body());

        // Token Rotation: 갱신마다 새 refresh token 저장
        session.setAttribute("accessToken", (String) tokens.get("access_token"));
        session.setAttribute("refreshToken", (String) tokens.get("refresh_token"));
        return true;
    }

    private Map<String, Object> fetchUserInfo(String accessToken) throws IOException, InterruptedException {
        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(oauthServerUrl + "/oauth/userinfo"))
                .header("Authorization", "Bearer " + accessToken)
                .GET()
                .build();
        HttpResponse<String> resp = http.send(req, HttpResponse.BodyHandlers.ofString());
        if (resp.statusCode() != 200) return null;
        return parseJson(resp.body());
    }

    private Map<String, Object> parseJson(String body) {
        try {
            @SuppressWarnings("unchecked")
            Map<String, Object> m = json.readValue(body, Map.class);
            return m;
        } catch (Exception e) {
            return Map.of();
        }
    }

    private String urlEncode(String s) {
        return URLEncoder.encode(s, StandardCharsets.UTF_8);
    }

    private String basicAuth(String user, String pass) {
        return Base64.getEncoder().encodeToString((user + ":" + pass).getBytes(StandardCharsets.UTF_8));
    }

    private String randomHex(int byteLen) {
        byte[] b = new byte[byteLen];
        random.nextBytes(b);
        return HexFormat.of().formatHex(b);
    }
}
