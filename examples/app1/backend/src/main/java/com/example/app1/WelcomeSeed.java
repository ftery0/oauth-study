package com.example.app1;

import com.example.app1.domain.Note;
import com.example.app1.domain.NoteRepository;
import com.example.app1.domain.Notebook;
import com.example.app1.domain.NotebookRepository;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * 신규 사용자 (또는 노트북이 0 개인 사용자) 에게 예시 노트북·노트를 한 번 만들어 둔다.
 * OAuthController.callback 에서 호출. 이미 노트북이 있으면 no-op.
 */
@Component
public class WelcomeSeed {

    private final NotebookRepository notebooks;
    private final NoteRepository notes;

    public WelcomeSeed(NotebookRepository notebooks, NoteRepository notes) {
        this.notebooks = notebooks;
        this.notes = notes;
    }

    public void seedIfEmpty(String ownerSub) {
        if (!notebooks.findByOwnerSubOrderByCreatedAtDesc(ownerSub).isEmpty()) return;

        Notebook gettingStarted = notebooks.save(new Notebook(ownerSub, "시작하기"));
        for (var spec : GETTING_STARTED) {
            notes.save(new Note(gettingStarted.getId(), ownerSub, spec.title, spec.body));
        }

        Notebook project = notebooks.save(new Notebook(ownerSub, "프로젝트 메모"));
        for (var spec : PROJECT) {
            notes.save(new Note(project.getId(), ownerSub, spec.title, spec.body));
        }
    }

    private record Spec(String title, String body) {}

    private static final List<Spec> GETTING_STARTED = List.of(
            new Spec("환영합니다",
                    """
                    # 환영합니다 👋

                    이 노트북은 **Markdown** 으로 자유롭게 쓸 수 있어요.

                    - 좌측 사이드바에서 다른 노트북으로 이동
                    - 가운데 목록에서 노트 선택
                    - 우측은 편집기 + 미리보기

                    > 자동 저장 (1.2 초 디바운스) 이라 따로 저장 버튼을 누를 필요가 없어요.
                    """),
            new Spec("마크다운 치트시트",
                    """
                    # 마크다운 빠른 정리

                    ## 텍스트 강조
                    - **굵게** / *기울임* / ~~취소선~~
                    - `인라인 코드`

                    ## 목록
                    1. 번호 목록
                    2. 항목 둘
                       - 중첩 가능

                    ## 코드 블록
                    ```ts
                    function greet(name: string) {
                      return `Hello, ${name}!`
                    }
                    ```

                    ## 인용
                    > 한 줄 인용
                    >
                    > 여러 줄도 가능

                    ## 링크 / 이미지
                    [OAuth 2.0 RFC 6749](https://datatracker.ietf.org/doc/html/rfc6749)
                    """),
            new Spec("팁",
                    """
                    # 자주 쓰는 단축 흐름

                    - 새 노트북: 좌측 상단 `+`
                    - 새 노트: 가운데 상단 `+`
                    - 검색: 우측 상단 입력칸 (200ms 디바운스)
                    - 노트북 이름 바꾸기: 마우스 올리면 `✎` / 삭제는 `✕`
                    - 헤더의 아바타 클릭 → 프로필 페이지
                    """)
    );

    private static final List<Spec> PROJECT = List.of(
            new Spec("회의록 템플릿",
                    """
                    # 회의록 — yyyy-mm-dd

                    **참석자:** 이름1, 이름2
                    **장소:** 회의실 A
                    **시간:** 13:00 — 14:00

                    ## 안건
                    1. 이슈 A
                    2. 이슈 B

                    ## 결정 사항
                    - [ ] 액션 1 (담당: ___, 마감: ___)
                    - [ ] 액션 2

                    ## 다음 일정
                    yyyy-mm-dd
                    """),
            new Spec("아이디어 보관함",
                    """
                    # 아이디어

                    | 우선순위 | 항목 | 비고 |
                    |---|---|---|
                    | 🔥 | 검색 디바운스 200 → 150ms 로 줄여보기 | 체감 빠를 듯 |
                    | ⭐ | 노트 즐겨찾기 별표 | 사이드바 상단 고정 |
                    | 💡 | 코드 블록 syntax highlighting | highlight.js? |
                    """)
    );
}
