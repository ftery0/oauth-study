package com.example.app1;

import com.example.app1.domain.User;
import com.example.app1.domain.UserRepository;
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
import java.util.HashMap;
import java.util.HexFormat;
import java.util.Map;

@RestController
public class OAuthController {

    @Value("${oauth.server-url}") private String oauthServerUrl;
    @Value("${oauth.internal-url}") private String oauthInternalUrl;
    @Value("${oauth.client-id}") private String clientId;
    @Value("${oauth.client-secret}") private String clientSecret;
    @Value("${oauth.redirect-uri}") private String redirectUri;
    @Value("${frontend.url}") private String frontendUrl;

    private final HttpClient http = HttpClient.newHttpClient();
    private final ObjectMapper json = new ObjectMapper();
    private final SecureRandom random = new SecureRandom();
    private final UserRepository users;
    private final WelcomeSeed welcomeSeed;

    public OAuthController(UserRepository users, WelcomeSeed welcomeSeed) {
        this.users = users;
        this.welcomeSeed = welcomeSeed;
    }

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

        String accessToken = (String) tokens.get("access_token");
        session.setAttribute("accessToken", accessToken);
        session.setAttribute("refreshToken", (String) tokens.get("refresh_token"));
        // id_token: RP-initiated logout 의 id_token_hint 로 사용 (openid scope 일 때만 IdP 가 발급)
        if (tokens.get("id_token") instanceof String idToken) {
            session.setAttribute("idToken", idToken);
        }

        // 자동 프로비저닝: IdP userinfo → 내부 user 행 upsert + session 에 sub 박음.
        // IdP 가 name (display_name) 을 주면 우선 사용, 없으면 sub fallback.
        Map<String, Object> userInfo = fetchUserInfo(accessToken);
        if (userInfo != null && userInfo.get("sub") instanceof String sub) {
            String name = resolveDisplayName(userInfo, sub);
            users.findById(sub).ifPresentOrElse(
                u -> {
                    if (!name.equals(u.getDisplayName())) {
                        u.setDisplayName(name);
                        users.save(u);
                    }
                },
                () -> users.save(new User(sub, name))
            );
            // 노트북이 0 개인 사용자에게 예시 콘텐츠 한 번 시드 (이미 있으면 no-op).
            welcomeSeed.seedIfEmpty(sub);
            session.setAttribute("userSub", sub);
        }

        res.sendRedirect(frontendUrl);
    }

    // IdP userinfo 응답에서 표시명 결정. name > preferred_username > sub.
    private String resolveDisplayName(Map<String, Object> userInfo, String sub) {
        if (userInfo.get("name") instanceof String n && !n.isBlank()) return n;
        if (userInfo.get("preferred_username") instanceof String u && !u.isBlank()) return u;
        return sub;
    }

    @GetMapping("/api/me")
    public ResponseEntity<?> me(HttpServletRequest req) throws IOException, InterruptedException {
        HttpSession session = req.getSession(false);
        if (session == null || session.getAttribute("accessToken") == null) {
            return ResponseEntity.status(401).body(Map.of("error", "Not authenticated"));
        }

        String accessToken = (String) session.getAttribute("accessToken");
        Map<String, Object> userInfo = fetchUserInfo(accessToken);

        if (userInfo == null) {
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

        // /api/me 응답: 프론트가 실제로 쓰는 필드만 화이트리스트로 노출 (least info).
        // 부수효과: IdP 의 name 으로 내부 user.display_name 동기화 (UUID 박힌 레거시 행 자동 보정).
        Map<String, Object> body = new HashMap<>();
        if (userInfo.get("sub") instanceof String sub) {
            body.put("sub", sub);
            String name = resolveDisplayName(userInfo, sub);
            users.findById(sub).ifPresent(u -> {
                if (!name.equals(u.getDisplayName())) {
                    u.setDisplayName(name);
                    users.save(u);
                }
            });
            if (userInfo.get("preferred_username") instanceof String pu) body.put("preferred_username", pu);
            if (userInfo.get("name") instanceof String n) body.put("name", n);
            session.setAttribute("userSub", sub);
        }
        return ResponseEntity.ok(body);
    }

    // 로그아웃: app1 세션 무효화 + IdP 의 RP-initiated logout 으로 리다이렉트.
    // 단순히 app1 세션만 끊으면 silent SSO 가 다시 자동 로그인시키므로 IdP 세션까지 같이 정리해야 진짜 로그아웃.
    @GetMapping("/api/logout")
    public void logout(HttpServletRequest req, HttpServletResponse res) throws IOException {
        String idTokenHint = null;
        HttpSession session = req.getSession(false);
        if (session != null) {
            if (session.getAttribute("idToken") instanceof String t) {
                idTokenHint = t;
            }
            session.invalidate();
        }

        // 프론트는 비로그인 시 공개 메인 표시. logout=1 hint 같은 거 불필요 — 그냥 메인으로.
        StringBuilder url = new StringBuilder(oauthServerUrl).append("/oauth/logout")
                .append("?post_logout_redirect_uri=").append(urlEncode(frontendUrl));
        if (idTokenHint != null) {
            url.append("&id_token_hint=").append(urlEncode(idTokenHint));
        }
        res.sendRedirect(url.toString());
    }

    private Map<String, Object> exchangeToken(String code) throws IOException, InterruptedException {
        String body = "grant_type=authorization_code"
                + "&code=" + urlEncode(code)
                + "&redirect_uri=" + urlEncode(redirectUri);

        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(oauthInternalUrl + "/oauth/token"))
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
                .uri(URI.create(oauthInternalUrl + "/oauth/token"))
                .header("Content-Type", "application/x-www-form-urlencoded")
                .header("Authorization", "Basic " + basicAuth(clientId, clientSecret))
                .POST(HttpRequest.BodyPublishers.ofString(body))
                .build();

        HttpResponse<String> resp = http.send(req, HttpResponse.BodyHandlers.ofString());
        if (resp.statusCode() != 200) return false;
        Map<String, Object> tokens = parseJson(resp.body());

        session.setAttribute("accessToken", (String) tokens.get("access_token"));
        session.setAttribute("refreshToken", (String) tokens.get("refresh_token"));
        return true;
    }

    private Map<String, Object> fetchUserInfo(String accessToken) throws IOException, InterruptedException {
        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(oauthInternalUrl + "/oauth/userinfo"))
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

    private String urlEncode(String s) { return URLEncoder.encode(s, StandardCharsets.UTF_8); }
    private String basicAuth(String user, String pass) {
        return Base64.getEncoder().encodeToString((user + ":" + pass).getBytes(StandardCharsets.UTF_8));
    }
    private String randomHex(int byteLen) {
        byte[] b = new byte[byteLen];
        random.nextBytes(b);
        return HexFormat.of().formatHex(b);
    }
}
