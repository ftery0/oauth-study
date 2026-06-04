package com.example.app1.web;

import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpSession;
import org.springframework.web.server.ResponseStatusException;
import org.springframework.http.HttpStatus;

public final class CurrentUser {
    private CurrentUser() {}

    public static String requireSub(HttpServletRequest req) {
        HttpSession session = req.getSession(false);
        Object sub = session == null ? null : session.getAttribute("userSub");
        if (sub instanceof String s && !s.isEmpty()) return s;
        throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Not authenticated");
    }
}
