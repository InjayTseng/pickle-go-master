'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import { useAuthContext } from '@/contexts/AuthContext';
import { useRegisterForEvent, useCancelRegistration, isEventFull } from '@/hooks/useEvents';
import { Event, RegistrationResponse } from '@/lib/api-client';
import { getLineLoginURL } from '@/lib/auth';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { CheckCircle, XCircle, Clock, Users, AlertCircle } from 'lucide-react';

interface RegistrationButtonProps {
  event: Event;
  userRegistration?: {
    status: string;
    waitlist_position?: number;
  } | null;
  onSuccess?: (result: RegistrationResponse) => void;
  onCancel?: () => void;
  className?: string;
  size?: 'default' | 'sm' | 'lg';
}

type ModalState = 'none' | 'success' | 'waitlist' | 'cancelled' | 'error';

export function RegistrationButton({
  event,
  userRegistration,
  onSuccess,
  onCancel,
  className = '',
  size = 'default',
}: RegistrationButtonProps) {
  const router = useRouter();
  const { user, isAuthenticated, isLoading: authLoading } = useAuthContext();
  const registerMutation = useRegisterForEvent();
  const cancelMutation = useCancelRegistration();

  const [modalState, setModalState] = useState<ModalState>('none');
  const [modalData, setModalData] = useState<{ message?: string; position?: number }>({});

  const isHost = user?.id === event.host.id;
  const isRegistered = userRegistration?.status === 'confirmed';
  const isWaitlisted = userRegistration?.status === 'waitlist';
  const isFull = isEventFull(event);
  const isEventOver = event.status === 'completed' || event.status === 'cancelled';

  const handleRegister = async () => {
    if (!isAuthenticated) {
      // Store return URL in sessionStorage for after login
      const returnUrl = window.location.pathname;
      if (typeof window !== 'undefined') {
        sessionStorage.setItem('login_redirect', returnUrl);
      }
      // Redirect to Line login with CSRF-protected state
      const loginUrl = await getLineLoginURL();
      window.location.href = loginUrl;
      return;
    }

    try {
      const result = await registerMutation.mutateAsync(event.id);

      if (result.status === 'waitlist') {
        setModalData({ message: result.message, position: result.waitlist_position });
        setModalState('waitlist');
      } else {
        setModalData({ message: result.message });
        setModalState('success');
      }

      onSuccess?.(result);
    } catch (error: any) {
      setModalData({ message: error.message || 'Registration failed' });
      setModalState('error');
    }
  };

  const handleCancel = async () => {
    try {
      await cancelMutation.mutateAsync(event.id);
      setModalState('cancelled');
      onCancel?.();
    } catch (error: any) {
      setModalData({ message: error.message || 'Cancellation failed' });
      setModalState('error');
    }
  };

  const closeModal = () => {
    setModalState('none');
    setModalData({});
  };

  const isLoading = registerMutation.isPending || cancelMutation.isPending || authLoading;

  // Determine button state and content
  const getButtonContent = () => {
    if (isLoading) {
      return (
        <>
          <Spinner className="mr-2 h-4 w-4" />
          處理中...
        </>
      );
    }

    if (isEventOver) {
      return (
        <>
          <XCircle className="mr-2 h-4 w-4" />
          {event.status === 'cancelled' ? '已取消' : '已結束'}
        </>
      );
    }

    if (isHost) {
      return (
        <>
          <Users className="mr-2 h-4 w-4" />
          你是主辦人
        </>
      );
    }

    if (isRegistered) {
      return (
        <>
          <CheckCircle className="mr-2 h-4 w-4" />
          已報名
        </>
      );
    }

    if (isWaitlisted) {
      return (
        <>
          <Clock className="mr-2 h-4 w-4" />
          候補中 (第 {userRegistration?.waitlist_position} 位)
        </>
      );
    }

    if (!isAuthenticated) {
      return '+1 參加';
    }

    if (isFull) {
      return (
        <>
          <Clock className="mr-2 h-4 w-4" />
          排候補
        </>
      );
    }

    return '+1 參加';
  };

  const getButtonVariant = () => {
    if (isEventOver || isHost) return 'secondary';
    if (isRegistered) return 'outline';
    if (isWaitlisted) return 'outline';
    return 'default';
  };

  const isDisabled = isEventOver || isHost || isLoading;

  return (
    <>
      <div className={`flex gap-2 ${className}`}>
        {/* Main action button */}
        {(isRegistered || isWaitlisted) ? (
          <>
            <Button
              variant={getButtonVariant()}
              size={size}
              disabled={isDisabled}
              className="flex-1"
            >
              {getButtonContent()}
            </Button>
            <Button
              variant="destructive"
              size={size}
              disabled={isLoading}
              onClick={handleCancel}
            >
              {cancelMutation.isPending ? (
                <Spinner className="h-4 w-4" />
              ) : (
                '取消報名'
              )}
            </Button>
          </>
        ) : (
          <Button
            variant={getButtonVariant()}
            size={size}
            disabled={isDisabled}
            onClick={handleRegister}
            className="flex-1"
          >
            {getButtonContent()}
          </Button>
        )}
      </div>

      {/* Success Modal */}
      <Dialog open={modalState === 'success'} onOpenChange={closeModal}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
              <CheckCircle className="h-10 w-10 text-green-600" />
            </div>
            <DialogTitle className="text-center text-xl">報名成功!</DialogTitle>
            <DialogDescription className="text-center">
              {modalData.message || '你已成功報名此活動'}
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-center mt-4">
            <Button onClick={closeModal}>確定</Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Waitlist Modal */}
      <Dialog open={modalState === 'waitlist'} onOpenChange={closeModal}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-amber-100">
              <Clock className="h-10 w-10 text-amber-600" />
            </div>
            <DialogTitle className="text-center text-xl">已加入候補!</DialogTitle>
            <DialogDescription className="text-center">
              {modalData.message || `你是第 ${modalData.position} 位候補`}
              <br />
              <span className="text-sm text-muted-foreground mt-2 block">
                有人取消報名時，系統會自動遞補並通知你
              </span>
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-center mt-4">
            <Button onClick={closeModal}>確定</Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Cancelled Modal */}
      <Dialog open={modalState === 'cancelled'} onOpenChange={closeModal}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100">
              <XCircle className="h-10 w-10 text-gray-600" />
            </div>
            <DialogTitle className="text-center text-xl">已取消報名</DialogTitle>
            <DialogDescription className="text-center">
              你已成功取消此活動的報名
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-center mt-4">
            <Button onClick={closeModal}>確定</Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Error Modal */}
      <Dialog open={modalState === 'error'} onOpenChange={closeModal}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
              <AlertCircle className="h-10 w-10 text-red-600" />
            </div>
            <DialogTitle className="text-center text-xl">操作失敗</DialogTitle>
            <DialogDescription className="text-center">
              {modalData.message || '發生錯誤，請稍後再試'}
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-center mt-4">
            <Button onClick={closeModal}>確定</Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

export default RegistrationButton;
