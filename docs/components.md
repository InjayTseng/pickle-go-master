# Pickle Go 前端元件文件

## 概述

Pickle Go 前端使用 Next.js 14 App Router 架構，搭配 shadcn/ui 元件庫和 TanStack Query 進行狀態管理。所有元件採用 TypeScript 開發，並遵循響應式設計原則。

**技術棧**:
- Next.js 14 (App Router)
- React 18
- TypeScript
- Tailwind CSS
- shadcn/ui
- TanStack Query
- Google Maps API

---

## 目錄結構

```
apps/web/src/
├── components/
│   ├── ui/              # shadcn/ui 基礎元件
│   ├── layout/          # 版面配置元件
│   ├── event/           # 活動相關元件
│   ├── map/             # 地圖相關元件
│   ├── form/            # 表單元件
│   ├── seo/             # SEO 元件
│   └── analytics/       # 分析追蹤元件
├── app/                 # Next.js App Router 頁面
├── lib/                 # 工具函式和 API Client
├── hooks/               # 自訂 React Hooks
├── contexts/            # React Context
└── types/               # TypeScript 型別定義
```

---

## 1. UI 基礎元件 (components/ui)

基於 shadcn/ui 的可重用基礎元件。

### 1.1 Button

按鈕元件，支援多種樣式和尺寸。

**檔案**: `components/ui/button.tsx`

#### 使用範例

```tsx
import { Button } from '@/components/ui/button';

// 基本用法
<Button>Click me</Button>

// 不同樣式
<Button variant="default">Default</Button>
<Button variant="destructive">Delete</Button>
<Button variant="outline">Outline</Button>
<Button variant="ghost">Ghost</Button>

// 不同尺寸
<Button size="sm">Small</Button>
<Button size="default">Default</Button>
<Button size="lg">Large</Button>

// 禁用狀態
<Button disabled>Disabled</Button>
```

#### Props

| 屬性 | 型別 | 預設值 | 說明 |
|-----|------|-------|------|
| `variant` | `'default' \| 'destructive' \| 'outline' \| 'ghost'` | `'default'` | 按鈕樣式 |
| `size` | `'sm' \| 'default' \| 'lg'` | `'default'` | 按鈕尺寸 |
| `disabled` | `boolean` | `false` | 是否禁用 |

---

### 1.2 Card

卡片容器元件，用於包裝內容區塊。

**檔案**: `components/ui/card.tsx`

#### 使用範例

```tsx
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

<Card>
  <CardHeader>
    <CardTitle>標題</CardTitle>
  </CardHeader>
  <CardContent>
    內容區域
  </CardContent>
</Card>
```

---

### 1.3 Input

輸入框元件。

**檔案**: `components/ui/input.tsx`

#### 使用範例

```tsx
import { Input } from '@/components/ui/input';

<Input
  type="text"
  placeholder="請輸入..."
  value={value}
  onChange={(e) => setValue(e.target.value)}
/>
```

---

### 1.4 Select

下拉選單元件。

**檔案**: `components/ui/select.tsx`

#### 使用範例

```tsx
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

<Select value={value} onValueChange={setValue}>
  <SelectTrigger>
    <SelectValue placeholder="請選擇" />
  </SelectTrigger>
  <SelectContent>
    <SelectItem value="option1">選項 1</SelectItem>
    <SelectItem value="option2">選項 2</SelectItem>
  </SelectContent>
</Select>
```

---

### 1.5 Textarea

多行文字輸入框。

**檔案**: `components/ui/textarea.tsx`

#### 使用範例

```tsx
import { Textarea } from '@/components/ui/textarea';

<Textarea
  placeholder="請輸入備註..."
  rows={3}
  value={value}
  onChange={(e) => setValue(e.target.value)}
/>
```

---

### 1.6 Badge

徽章元件，用於顯示標籤或狀態。

**檔案**: `components/ui/badge.tsx`

#### 使用範例

```tsx
import { Badge } from '@/components/ui/badge';

<Badge>新手友善</Badge>
<Badge variant="outline">中階</Badge>
```

---

### 1.7 Dialog

對話框元件，用於顯示模態視窗。

**檔案**: `components/ui/dialog.tsx`

#### 使用範例

```tsx
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';

<Dialog>
  <DialogTrigger asChild>
    <Button>開啟對話框</Button>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>對話框標題</DialogTitle>
    </DialogHeader>
    <div>對話框內容</div>
  </DialogContent>
</Dialog>
```

---

### 1.8 Avatar

頭像元件。

**檔案**: `components/ui/avatar.tsx`

#### 使用範例

```tsx
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';

<Avatar>
  <AvatarImage src={user.avatar_url} alt={user.display_name} />
  <AvatarFallback>{user.display_name[0]}</AvatarFallback>
</Avatar>
```

---

### 1.9 Spinner

載入動畫元件。

**檔案**: `components/ui/spinner.tsx`

#### 使用範例

```tsx
import { Spinner } from '@/components/ui/spinner';

<Spinner />
<Spinner className="h-8 w-8" />
```

---

### 1.10 OptimizedImage

優化的圖片元件，支援 Next.js Image 優化。

**檔案**: `components/ui/optimized-image.tsx`

#### 使用範例

```tsx
import { OptimizedImage } from '@/components/ui/optimized-image';

<OptimizedImage
  src="/images/hero.jpg"
  alt="Hero image"
  width={1200}
  height={600}
  priority
/>
```

---

### 1.11 OptimizedAvatar

優化的頭像元件，結合 Avatar 和 OptimizedImage。

**檔案**: `components/ui/optimized-avatar.tsx`

#### 使用範例

```tsx
import { OptimizedAvatar } from '@/components/ui/optimized-avatar';

<OptimizedAvatar
  src={user.avatar_url}
  alt={user.display_name}
  fallback={user.display_name[0]}
/>
```

---

## 2. 版面配置元件 (components/layout)

### 2.1 Header

網站頂部導航列。

**檔案**: `components/layout/Header.tsx`

#### 功能

- 顯示網站 Logo 和名稱
- 使用者登入狀態顯示
- 桌面版導航選單
- 移動版漢堡選單
- 登入/登出功能

#### 使用範例

```tsx
import { Header } from '@/components/layout/Header';

export default function Layout({ children }) {
  return (
    <div>
      <Header />
      <main>{children}</main>
    </div>
  );
}
```

---

### 2.2 MobileNav

移動版導航選單。

**檔案**: `components/layout/MobileNav.tsx`

#### 功能

- 響應式側邊欄選單
- 使用者資訊顯示
- 導航連結
- 登出功能

#### Props

| 屬性 | 型別 | 說明 |
|-----|------|------|
| `isOpen` | `boolean` | 選單是否開啟 |
| `onClose` | `() => void` | 關閉選單回呼 |

---

## 3. 活動相關元件 (components/event)

### 3.1 EventForm

活動建立/編輯表單。

**檔案**: `components/event/EventForm.tsx`

#### 功能

- 活動基本資訊輸入（標題、說明）
- 日期時間選擇器
- 地點搜尋（整合 Google Places API）
- 人數設定（4-20 人）
- 技能等級選擇
- 費用設定
- 表單驗證
- 提交處理

#### 使用範例

```tsx
import { EventForm } from '@/components/event/EventForm';

export default function CreateEventPage() {
  return (
    <div className="container mx-auto py-8">
      <h1 className="text-2xl font-bold mb-6">建立活動</h1>
      <EventForm />
    </div>
  );
}
```

#### 表單驗證規則

- **活動日期**: 不可早於今天
- **開始時間**: 必填
- **地點**: 必填
- **人數**: 4-20 人
- **技能等級**: 必填
- **費用**: 0-9999 元

---

### 3.2 EventDetail

活動詳細資訊顯示。

**檔案**: `components/event/EventDetail.tsx`

#### 功能

- 顯示活動完整資訊
- 主辦人資訊
- 報名按鈕
- 報名狀態顯示
- 分享功能
- 地圖位置顯示

#### Props

| 屬性 | 型別 | 說明 |
|-----|------|------|
| `event` | `Event` | 活動資料物件 |

#### 使用範例

```tsx
import { EventDetail } from '@/components/event/EventDetail';

export default function EventPage({ params }) {
  const { data: event } = useQuery({
    queryKey: ['event', params.id],
    queryFn: () => apiClient.getEvent(params.id),
  });

  if (!event) return <Spinner />;

  return <EventDetail event={event} />;
}
```

---

### 3.3 RegistrationButton

活動報名按鈕。

**檔案**: `components/event/RegistrationButton.tsx`

#### 功能

- 處理報名邏輯
- 顯示報名狀態（未報名/已報名/候補）
- 處理取消報名
- 登入狀態檢查
- 載入狀態顯示

#### Props

| 屬性 | 型別 | 說明 |
|-----|------|------|
| `eventId` | `string` | 活動 ID |
| `status` | `string` | 活動狀態 |
| `isHost` | `boolean` | 是否為主辦人 |

#### 使用範例

```tsx
import { RegistrationButton } from '@/components/event/RegistrationButton';

<RegistrationButton
  eventId={event.id}
  status={event.status}
  isHost={event.host.id === currentUser?.id}
/>
```

---

## 4. 地圖相關元件 (components/map)

### 4.1 EventMap

Google Maps 地圖元件，顯示活動位置。

**檔案**: `components/map/EventMap.tsx`

#### 功能

- 顯示 Google Maps
- 標記活動位置
- 支援地圖拖曳和縮放
- 點擊標記顯示活動卡片
- 當前位置按鈕
- 響應式設計

#### Props

| 屬性 | 型別 | 說明 |
|-----|------|------|
| `events` | `Event[]` | 活動列表 |
| `center` | `{ lat: number; lng: number }` | 地圖中心點 |
| `zoom` | `number` | 縮放等級（預設 14） |
| `onBoundsChange` | `(bounds: LatLngBounds) => void` | 地圖範圍變更回呼 |
| `onCenterChange` | `(center: LatLng) => void` | 地圖中心變更回呼 |
| `loading` | `boolean` | 載入狀態 |

#### 使用範例

```tsx
import { EventMap } from '@/components/map/EventMap';

export default function MapPage() {
  const [center, setCenter] = useState({ lat: 25.0330, lng: 121.5654 });
  const { data } = useQuery({
    queryKey: ['events', center],
    queryFn: () => apiClient.listEvents({
      lat: center.lat,
      lng: center.lng,
      radius: 5000,
    }),
  });

  return (
    <EventMap
      events={data?.events || []}
      center={center}
      zoom={14}
      onCenterChange={setCenter}
    />
  );
}
```

---

### 4.2 EventPin

地圖上的活動標記。

**檔案**: `components/map/EventPin.tsx`

#### 功能

- 自訂標記樣式
- 顯示技能等級顏色
- 點擊事件處理
- 選中狀態樣式

#### Props

| 屬性 | 型別 | 說明 |
|-----|------|------|
| `event` | `Event` | 活動資料 |
| `onClick` | `(event: Event) => void` | 點擊回呼 |
| `isSelected` | `boolean` | 是否被選中 |

---

### 4.3 EventCard

活動資訊卡片（用於地圖上）。

**檔案**: `components/map/EventCard.tsx`

#### 功能

- 顯示活動摘要資訊
- 導航到活動詳情頁
- 關閉按鈕
- 響應式設計（桌面/移動版）

#### Props

| 屬性 | 型別 | 說明 |
|-----|------|------|
| `event` | `Event` | 活動資料 |
| `onClose` | `() => void` | 關閉回呼 |
| `compact` | `boolean` | 精簡模式 |

---

### 4.4 SkillLevelFilter

技能等級篩選器。

**檔案**: `components/map/SkillLevelFilter.tsx`

#### 功能

- 顯示所有技能等級選項
- 多選功能
- 顏色標示

#### Props

| 屬性 | 型別 | 說明 |
|-----|------|------|
| `value` | `string[]` | 已選擇的等級 |
| `onChange` | `(levels: string[]) => void` | 變更回呼 |

---

## 5. 表單元件 (components/form)

### 5.1 PlacesAutocomplete

Google Places 地點自動完成輸入框。

**檔案**: `components/form/PlacesAutocomplete.tsx`

#### 功能

- 整合 Google Places Autocomplete API
- 即時搜尋地點
- 取得地點詳細資訊（名稱、地址、經緯度）
- 錯誤處理

#### Props

| 屬性 | 型別 | 說明 |
|-----|------|------|
| `value` | `string` | 輸入值 |
| `onPlaceSelect` | `(place: PlaceResult) => void` | 地點選擇回呼 |
| `placeholder` | `string` | 佔位文字 |
| `className` | `string` | CSS 類名 |

#### PlaceResult 型別

```typescript
interface PlaceResult {
  name: string;
  address: string;
  lat: number;
  lng: number;
  placeId: string;
}
```

#### 使用範例

```tsx
import { PlacesAutocomplete, PlaceResult } from '@/components/form/PlacesAutocomplete';

const [location, setLocation] = useState<PlaceResult | null>(null);

<PlacesAutocomplete
  value={location?.name || ''}
  onPlaceSelect={setLocation}
  placeholder="搜尋球場或地點..."
/>
```

---

## 6. SEO 元件 (components/seo)

### 6.1 EventJsonLd

活動結構化資料元件（JSON-LD）。

**檔案**: `components/seo/EventJsonLd.tsx`

#### 功能

- 生成 Schema.org Event 結構化資料
- 提升 SEO 和搜尋結果豐富度
- Google Search 活動標記支援

#### Props

| 屬性 | 型別 | 說明 |
|-----|------|------|
| `event` | `Event` | 活動資料 |

#### 使用範例

```tsx
import { EventJsonLd } from '@/components/seo/EventJsonLd';

export default function EventPage({ event }) {
  return (
    <>
      <EventJsonLd event={event} />
      <EventDetail event={event} />
    </>
  );
}
```

---

## 7. 分析元件 (components/analytics)

### 7.1 GoogleAnalytics

Google Analytics 4 整合元件。

**檔案**: `components/analytics/GoogleAnalytics.tsx`

#### 功能

- 整合 GA4 追蹤
- 頁面瀏覽追蹤
- 事件追蹤

#### 使用範例

```tsx
import { GoogleAnalytics } from '@/components/analytics/GoogleAnalytics';

// 在 app/layout.tsx 中使用
export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <GoogleAnalytics />
      </body>
    </html>
  );
}
```

---

## 8. Context Providers

### 8.1 AuthContext

使用者認證狀態管理。

**檔案**: `contexts/AuthContext.tsx`

#### 功能

- 管理使用者登入狀態
- 提供登入/登出方法
- Token 管理
- 自動更新 Token

#### 使用範例

```tsx
import { useAuth } from '@/contexts/AuthContext';

function MyComponent() {
  const { user, isAuthenticated, login, logout } = useAuth();

  if (!isAuthenticated) {
    return <Button onClick={login}>登入</Button>;
  }

  return (
    <div>
      <p>歡迎, {user.display_name}</p>
      <Button onClick={logout}>登出</Button>
    </div>
  );
}
```

---

## 9. Custom Hooks

專案中使用的自訂 React Hooks。

### 9.1 useCurrentUser

取得目前登入使用者資訊。

```tsx
import { useCurrentUser } from '@/hooks/useCurrentUser';

function MyComponent() {
  const { data: user, isLoading } = useCurrentUser();

  if (isLoading) return <Spinner />;
  if (!user) return <LoginButton />;

  return <div>Hello, {user.display_name}</div>;
}
```

### 9.2 useEvents

取得活動列表。

```tsx
import { useEvents } from '@/hooks/useEvents';

function EventList() {
  const { data, isLoading, error } = useEvents({
    lat: 25.0330,
    lng: 121.5654,
    radius: 5000,
  });

  if (isLoading) return <Spinner />;
  if (error) return <Error />;

  return (
    <div>
      {data.events.map(event => (
        <EventCard key={event.id} event={event} />
      ))}
    </div>
  );
}
```

---

## 10. API Client

### apiClient

封裝 API 請求的客戶端。

**檔案**: `lib/api-client.ts`

#### 功能

- 統一的 API 請求處理
- 自動加入 JWT Token
- 錯誤處理
- Token 自動更新
- 型別安全

#### 使用範例

```tsx
import { apiClient } from '@/lib/api-client';

// 取得活動
const event = await apiClient.getEvent(eventId);

// 建立活動
const result = await apiClient.createEvent({
  title: '週末輕鬆打',
  event_date: '2026-01-25',
  start_time: '19:00',
  location: { name: '大安森林公園', lat: 25.0292, lng: 121.5367 },
  capacity: 8,
  skill_level: 'intermediate',
});

// 報名活動
await apiClient.registerForEvent(eventId);

// 取消報名
await apiClient.cancelRegistration(eventId);
```

---

## 11. 型別定義

主要 TypeScript 型別定義。

**檔案**: `types/index.ts`

### Event

```typescript
interface Event {
  id: string;
  host: User;
  title?: string;
  description?: string;
  event_date: string;
  start_time: string;
  end_time?: string;
  location: {
    name: string;
    address?: string;
    lat: number;
    lng: number;
    google_place_id?: string;
  };
  capacity: number;
  confirmed_count: number;
  waitlist_count: number;
  skill_level: string;
  skill_level_label: string;
  fee: number;
  status: string;
}
```

### User

```typescript
interface User {
  id: string;
  display_name: string;
  avatar_url?: string;
  email?: string;
}
```

---

## 12. 樣式指南

### Tailwind CSS 使用規範

#### 間距

```tsx
// 一致的間距系統
<div className="space-y-4">  {/* 垂直間距 */}
<div className="space-x-2">  {/* 水平間距 */}
<div className="p-4">        {/* padding */}
<div className="m-4">        {/* margin */}
```

#### 響應式設計

```tsx
// Mobile-first 方法
<div className="w-full md:w-1/2 lg:w-1/3">
  {/* 手機全寬，平板半寬，桌面三分之一寬 */}
</div>
```

#### 顏色系統

```tsx
// 使用語義化顏色
<div className="bg-primary text-primary-foreground">
<div className="bg-destructive text-destructive-foreground">
<div className="text-muted-foreground">
```

---

## 13. 效能優化建議

### 圖片優化

```tsx
// 使用 OptimizedImage 而非 <img>
<OptimizedImage
  src="/images/hero.jpg"
  alt="Hero"
  width={1200}
  height={600}
  priority  // 首屏圖片使用 priority
/>
```

### 程式碼分割

```tsx
// 動態載入非關鍵元件
import dynamic from 'next/dynamic';

const EventMap = dynamic(() => import('@/components/map/EventMap'), {
  loading: () => <Spinner />,
  ssr: false,  // 地圖不需要 SSR
});
```

### React Query 快取

```tsx
// 設定適當的 staleTime 和 cacheTime
const { data } = useQuery({
  queryKey: ['events', filters],
  queryFn: () => apiClient.listEvents(filters),
  staleTime: 5 * 60 * 1000,  // 5 分鐘
  cacheTime: 10 * 60 * 1000, // 10 分鐘
});
```

---

## 14. 無障礙設計 (Accessibility)

### ARIA 屬性

```tsx
// 為互動元素提供 ARIA 標籤
<button aria-label="關閉對話框" onClick={onClose}>
  <X />
</button>

<input
  type="search"
  placeholder="搜尋活動"
  aria-label="搜尋活動"
/>
```

### 鍵盤導航

```tsx
// 支援 ESC 鍵關閉
useEffect(() => {
  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose();
    }
  };
  document.addEventListener('keydown', handleKeyDown);
  return () => document.removeEventListener('keydown', handleKeyDown);
}, [onClose]);
```

---

## 15. 測試建議

### 元件測試

```tsx
// 使用 React Testing Library
import { render, screen } from '@testing-library/react';
import { Button } from '@/components/ui/button';

test('renders button with text', () => {
  render(<Button>Click me</Button>);
  expect(screen.getByText('Click me')).toBeInTheDocument();
});
```

### 整合測試

```tsx
// 測試完整使用者流程
test('user can create event', async () => {
  render(<EventForm />);

  // 填寫表單
  await userEvent.type(screen.getByLabelText('活動日期'), '2026-01-25');
  await userEvent.type(screen.getByLabelText('開始時間'), '19:00');
  // ...

  // 提交
  await userEvent.click(screen.getByText('建立活動'));

  // 驗證
  expect(screen.getByText('活動建立成功')).toBeInTheDocument();
});
```

---

## 16. 元件開發最佳實踐

### 1. 使用 TypeScript

所有元件都應該有明確的型別定義。

```tsx
interface ButtonProps {
  children: React.ReactNode;
  onClick?: () => void;
  variant?: 'default' | 'outline';
  disabled?: boolean;
}

export function Button({ children, onClick, variant = 'default', disabled }: ButtonProps) {
  // ...
}
```

### 2. 使用 React.memo 優化

對於頻繁重渲染的元件使用 memo。

```tsx
export const EventCard = React.memo(({ event }: { event: Event }) => {
  // ...
});
```

### 3. 提取可重用邏輯到 Hooks

```tsx
// ❌ 不好
function MyComponent() {
  const [isOpen, setIsOpen] = useState(false);
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.key === 'Escape') setIsOpen(false);
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);
}

// ✅ 好
function useEscapeKey(callback: () => void) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') callback();
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [callback]);
}

function MyComponent() {
  const [isOpen, setIsOpen] = useState(false);
  useEscapeKey(() => setIsOpen(false));
}
```

### 4. 使用 Server Components

盡可能使用 Next.js Server Components。

```tsx
// app/events/page.tsx
// 這是 Server Component，可以直接 fetch 資料
export default async function EventsPage() {
  const events = await fetch('https://api.picklego.tw/api/v1/events')
    .then(res => res.json());

  return <EventList events={events} />;
}
```

---

## 17. 常見問題 (FAQ)

### Q1: 如何在元件中使用環境變數？

```tsx
// 必須以 NEXT_PUBLIC_ 開頭才能在客戶端使用
const apiUrl = process.env.NEXT_PUBLIC_API_URL;
const mapsKey = process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY;
```

### Q2: 如何處理表單驗證？

```tsx
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';

const schema = z.object({
  title: z.string().min(1).max(200),
  capacity: z.number().min(4).max(20),
});

function MyForm() {
  const { register, handleSubmit, formState: { errors } } = useForm({
    resolver: zodResolver(schema),
  });

  const onSubmit = (data) => {
    // 處理提交
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register('title')} />
      {errors.title && <span>{errors.title.message}</span>}
    </form>
  );
}
```

### Q3: 如何實作無限滾動？

```tsx
import { useInfiniteQuery } from '@tanstack/react-query';
import { useInView } from 'react-intersection-observer';

function EventList() {
  const { ref, inView } = useInView();

  const {
    data,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery({
    queryKey: ['events'],
    queryFn: ({ pageParam = 0 }) => apiClient.listEvents({ offset: pageParam }),
    getNextPageParam: (lastPage, pages) =>
      lastPage.has_more ? pages.length * 20 : undefined,
  });

  useEffect(() => {
    if (inView && hasNextPage) {
      fetchNextPage();
    }
  }, [inView, hasNextPage, fetchNextPage]);

  return (
    <div>
      {data?.pages.map(page =>
        page.events.map(event => <EventCard key={event.id} event={event} />)
      )}
      <div ref={ref} />
    </div>
  );
}
```

---

## 18. 資源連結

- [Next.js 文件](https://nextjs.org/docs)
- [shadcn/ui 文件](https://ui.shadcn.com/)
- [TanStack Query 文件](https://tanstack.com/query/latest)
- [Tailwind CSS 文件](https://tailwindcss.com/docs)
- [Google Maps API](https://developers.google.com/maps/documentation)

---

**版本**: 1.0.0
**最後更新**: 2026-01-21
