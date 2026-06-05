export function LoginPage({ error }: { error: string | null }) {
  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4 antialiased text-slate-900">
      <main className="w-full max-w-md bg-white rounded-2xl shadow-lg p-6 sm:p-8">
        <header className="mb-6">
          <div className="inline-flex items-center gap-2 mb-2">
            <span className="inline-block w-2 h-2 rounded-full bg-blue-500" />
            <span className="text-xs font-medium text-blue-700 uppercase tracking-wide">group-a</span>
          </div>
          <h1 className="text-2xl font-semibold">Notebook</h1>
          <p className="mt-1 text-sm text-slate-500">사내 노트 · Spring Boot + React</p>
        </header>

        {error && (
          <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-sm text-red-700" role="alert">
            오류: {error}
          </div>
        )}

        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-5 text-sm text-blue-900">
          <p className="font-medium mb-1">silent SSO 시연</p>
          <p>이미 같은 그룹(group-a)의 TaskBoard 에 로그인되어 있다면 폼 없이 자동 로그인됩니다.</p>
        </div>

        <a
          href="/login"
          className="block w-full text-center rounded-lg bg-blue-600 hover:bg-blue-700 active:bg-blue-800 text-white font-medium px-4 py-3 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          OAuth 로그인
        </a>
      </main>
    </div>
  )
}
