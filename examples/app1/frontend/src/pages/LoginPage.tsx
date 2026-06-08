export function LoginPage({ error }: { error: string | null }) {
  return (
    <div className="min-h-screen bg-paper-100 flex items-center justify-center p-6 text-ink-900">
      <main className="w-full max-w-lg">
        <div className="mb-10 text-center">
          <div className="inline-flex items-center gap-2 text-[11px] uppercase tracking-[0.18em] text-emerald-800/70 mb-4">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-700" />
            group · a
          </div>
          <h1 className="font-serif text-5xl text-ink-900 mb-3">Notebook</h1>
          <p className="text-ink-500 italic">
            "쓰는 동안 생각이 정리됩니다."
          </p>
        </div>

        <div className="bg-paper-50 border border-paper-200 rounded-lg shadow-sm px-8 py-7">
          {error && (
            <div className="mb-5 border-l-2 border-red-400 bg-red-50/60 pl-3 py-2 text-sm text-red-800" role="alert">
              {error}
            </div>
          )}

          <p className="text-sm text-ink-700 leading-relaxed mb-6">
            이 노트북은 OAuth 로 보호됩니다. 같은 그룹의 다른 앱(TaskBoard)에 이미
            로그인되어 있다면, 별도 입력 없이 곧바로 연결됩니다.
          </p>

          <a
            href="/login"
            className="block w-full text-center rounded-md bg-emerald-800 hover:bg-emerald-900 text-paper-50 font-medium px-4 py-3 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-emerald-700 focus:ring-offset-paper-50"
          >
            OAuth 로 계속하기
          </a>

          <p className="mt-4 text-[11px] text-ink-400 text-center">
            계정이 없으면 로그인 화면에서 바로 만들 수 있어요.
          </p>
        </div>
      </main>
    </div>
  )
}
