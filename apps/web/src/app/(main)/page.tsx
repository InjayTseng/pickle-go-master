import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Pickle Go - 找球局、揪球友',
  description: '台灣最方便的匹克球揪團平台，30 秒找到附近球局並報名',
};

export default function HomePage() {
  return (
    <main className="flex min-h-screen flex-col">
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-4xl font-bold text-pickle-600 mb-4">
            Pickle Go
          </h1>
          <p className="text-lg text-muted-foreground">
            台灣最方便的匹克球揪團平台
          </p>
          <p className="text-sm text-muted-foreground mt-2">
            30 秒找到附近球局並報名
          </p>
        </div>
      </div>
    </main>
  );
}
