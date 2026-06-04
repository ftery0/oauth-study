package com.example.app1.domain;

import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;
import java.util.Optional;

public interface NotebookRepository extends JpaRepository<Notebook, Long> {
    List<Notebook> findByOwnerSubOrderByCreatedAtDesc(String ownerSub);
    Optional<Notebook> findByIdAndOwnerSub(Long id, String ownerSub);
}
