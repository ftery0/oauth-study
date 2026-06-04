package com.example.app1.web;

import com.example.app1.domain.Note;
import com.example.app1.domain.NoteRepository;
import com.example.app1.domain.NotebookRepository;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.server.ResponseStatusException;

import java.util.List;

@RestController
@RequestMapping("/api/notes")
public class NoteController {

    private final NoteRepository notes;
    private final NotebookRepository notebooks;

    public NoteController(NoteRepository notes, NotebookRepository notebooks) {
        this.notes = notes;
        this.notebooks = notebooks;
    }

    public record CreateRequest(@NotNull Long notebookId, @NotBlank String title, String bodyMd) {}
    public record UpdateRequest(String title, String bodyMd) {}
    public record NoteDto(Long id, Long notebookId, String title, String bodyMd,
                          String createdAt, String updatedAt) {
        static NoteDto from(Note n) {
            return new NoteDto(n.getId(), n.getNotebookId(), n.getTitle(),
                    n.getBodyMd() == null ? "" : n.getBodyMd(),
                    n.getCreatedAt().toString(), n.getUpdatedAt().toString());
        }
    }

    @GetMapping
    public List<NoteDto> list(@RequestParam Long notebookId, HttpServletRequest req) {
        String sub = CurrentUser.requireSub(req);
        notebooks.findByIdAndOwnerSub(notebookId, sub)
                .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND));
        return notes.findByNotebookIdAndOwnerSubOrderByUpdatedAtDesc(notebookId, sub).stream()
                .map(NoteDto::from).toList();
    }

    @GetMapping("/search")
    public List<NoteDto> search(@RequestParam String q, HttpServletRequest req) {
        String sub = CurrentUser.requireSub(req);
        if (q == null || q.trim().isEmpty()) return List.of();
        return notes.searchByOwner(sub, q.trim()).stream().map(NoteDto::from).toList();
    }

    @GetMapping("/{id}")
    public NoteDto get(@PathVariable Long id, HttpServletRequest req) {
        String sub = CurrentUser.requireSub(req);
        return notes.findByIdAndOwnerSub(id, sub).map(NoteDto::from)
                .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND));
    }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public NoteDto create(@Valid @RequestBody CreateRequest body, HttpServletRequest req) {
        String sub = CurrentUser.requireSub(req);
        notebooks.findByIdAndOwnerSub(body.notebookId(), sub)
                .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "notebook not found"));
        Note saved = notes.save(new Note(body.notebookId(), sub, body.title().trim(),
                body.bodyMd() == null ? "" : body.bodyMd()));
        return NoteDto.from(saved);
    }

    @PatchMapping("/{id}")
    public NoteDto update(@PathVariable Long id, @RequestBody UpdateRequest body,
                          HttpServletRequest req) {
        String sub = CurrentUser.requireSub(req);
        Note n = notes.findByIdAndOwnerSub(id, sub)
                .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND));
        if (body.title() != null && !body.title().trim().isEmpty()) n.setTitle(body.title().trim());
        if (body.bodyMd() != null) n.setBodyMd(body.bodyMd());
        return NoteDto.from(notes.save(n));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> delete(@PathVariable Long id, HttpServletRequest req) {
        String sub = CurrentUser.requireSub(req);
        Note n = notes.findByIdAndOwnerSub(id, sub)
                .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND));
        notes.delete(n);
        return ResponseEntity.noContent().build();
    }
}
