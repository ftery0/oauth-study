package com.example.app1.web;

import com.example.app1.domain.Notebook;
import com.example.app1.domain.NotebookRepository;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.server.ResponseStatusException;

import java.util.List;

@RestController
@RequestMapping("/api/notebooks")
public class NotebookController {

    private final NotebookRepository notebooks;

    public NotebookController(NotebookRepository notebooks) {
        this.notebooks = notebooks;
    }

    public record CreateRequest(@NotBlank String title) {}
    public record NotebookDto(Long id, String title, String createdAt, String updatedAt) {
        static NotebookDto from(Notebook n) {
            return new NotebookDto(n.getId(), n.getTitle(),
                    n.getCreatedAt().toString(), n.getUpdatedAt().toString());
        }
    }

    @GetMapping
    public List<NotebookDto> list(HttpServletRequest req) {
        String sub = CurrentUser.requireSub(req);
        return notebooks.findByOwnerSubOrderByCreatedAtDesc(sub).stream()
                .map(NotebookDto::from).toList();
    }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public NotebookDto create(@Valid @RequestBody CreateRequest body, HttpServletRequest req) {
        String sub = CurrentUser.requireSub(req);
        Notebook saved = notebooks.save(new Notebook(sub, body.title().trim()));
        return NotebookDto.from(saved);
    }

    @PatchMapping("/{id}")
    public NotebookDto rename(@PathVariable Long id, @Valid @RequestBody CreateRequest body,
                              HttpServletRequest req) {
        String sub = CurrentUser.requireSub(req);
        Notebook n = notebooks.findByIdAndOwnerSub(id, sub)
                .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND));
        n.setTitle(body.title().trim());
        return NotebookDto.from(notebooks.save(n));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> delete(@PathVariable Long id, HttpServletRequest req) {
        String sub = CurrentUser.requireSub(req);
        Notebook n = notebooks.findByIdAndOwnerSub(id, sub)
                .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND));
        notebooks.delete(n);
        return ResponseEntity.noContent().build();
    }
}
