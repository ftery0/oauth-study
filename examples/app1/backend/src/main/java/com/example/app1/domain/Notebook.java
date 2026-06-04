package com.example.app1.domain;

import jakarta.persistence.*;
import java.time.Instant;

@Entity
@Table(name = "notebooks", indexes = {
        @Index(name = "idx_notebooks_owner", columnList = "owner_sub")
})
public class Notebook {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "owner_sub", nullable = false, length = 64)
    private String ownerSub;

    @Column(nullable = false, length = 200)
    private String title;

    @Column(name = "created_at", nullable = false)
    private Instant createdAt;

    @Column(name = "updated_at", nullable = false)
    private Instant updatedAt;

    public Notebook() {}

    public Notebook(String ownerSub, String title) {
        this.ownerSub = ownerSub;
        this.title = title;
        Instant now = Instant.now();
        this.createdAt = now;
        this.updatedAt = now;
    }

    public Long getId() { return id; }
    public String getOwnerSub() { return ownerSub; }
    public String getTitle() { return title; }
    public Instant getCreatedAt() { return createdAt; }
    public Instant getUpdatedAt() { return updatedAt; }

    public void setTitle(String title) {
        this.title = title;
        this.updatedAt = Instant.now();
    }
}
