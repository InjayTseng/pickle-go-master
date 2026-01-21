'use client';

import React from 'react';
import { OverlayView, OverlayViewF } from '@react-google-maps/api';
import { Event } from '@/lib/api-client';
import { getEventPinColor, isEventFull } from '@/hooks/useEvents';

interface EventPinProps {
  event: Event;
  onClick?: (event: Event) => void;
  isSelected?: boolean;
}

export function EventPin({ event, onClick, isSelected = false }: EventPinProps) {
  const color = getEventPinColor(event);
  const isFull = isEventFull(event);

  const getBackgroundColor = () => {
    switch (color) {
      case 'green':
        return '#22c55e'; // green-500
      case 'red':
        return '#ef4444'; // red-500
      case 'gray':
        return '#9ca3af'; // gray-400
      default:
        return '#22c55e';
    }
  };

  const getBorderColor = () => {
    if (isSelected) {
      return '#1d4ed8'; // blue-700
    }
    switch (color) {
      case 'green':
        return '#16a34a'; // green-600
      case 'red':
        return '#dc2626'; // red-600
      case 'gray':
        return '#6b7280'; // gray-500
      default:
        return '#16a34a';
    }
  };

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    onClick?.(event);
  };

  return (
    <OverlayViewF
      position={{ lat: event.location.lat, lng: event.location.lng }}
      mapPaneName={OverlayView.OVERLAY_MOUSE_TARGET}
    >
      <div
        className="relative cursor-pointer transform -translate-x-1/2 -translate-y-full"
        onClick={handleClick}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            onClick?.(event);
          }
        }}
      >
        {/* Pin body */}
        <div
          className={`
            relative w-10 h-10 rounded-full flex items-center justify-center
            shadow-lg transition-transform duration-200
            hover:scale-110 active:scale-95
            ${isSelected ? 'scale-125 z-10' : ''}
          `}
          style={{
            backgroundColor: getBackgroundColor(),
            border: `3px solid ${getBorderColor()}`,
          }}
        >
          {/* Inner content - spots remaining or icon */}
          <span className="text-white font-bold text-sm">
            {color === 'gray' ? (
              <XIcon />
            ) : isFull ? (
              <FullIcon />
            ) : (
              event.capacity - event.confirmed_count
            )}
          </span>
        </div>

        {/* Pin pointer */}
        <div
          className="absolute left-1/2 -translate-x-1/2 w-0 h-0"
          style={{
            borderLeft: '8px solid transparent',
            borderRight: '8px solid transparent',
            borderTop: `10px solid ${getBorderColor()}`,
            top: '36px',
          }}
        />

        {/* Shadow */}
        <div
          className="absolute left-1/2 -translate-x-1/2 w-4 h-1 rounded-full opacity-30"
          style={{
            backgroundColor: '#000',
            top: '44px',
          }}
        />
      </div>
    </OverlayViewF>
  );
}

// X icon for completed/cancelled events
function XIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="3"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <line x1="18" y1="6" x2="6" y2="18" />
      <line x1="6" y1="6" x2="18" y2="18" />
    </svg>
  );
}

// Full icon (checkmark) for full events
function FullIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="3"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <polyline points="20 6 9 17 4 12" />
    </svg>
  );
}

export default EventPin;
