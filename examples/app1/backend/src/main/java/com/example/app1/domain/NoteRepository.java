package com.example.app1.domain;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.util.List;
import java.util.Optional;

public interface NoteRepository extends JpaRepository<Note, Long> {
    List<Note> findByNotebookIdAndOwnerSubOrderByUpdatedAtDesc(Long notebookId, String ownerSub);
    Optional<Note> findByIdAndOwnerSub(Long id, String ownerSub);

    @Query("""
        select n from Note n
        where n.ownerSub = :sub
          and (lower(n.title) like lower(concat('%', :q, '%'))
               or lower(n.bodyMd) like lower(concat('%', :q, '%')))
        order by n.updatedAt desc
    """)
    List<Note> searchByOwner(@Param("sub") String ownerSub, @Param("q") String q);
}
