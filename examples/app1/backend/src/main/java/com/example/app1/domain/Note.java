package com.example.app1.domain;

import jakarta.persistence.*;
import java.time.Instant;

@Entity
@Table(name = "notes", indexes = {
        @Index(name = "idx_notes_notebook", columnList = "notebook_id"),
        @Index(name = "idx_notes_owner", columnList = "owner_sub")
})
public class Note {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "notebook_id", nullable = false)
    private Long notebookId;

    @Column(name = "owner_sub", nullable = false, length = 64)
    private String ownerSub;

    @Column(nullable = false, length = 200)
    private String title;

    @Column(name = "body_md", columnDefinition = "MEDIUMTEXT")
    private String bodyMd;

    @Column(name = "created_at", nullable = false)
    private Instant createdAt;

    @Column(name = "updated_at", nullable = false)
    private Instant updatedAt;

    public Note() {}

    public Note(Long notebookId, String ownerSub, String title, String bodyMd) {
        this.notebookId = notebookId;
        this.ownerSub = ownerSub;
        this.title = title;
        this.bodyMd = bodyMd;
        Instant now = Instant.now();
        this.createdAt = now;
        this.updatedAt = now;
    }

    public Long getId() { return id; }
    public Long getNotebookId() { return notebookId; }
    public String getOwnerSub() { return ownerSub; }
    public String getTitle() { return title; }
    public String getBodyMd() { return bodyMd; }
    public Instant getCreatedAt() { return createdAt; }
    public Instant getUpdatedAt() { return updatedAt; }

    public void setTitle(String title) {
        this.title = title;
        this.updatedAt = Instant.now();
    }

    public void setBodyMd(String bodyMd) {
        this.bodyMd = bodyMd;
        this.updatedAt = Instant.now();
    }
}
