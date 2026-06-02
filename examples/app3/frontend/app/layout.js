import './globals.css'

export const metadata = {
  title: 'App 3',
}

export default function RootLayout({ children }) {
  return (
    <html lang="ko">
      <body>{children}</body>
    </html>
  )
}
