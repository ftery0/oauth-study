package com.example.app1.domain;

import jakarta.persistence.*;
import java.time.Instant;

@Entity
@Table(name = "users")
public class User {

    @Id
    @Column(length = 64)
    private String sub;

    @Column(name = "display_name", nullable = false, length = 128)
    private String displayName;

    @Column(name = "created_at", nullable = false)
    private Instant createdAt;

    public User() {}

    public User(String sub, String displayName) {
        this.sub = sub;
        this.displayName = displayName;
        this.createdAt = Instant.now();
    }

    public String getSub() { return sub; }
    public String getDisplayName() { return displayName; }
    public Instant getCreatedAt() { return createdAt; }

    public void setDisplayName(String displayName) { this.displayName = displayName; }
}
