import Link from 'next/link';
import { Home, ArrowLeft } from 'lucide-react';

import { Button } from '@/components/ui/button';

export default function NotFound() {
  return (
    <div className="container flex flex-col items-center justify-center min-h-[60vh] text-center px-4">
      <h1 className="text-6xl font-bold text-muted-foreground/30 mb-4">404</h1>
      <h2 className="text-2xl font-semibold mb-2">找不到頁面</h2>
      <p className="text-muted-foreground mb-8 max-w-md">
        抱歉，你要找的頁面不存在或已被移除。
      </p>
      <div className="flex gap-4">
        <Link href="/">
          <Button>
            <Home className="mr-2 h-4 w-4" />
            返回首頁
          </Button>
        </Link>
      </div>
    </div>
  );
}
