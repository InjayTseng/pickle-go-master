'use client';

import React from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';

export type SkillLevel = 'all' | 'beginner' | 'intermediate' | 'advanced' | 'expert' | 'any';

interface SkillLevelFilterProps {
  value: SkillLevel;
  onChange: (value: SkillLevel) => void;
  className?: string;
}

const skillLevels: { value: SkillLevel; label: string; color: string }[] = [
  { value: 'all', label: '全部', color: 'bg-gray-100 text-gray-800 hover:bg-gray-200' },
  { value: 'beginner', label: '新手友善', color: 'bg-emerald-100 text-emerald-800 hover:bg-emerald-200' },
  { value: 'intermediate', label: '中階', color: 'bg-blue-100 text-blue-800 hover:bg-blue-200' },
  { value: 'advanced', label: '進階', color: 'bg-purple-100 text-purple-800 hover:bg-purple-200' },
  { value: 'expert', label: '高階', color: 'bg-red-100 text-red-800 hover:bg-red-200' },
  { value: 'any', label: '不限程度', color: 'bg-amber-100 text-amber-800 hover:bg-amber-200' },
];

export function SkillLevelFilter({ value, onChange, className }: SkillLevelFilterProps) {
  return (
    <div className={cn('flex flex-wrap gap-2', className)}>
      {skillLevels.map((level) => (
        <button
          key={level.value}
          onClick={() => onChange(level.value)}
          className={cn(
            'px-3 py-1.5 rounded-full text-sm font-medium transition-all',
            'border-2 border-transparent',
            level.color,
            value === level.value && 'ring-2 ring-offset-1 ring-primary border-primary'
          )}
        >
          {level.label}
        </button>
      ))}
    </div>
  );
}

// Compact version for mobile
export function SkillLevelFilterCompact({ value, onChange, className }: SkillLevelFilterProps) {
  return (
    <div className={cn('overflow-x-auto pb-2', className)}>
      <div className="flex gap-2 min-w-max px-1">
        {skillLevels.map((level) => (
          <button
            key={level.value}
            onClick={() => onChange(level.value)}
            className={cn(
              'px-3 py-1 rounded-full text-xs font-medium transition-all whitespace-nowrap',
              'border border-transparent',
              level.color,
              value === level.value && 'ring-2 ring-primary'
            )}
          >
            {level.label}
          </button>
        ))}
      </div>
    </div>
  );
}

export default SkillLevelFilter;
